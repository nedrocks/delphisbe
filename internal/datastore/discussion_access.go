package datastore

import (
	"context"
	"database/sql"
	"io"

	"github.com/lib/pq"

	"github.com/sirupsen/logrus"

	"github.com/delphis-inc/delphisbe/graph/model"
)

func (d *delphisDB) GetDiscussionsByUserAccess(ctx context.Context, userID string, state model.DiscussionUserAccessState) DiscussionIter {
	logrus.Debug("GetDiscussionsByUserAccess::SQL Query")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("GetDiscussionsByUserAccess::failed to initialize statements")
		return &discussionIter{err: err}
	}

	rows, err := d.prepStmts.getDiscussionsByUserAccessStmt.QueryContext(
		ctx,
		userID,
		state,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to query GetDiscussionsByUserAccess")
		return &discussionIter{err: err}
	}

	return &discussionIter{
		ctx:  ctx,
		rows: rows,
	}
}

func (d *delphisDB) GetDiscussionUserAccess(ctx context.Context, discussionID, userID string) (*model.DiscussionUserAccess, error) {
	logrus.Debug("GetDiscussionUserAccess::SQL Query")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("GetDiscussionUserAccess::failed to initialize statements")
		return nil, err
	}

	dua := model.DiscussionUserAccess{}
	if err := d.prepStmts.getDiscussionUserAccessStmt.QueryRowContext(
		ctx,
		discussionID,
		userID,
	).Scan(
		&dua.DiscussionID,
		&dua.UserID,
		&dua.State,
		&dua.RequestID,
		&dua.NotifSetting,
		&dua.CreatedAt,
		&dua.UpdatedAt,
		&dua.DeletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logrus.WithError(err).Error("failed to query GetDiscussionsByUserAccess")
		return nil, err
	}

	return &dua, nil
}

func (d *delphisDB) GetDUAForEverythingNotifications(ctx context.Context, discussionID, userID string) DiscussionUserAccessIter {
	logrus.Debug("GetDUAForEverythingNotifications::SQL Query")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("GetDUAForEverythingNotifications::failed to initialize statements")
		return &duaIter{err: err}
	}

	rows, err := d.prepStmts.getDUAForEverythingNotificationsStmt.QueryContext(
		ctx,
		discussionID,
		userID,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to query getDUAForEverythingNotificationsString")
		return &duaIter{err: err}
	}

	return &duaIter{
		ctx:  ctx,
		rows: rows,
	}
}

func (d *delphisDB) GetDUAForMentionNotifications(ctx context.Context, discussionID string, userID string, mentionedUserIDs []string) DiscussionUserAccessIter {
	logrus.Debug("GetDUAForMentionNotifications::SQL Query")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("GetDUAForMentionNotifications::failed to initialize statements")
		return &duaIter{err: err}
	}

	rows, err := d.prepStmts.getDUAForMentionNotificationsStmt.QueryContext(
		ctx,
		discussionID,
		userID,
		pq.Array(mentionedUserIDs),
	)
	if err != nil {
		logrus.WithError(err).Error("failed to query getDUAForMentionNotificationsString")
		return &duaIter{err: err}
	}

	return &duaIter{
		ctx:  ctx,
		rows: rows,
	}
}

func (d *delphisDB) UpsertDiscussionUserAccess(ctx context.Context, tx *sql.Tx, dua model.DiscussionUserAccess) (*model.DiscussionUserAccess, error) {
	logrus.Debug("UpsertDiscussionUserAccess::SQL Create")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("UpsertDiscussionUserAccess::failed to initialize statements")
		return nil, err
	}

	if err := tx.StmtContext(ctx, d.prepStmts.upsertDiscussionUserAccessStmt).QueryRowContext(
		ctx,
		dua.DiscussionID,
		dua.UserID,
		dua.State,
		dua.RequestID,
		dua.NotifSetting,
	).Scan(
		&dua.DiscussionID,
		&dua.UserID,
		&dua.State,
		&dua.RequestID,
		&dua.NotifSetting,
		&dua.CreatedAt,
		&dua.UpdatedAt,
		&dua.DeletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return &model.DiscussionUserAccess{}, nil
		}
		logrus.WithError(err).Error("failed to execute upsertDiscussionUserAccess")
		return nil, err
	}

	return &dua, nil
}

func (d *delphisDB) DeleteDiscussionUserAccess(ctx context.Context, tx *sql.Tx, discussionID, userID string) (*model.DiscussionUserAccess, error) {
	logrus.Debug("DeleteDiscussionUserAccess::SQL Delete")
	if err := d.initializeStatements(ctx); err != nil {
		logrus.WithError(err).Error("DeleteDiscussionUserAccess::failed to initialize statements")
		return nil, err
	}

	dua := model.DiscussionUserAccess{}
	if err := tx.StmtContext(ctx, d.prepStmts.deleteDiscussionUserAccessStmt).QueryRowContext(
		ctx,
		discussionID,
		userID,
	).Scan(
		&dua.DiscussionID,
		&dua.UserID,
		&dua.CreatedAt,
		&dua.UpdatedAt,
		&dua.DeletedAt,
	); err != nil {
		logrus.WithError(err).Error("failed to execute deleteDiscussionUserAccessStmt")
		return nil, err
	}

	return &dua, nil
}

func (d *delphisDB) DuaIterCollect(ctx context.Context, iter DiscussionUserAccessIter) ([]*model.DiscussionUserAccess, error) {
	var duaArr []*model.DiscussionUserAccess
	dua := model.DiscussionUserAccess{}

	defer iter.Close()

	for iter.Next(&dua) {
		tempDua := dua

		duaArr = append(duaArr, &tempDua)
	}

	if err := iter.Close(); err != nil && err != io.EOF {
		logrus.WithError(err).Error("failed to close iter")
		return nil, err
	}

	return duaArr, nil
}

type duaIter struct {
	err  error
	ctx  context.Context
	rows *sql.Rows
}

func (iter *duaIter) Next(dua *model.DiscussionUserAccess) bool {
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

	if iter.err = iter.rows.Scan(
		&dua.DiscussionID,
		&dua.UserID,
		&dua.State,
		&dua.RequestID,
		&dua.NotifSetting,
		&dua.CreatedAt,
		&dua.UpdatedAt,
		&dua.DeletedAt,
	); iter.err != nil {
		logrus.WithError(iter.err).Error("iterator failed to scan row")
		return false
	}

	return true
}

func (iter *duaIter) Close() error {
	if err := iter.err; err != nil {
		logrus.WithError(err).Error("iter error on close")
		return err
	}
	if err := iter.rows.Close(); err != nil {
		logrus.WithError(err).Error("iter rows close on close")
		return err
	}

	return nil
}
