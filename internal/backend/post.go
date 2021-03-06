package backend

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/multierr"

	"github.com/delphis-inc/delphisbe/graph/model"
	"github.com/delphis-inc/delphisbe/internal/util"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	PostPerPageLimit = 50
	PutPostMaxRetry  = 3
)

func (d *delphisBackend) CreatePost(ctx context.Context, discussionID string, userID string, participantID string, input model.PostContentInput) (*model.Post, error) {
	// Validate post params
	if err := validatePostParams(ctx, input); err != nil {
		logrus.WithError(err).Error("failed to validate post params")
		return nil, err
	}

	postContent := model.PostContent{
		ID:                util.UUIDv4(),
		Content:           input.PostText,
		MentionedEntities: input.MentionedEntities,
	}

	post := model.Post{
		ID:            util.UUIDv4(),
		PostType:      input.PostType,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		DiscussionID:  &discussionID,
		ParticipantID: &participantID,
		PostContentID: &postContent.ID,
		PostContent:   &postContent,
		QuotedPostID:  input.QuotedPostID,
		MediaID:       input.MediaID,
	}

	retryAttempts := 0
	for true {
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
			pqError, isPqError := err.(*pq.Error)
			if isPqError && retryAttempts < PutPostMaxRetry && pqError.Code == "23505" {
				retryAttempts++
				logrus.WithError(err).Error("failed to PutPost, retrying with attempt #" + strconv.Itoa(retryAttempts))
				// Note for the future: should we backoff a little?
				continue
			} else {
				logrus.WithError(err).Error("failed to PutPost")
				// Rollback on errors
				if txErr := d.db.RollbackTx(ctx, tx); txErr != nil {
					logrus.WithError(txErr).Error("failed to rollback tx")
					return nil, multierr.Append(err, txErr)
				}
				return nil, err
			}
		}

		// Put Activity
		if err := d.db.PutActivity(ctx, tx, postObj); err != nil {
			logrus.WithError(err).Error("failed to PutActivity")

			// We don't want to rollback the whole transaction if we mess up the recording of mentions.
			// Ideally we'd push it to a queue to be re-ran later
		}

		// Commit transaction
		if err := d.db.CommitTx(ctx, tx); err != nil {
			logrus.WithError(err).Error("failed to commit post tx")
			return nil, err
		}

		discInput := model.DiscussionInput{
			LastPostID:        &post.ID,
			LastPostCreatedAt: &post.CreatedAt,
		}

		discussion, err := d.UpdateDiscussion(ctx, discussionID, discInput)
		// If we reach this point then the transaction is succesfully committed and we should not retry
		if err != nil {
			logrus.WithError(err).Debugf("Skipping notification to subscribers because of an error")
		} else {
			_, err := d.SendNotificationsToSubscribers(ctx, userID, discussion, &post, input.Preview)
			if err != nil {
				logrus.WithError(err).Warn("Failed to send push notifications on createPost")
			}
		}

		if err := d.NotifySubscribersOfCreatedPost(ctx, postObj, discussionID); err != nil {
			// Silently ignore this
			logrus.Warnf("Failed to notify subscribers of created post")
		}

		return postObj, nil
	}
	return nil, errors.New("Unknown error, this code should not be reachable")
}

func (d *delphisBackend) CreateWelcomeAlertPost(ctx context.Context, discussionID string, participantID string, userObj *model.User, isAnonymous bool) (*model.Post, error) {
	// Do not create an alert post when the concierge joins
	if userObj.ID == model.ConciergeUser {
		return nil, nil
	}

	discObj, err := d.GetDiscussionByID(ctx, discussionID)
	if err != nil {
		logrus.WithError(err).Error("failed to get discussion by ID")
		return nil, err
	}

	// Create post string based on anonymity. Need proper copy from Ned/Chris
	if userObj.UserProfile == nil {
		logrus.Infof("userProfile should not be nil here")
		return nil, fmt.Errorf("userProfile should not be nil")
	}

	displayName := userObj.UserProfile.DisplayName
	if isAnonymous {
		hashAsInt64 := util.GenerateParticipantSeed(discussionID, participantID, discObj.ShuffleCount)
		displayName = util.GenerateFullDisplayName(hashAsInt64)
	}

	welcomeStr := fmt.Sprintf("Welcome %v to the chat", displayName)

	// Get concierge participant
	resp, err := d.GetParticipantsByDiscussionIDUserID(ctx, discussionID, model.ConciergeUser)
	if err != nil {
		logrus.WithError(err).Error("failed to fetch concierge participant")
		return nil, err
	}
	if resp.NonAnon == nil {
		return nil, fmt.Errorf("discussion is missing a concierge participant")
	}

	input := model.PostContentInput{
		PostText: welcomeStr,
		PostType: model.PostTypeAlert,
	}

	return d.CreatePost(ctx, discussionID, model.ConciergeUser, resp.NonAnon.ID, input)
}

