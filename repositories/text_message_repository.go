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

type TextMessageRepository interface {
	FindTextMessages(
		ctx context.Context,
		request *models.TextMessageFilter,
	) ([]*models.TextMessage, int64, int64, error)
	FindTextMessage(
		ctx context.Context,
		request *models.TextMessageFilter,
	) (*models.TextMessage, error)
	CreateTextMessage(
		ctx context.Context,
		tx pgx.Tx,
		request *models.TextMessage,
	) (*models.TextMessage, error)
	UpdateTextMessage(
		ctx context.Context,
		tx pgx.Tx,
		request *models.TextMessage,
	) (*models.TextMessage, error)
}

type textMessageRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewTextMessageRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) TextMessageRepository {
	return &textMessageRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *textMessageRepository) FindTextMessages(
	ctx context.Context,
	request *models.TextMessageFilter,
) ([]*models.TextMessage, int64, int64, error) {
	ctxt := "TextMessageRepository-FindTextMessages"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM text_messages %s", condition)
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
		, mobile_phone
		, message_type
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
	var response []*models.TextMessage
	for rows.Next() {
		var textMessage models.TextMessage
		if err = rows.Scan(
			&textMessage.ID,
			&textMessage.AdminID,
			&textMessage.SellerID,
			&textMessage.BuyerID,
			&textMessage.MobilePhone,
			&textMessage.MessageType,
			&textMessage.Body,
			&textMessage.DeliveryStatus,
			&textMessage.MessageID,
			&textMessage.CreatedBy,
			&textMessage.CreatedAt,
			&textMessage.UpdatedAt,
			&textMessage.SentAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		textMessage.RowNo = offset
		response = append(response, &textMessage)
	}
	return response, total, pages, nil
}

func (q *textMessageRepository) FindTextMessage(
	ctx context.Context,
	request *models.TextMessageFilter,
) (*models.TextMessage, error) {
	ctxt := "TextMessageRepository-FindTextMessage"
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
			, mobile_phone
			, message_type
			, body
			, delivery_status
			, message_id
			, created_by
			, created_at
			, updated_at
			, sent_at
		FROM text_messages %s`,
		condition,
	)
	var response models.TextMessage
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.MobilePhone,
		&response.MessageType,
		&response.Body,
		&response.DeliveryStatus,
		&response.MessageID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.SentAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrTextMessageNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *textMessageRepository) CreateTextMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TextMessage,
) (*models.TextMessage, error) {
	ctxt := "TextMessageRepository-CreateTextMessage"
	var response models.TextMessage
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO text_messages (
			admin_id
			, seller_id
			, buyer_id
			, mobile_phone
			, message_type
			, body
			, created_by
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
			, admin_id
			, seller_id
			, buyer_id
			, mobile_phone
			, message_type
			, body
			, delivery_status
			, message_id
			, created_by
			, created_at
			, updated_at
			, sent_at`,
		request.AdminID,
		request.SellerID,
		request.BuyerID,
		request.MobilePhone,
		request.MessageType,
		request.Body,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.MobilePhone,
		&response.MessageType,
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

func (q *textMessageRepository) UpdateTextMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TextMessage,
) (*models.TextMessage, error) {
	ctxt := "TextMessageRepository-UpdateTextMessage"
	var response models.TextMessage
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, admin_id
			, seller_id
			, buyer_id
			, mobile_phone
			, message_type
			, body
			, delivery_status
			, message_id
			, created_by
			, created_at
			, updated_at
			, sent_at
		FROM text_messages
		WHERE id = $1
		FOR NO KEY UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.MobilePhone,
		&response.MessageType,
		&response.Body,
		&response.DeliveryStatus,
		&response.MessageID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.SentAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrTextMessageNotFound
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
	_, _ = sb.WriteString("UPDATE text_messages SET")
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
			, admin_id
			, seller_id
			, buyer_id
			, mobile_phone
			, message_type
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
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.MobilePhone,
		&response.MessageType,
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

func (q *textMessageRepository) buildCondition(request *models.TextMessageFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.TextMessageIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, textMessageID := range request.TextMessageIDs {
			params = append(params, textMessageID)
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
	if len(request.MessageTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("message_type IN (")
		for i, messageType := range request.MessageTypes {
			params = append(params, messageType)
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
				WHERE admins.id = text_messages.admin_id
					AND (
						LOWER(admins.name) LIKE $%d
						OR admins.username LIKE $%d
						OR admins.email LIKE $%d
						OR admins.mobile_phone LIKE $%d
					)
			) OR EXISTS(
				SELECT 1
				FROM sellers
				WHERE sellers.id = text_messages.seller_id
					AND (
						LOWER(sellers.name) LIKE $%d
						OR sellers.username LIKE $%d
						OR sellers.email LIKE $%d
						OR sellers.mobile_phone LIKE $%d
					)
			) OR EXISTS(
				SELECT 1
				FROM buyers
				WHERE buyers.id = text_messages.buyer_id
					AND (
						LOWER(buyers.name) LIKE $%d
						OR buyers.username LIKE $%d
						OR buyers.email LIKE $%d
						OR buyers.mobile_phone LIKE $%d
					)
			))`,
			n, n, n, n, n, n, n, n, n, n, n, n,
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
