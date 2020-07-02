package backend

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nedrocks/delphisbe/graph/model"

	"github.com/nedrocks/delphisbe/internal/cache"
	"github.com/nedrocks/delphisbe/internal/config"
	"github.com/nedrocks/delphisbe/internal/util"
	"github.com/nedrocks/delphisbe/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestDelphisBackend_AutoPostContent(t *testing.T) {
	ctx := context.Background()

	discussionID := "discussionID"
	limit := 10
	userID := model.ConciergeUser
	contentID := "contentID"

	apObj := model.DiscussionAutoPost{
		ID:          discussionID,
		IdleMinutes: limit,
	}

	icObj := model.ImportedContent{
		ID: contentID,
	}

	parObj := model.Participant{
		ID: "participantID",
	}

	tx := sql.Tx{}

	discObj := model.Discussion{ID: discussionID}
	matchingTags := []string{"tag1"}

	Convey("AutoPostContent", t, func() {
		now := time.Now()
		cacheObj := cache.NewInMemoryCache()
		mockAuth := &mocks.DelphisAuth{}
		mockDB := &mocks.Datastore{}
		backendObj := &delphisBackend{
			db:              mockDB,
			auth:            mockAuth,
			cache:           cacheObj,
			discussionMutex: sync.Mutex{},
			config:          config.Config{},
			timeProvider:    &util.FrozenTime{NowTime: now},
		}

		Convey("when GetDiscussionsForAutoPost errors outs", func() {
			expectedError := fmt.Errorf("Some Error")
			mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
			mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return(nil, expectedError)

			backendObj.AutoPostContent()
		})

		Convey("when checkIdleTime errors outs", func() {
			expectedError := fmt.Errorf("Some Error")
			mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
			mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
			mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, expectedError)

			backendObj.AutoPostContent()
		})

		Convey("when postNextContent errors outs", func() {
			Convey("when GetScheduledImportedContentByDiscussionID errors outs", func() {
				expectedError := fmt.Errorf("Some Error")
				mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
				mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
				mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, nil)
				mockDB.On("GetScheduledImportedContentByDiscussionID", ctx, discussionID).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return(nil, expectedError)

				backendObj.AutoPostContent()
			})

			Convey("when GetImportedContentByDiscussionID errors outs", func() {
				expectedError := fmt.Errorf("Some Error")
				mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
				mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
				mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, nil)
				mockDB.On("GetScheduledImportedContentByDiscussionID", ctx, discussionID).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return(nil, nil)
				mockDB.On("GetImportedContentByDiscussionID", ctx, discussionID, limit).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return(nil, expectedError)

				backendObj.AutoPostContent()
			})

			Convey("when GetImportedContentByDiscussionID returns 0 content", func() {
				mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
				mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
				mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, nil)
				mockDB.On("GetScheduledImportedContentByDiscussionID", ctx, discussionID).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return(nil, nil)
				mockDB.On("GetImportedContentByDiscussionID", ctx, discussionID, limit).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return(nil, nil)

				backendObj.AutoPostContent()
			})

			Convey("when GetParticipantsByDiscussionIDUserID errors outs", func() {
				expectedError := fmt.Errorf("Some Error")
				mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
				mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
				mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, nil)
				mockDB.On("GetScheduledImportedContentByDiscussionID", ctx, discussionID).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return([]*model.ImportedContent{&icObj}, nil)
				mockDB.On("GetParticipantsByDiscussionIDUserID", ctx, discussionID, userID).Return(nil, expectedError)

				backendObj.AutoPostContent()
			})

			Convey("when GetParticipantsByDiscussionIDUserID returns only an anonymous concierge", func() {
				tempParObj := parObj
				tempParObj.IsAnonymous = true
				mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
				mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
				mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, nil)
				mockDB.On("GetScheduledImportedContentByDiscussionID", ctx, discussionID).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return([]*model.ImportedContent{&icObj}, nil)
				mockDB.On("GetParticipantsByDiscussionIDUserID", ctx, discussionID, userID).Return([]model.Participant{tempParObj}, nil)

				backendObj.AutoPostContent()
			})

			Convey("when PostImportedContent errors out", func() {
				expectedError := fmt.Errorf("Some Error")
				mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
				mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
				mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, nil)
				mockDB.On("GetScheduledImportedContentByDiscussionID", ctx, discussionID).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return([]*model.ImportedContent{&icObj}, nil)
				mockDB.On("GetParticipantsByDiscussionIDUserID", ctx, discussionID, userID).Return([]model.Participant{parObj}, nil)
				mockDB.On("GetImportedContentByID", ctx, contentID).Return(nil, expectedError)

				backendObj.AutoPostContent()
			})

			Convey("when auto posting succeeds", func() {
				mockDB.On("GetDiscussionsAutoPost", ctx).Return(&mockDiscAutoPostIter{})
				mockDB.On("DiscussionAutoPostIterCollect", ctx, mock.Anything).Return([]*model.DiscussionAutoPost{&apObj}, nil)
				mockDB.On("GetLastPostByDiscussionID", ctx, discussionID, limit).Return(nil, nil)
				mockDB.On("GetScheduledImportedContentByDiscussionID", ctx, discussionID).Return(&mockImportedContentIter{})
				mockDB.On("ContentIterCollect", ctx, mock.Anything).Return([]*model.ImportedContent{&icObj}, nil)
				mockDB.On("GetParticipantsByDiscussionIDUserID", ctx, discussionID, userID).Return([]model.Participant{parObj}, nil)

				mockDB.On("GetImportedContentByID", ctx, contentID).Return(&icObj, nil)

				// Create post functions
				mockDB.On("BeginTx", ctx).Return(&tx, nil)
				mockDB.On("PutPostContent", ctx, mock.Anything, mock.Anything).Return(nil)
				mockDB.On("PutPost", ctx, mock.Anything, mock.Anything).Return(&model.Post{ID: "post123"}, nil)
				mockDB.On("PutActivity", ctx, mock.Anything, mock.Anything).Return(nil)
				mockDB.On("CommitTx", ctx, mock.Anything).Return(nil)
				mockDB.On("GetDiscussionByID", ctx, mock.Anything).Return(&discObj, nil)
				mockDB.On("GetParticipantsByDiscussionID", ctx, mock.Anything, mock.Anything).Return([]model.Participant{parObj}, nil)

				// Put Imported Content Queue
				mockDB.On("GetMatchingTags", ctx, discussionID, contentID).Return(matchingTags, nil)
				mockDB.On("UpdateImportedContentDiscussionQueue", ctx, discussionID, contentID, mock.Anything).Return(
					&model.ContentQueueRecord{DiscussionID: discussionID}, nil)
				mockDB.On("PutImportedContentDiscussionQueue", ctx, discussionID, contentID, time.Now(), matchingTags).Return(&icObj, nil)
				backendObj.AutoPostContent()
			})
		})
	})
}