func (d *delphisBackend) CreateShuffleAlertPost(ctx context.Context, discussionID string) (*model.Post, error) {
	postStr := "The moderator has shuffled all aliases in the discussion"

	// Get concierge participant
	resp, err := d.GetParticipantsByDiscussionIDUserID(ctx, discussionID, model.ConciergeUser)
	if err != nil {
		logrus.WithError(err).Error("failed to fetch concierge participant")
		return nil, err
	}
	if resp.NonAnon == nil {
		return nil, fmt.Errorf("discussion is missing a concierge participant")
	}

	input := model.PostContentInput{
		PostText: postStr,
		PostType: model.PostTypeAlert,
	}

	return d.CreatePost(ctx, discussionID, model.ConciergeUser, resp.NonAnon.ID, input)
}

func (d *delphisBackend) notifySubscribersOfEvent(ctx context.Context, event *model.DiscussionSubscriptionEvent, discussionID string) error {
	cacheKey := fmt.Sprintf(discussionEventSubscriberKey, discussionID)
	d.discussionMutex.Lock()
	defer d.discussionMutex.Unlock()
	currentSubsIface, found := d.cache.Get(cacheKey)
	if !found {
		currentSubsIface = map[string]chan *model.DiscussionSubscriptionEvent{}
	}
	var currentSubs map[string]chan *model.DiscussionSubscriptionEvent
	var ok bool
	if currentSubs, ok = currentSubsIface.(map[string]chan *model.DiscussionSubscriptionEvent); !ok {
		currentSubs = map[string]chan *model.DiscussionSubscriptionEvent{}
	}
	for userID, channel := range currentSubs {
		if channel != nil {
			select {
			case channel <- event:
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

func (d *delphisBackend) NotifySubscribersOfCreatedPost(ctx context.Context, post *model.Post, discussionID string) error {
	event := &model.DiscussionSubscriptionEvent{
		EventType: model.DiscussionSubscriptionEventTypePostAdded,
		Entity:    post,
	}
	return d.notifySubscribersOfEvent(ctx, event, discussionID)
}

func (d *delphisBackend) NotifySubscribersOfDeletedPost(ctx context.Context, post *model.Post, discussionID string) error {
	event := &model.DiscussionSubscriptionEvent{
		EventType: model.DiscussionSubscriptionEventTypePostDeleted,
		Entity:    post,
	}
	return d.notifySubscribersOfEvent(ctx, event, discussionID)
}

func (d *delphisBackend) NotifySubscribersOfBannedParticipant(ctx context.Context, participant *model.Participant, discussionID string) error {
	event := &model.DiscussionSubscriptionEvent{
		EventType: model.DiscussionSubscriptionEventTypeParticipantBanned,
		Entity:    participant,
	}
	return d.notifySubscribersOfEvent(ctx, event, discussionID)
}

func (d *delphisBackend) GetPostsByDiscussionID(ctx context.Context, discussionID string) ([]*model.Post, error) {
	iter := d.db.GetPostsByDiscussionIDIter(ctx, discussionID)

	posts, err := d.db.PostIterCollect(ctx, iter)
	if err != nil {
		logrus.WithError(err).Error("failed to get posts by discussionID")
		return nil, err
	}

	return posts, nil
}

func (d *delphisBackend) GetLastPostByDiscussionID(ctx context.Context, discussionID string) (*model.Post, error) {
	return d.db.GetLastPostByDiscussionID(ctx, discussionID)
}

func (d *delphisBackend) GetPostsConnectionByDiscussionID(ctx context.Context, discussionID string, cursor string, limit int) (*model.PostsConnection, error) {
	if limit < 2 || limit > PostPerPageLimit {
		return nil, errors.New("Values of 'limit' is illegal")
	}

	connection, err := d.db.GetPostsConnectionByDiscussionID(ctx, discussionID, cursor, limit)
	if err != nil {
		return nil, err
	}

	return connection, err
}

func (d *delphisBackend) GetMentionedEntities(ctx context.Context, entityIDs []string) (map[string]model.Entity, error) {
	entities := map[string]model.Entity{}
	var participantIDs []string
	var discussionIDs []string

	// Iterate over mentioned entities and divide into participants and discussions
	for _, entityID := range entityIDs {
		entity, err := util.ReturnParsedEntityID(entityID)
		if err != nil {
			logrus.WithError(err).Error("failed to parse entityID")
			continue
		}
		if entity.Type == model.ParticipantPrefix {
			participantIDs = append(participantIDs, entity.ID)
		} else if entity.Type == model.DiscussionPrefix {
			discussionIDs = append(discussionIDs, entity.ID)
		} else {
			// TODO: Log to cloudwatch
			logrus.Debugf("MentionedEntity using an unsupported type: %v\n", entityID)
			continue
		}
	}

	if len(participantIDs) == 0 && len(discussionIDs) == 0 {
		return nil, nil
	}

	participants, err := d.GetParticipantsByIDs(ctx, participantIDs)
	if err != nil {
		logrus.WithError(err).Error("failed to GetParticipantsWithIDs")
		return nil, err
	}

	discussions, err := d.GetDiscussionsByIDs(ctx, discussionIDs)
	if err != nil {
		logrus.WithError(err).Error("failed to GetDiscussionsByIDs")
		return nil, err
	}

	for k, v := range participants {
		if v != nil {
			key := strings.Join([]string{model.ParticipantPrefix, k}, ":")
			entities[key] = v
		}
	}

	for k, v := range discussions {
		if v != nil {
			key := strings.Join([]string{model.DiscussionPrefix, k}, ":")
			entities[key] = v
		}
	}

	return entities, nil
}

func (d *delphisBackend) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	return d.db.GetPostByID(ctx, id)
}

// This filters the post to ensure it belongs to the correct discussionID!
func (d *delphisBackend) GetPostByDiscussionPostID(ctx context.Context, discussionID, postID string) (*model.Post, error) {
	post, err := d.db.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post != nil && (post.DiscussionID == nil || *post.DiscussionID != discussionID) {
		return nil, nil
	}

	return post, nil
}

func (d *delphisBackend) DeletePostByID(ctx context.Context, discussionID string, postID string, requestingUserID string) (*model.Post, error) {
	disc, err := d.GetDiscussionByID(ctx, discussionID)
	if err != nil || disc == nil {
		return nil, fmt.Errorf("Discussion not found")
	}

	moderatorObj, err := d.GetModeratorByID(ctx, *disc.ModeratorID)
	if err != nil || moderatorObj == nil || moderatorObj.UserProfileID == nil {
		return nil, fmt.Errorf("Failed to retrieve discussion")
	}

	userProfile, err := d.GetUserProfileByID(ctx, *moderatorObj.UserProfileID)
	if err != nil || userProfile == nil || userProfile.UserID == nil {
		return nil, fmt.Errorf("Failed to retrieve discussion")
	}

	post, err := d.GetPostByID(ctx, postID)
	if err != nil || post == nil || post.ParticipantID == nil {
		return nil, fmt.Errorf("Post not found")
	}

	participant, err := d.GetParticipantByID(ctx, *post.ParticipantID)
	if err != nil || participant == nil {
		return nil, fmt.Errorf("Participant not found")
	}

	isModerator := *userProfile.UserID == requestingUserID
	isParticipant := participant.UserID != nil && *participant.UserID == requestingUserID

	// Only the moderator or author can delete a post
	if !isModerator && !isParticipant {
		return nil, fmt.Errorf("Only moderator or author can delete a post")
	}

	if post.DeletedAt != nil {
		// This has already been deleted. Make this idempotent.
		return post, nil
	}

	deletedReasonCode := model.PostDeletedReasonModeratorRemoved
	if isParticipant {
		deletedReasonCode = model.PostDeletedReasonParticipantRemoved
	}

	return d.db.DeletePostByID(ctx, postID, deletedReasonCode)
}

func validatePostParams(ctx context.Context, input model.PostContentInput) error {
	// Validate post type
	if input.PostType == "" {
		return fmt.Errorf("PostType must not be empty")
	}

	// Validate mentionedEntities
	if input.MentionedEntities != nil {
		tokens := regexp.MustCompile(`\<(\d+)\>`).FindAllStringSubmatch(input.PostText, -1)
		if len(tokens) != len(input.MentionedEntities) {
			return errors.New("tokens did not match entities")
		}
	}

	return nil
}
