package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type NotificationRepository interface {
	FindNotifications(
		ctx context.Context,
		request *models.NotificationFilter,
	) ([]*models.Notification, int64, int64, error)
	FindNotification(
		ctx context.Context,
		request *models.NotificationFilter,
	) (*models.Notification, error)
	CreateNotification(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Notification,
	) (*models.Notification, error)
	UpdateNotification(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Notification,
	) (*models.Notification, error)
}

type notificationRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewNotificationRepository(
	dbRead,
	dbWrite *sql.DB,
) NotificationRepository {
	return &notificationRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *notificationRepository) FindNotifications(
	ctx context.Context,
	request *models.NotificationFilter,
) ([]*models.Notification, int64, int64, error) {
	ctxt := "NotificationRepository-FindNotifications"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM notifications %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRowContext(ctx, query, helper.SQLArgs(params...)...).Scan(helper.SQLScan(&total)...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return nil, 0, 0, err
	}
	if total == 0 {
		return nil, 0, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`id
		, user_id
		, notification_type
		, link
		, link_id
		, subject
		, body
		, photo
		, params
		, created_by
		, created_at
		, read_at`,
	)
	sb.Reset()
	_, _ = sb.WriteString(query)
	sortBy := request.Sortings
	if len(sortBy) == 0 {
		sortBy = append(sortBy, models.DefaultSortBy)
	}
	_, _ = sb.WriteString(" ORDER BY")
	for i, item := range sortBy {
		if i > 0 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " %s %s", sortExpression(item.Field), item.Direction)
	}
	pages := int64(1)
	var offset int64
	if request.Limit > 0 {
		totalDecimal, err := decimal.New(total, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		perPageDecimal, err := decimal.New(request.Limit, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		pagesDecimal, err := totalDecimal.Quo(perPageDecimal)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuo")
			return nil, 0, 0, err
		}
		pages, _, _ = pagesDecimal.Ceil(0).Int64(0)
		offset = request.Page * request.Limit
		_, _ = fmt.Fprintf(sb, " LIMIT %d OFFSET %d", request.Limit, offset)
	}
	rows, err := q.dbRead.QueryContext(ctx, sb.String(), helper.SQLArgs(params...)...)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.Notification
	for rows.Next() {
		var notification models.Notification
		if err = rows.Scan(
			helper.SQLScan(
				&notification.ID,
				&notification.UserID,
				&notification.NotificationType,
				&notification.Link,
				&notification.LinkID,
				&notification.Subject,
				&notification.Body,
				&notification.Photo,
				&notification.Params,
				&notification.CreatedBy,
				&notification.CreatedAt,
				&notification.ReadAt,
			)...,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		notification.RowNo = offset
		response = append(response, &notification)
	}
	return response, total, pages, nil
}

func (q *notificationRepository) FindNotification(
	ctx context.Context,
	request *models.NotificationFilter,
) (*models.Notification, error) {
	ctxt := "NotificationRepository-FindNotification"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`SELECT
			id
			, user_id
			, notification_type
			, link
			, link_id
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
			, read_at
		FROM notifications %s`,
		condition,
	)
	var response models.Notification
	err := q.dbRead.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.UserID,
			&response.NotificationType,
			&response.Link,
			&response.LinkID,
			&response.Subject,
			&response.Body,
			&response.Photo,
			&response.Params,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.ReadAt,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotificationNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *notificationRepository) CreateNotification(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Notification,
) (*models.Notification, error) {
	ctxt := "NotificationRepository-CreateNotification"
	var response models.Notification
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO notifications (
			user_id
			, notification_type
			, link
			, link_id
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
			, user_id
			, notification_type
			, link
			, link_id
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
			, sent_at
			, read_at`,
		helper.SQLArgs(
			request.UserID,
			request.NotificationType,
			request.Link,
			request.LinkID,
			request.Subject,
			request.Body,
			request.Photo,
			request.Params,
			request.CreatedBy,
			request.CreatedAt,
		)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.UserID,
			&response.NotificationType,
			&response.Link,
			&response.LinkID,
			&response.Subject,
			&response.Body,
			&response.Photo,
			&response.Params,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.SentAt,
			&response.ReadAt,
		)...,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *notificationRepository) UpdateNotification(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Notification,
) (*models.Notification, error) {
	ctxt := "NotificationRepository-UpdateNotification"
	var response models.Notification
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, user_id
			, notification_type
			, link
			, link_id
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
			, sent_at
			, read_at
		FROM notifications
		WHERE id = ?
		FOR UPDATE`,
		helper.SQLArgs(request.ID)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.UserID,
			&response.NotificationType,
			&response.Link,
			&response.LinkID,
			&response.Subject,
			&response.Body,
			&response.Photo,
			&response.Params,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.SentAt,
			&response.ReadAt,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotificationNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE notifications SET")
	if (response.SentAt == nil && request.SentAt != nil) ||
		(response.SentAt != nil && request.SentAt == nil) ||
		(response.SentAt != nil && request.SentAt != nil &&
			!response.SentAt.Equal(*request.SentAt)) {
		params = append(params, request.SentAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" sent_at = ?")
	}
	if (response.ReadAt == nil && request.ReadAt != nil) ||
		(response.ReadAt != nil && request.ReadAt == nil) ||
		(response.ReadAt != nil && request.ReadAt != nil &&
			!response.ReadAt.Equal(*request.ReadAt)) {
		params = append(params, request.ReadAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" read_at = ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.ID)
	_, _ = sb.WriteString(` WHERE id = ?
		RETURNING id
			, user_id
			, notification_type
			, link
			, link_id
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
			, sent_at
			, read_at`)
	if err = tx.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.UserID,
			&response.NotificationType,
			&response.Link,
			&response.LinkID,
			&response.Subject,
			&response.Body,
			&response.Photo,
			&response.Params,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.SentAt,
			&response.ReadAt,
		)...,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *notificationRepository) buildCondition(request *models.NotificationFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.NotificationIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, notificationID := range request.NotificationIDs {
			params = append(params, notificationID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("user_id IN (")
		for i, userID := range request.UserIDs {
			params = append(params, userID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.NotificationTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("notification_type IN (")
		for i, notificationType := range request.NotificationTypes {
			params = append(params, notificationType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.SentAtBlank != nil {
		if *request.SentAtBlank {
			conds = append(conds, "sent_at IS NULL")
		} else {
			conds = append(conds, "sent_at IS NOT NULL")
		}
	}
	if request.ReadAtBlank != nil {
		if *request.ReadAtBlank {
			conds = append(conds, "read_at IS NULL")
		} else {
			conds = append(conds, "read_at IS NOT NULL")
		}
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		keyword := sb.String()
		params = append(params, keyword, keyword, keyword, keyword)
		sb.Reset()
		_, _ = sb.WriteString(`EXISTS(
				SELECT 1
				FROM users
				WHERE users.id = notifications.user_id
					AND (
						LOWER(users.name) LIKE ?
						OR users.username LIKE ?
						OR users.email LIKE ?
						OR users.mobile_phone LIKE ?
					)
			)`)
		conds = append(conds, sb.String())
	}
	if len(conds) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = fmt.Fprintf(sb, " %s", cond)
		}
		condition = sb.String()
	}
	return
}
