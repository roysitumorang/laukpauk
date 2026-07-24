package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	return &deliveryPlanRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
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
		, weekday
		, lower(transaction_window)
		, upper(transaction_window)
		, same_day
		, lower(same_day_window)
		, upper(same_day_window)
		, next_day
		, lower(next_day_window)
		, upper(next_day_window)
		, next_two_day
		, lower(next_two_day_window)
		, upper(next_two_day_window)
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
			&deliveryPlan.Weekday,
			&deliveryPlan.TransactionWindowMin,
			&deliveryPlan.TransactionWindowMax,
			&deliveryPlan.SameDay,
			&deliveryPlan.SameDayWindowMin,
			&deliveryPlan.SameDayWindowMax,
			&deliveryPlan.NextDay,
			&deliveryPlan.NextDayWindowMin,
			&deliveryPlan.NextDayWindowMax,
			&deliveryPlan.NextTwoDay,
			&deliveryPlan.NextTwoDayWindowMin,
			&deliveryPlan.NextTwoDayWindowMax,
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
		response = append(response, (&deliveryPlan).UpdateDate(request.DeliveryTimeInterval))
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, 0, 0, err
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
			, weekday
			, lower(transaction_window)
			, upper(transaction_window)
			, same_day
			, lower(same_day_window)
			, upper(same_day_window)
			, next_day
			, lower(next_day_window)
			, upper(next_day_window)
			, next_two_day
			, lower(next_two_day_window)
			, upper(next_two_day_window)
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
	var response models.DeliveryPlan
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Weekday,
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.SameDay,
		&response.SameDayWindowMin,
		&response.SameDayWindowMax,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
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
	return (&response).UpdateDate(request.DeliveryTimeInterval), nil
}

