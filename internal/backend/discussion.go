package backend

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/multierr"

	"github.com/sirupsen/logrus"

	"github.com/delphis-inc/delphisbe/graph/model"
	"github.com/delphis-inc/delphisbe/internal/util"
)

const discussionSubscriberKey = "discussion_subscribers-%s"

func (d *delphisBackend) CreateNewDiscussion(ctx context.Context, creatingUser *model.User, anonymityType model.AnonymityType, title string, publicAccess bool) (*model.Discussion, error) {
	moderatorObj := model.Moderator{
		ID:            util.UUIDv4(),
		UserProfileID: &creatingUser.UserProfile.ID,
	}
	_, err := d.db.CreateModerator(ctx, moderatorObj)
	if err != nil {
		return nil, err
	}

	discussionID := util.UUIDv4()
	discussionObj := model.Discussion{
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ID:            discussionID,
		AnonymityType: anonymityType,
		Title:         title,
		ModeratorID:   &moderatorObj.ID,
		PublicAccess:  publicAccess,
	}

	_, err = d.db.UpsertDiscussion(ctx, discussionObj)
	if err != nil {
		return nil, err
	}

	// Create concierge participant
	trueObj := true
	if _, err := d.CreateParticipantForDiscussion(ctx, discussionObj.ID, model.ConciergeUser, model.AddDiscussionParticipantInput{HasJoined: &trueObj}); err != nil {
		logrus.WithError(err).Error("failed to create concierge user")
		return nil, err
	}

	// Create invite links for discussion
	if _, err := d.UpsertInviteLinksByDiscussionID(ctx, discussionID); err != nil {
		logrus.WithError(err).Error("failed to create invite links")
		return nil, err
	}

	return &discussionObj, nil
}

func (d *delphisBackend) UpdateDiscussion(ctx context.Context, id string, input model.DiscussionInput) (*model.Discussion, error) {
	discObj, err := d.db.GetDiscussionByID(ctx, id)
	if err != nil {
		logrus.WithError(err).Error("failed to get discussion by ID")
		return nil, err
	}

	updateDiscussionObj(discObj, input)

	return d.db.UpsertDiscussion(ctx, *discObj)
}

func (d *delphisBackend) GetDiscussionByID(ctx context.Context, id string) (*model.Discussion, error) {
	return d.db.GetDiscussionByID(ctx, id)
}

func (d *delphisBackend) GetDiscussionsByIDs(ctx context.Context, ids []string) (map[string]*model.Discussion, error) {
	return d.db.GetDiscussionsByIDs(ctx, ids)
}

func (d *delphisBackend) GetDiscussionByModeratorID(ctx context.Context, moderatorID string) (*model.Discussion, error) {
	return d.db.GetDiscussionByModeratorID(ctx, moderatorID)
}

func (d *delphisBackend) GetDiscussionsForAutoPost(ctx context.Context) ([]*model.DiscussionAutoPost, error) {
	iter := d.db.GetDiscussionsAutoPost(ctx)
	return d.db.DiscussionAutoPostIterCollect(ctx, iter)
}

func (d *delphisBackend) ListDiscussions(ctx context.Context) (*model.DiscussionsConnection, error) {
	return d.db.ListDiscussions(ctx)
}

func (d *delphisBackend) SubscribeToDiscussion(ctx context.Context, subscriberUserID string, postChannel chan *model.Post, discussionID string) error {
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
	currentSubs[subscriberUserID] = postChannel
	d.cache.Set(cacheKey, currentSubs, time.Hour)
	return nil
}

func (d *delphisBackend) UnSubscribeFromDiscussion(ctx context.Context, subscriberUserID string, discussionID string) error {
	cacheKey := fmt.Sprintf(discussionSubscriberKey, discussionID)
	d.discussionMutex.Lock()
	defer d.discussionMutex.Unlock()
	currentSubsIface, found := d.cache.Get(cacheKey)
	if !found {
		return nil
	}
	var currentSubs map[string]chan *model.Post
	var ok bool
	if currentSubs, ok = currentSubsIface.(map[string]chan *model.Post); !ok {
		currentSubs = map[string]chan *model.Post{}
	}
	delete(currentSubs, subscriberUserID)
	d.cache.Set(cacheKey, currentSubs, time.Hour)
	return nil
}

