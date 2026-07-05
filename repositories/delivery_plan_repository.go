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
)

type (
	DeliveryPlanRepository interface {
		FindDeliveryPlans(
			ctx context.Context,
			request *models.DeliveryPlanFilter,
		) ([]*models.DeliveryPlan, int64, int64, error)
		FindDeliveryPlan(
			ctx context.Context,
			request *models.DeliveryPlanFilter,
		) (*models.DeliveryPlan, error)
		CreateDeliveryPlan(
			ctx context.Context,
			tx pgx.Tx,
			request *models.DeliveryPlan,
		) (*models.DeliveryPlan, error)
		UpdateDeliveryPlan(
			ctx context.Context,
			tx pgx.Tx,
			request *models.DeliveryPlan,
		) (*models.DeliveryPlan, error)
	}

	deliveryPlanRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewDeliveryPlanRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) DeliveryPlanRepository {
	return new(deliveryPlanRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *deliveryPlanRepository) FindDeliveryPlans(
	ctx context.Context,
	request *models.DeliveryPlanFilter,
) ([]*models.DeliveryPlan, int64, int64, error) {
	ctxt := "DeliveryPlanRepository-FindDeliveryPlans"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM delivery_plans %s", condition)
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
		, record_status
		, name
		, transaction_time_min
		, transaction_time_max
		, transaction_time_description
		, delivery_time_min
		, delivery_time_max
		, delivery_time_description
		, delivery_daytime
		, same_day
		, next_day
		, next_two_day
		, created_by
		, created_at
		, updated_by
		, updated_at
		, deleted_by
		, deleted_at
		, deletion_reason`,
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
		_, _ = fmt.Fprintf(&sb, " %s %s", item.Field, item.Direction)
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
		_, _ = fmt.Fprintf(&sb, " LIMIT %d OFFSET %d", request.Limit, offset)
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
	var response []*models.DeliveryPlan
	for rows.Next() {
		var deliveryPlan models.DeliveryPlan
		if err = rows.Scan(
			&deliveryPlan.ID,
			&deliveryPlan.RecordStatus,
			&deliveryPlan.Name,
			&deliveryPlan.TransactionTimeMin,
			&deliveryPlan.TransactionTimeMax,
			&deliveryPlan.TransactionTimeDescription,
			&deliveryPlan.DeliveryTimeMin,
			&deliveryPlan.DeliveryTimeMax,
			&deliveryPlan.DeliveryTimeDescription,
			&deliveryPlan.DeliveryDayTime,
			&deliveryPlan.SameDay,
			&deliveryPlan.NextDay,
			&deliveryPlan.NextTwoDay,
			&deliveryPlan.CreatedBy,
			&deliveryPlan.CreatedAt,
			&deliveryPlan.UpdatedBy,
			&deliveryPlan.UpdatedAt,
			&deliveryPlan.DeletedBy,
			&deliveryPlan.DeletedAt,
			&deliveryPlan.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		deliveryPlan.RowNo = offset
		response = append(response, new(deliveryPlan).Populate())
	}
	return response, total, pages, nil
}

func (q *deliveryPlanRepository) FindDeliveryPlan(
	ctx context.Context,
	request *models.DeliveryPlanFilter,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-FindDeliveryPlan"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, record_status
			, name
			, transaction_time_min
			, transaction_time_max
			, transaction_time_description
			, delivery_time_min
			, delivery_time_max
			, delivery_time_description
			, delivery_daytime
			, same_day
			, next_day
			, next_two_day
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM delivery_plans %s`,
		condition,
	)
	query := sb.String()
	var response models.DeliveryPlan
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.TransactionTimeMin,
		&response.TransactionTimeMax,
		&response.TransactionTimeDescription,
		&response.DeliveryTimeMin,
		&response.DeliveryTimeMax,
		&response.DeliveryTimeDescription,
		&response.DeliveryDayTime,
		&response.SameDay,
		&response.NextDay,
		&response.NextTwoDay,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrDeliveryPlanNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response).Populate(), nil
}

