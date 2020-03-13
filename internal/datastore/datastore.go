package datastore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/nedrocks/delphisbe/graph/model"
	"github.com/nedrocks/delphisbe/internal/config"
)

type Datastore interface {
	GetDiscussionByID(ctx context.Context, id string) (*model.Discussion, error)
	GetDiscussionsByIDs(ctx context.Context, ids []string) (map[string]*model.Discussion, error)
	ListDiscussions(ctx context.Context) (*model.DiscussionsConnection, error)
	PutDiscussion(ctx context.Context, discussion model.Discussion) (*model.Discussion, error)
	GetParticipantsByDiscussionID(ctx context.Context, id string) ([]model.Participant, error)
	GetParticipantsByIDs(ctx context.Context, discussionParticipantKeys []model.DiscussionParticipantKey) (map[model.DiscussionParticipantKey]*model.Participant, error)
	PutParticipant(ctx context.Context, participant model.Participant) (*model.Participant, error)
	AddParticipantToUser(ctx context.Context, userID string, discussionParticipant model.DiscussionParticipantKey) (*model.User, error)
	GetPostsByDiscussionID(ctx context.Context, discussionID string) ([]*model.Post, error)
	PutPost(ctx context.Context, post model.Post) (*model.Post, error)
	AddViewerToUser(ctx context.Context, userID string, discussionViewerKey model.DiscussionViewerKey) (*model.User, error)
	GetUserProfileByID(ctx context.Context, id string) (*model.UserProfile, error)
	AddModeratedDiscussionToUserProfile(ctx context.Context, userProfileID string, discussionID string) (*model.UserProfile, error)
	CreateOrUpdateUserProfile(ctx context.Context, userProfile model.UserProfile) (*model.UserProfile, bool, error)
	UpdateUserProfileTwitterInfo(ctx context.Context, userProfile model.UserProfile, twitterInfo model.SocialInfo) (*model.UserProfile, error)
	UpdateUserProfileUserID(ctx context.Context, userProfileID string, userID string) (*model.UserProfile, error)
	PutUser(ctx context.Context, user model.User) (*model.User, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	GetViewersByIDs(ctx context.Context, discussionViewerKeys []model.DiscussionViewerKey) (map[model.DiscussionViewerKey]*model.Viewer, error)
	PutViewer(ctx context.Context, viewer model.Viewer) (*model.Viewer, error)

	marshalMap(in interface{}) (map[string]*dynamodb.AttributeValue, error)
}

type db struct {
	dynamo   dynamodbiface.DynamoDBAPI
	dbConfig config.TablesConfig
	encoder  *dynamodbattribute.Encoder
}

//MarshalMap wraps the dynamodbattribute.MarshalMap with a defined encoder.
func (d *db) marshalMap(in interface{}) (map[string]*dynamodb.AttributeValue, error) {
	av, err := d.encoder.Encode(in)
	if err != nil || av == nil || av.M == nil {
		return map[string]*dynamodb.AttributeValue{}, err
	}

	return av.M, nil
}

func NewDatastore(dbConfig config.DBConfig, awsSession *session.Session) Datastore {
	mySession := awsSession
	if dbConfig.Host != "" && dbConfig.Port != 0 {
		mySession = mySession.Copy(awsSession.Config.WithEndpoint(fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port)))
		logrus.Debugf("endpoint: %s", mySession.Config.Endpoint)
	}
	dbSvc := dynamodb.New(mySession)
	return &db{
		dbConfig: dbConfig.TablesConfig,
		dynamo:   dbSvc,
		encoder: &dynamodbattribute.Encoder{
			MarshalOptions: dynamodbattribute.MarshalOptions{
				SupportJSONTags: false,
			},
			NullEmptyString: true,
		},
	}
}