func (q *deliveryPlanRepository) CreateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-CreateDeliveryPlan"
	var response models.DeliveryPlan
	var sameDayInterval,
		nextDayInterval,
		nextTwoDayInterval string
	if request.SameDay {
		sameDayInterval = ", '[]'"
	}
	if request.NextDay {
		nextDayInterval = ", '[]'"
	}
	if request.NextTwoDay {
		nextTwoDayInterval = ", '[]'"
	}
	var sb strings.Builder
	_, _ = fmt.Fprintf(
		&sb,
		`INSERT INTO delivery_plans (
			record_status
			, name
			, weekday
			, transaction_window
			, same_day
			, same_day_window
			, next_day
			, next_day_window
			, next_two_day
			, next_two_day_window
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (
			$1
			, $2
			, $3
			, tstzrange($4, $5, '[]')
			, $6
			, tstzrange($7, $8%s)
			, $9
			, tstzrange($10, $11%s)
			, $12
			, tstzrange($13, $14%s)
			, $15
			, $16
			, $17
			, $18
		)
		RETURNING id
			, record_status
			, name
			, weekday
			, lower(transaction_window)
			, upper(transaction_window)
			, same_day
			, lower(same_day_window)
			, upper(same_day_window)
			, next_day
			, lower(next_day_window)
			, upper(next_day_window)
			, next_two_day
			, lower(next_two_day_window)
			, upper(next_two_day_window)
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		sameDayInterval,
		nextDayInterval,
		nextTwoDayInterval,
	)
	err := tx.QueryRow(
		ctx,
		sb.String(),
		request.RecordStatus,
		request.Name,
		request.Weekday,
		request.TransactionWindowMin,
		request.TransactionWindowMax,
		request.SameDay,
		request.SameDayWindowMin,
		request.SameDayWindowMax,
		request.NextDay,
		request.NextDayWindowMin,
		request.NextDayWindowMax,
		request.NextTwoDay,
		request.NextTwoDayWindowMin,
		request.NextTwoDayWindowMax,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Weekday,
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.SameDay,
		&response.SameDayWindowMin,
		&response.SameDayWindowMax,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.ExclusionViolation &&
			pgxErr.ConstraintName == "no_overlapping_delivery_plans_transaction_window" {
			err = models.ErrUniqueDeliveryPlanTransactionWindowViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *deliveryPlanRepository) UpdateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-UpdateDeliveryPlan"
	var response models.DeliveryPlan
	var sameDayInterval,
		nextDayInterval,
		nextTwoDayInterval string
	if request.SameDay {
		sameDayInterval = ", '[]'"
	}
	if request.NextDay {
		nextDayInterval = ", '[]'"
	}
	if request.NextTwoDay {
		nextTwoDayInterval = ", '[]'"
	}
	err := tx.QueryRow(
		ctx,
		"SELECT id FROM delivery_plans WHERE id = $1 FOR UPDATE",
		request.ID,
	).Scan(&response.ID)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var sb strings.Builder
	_, _ = fmt.Fprintf(
		&sb,
		`UPDATE delivery_plans SET
			record_status = $1
			, name = $2
			, weekday = $3
			, transaction_window = tstzrange($4, $5, '[]')
			, same_day = $6
			, same_day_window = tstzrange($7, $8%s)
			, next_day = $9
			, next_day_window = tstzrange($10, $11%s)
			, next_two_day = $12
			, next_two_day_window = tstzrange($13, $14%s)
			, updated_by = $15
			, updated_at = $16
			, deleted_by = $17
			, deleted_at = $18
			, deletion_reason = $19
		WHERE id = $20
		RETURNING id
			, record_status
			, name
			, weekday
			, lower(transaction_window)
			, upper(transaction_window)
			, same_day
			, lower(same_day_window)
			, upper(same_day_window)
			, next_day
			, lower(next_day_window)
			, upper(next_day_window)
			, next_two_day
			, lower(next_two_day_window)
			, upper(next_two_day_window)
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		sameDayInterval,
		nextDayInterval,
		nextTwoDayInterval,
	)
	if err = tx.QueryRow(
		ctx,
		sb.String(),
		request.RecordStatus,
		request.Name,
		request.Weekday,
		request.TransactionWindowMin,
		request.TransactionWindowMax,
		request.SameDay,
		request.SameDayWindowMin,
		request.SameDayWindowMax,
		request.NextDay,
		request.NextDayWindowMin,
		request.NextDayWindowMax,
		request.NextTwoDay,
		request.NextTwoDayWindowMin,
		request.NextTwoDayWindowMax,
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
		&response.Weekday,
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.SameDay,
		&response.SameDayWindowMin,
		&response.SameDayWindowMax,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.ExclusionViolation &&
			pgxErr.ConstraintName == "no_overlapping_delivery_plans_transaction_window" {
			err = models.ErrUniqueDeliveryPlanTransactionWindowViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
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
	if len(request.DeliveryWeekdays) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("weekday IN (")
		for i, deliveryWeekday := range request.DeliveryWeekdays {
			params = append(params, deliveryWeekday)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if !request.TransactionDatetime.IsZero() {
		t := time.Date(
			2000,
			1,
			1,
			request.TransactionDatetime.Hour(),
			request.TransactionDatetime.Minute(),
			request.TransactionDatetime.Second(),
			request.TransactionDatetime.Nanosecond(),
			request.TransactionDatetime.Location(),
		)
		u := time.Date(
			request.TransactionDatetime.Year(),
			request.TransactionDatetime.Month(),
			request.TransactionDatetime.Day(),
			0,
			0,
			0,
			0,
			request.TransactionDatetime.Location(),
		)
		params = append(params, t, u, models.RecordStatusActive)
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(
			sb,
			`(transaction_window @> $%d::timestamptz AND NOT EXISTS(
				SELECT 1
				FROM skipped_delivery_plans
				WHERE skipped_delivery_plans.delivery_plan_id = delivery_plans.id
					AND skipped_delivery_plans.date = $%d
					AND skipped_delivery_plans.record_status = $%d
			))`,
			n-2, n-1, n,
		)
		conds = append(conds, sb.String())
	}
	if request.TransactionWindow != nil {
		params = append(params, request.TransactionWindow.Min, request.TransactionWindow.Max)
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "transaction_window && tstzrange($%d, $%d, '[]')", n-1, n)
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
