package datastore

import (
	"context"
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/nedrocks/delphisbe/graph/model"
	"github.com/sirupsen/logrus"
)

func (d *delphisDB) PutPost(ctx context.Context, tx *sql.Tx, post model.Post) (*model.Post, error) {
	logrus.Debug("PutPost::SQL Create")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("PutPost::failed to initialize statements")
		return nil, err
	}

	err := tx.StmtContext(ctx, d.prepStmts.putPostStmt).QueryRowContext(
		ctx,
		post.ID,
		post.DiscussionID,
		post.ParticipantID,
		post.PostContent.ID,
		post.QuotedPostID,
		post.MediaID,
		post.ImportedContentID,
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.DiscussionID,
		&post.ParticipantID,
		&post.PostContentID,
		&post.QuotedPostID,
		&post.MediaID,
		&post.ImportedContentID,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to execute putPostStmt")
		return nil, err
	}

	logrus.Infof("Post: %v\n", post)

	// Do we want to query for the post or just return the post?
	return &post, nil
}

// TODO: Add created_at and limit
func (d *delphisDB) GetPostsByDiscussionIDIter(ctx context.Context, discussionID string) PostIter {
	logrus.Debug("GetPostsByDiscussionID::SQL Query")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("GetPostsByDiscussionIDIter::failed to initialize statements")
		return &postIter{err: err}
	}

	rows, err := d.prepStmts.getPostsByDiscussionIDStmt.QueryContext(
		ctx,
		discussionID,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to query GetPostsByDiscussionID")
		return &postIter{err: err}
	}

	return &postIter{
		ctx:  ctx,
		rows: rows,
	}
}

func (d *delphisDB) GetLastPostByDiscussionID(ctx context.Context, discussionID string, minutes int) (*model.Post, error) {
	logrus.Debug("GetLastPostByDiscussionID::SQL Query")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("GetLastPostByDiscussionID::failed to initialize statements")
		return nil, err
	}

	post := model.Post{}
	postContent := model.PostContent{}
	if err := d.prepStmts.getLastPostByDiscussionIDStmt.QueryRowContext(
		ctx,
		discussionID,
		minutes,
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.DeletedAt,
		&post.DeletedReasonCode,
		&post.DiscussionID,
		&post.ParticipantID,
		&post.QuotedPostID,
		&post.MediaID,
		&post.ImportedContentID,
		&postContent.ID,
		&postContent.Content,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logrus.WithError(err).Error("failed to get last post")
		return nil, err
	}

	post.PostContent = &postContent

	return &post, nil
}

func (d *delphisDB) GetPostsByDiscussionID(ctx context.Context, discussionID string) ([]*model.Post, error) {
	logrus.Debug("GetPostsByDiscussionID::SQL Query")
	posts := []model.Post{}
	if err := d.sql.Where(model.Post{DiscussionID: &discussionID}).Preload("PostContent").Find(&posts).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Not sure if this will return not found error... If the discussion is empty maybe?
			// Should this be nil, nil?
			return []*model.Post{}, nil
		}
		logrus.WithError(err).Errorf("Failed to get posts by discussionID")
		return nil, err
	}

	logrus.Debugf("Found posts: %+v", posts)

	returnedPosts := []*model.Post{}
	for i := range posts {
		if posts[i].QuotedPostID != nil {
			var err error
			posts[i].QuotedPost, err = d.GetPostByID(ctx, *posts[i].QuotedPostID)
			if err != nil {
				// Do we want to fail the whole discussion if we can't get a quote?
				return nil, err
			}
		}
		returnedPosts = append(returnedPosts, &posts[i])

	}

	return returnedPosts, nil
}

// TODO: rewrite for single posts
func (d *delphisDB) GetPostByID(ctx context.Context, postID string) (*model.Post, error) {
	logrus.Debug("GetPostByID::SQL Query")
	post := model.Post{}
	// TODO: Clean up for single queries
	if err := d.sql.Where([]string{postID}).Preload("PostContent").Find(&post).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		logrus.WithError(err).Errorf("Failed to get Post by ID")
		return nil, err
	}

	return &post, nil
}

type postIter struct {
	err  error
	ctx  context.Context
	rows *sql.Rows
}

func (iter *postIter) Next(post *model.Post) bool {
	if iter.err != nil {
		logrus.WithError(iter.err).Error("iterator error")
		return false
	}

	if iter.err = iter.ctx.Err(); iter.err != nil {
		logrus.WithError(iter.err).Error("iterator context error")
		return false
	}

	if !iter.rows.Next() {
		return false
	}
	postContent := model.PostContent{}

	if iter.err = iter.rows.Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.DeletedAt,
		&post.DeletedReasonCode,
		&post.DiscussionID,
		&post.ParticipantID,
		&post.QuotedPostID,
		&post.MediaID,
		&post.ImportedContentID,
		&postContent.ID,
		&postContent.Content,
	); iter.err != nil {
		logrus.WithError(iter.err).Error("iterator failed to scan row")
		return false
	}

	post.PostContent = &postContent

	return true

}

func (iter *postIter) Close() error {
	if err := iter.err; err != nil {
		logrus.WithError(err).Error("iter error on close")
		return err
	}
	if err := iter.rows.Close(); err != nil {
		logrus.WithError(err).Error("iter rows close on close")
	}

	return nil
}

///////////////
// Dynamo functions
///////////////

// func (d *db) PutPostDynamo(ctx context.Context, post model.Post) (*model.Post, error) {
// 	logrus.Debug("PutPost: DynamoDB PutItem")
// 	av, err := d.marshalMap(post)
// 	if err != nil {
// 		logrus.WithError(err).Errorf("PutPost: Failed to marshal post object: %+v", post)
// 		return nil, err
// 	}
// 	_, err = d.dynamo.PutItem(&dynamodb.PutItemInput{
// 		TableName: aws.String(d.dbConfig.Posts.TableName),
// 		Item:      av,
// 	})

// 	if err != nil {
// 		logrus.WithError(err).Errorf("PutPost: Failed to put post object: %+v", av)
// 		return nil, err
// 	}

// 	return &post, nil
// }

// func (d *db) GetPostsByDiscussionIDDynamo(ctx context.Context, discussionID string) ([]*model.Post, error) {
// 	logrus.Debug("GetPostsByDiscussionID: DynamoDB Query")
// 	res, err := d.dynamo.Query(&dynamodb.QueryInput{
// 		TableName: aws.String(d.dbConfig.Posts.TableName),
// 		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
// 			":did": {
// 				S: aws.String(discussionID),
// 			},
// 		},
// 		KeyConditionExpression: a
