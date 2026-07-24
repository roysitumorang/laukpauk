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
		, delivery_time_interval
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
			&deliveryPlan.DeliveryTimeInterval,
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
		response = append(response, (&deliveryPlan).Populate())
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
			, delivery_time_interval
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
		&response.DeliveryTimeInterval,
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
	return (&response).Populate(), nil
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
			, delivery_time_interval
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
			, $7
			, tstzrange($8, $9%s)
			, $10
			, tstzrange($11, $12%s)
			, $13
			, tstzrange($14, $15%s)
			, $16
			, $17
			, $18
			, $19
		)
		RETURNING id
			, record_status
			, name
			, weekday
			, lower(transaction_window)
			, upper(transaction_window)
			, delivery_time_interval
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
		request.DeliveryTimeInterval,
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
		&response.DeliveryTimeInterval,
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
	return (&response).Populate(), nil
}

func (q *deliveryPlanRepository) UpdateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-UpdateDeliveryPlan"
	var response models.DeliveryPlan
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, name
			, weekday
			, lower(transaction_window)
			, upper(transaction_window)
			, delivery_time_interval
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
		FROM delivery_plans
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Weekday,
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.DeliveryTimeInterval,
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	_, _ = sb.WriteString("UPDATE delivery_plans SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " record_status = $%d", n)
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " name = $%d", n)
	}
	if response.Weekday != request.Weekday {
		params = append(params, request.Weekday)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " weekday = $%d", n)
	}
	if response.Weekday != request.Weekday {
		params = append(params, request.Weekday)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " weekday = $%d", n)
	}
	if !response.TransactionWindowMin.Equal(request.TransactionWindowMin) ||
		!response.TransactionWindowMax.Equal(request.TransactionWindowMax) {
		params = append(params, request.TransactionWindowMin, request.TransactionWindowMax)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " transaction_window = tstzrange($%d, $%d, '[]')", n-1, n)
	}
	if response.DeliveryTimeInterval != request.DeliveryTimeInterval {
		params = append(params, request.DeliveryTimeInterval)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " delivery_time_interval = $%d", n)
	}
	if response.SameDay != request.SameDay {
		params = append(params, request.SameDay)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " same_day = $%d", n)
	}
	if (response.SameDayWindowMin != nil && request.SameDayWindowMin == nil) ||
		(response.SameDayWindowMin == nil && request.SameDayWindowMin != nil) ||
		(response.SameDayWindowMin != nil && request.SameDayWindowMin != nil &&
			!response.SameDayWindowMin.Equal(*request.SameDayWindowMin)) ||
		(response.SameDayWindowMax != nil && request.SameDayWindowMax == nil) ||
		(response.SameDayWindowMax == nil && request.SameDayWindowMax != nil) ||
		(response.SameDayWindowMax != nil && request.SameDayWindowMax != nil &&
			!response.SameDayWindowMax.Equal(*request.SameDayWindowMax)) {
		params = append(params, request.SameDayWindowMin, request.SameDayWindowMax)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " same_day_window = tstzrange($%d, $%d", n-1, n)
		if request.SameDay {
			_, _ = sb.WriteString(", '[]'")
		}
		_, _ = sb.WriteString(")")
	}
	if response.NextDay != request.NextDay {
		params = append(params, request.NextDay)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " next_day = $%d", n)
	}
	if (response.NextDayWindowMin != nil && request.NextDayWindowMin == nil) ||
		(response.NextDayWindowMin == nil && request.NextDayWindowMin != nil) ||
		(response.NextDayWindowMin != nil && request.NextDayWindowMin != nil &&
			!response.NextDayWindowMin.Equal(*request.NextDayWindowMin)) ||
		(response.NextDayWindowMax != nil && request.NextDayWindowMax == nil) ||
		(response.NextDayWindowMax == nil && request.NextDayWindowMax != nil) ||
		(response.NextDayWindowMax != nil && request.NextDayWindowMax != nil &&
			!response.NextDayWindowMax.Equal(*request.NextDayWindowMax)) {
		params = append(params, request.NextDayWindowMin, request.NextDayWindowMax)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " next_day_window = tstzrange($%d, $%d", n-1, n)
		if request.NextDay {
			_, _ = sb.WriteString(", '[]'")
		}
		_, _ = sb.WriteString(")")
	}
	if response.NextTwoDay != request.NextTwoDay {
		params = append(params, request.NextTwoDay)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " next_two_day = $%d", n)
	}
	if (response.NextTwoDayWindowMin != nil && request.NextTwoDayWindowMin == nil) ||
		(response.NextTwoDayWindowMin == nil && request.NextTwoDayWindowMin != nil) ||
		(response.NextTwoDayWindowMin != nil && request.NextTwoDayWindowMin != nil &&
			!response.NextTwoDayWindowMin.Equal(*request.NextTwoDayWindowMin)) ||
		(response.NextTwoDayWindowMax != nil && request.NextTwoDayWindowMax == nil) ||
		(response.NextTwoDayWindowMax == nil && request.NextTwoDayWindowMax != nil) ||
		(response.NextTwoDayWindowMax != nil && request.NextTwoDayWindowMax != nil &&
			!response.NextTwoDayWindowMax.Equal(*request.NextTwoDayWindowMax)) {
		params = append(params, request.NextTwoDayWindowMin, request.NextTwoDayWindowMax)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " next_two_day_window = tstzrange($%d, $%d", n-1, n)
		if request.NextTwoDay {
			_, _ = sb.WriteString(", '[]'")
		}
		_, _ = sb.WriteString(")")
	}
	if len(params) == 0 {
		return &response, nil
	}
	if request.RecordStatus == models.RecordStatusDeleted {
		params = append(params, request.DeletedBy, request.DeletedAt, request.DeletionReason)
		n := len(params)
		_, _ = fmt.Fprintf(&sb, ", deleted_by = $%d, deleted_at = $%d, deletion_reason = $%d", n-2, n-1, n)
	} else {
		params = append(params, request.UpdatedBy, request.UpdatedAt)
		n := len(params)
		_, _ = fmt.Fprintf(&sb, ", updated_by = $%d, updated_at = $%d", n-1, n)
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		&sb,
		` WHERE id = $%d
		RETURNING id
			, record_status
			, name
			, weekday
			, lower(transaction_window)
			, upper(transaction_window)
			, delivery_time_interval
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
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Weekday,
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.DeliveryTimeInterval,
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
	return (&response).Populate(), nil
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
		params = append(params, request.TransactionDatetime)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "transaction_window @> $%d::timestamptz", len(params))
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