func (q *deliveryPlanRepository) CreateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-CreateDeliveryPlan"
	var response models.DeliveryPlan
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO delivery_plans (
			record_status
			, name
			, transaction_time_min
			, transaction_time_max
			, transaction_time_description
			, delivery_time_min
			, delivery_time_max
			, delivery_time_description
			, delivery_daytime
			, same_day
			, next_day
			, next_two_day
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
			, record_status
			, name
			, transaction_time_min
			, transaction_time_max
			, transaction_time_description
			, delivery_time_min
			, delivery_time_max
			, delivery_time_description
			, delivery_daytime
			, same_day
			, next_day
			, next_two_day
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.TransactionTimeMin,
		request.TransactionTimeMax,
		request.TransactionTimeDescription,
		request.DeliveryTimeMin,
		request.DeliveryTimeMax,
		request.DeliveryTimeDescription,
		request.DeliveryDayTime,
		request.SameDay,
		request.NextDay,
		request.NextTwoDay,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.TransactionTimeMin,
		&response.TransactionTimeMax,
		&response.TransactionTimeDescription,
		&response.DeliveryTimeMin,
		&response.DeliveryTimeMax,
		&response.DeliveryTimeDescription,
		&response.DeliveryDayTime,
		&response.SameDay,
		&response.NextDay,
		&response.NextTwoDay,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response).Populate(), nil
}

func (q *deliveryPlanRepository) UpdateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-UpdateDeliveryPlan"
	var response models.DeliveryPlan
	if err := tx.QueryRow(
		ctx,
		`UPDATE delivery_plans SET
			record_status = $1
			, name = $2
			, transaction_time_min = $3
			, transaction_time_max = $4
			, transaction_time_description = $5
			, delivery_time_min = $6
			, delivery_time_max = $7
			, delivery_time_description = $8
			, delivery_daytime = $9
			, same_day = $10
			, next_day = $11
			, next_two_day = $12
			, updated_by = $13
			, updated_at = $14
			, deleted_by = $15
			, deleted_at = $16
			, deletion_reason = $17
		WHERE id = $18
		RETURNING id
			, record_status
			, name
			, transaction_time_min
			, transaction_time_max
			, transaction_time_description
			, delivery_time_min
			, delivery_time_max
			, delivery_time_description
			, delivery_daytime
			, same_day
			, next_day
			, next_two_day
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.TransactionTimeMin,
		request.TransactionTimeMax,
		request.TransactionTimeDescription,
		request.DeliveryTimeMin,
		request.DeliveryTimeMax,
		request.DeliveryTimeDescription,
		request.DeliveryDayTime,
		request.SameDay,
		request.NextDay,
		request.NextTwoDay,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.TransactionTimeMin,
		&response.TransactionTimeMax,
		&response.TransactionTimeDescription,
		&response.DeliveryTimeMin,
		&response.DeliveryTimeMax,
		&response.DeliveryTimeDescription,
		&response.DeliveryDayTime,
		&response.SameDay,
		&response.NextDay,
		&response.NextTwoDay,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response).Populate(), nil
}

func (q *deliveryPlanRepository) buildCondition(request *models.DeliveryPlanFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.DeliveryPlanIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, deliveryPlanID := range request.DeliveryPlanIDs {
			params = append(params, deliveryPlanID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedDeliveryPlanIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedDeliveryPlanID := range request.ExcludedDeliveryPlanIDs {
			params = append(params, excludedDeliveryPlanID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecordStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("record_status IN (")
		for i, recordStatus := range request.RecordStatuses {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Name != "" {
		params = append(params, strings.ToLower(request.Name))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(name) = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.TransactionTimeMin != nil {
		params = append(params, request.TransactionTimeMin)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "transaction_time_min = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.TransactionTimeMax != nil {
		params = append(params, request.TransactionTimeMax)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "transaction_time_max = $%d", len(params))
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
