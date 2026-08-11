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

type SmsMessageRepository interface {
	FindSmsMessages(
		ctx context.Context,
		request *models.SmsMessageFilter,
	) ([]*models.SmsMessage, int64, int64, error)
	FindSmsMessage(
		ctx context.Context,
		request *models.SmsMessageFilter,
	) (*models.SmsMessage, error)
	CreateSmsMessage(
		ctx context.Context,
		tx pgx.Tx,
		request *models.SmsMessage,
	) (*models.SmsMessage, error)
	UpdateSmsMessage(
		ctx context.Context,
		tx pgx.Tx,
		request *models.SmsMessage,
	) (*models.SmsMessage, error)
}

type smsMessageRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewSmsMessageRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) SmsMessageRepository {
	return &smsMessageRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *smsMessageRepository) FindSmsMessages(
	ctx context.Context,
	request *models.SmsMessageFilter,
) ([]*models.SmsMessage, int64, int64, error) {
	ctxt := "SmsMessageRepository-FindSmsMessages"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM sms_messages %s", condition)
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
		, user_id
		, mobile_phone
		, body
		, delivery_status
		, message_id
		, created_by
		, created_at
		, updated_at
		, sent_at`,
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
	var response []*models.SmsMessage
	for rows.Next() {
		var smsMessage models.SmsMessage
		if err = rows.Scan(
			&smsMessage.ID,
			&smsMessage.UserID,
			&smsMessage.MobilePhone,
			&smsMessage.Body,
			&smsMessage.DeliveryStatus,
			&smsMessage.MessageID,
			&smsMessage.CreatedBy,
			&smsMessage.CreatedAt,
			&smsMessage.UpdatedAt,
			&smsMessage.SentAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		smsMessage.RowNo = offset
		response = append(response, &smsMessage)
	}
	return response, total, pages, nil
}

func (q *smsMessageRepository) FindSmsMessage(
	ctx context.Context,
	request *models.SmsMessageFilter,
) (*models.SmsMessage, error) {
	ctxt := "SmsMessageRepository-FindSmsMessage"
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
			, mobile_phone
			, body
			, delivery_status
			, message_id
			, created_by
			, created_at
			, updated_at
			, sent_at
		FROM sms_messages %s`,
		condition,
	)
	query := sb.String()
	var response models.SmsMessage
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.UserID,
		&response.MobilePhone,
		&response.Body,
		&response.DeliveryStatus,
		&response.MessageID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.SentAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrSmsMessageNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *smsMessageRepository) CreateSmsMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.SmsMessage,
) (*models.SmsMessage, error) {
	ctxt := "SmsMessageRepository-CreateSmsMessage"
	var response models.SmsMessage
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO sms_messages (
			user_id
			, mobile_phone
			, body
			, created_by
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
			, user_id
			, mobile_phone
			, body
			, delivery_status
			, message_id
			, created_by
			, created_at
			, updated_at
			, sent_at`,
		request.UserID,
		request.MobilePhone,
		request.Body,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.UserID,
		&response.MobilePhone,
		&response.Body,
		&response.DeliveryStatus,
		&response.MessageID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.SentAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *smsMessageRepository) UpdateSmsMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.SmsMessage,
) (*models.SmsMessage, error) {
	ctxt := "SmsMessageRepository-UpdateSmsMessage"
	var response models.SmsMessage
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, user_id
			, mobile_phone
			, body
			, delivery_status
			, message_id
			, created_by
			, created_at
			, updated_at
			, sent_at
		FROM sms_messages
		WHERE id = $1
		FOR NO KEY UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.UserID,
		&response.MobilePhone,
		&response.Body,
		&response.DeliveryStatus,
		&response.MessageID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.SentAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrSmsMessageNotFound
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
	_, _ = sb.WriteString("UPDATE sms_messages SET")
	if (response.DeliveryStatus == nil && request.DeliveryStatus != nil) ||
		(response.DeliveryStatus != nil && request.DeliveryStatus == nil) ||
		(response.DeliveryStatus != nil && request.DeliveryStatus != nil &&
			*response.DeliveryStatus != *request.DeliveryStatus) {
		params = append(params, request.DeliveryStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " delivery_status = $%d", n)
	}
	if (response.MessageID == nil && request.MessageID != nil) ||
		(response.MessageID != nil && request.MessageID == nil) ||
		(response.MessageID != nil && request.MessageID != nil &&
			*response.MessageID != *request.MessageID) {
		params = append(params, request.MessageID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " message_id = $%d", n)
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_at = $%d", n)
	}
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
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, user_id
			, mobile_phone
			, body
			, delivery_status
			, message_id
			, created_by
			, created_at
			, updated_at
			, sent_at`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.UserID,
		&response.MobilePhone,
		&response.Body,
		&response.DeliveryStatus,
		&response.MessageID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.SentAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *smsMessageRepository) buildCondition(request *models.SmsMessageFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.SmsMessageIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, smsMessageID := range request.SmsMessageIDs {
			params = append(params, smsMessageID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.DeliveryStatusBlank {
		conds = append(conds, "delivery_status IS NULL")
	}
	if len(request.DeliveryStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("delivery_status IN (")
		for i, deliveryStatus := range request.DeliveryStatuses {
			params = append(params, deliveryStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.MessageIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("message_id IN (")
		for i, messageID := range request.MessageIDs {
			params = append(params, messageID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
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
