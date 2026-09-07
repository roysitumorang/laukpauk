package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		tx pgx.Tx,
		request *models.Notification,
	) (*models.Notification, error)
	UpdateNotification(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Notification,
	) (*models.Notification, error)
}

type notificationRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewNotificationRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
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
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(&total)
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
		, admin_id
		, seller_id
		, buyer_id
		, notification_type
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
		_, _ = fmt.Fprintf(sb, " %s %s", item.Field, item.Direction)
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
	rows, err := q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
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
			&notification.ID,
			&notification.AdminID,
			&notification.SellerID,
			&notification.BuyerID,
			&notification.NotificationType,
			&notification.Subject,
			&notification.Body,
			&notification.Photo,
			&notification.Params,
			&notification.CreatedBy,
			&notification.CreatedAt,
			&notification.ReadAt,
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
			, admin_id
			, seller_id
			, buyer_id
			, notification_type
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
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.NotificationType,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.Params,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.ReadAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
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
	tx pgx.Tx,
	request *models.Notification,
) (*models.Notification, error) {
	ctxt := "NotificationRepository-CreateNotification"
	var response models.Notification
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO notifications (
			admin_id
			, seller_id
			, buyer_id
			, notification_type
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
			, admin_id
			, seller_id
			, buyer_id
			, notification_type
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
			, sent_at
			, read_at`,
		request.AdminID,
		request.SellerID,
		request.BuyerID,
		request.NotificationType,
		request.Subject,
		request.Body,
		request.Photo,
		request.Params,
		request.CreatedBy,
		request.CreatedAt,
	).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.NotificationType,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.Params,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.SentAt,
		&response.ReadAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *notificationRepository) UpdateNotification(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Notification,
) (*models.Notification, error) {
	ctxt := "NotificationRepository-UpdateNotification"
	var response models.Notification
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, admin_id
			, seller_id
			, buyer_id
			, notification_type
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
			, sent_at
			, read_at
		FROM notifications
		WHERE id = $1
		FOR NO KEY UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.NotificationType,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.Params,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.SentAt,
		&response.ReadAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
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
		_, _ = fmt.Fprintf(sb, " sent_at = $%d", n)
	}
	if (response.ReadAt == nil && request.ReadAt != nil) ||
		(response.ReadAt != nil && request.ReadAt == nil) ||
		(response.ReadAt != nil && request.ReadAt != nil &&
			!response.ReadAt.Equal(*request.ReadAt)) {
		params = append(params, request.ReadAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " read_at = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, admin_id
			, seller_id
			, buyer_id
			, notification_type
			, subject
			, body
			, photo
			, params
			, created_by
			, created_at
			, sent_at
			, read_at`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.NotificationType,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.Params,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.SentAt,
		&response.ReadAt,
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.AdminIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("admin_id IN (")
		for i, adminID := range request.AdminIDs {
			params = append(params, adminID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.SellerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("seller_id IN (")
		for i, sellerID := range request.SellerIDs {
			params = append(params, sellerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.BuyerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("buyer_id IN (")
		for i, buyerID := range request.BuyerIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.NotificationTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("message_type IN (")
		for i, notificationType := range request.NotificationTypes {
			params = append(params, notificationType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
		params = append(params, sb.String())
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(
			sb,
			`(EXISTS(
				SELECT 1
				FROM admins
				WHERE admins.id = notifications.admin_id
					AND (
						LOWER(admins.name) LIKE $%d
						OR admins.username LIKE $%d
						OR admins.email LIKE $%d
						OR admins.mobile_phone LIKE $%d
					)
			) OR EXISTS(
				SELECT 1
				FROM sellers
				WHERE sellers.id = notifications.seller_id
					AND (
						LOWER(sellers.name) LIKE $%d
						OR sellers.username LIKE $%d
						OR sellers.email LIKE $%d
						OR sellers.mobile_phone LIKE $%d
					)
			) OR EXISTS(
				SELECT 1
				FROM buyers
				WHERE buyers.id = notifications.buyer_id
					AND (
						LOWER(buyers.name) LIKE $%d
						OR buyers.username LIKE $%d
						OR buyers.mobile_phone LIKE $%d
					)
			))`,
			n, n, n, n, n, n, n, n, n, n, n,
		)
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
