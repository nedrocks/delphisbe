package backend

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/multierr"

	"github.com/nedrocks/delphisbe/graph/model"
	"github.com/nedrocks/delphisbe/internal/datastore"
	"github.com/nedrocks/delphisbe/internal/util"
	"github.com/sirupsen/logrus"
)

func (d *delphisBackend) CreatePost(ctx context.Context, discussionID string, participantID string, input model.PostContentInput) (*model.Post, error) {

	postContent := model.PostContent{
		ID:                util.UUIDv4(),
		Content:           input.PostText,
		MentionedEntities: input.MentionedEntities,
	}

	logrus.Infof("PostContent: %+v\n", postContent)

	post := model.Post{
		ID:            util.UUIDv4(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		DiscussionID:  &discussionID,
		ParticipantID: &participantID,
		PostContentID: &postContent.ID,
		PostContent:   &postContent,
		QuotedPostID:  input.QuotedPostID,
	}

	// Begin tx
	tx, err := d.db.BeginTx(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to begin tx")
		return nil, err
	}
	// Put post contents
	if err := d.db.PutPostContent(ctx, tx, postContent); err != nil {
		logrus.WithError(err).Error("failed to PutPostContent")

		// Rollback on errors
		if txErr := d.db.RollbackTx(ctx, tx); txErr != nil {
			logrus.WithError(txErr).Error("failed to rollback tx")
			return nil, multierr.Append(err, txErr)
		}
		return nil, err
	}

	// Put post
	postObj, err := d.db.PutPost(ctx, tx, post)
	if err != nil {
		logrus.WithError(err).Error("failed to PutPost")

		// Rollback on errors
		if txErr := d.db.RollbackTx(ctx, tx); txErr != nil {
			logrus.WithError(txErr).Error("failed to rollback tx")
			return nil, multierr.Append(err, txErr)
		}
		return nil, err
	}

	// Put Mentions
	if err := d.db.PutMention(ctx, tx, postObj); err != nil {
		logrus.WithError(err).Error("failed to PutMention")

		// We don't want to rollback the whole transaction if we mess up the recording of mentions.
		// Ideally we'd push it to a queue to be re-ran later
	}

	// Commit transaction
	if err := d.db.CommitTx(ctx, tx); err != nil {
		logrus.WithError(err).Error("failed to commit post tx")
		return nil, err
	}

	logrus.Infof("Post: %+v\n", postObj.PostContent)

	discussion, err := d.db.GetDiscussionByID(ctx, discussionID)
	if err != nil {
		logrus.WithError(err).Debugf("Skipping notification to subscribers because of an error")
	} else {
		_, err := d.SendNotificationsToSubscribers(ctx, discussion, &post)
		if err != nil {
			logrus.WithError(err).Warn("Failed to send push notifications on createPost")
		}
	}

	return postObj, nil
}

func (d *delphisBackend) NotifySubscribersOfCreatedPost(ctx context.Context, post *model.Post, discussionID string) error {
	cacheKey := fmt.Sprintf(discussionSubscriberKey, discussionID)
	d.discussionMutex.Lock()
	defer d.discussionMutex.Unlock()
	currentSubsIface, found := d.cache.Get(cacheKey)
	if !found {
		currentSubsIface = map[string]chan *model.Post{}
	}
	var currentSubs map[string]chan *model.Post
	var ok bool
	if currentSubs, ok = currentSubsIface.(map[string]chan *model.Post); !ok {
		currentSubs = map[string]chan *model.Post{}
	}
	for userID, channel := range currentSubs {
		if channel != nil {
			select {
			case channel <- post:
				logrus.Debugf("Sent message to channel for user ID: %s", userID)
			default:
				logrus.Debugf("No message was sent. Unsubscribing the user")
				delete(currentSubs, userID)
			}
		}
	}
	d.cache.Set(cacheKey, currentSubs, time.Hour)
	return nil
}

func (d *delphisBackend) GetPostsByDiscussionID(ctx context.Context, discussionID string) ([]*model.Post, error) {
	iter := d.db.GetPostsByDiscussionIDIter(ctx, discussionID)
	return d.iterToPosts(ctx, iter)
	//return d.db.GetPostsByDiscussionID(ctx, discussionID)
}

// Testing function to keep functionality
func (d *delphisBackend) iterToPosts(ctx context.Context, iter datastore.PostIter) ([]*model.Post, error) {
	var posts []*model.Post
	post := model.Post{}

	defer iter.Close()

	for iter.Next(&post) {
		tempPost := post

		// Check if there is a quotedPostID. Fetch if so
		if tempPost.QuotedPostID != nil {
			var err error
			// TODO: potentially optimize into joins
			tempPost.QuotedPost, err = d.db.GetPostByID(ctx, *tempPost.QuotedPostID)
			if err != nil {
				// Do we want to fail the whole discussion if we can't get a quote?
				return nil, err
			}
		}

		posts = append(posts, &tempPost)
	}

	if err := iter.Close(); err != nil && err != io.EOF {
		logrus.WithError(err).Error("failed to close iter")
		return nil, err
	}

	return posts, nil
}
