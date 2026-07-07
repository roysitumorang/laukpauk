package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	NonDeliveryDayRepository interface {
		FindNonDeliveryDays(
			ctx context.Context,
			request *models.NonDeliveryDayFilter,
		) ([]*models.NonDeliveryDay, int64, int64, error)
		FindNonDeliveryDay(
			ctx context.Context,
			request *models.NonDeliveryDayFilter,
		) (*models.NonDeliveryDay, error)
		CreateNonDeliveryDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.NonDeliveryDay,
		) (*models.NonDeliveryDay, error)
		UpdateNonDeliveryDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.NonDeliveryDay,
		) (*models.NonDeliveryDay, error)
	}

	nonDeliveryDayRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewNonDeliveryDayRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) NonDeliveryDayRepository {
	return new(nonDeliveryDayRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *nonDeliveryDayRepository) FindNonDeliveryDays(
	ctx context.Context,
	request *models.NonDeliveryDayFilter,
) ([]*models.NonDeliveryDay, int64, int64, error) {
	ctxt := "NonDeliveryDayRepository-FindNonDeliveryDays"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT COUNT(1)
		FROM non_delivery_days n
		JOIN delivery_plans d ON n.delivery_plan_id = d.id %s`,
		condition,
	)
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
		`n.id
		, n.record_status
		, n.date
		, n.delivery_plan_id
		, d.name
		, n.created_by
		, n.created_at
		, n.deleted_by
		, n.deleted_at
		, n.deletion_reason`,
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
		_, _ = fmt.Fprintf(&sb, " n.%s %s", item.Field, item.Direction)
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
	var response []*models.NonDeliveryDay
	for rows.Next() {
		var nonDeliveryDay models.NonDeliveryDay
		if err = rows.Scan(
			&nonDeliveryDay.ID,
			&nonDeliveryDay.RecordStatus,
			&nonDeliveryDay.Date,
			&nonDeliveryDay.DeliveryPlanID,
			&nonDeliveryDay.DeliveryPlanName,
			&nonDeliveryDay.CreatedBy,
			&nonDeliveryDay.CreatedAt,
			&nonDeliveryDay.DeletedBy,
			&nonDeliveryDay.DeletedAt,
			&nonDeliveryDay.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		nonDeliveryDay.RowNo = offset
		response = append(response, (&nonDeliveryDay).Populate())
	}
	return response, total, pages, nil
}

func (q *nonDeliveryDayRepository) FindNonDeliveryDay(
	ctx context.Context,
	request *models.NonDeliveryDayFilter,
) (*models.NonDeliveryDay, error) {
	ctxt := "NonDeliveryDayRepository-FindNonDeliveryDay"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			n.id
			, n.record_status
			, n.date
			, n.delivery_plan_id
			, d.name
			, n.created_by
			, n.created_at
			, n.deleted_by
			, n.deleted_at
			, n.deletion_reason
		FROM non_delivery_days n
		JOIN delivery_plans d ON n.delivery_plan_id = d.id %s`,
		condition,
	)
	query := sb.String()
	var response models.NonDeliveryDay
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Date,
		&response.DeliveryPlanID,
		&response.DeliveryPlanName,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrNonDeliveryDayNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *nonDeliveryDayRepository) CreateNonDeliveryDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NonDeliveryDay,
) (*models.NonDeliveryDay, error) {
	ctxt := "NonDeliveryDayRepository-CreateNonDeliveryDay"
	var response models.NonDeliveryDay
	err := tx.QueryRow(
		ctx,
		`WITH n AS (
			INSERT INTO non_delivery_days (
				record_status
				, date
				, delivery_plan_id
				, created_by
				, created_at
			) VALUES ($1, $2, $3, $4, $5)
			RETURNING *
		)
		SELECT
			n.id
			, n.record_status
			, n.date
			, n.delivery_plan_id
			, d.name
			, n.created_by
			, n.created_at
			, n.deleted_by
			, n.deleted_at
			, n.deletion_reason
		FROM n JOIN delivery_plans d ON n.delivery_plan_id = d.id`,
		request.RecordStatus,
		request.Date,
		request.DeliveryPlanID,
		request.CreatedBy,
		request.CreatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Date,
		&response.DeliveryPlanID,
		&response.DeliveryPlanName,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "non_delivery_days_date_delivery_plan_id_key" {
			err = models.ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *nonDeliveryDayRepository) UpdateNonDeliveryDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NonDeliveryDay,
) (*models.NonDeliveryDay, error) {
	ctxt := "NonDeliveryDayRepository-UpdateNonDeliveryDay"
	var response models.NonDeliveryDay
	err := tx.QueryRow(
		ctx,
		`WITH n AS (
			UPDATE non_delivery_days SET
				record_status = $1
				, date = $2
				, delivery_plan_id = $3
				, deleted_by = $4
				, deleted_at = $5
				, deletion_reason = $6
			WHERE id = $7
			RETURNING *
		)
		SELECT
			n.id
			, n.record_status
			, n.date
			, n.delivery_plan_id
			, d.name
			, n.created_by
			, n.created_at
			, n.deleted_by
			, n.deleted_at
			, n.deletion_reason
		FROM n JOIN delivery_plans d ON n.delivery_plan_id = d.id`,
		request.RecordStatus,
		request.Date,
		request.DeliveryPlanID,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Date,
		&response.DeliveryPlanID,
		&response.DeliveryPlanName,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "non_delivery_days_date_delivery_plan_id_key" {
			err = models.ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *nonDeliveryDayRepository) buildCondition(request *models.NonDeliveryDayFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.NonDeliveryDayIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("n.id IN (")
		for i, nonDeliveryDayID := range request.NonDeliveryDayIDs {
			params = append(params, nonDeliveryDayID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedNonDeliveryDayIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("n.id NOT IN (")
		for i, excludedNonDeliveryDayID := range request.ExcludedNonDeliveryDayIDs {
			params = append(params, excludedNonDeliveryDayID)
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
		_, _ = sb.WriteString("n.record_status IN (")
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
	if len(request.Dates) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("n.date IN (")
		for i, date := range request.Dates {
			params = append(params, date)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.DeliveryPlanIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("n.delivery_plan_id IN (")
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
