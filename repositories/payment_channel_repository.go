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

type PaymentChannelRepository interface {
	FindPaymentChannels(
		ctx context.Context,
		request *models.PaymentChannelFilter,
	) ([]*models.PaymentChannel, int64, int64, error)
	FindPaymentChannel(
		ctx context.Context,
		request *models.PaymentChannelFilter,
	) (*models.PaymentChannel, error)
}

type paymentChannelRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewPaymentChannelRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) PaymentChannelRepository {
	return &paymentChannelRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *paymentChannelRepository) FindPaymentChannels(
	ctx context.Context,
	request *models.PaymentChannelFilter,
) ([]*models.PaymentChannel, int64, int64, error) {
	ctxt := "PaymentChannelRepository-FindPaymentChannels"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM payment_channels %s", condition)
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
		, business_id
		, name
		, channel_code
		, country
		, currency
		, minimum_amount
		, maximum_amount
		, payment_request_expiry
		, refund
		, channel_type
		, channel_category
		, created_at`,
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
		field := item.Field
		if item.Field == models.PaymentChannelSortingFieldName {
			field = "LOWER(name)"
		}
		_, _ = fmt.Fprintf(sb, " %s %s", field, item.Direction)
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
	var response []*models.PaymentChannel
	for rows.Next() {
		var paymentChannel models.PaymentChannel
		if err = rows.Scan(
			&paymentChannel.ID,
			&paymentChannel.BusinessID,
			&paymentChannel.Name,
			&paymentChannel.ChannelCode,
			&paymentChannel.Country,
			&paymentChannel.Currency,
			&paymentChannel.MinimumAmount,
			&paymentChannel.MaximumAmount,
			&paymentChannel.PaymentRequestExpiry,
			&paymentChannel.Refund,
			&paymentChannel.ChannelType,
			&paymentChannel.ChannelCategory,
			&paymentChannel.CreatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		paymentChannel.RowNo = offset
		response = append(response, &paymentChannel)
	}
	return response, total, pages, nil
}

func (q *paymentChannelRepository) FindPaymentChannel(
	ctx context.Context,
	request *models.PaymentChannelFilter,
) (*models.PaymentChannel, error) {
	ctxt := "PaymentChannelRepository-FindPaymentChannel"
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
			, business_id
			, name
			, channel_code
			, country
			, currency
			, minimum_amount
			, maximum_amount
			, payment_request_expiry
			, refund
			, channel_type
			, channel_category
			, created_at
		FROM payment_channels %s`,
		condition,
	)
	query := sb.String()
	var response models.PaymentChannel
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.BusinessID,
		&response.Name,
		&response.ChannelCode,
		&response.Country,
		&response.Currency,
		&response.MinimumAmount,
		&response.MaximumAmount,
		&response.PaymentRequestExpiry,
		&response.Refund,
		&response.ChannelType,
		&response.ChannelCategory,
		&response.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPaymentChannelNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *paymentChannelRepository) buildCondition(request *models.PaymentChannelFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.PaymentChannelIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, paymentChannelID := range request.PaymentChannelIDs {
			params = append(params, paymentChannelID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Amount > 0 {
		params = append(params, request.Amount)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "($%d BETWEEN minimum_amount AND maximum_amount)", len(params))
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(name) LIKE $%d", len(params))
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
