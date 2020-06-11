package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	"github.com/nedrocks/delphisbe/graph/generated"
	"github.com/nedrocks/delphisbe/graph/model"
	"github.com/sirupsen/logrus"
)

func (r *postResolver) IsDeleted(ctx context.Context, obj *model.Post) (bool, error) {
	return obj.DeletedAt != nil, nil
}

func (r *postResolver) Content(ctx context.Context, obj *model.Post) (string, error) {
	if obj.PostContent == nil && obj.PostContentID != nil {
		postContent, err := r.DAOManager.GetPostContentByID(ctx, *obj.PostContentID)
		if err != nil {
			return "", err
		}
		obj.PostContent = postContent
	}
	if obj.PostContent == nil {
		// If it is still nil we should return an empty string I guess?
		logrus.Errorf("PostContent is nil for post ID: %s", obj.ID)
		return "", nil
	}
	return obj.PostContent.Content, nil
}

func (r *postResolver) Discussion(ctx context.Context, obj *model.Post) (*model.Discussion, error) {
	if obj.Discussion == nil && obj.DiscussionID != nil {
		res, err := r.DAOManager.GetDiscussionByID(ctx, *obj.DiscussionID)

		if err != nil {
			return nil, err
		}
		obj.Discussion = res
	}
	return obj.Discussion, nil
}

func (r *postResolver) Participant(ctx context.Context, obj *model.Post) (*model.Participant, error) {
	if obj.Participant == nil && obj.ParticipantID != nil {
		participant, err := r.DAOManager.GetParticipantByID(ctx, *obj.ParticipantID)

		if err != nil {
			return nil, err
		}

		obj.Participant = participant
	}
	return obj.Participant, nil
}

func (r *postResolver) CreatedAt(ctx context.Context, obj *model.Post) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *postResolver) UpdatedAt(ctx context.Context, obj *model.Post) (string, error) {
	return obj.UpdatedAt.Format(time.RFC3339), nil
}

func (r *postResolver) MentionedEntities(ctx context.Context, obj *model.Post) ([]model.Entity, error) {
	if len(obj.PostContent.MentionedEntities) == 0 {
		return nil, nil
	}

	mentionedEntities, err := r.DAOManager.GetMentionedEntities(ctx, obj.PostContent.MentionedEntities)
	if err != nil {
		return nil, err
	}

	// Iterate over IDs to return mentionedEntities in the proper order
	var entities []model.Entity
	for _, entityID := range obj.PostContent.MentionedEntities {
		if _, ok := mentionedEntities[entityID]; !ok {
			entities = append(entities, &model.UnknownEntity{})
		} else {
			entities = append(entities, mentionedEntities[entityID])
		}
	}

	return entities, nil
}

func (r *postResolver) Media(ctx context.Context, obj *model.Post) (*model.Media, error) {
	if obj.MediaID != nil {
		return r.DAOManager.GetMediaRecord(ctx, *obj.MediaID)
	}
	return nil, nil
}

func (r *postResolver) ImportedContent(ctx context.Context, obj *model.Post) (*model.ImportedContent, error) {
	if obj.ImportedContentID != nil {
		return r.DAOManager.GetImportedContentByID(ctx, *obj.ImportedContentID)
	}

	return nil, nil
}

// Post returns generated.PostResolver implementation.
func (r *Resolver) Post() generated.PostResolver { return &postResolver{r} }

type postResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *postResolver) PostType(ctx context.Context, obj *model.Post) (model.PostType, error) {
	panic(fmt.Errorf("not implemented"))
}
