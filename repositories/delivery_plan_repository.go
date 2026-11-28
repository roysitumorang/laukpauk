package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type DeliveryPlanRepository interface {
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
		tx *sql.Tx,
		request *models.DeliveryPlan,
	) (*models.DeliveryPlan, error)
	UpdateDeliveryPlan(
		ctx context.Context,
		tx *sql.Tx,
		request *models.DeliveryPlan,
	) (*models.DeliveryPlan, error)
}

type deliveryPlanRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewDeliveryPlanRepository(
	dbRead,
	dbWrite *sql.DB,
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
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM delivery_plans %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRowContext(ctx, query, params...).Scan(&total)
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
		, transaction_window_min
		, transaction_window_max
		, delivery_time_interval
		, same_day
		, same_day_window_min
		, same_day_window_max
		, next_day
		, next_day_window_min
		, next_day_window_max
		, next_two_day
		, next_two_day_window_min
		, next_two_day_window_max
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
	rows, err := q.dbRead.QueryContext(ctx, sb.String(), params...)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer func() { _ = rows.Close() }()
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
			, record_status
			, name
			, weekday
			, transaction_window_min
			, transaction_window_max
			, delivery_time_interval
			, same_day
			, same_day_window_min
			, same_day_window_max
			, next_day
			, next_day_window_min
			, next_day_window_max
			, next_two_day
			, next_two_day_window_min
			, next_two_day_window_max
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
	err := q.dbRead.QueryRowContext(ctx, sb.String(), params...).Scan(
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
	if errors.Is(err, sql.ErrNoRows) {
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
	tx *sql.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-CreateDeliveryPlan"
	var response models.DeliveryPlan
	if err := q.checkTransactionWindow(ctx, tx, nil, request); err != nil {
		return nil, err
	}
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`INSERT INTO delivery_plans (
			record_status
			, name
			, weekday
			, transaction_window_min
			, transaction_window_max
			, delivery_time_interval
			, same_day
			, same_day_window_min
			, same_day_window_max
			, next_day
			, next_day_window_min
			, next_day_window_max
			, next_two_day
			, next_two_day_window_min
			, next_two_day_window_max
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (
			?
			, ?
			, ?
			, ?, ?
			, ?
			, ?
			, ?, ?
			, ?
			, ?, ?
			, ?
			, ?, ?
			, ?
			, ?
			, ?
			, ?
		)
		RETURNING id
			, record_status
			, name
			, weekday
			, transaction_window_min
			, transaction_window_max
			, delivery_time_interval
			, same_day
			, same_day_window_min
			, same_day_window_max
			, next_day
			, next_day_window_min
			, next_day_window_max
			, next_two_day
			, next_two_day_window_min
			, next_two_day_window_max
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
	)
	err := tx.QueryRowContext(
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *deliveryPlanRepository) UpdateDeliveryPlan(
	ctx context.Context,
	tx *sql.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanRepository-UpdateDeliveryPlan"
	var response models.DeliveryPlan
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, record_status
			, name
			, weekday
			, transaction_window_min
			, transaction_window_max
			, delivery_time_interval
			, same_day
			, same_day_window_min
			, same_day_window_max
			, next_day
			, next_day_window_min
			, next_day_window_max
			, next_two_day
			, next_two_day_window_min
			, next_two_day_window_max
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM delivery_plans
		WHERE id = ?
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
	if response.Weekday != request.Weekday ||
		!response.TransactionWindowMin.Equal(request.TransactionWindowMin) ||
		!response.TransactionWindowMax.Equal(request.TransactionWindowMax) {
		if err = q.checkTransactionWindow(ctx, tx, &request.ID, request); err != nil {
			return nil, err
		}
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE delivery_plans SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" record_status = ?")
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" name = ?")
	}
	if response.Weekday != request.Weekday {
		params = append(params, request.Weekday)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" weekday = ?")
	}
	if !response.TransactionWindowMin.Equal(request.TransactionWindowMin) ||
		!response.TransactionWindowMax.Equal(request.TransactionWindowMax) {
		params = append(params, request.TransactionWindowMin, request.TransactionWindowMax)
		if n = len(params); n > 2 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" transaction_window_min = ?, transaction_window_max = ?")
	}
	if response.DeliveryTimeInterval != request.DeliveryTimeInterval {
		params = append(params, request.DeliveryTimeInterval)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" delivery_time_interval = ?")
	}
	if response.SameDay != request.SameDay {
		params = append(params, request.SameDay)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" same_day = ?")
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
		_, _ = sb.WriteString(" same_day_window_min = ?, same_day_window_max = ?")
	}
	if response.NextDay != request.NextDay {
		params = append(params, request.NextDay)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" next_day = ?")
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
		_, _ = sb.WriteString(" next_day_window_min = ?, next_day_window_max = ?")
	}
	if response.NextTwoDay != request.NextTwoDay {
		params = append(params, request.NextTwoDay)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" next_two_day = ?")
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
		_, _ = sb.WriteString(" next_two_day_window_min = ?, next_two_day_window_max = ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	if request.RecordStatus == models.RecordStatusDeleted {
		params = append(params, request.DeletedBy, request.DeletedAt, request.DeletionReason)
		_, _ = sb.WriteString(", deleted_by = ?, deleted_at = ?, deletion_reason = ?")
	} else {
		params = append(params, request.UpdatedBy, request.UpdatedAt)
		_, _ = sb.WriteString(", updated_by = ?, updated_at = ?")
	}
	params = append(params, request.ID)
	_, _ = sb.WriteString(` WHERE id = ?
		RETURNING id
			, record_status
			, name
			, weekday
			, transaction_window_min
			, transaction_window_max
			, delivery_time_interval
			, same_day
			, same_day_window_min
			, same_day_window_max
			, next_day
			, next_day_window_min
			, next_day_window_max
			, next_two_day
			, next_two_day_window_min
			, next_two_day_window_max
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`)
	if err = tx.QueryRowContext(ctx, sb.String(), params...).Scan(
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
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
			_, _ = sb.WriteString("?")
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
			_, _ = sb.WriteString("?")
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
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Name != "" {
		params = append(params, strings.ToLower(request.Name))
		sb.Reset()
		_, _ = sb.WriteString("LOWER(name) = ?")
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
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if !request.TransactionDatetime.IsZero() {
		params = append(params, request.TransactionDatetime, request.TransactionDatetime)
		sb.Reset()
		_, _ = sb.WriteString("transaction_window_min <= ? AND transaction_window_max >= ?")
		conds = append(conds, sb.String())
	}
	if request.TransactionWindow != nil {
		params = append(params, request.TransactionWindow.Max, request.TransactionWindow.Min)
		sb.Reset()
		_, _ = sb.WriteString("transaction_window_min <= ? AND transaction_window_max >= ?")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		sb.Reset()
		_, _ = sb.WriteString("LOWER(name) LIKE ?")
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

// checkTransactionWindow returns models.ErrUniqueDeliveryPlanTransactionWindowViolation when another delivery plan
// on the same weekday has a transaction window overlapping the plan's (both ends included); excludeID is the plan
// being updated. MariaDB has no exclusion constraints, so this check takes their place: FOR UPDATE locks the plans
// of that weekday (and, through the weekday index, the gap around them) until the transaction ends, so two
// transactions can't add overlapping plans at the same time.
func (q *deliveryPlanRepository) checkTransactionWindow(
	ctx context.Context,
	tx *sql.Tx,
	excludeID *uuid.UUID,
	request *models.DeliveryPlan,
) error {
	ctxt := "DeliveryPlanRepository-checkTransactionWindow"
	params := []any{request.Weekday, request.TransactionWindowMax, request.TransactionWindowMin}
	query := `SELECT id
		FROM delivery_plans
		WHERE weekday = ?
			AND transaction_window_min <= ?
			AND transaction_window_max >= ?`
	if excludeID != nil {
		params = append(params, *excludeID)
		query += " AND id != ?"
	}
	rows, err := tx.QueryContext(ctx, query+" FOR UPDATE", params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return err
	}
	defer func() { _ = rows.Close() }()
	overlapping := rows.Next()
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return err
	}
	if overlapping {
		return models.ErrUniqueDeliveryPlanTransactionWindowViolation
	}
	return nil
}