func (d *delphisBackend) GetDiscussionTags(ctx context.Context, id string) ([]*model.Tag, error) {
	iter := d.db.GetDiscussionTags(ctx, id)
	return d.db.TagIterCollect(ctx, iter)
}

func (d *delphisBackend) PutDiscussionTags(ctx context.Context, discussionID string, tags []string) ([]*model.Tag, error) {
	if len(tags) == 0 {
		return nil, fmt.Errorf("no tags to add")
	}

	var addedTags []*model.Tag

	// Begin tx
	tx, err := d.db.BeginTx(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to begin tx")
		return nil, err
	}

	tagObj := model.Tag{
		ID: discussionID,
	}
	for _, tag := range tags {
		tagObj.Tag = tag
		tagResp, err := d.db.PutDiscussionTags(ctx, tx, tagObj)
		if err != nil {
			logrus.WithError(err).Error("failed to PutDiscussionTags")

			// Rollback on errors
			if txErr := d.db.RollbackTx(ctx, tx); txErr != nil {
				logrus.WithError(txErr).Error("failed to rollback tx")
				return nil, multierr.Append(err, txErr)
			}
			return nil, err
		}
		addedTags = append(addedTags, tagResp)
	}

	// Commit transaction
	if err := d.db.CommitTx(ctx, tx); err != nil {
		logrus.WithError(err).Error("failed to commit post tx")
		return nil, err
	}

	return addedTags, nil
}

func (d *delphisBackend) DeleteDiscussionTags(ctx context.Context, discussionID string, tags []string) ([]*model.Tag, error) {
	if len(tags) == 0 {
		return nil, fmt.Errorf("no tags to delete")
	}

	var deletedTags []*model.Tag

	// Begin tx
	tx, err := d.db.BeginTx(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to begin tx")
		return nil, err
	}

	tagObj := model.Tag{
		ID: discussionID,
	}
	for _, tag := range tags {
		tagObj.Tag = tag
		tagResp, err := d.db.DeleteDiscussionTags(ctx, tx, tagObj)
		if err != nil {
			logrus.WithError(err).Error("failed to PutDiscussionTags")

			// Rollback on errors
			if txErr := d.db.RollbackTx(ctx, tx); txErr != nil {
				logrus.WithError(txErr).Error("failed to rollback tx")
				return nil, multierr.Append(err, txErr)
			}
			return nil, err
		}
		deletedTags = append(deletedTags, tagResp)
	}

	// Commit transaction
	if err := d.db.CommitTx(ctx, tx); err != nil {
		logrus.WithError(err).Error("failed to commit post tx")
		return nil, err
	}

	return deletedTags, nil
}

func updateDiscussionObj(disc *model.Discussion, input model.DiscussionInput) {
	if input.AnonymityType != nil {
		disc.AnonymityType = *input.AnonymityType
	}
	if input.Title != nil {
		disc.Title = *input.Title
	}
	if input.AutoPost != nil {
		disc.AutoPost = *input.AutoPost
	}
	if input.IdleMinutes != nil {
		disc.IdleMinutes = *input.IdleMinutes
	}
	if input.PublicAccess != nil {
		disc.PublicAccess = *input.PublicAccess
	}
	if input.IconURL != nil {
		disc.IconURL = input.IconURL
	}
}

func dedupeDiscussions(discussions []*model.Discussion) []*model.Discussion {
	hashMap := make(map[string]int)

	var results []*model.Discussion
	for _, val := range discussions {
		if _, ok := hashMap[val.ID]; !ok {
			results = append(results, val)
		}
		hashMap[val.ID]++
	}
	return results
}
