package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

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
		, lower(transaction_window)
		, upper(transaction_window)
		, same_day
		, lower(same_day_window)
		, upper(same_day_window)
		, same_day_interval
		, next_day
		, lower(next_day_window)
		, upper(next_day_window)
		, next_day_interval
		, next_two_day
		, lower(next_two_day_window)
		, upper(next_two_day_window)
		, next_two_day_interval
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
	var (
		response        []*models.DeliveryPlan
		deliveryPlanIDs []any
	)
	mapDeliveryPlanIDs := map[uuid.UUID]int{}
	for rows.Next() {
		var deliveryPlan models.DeliveryPlan
		if err = rows.Scan(
			&deliveryPlan.ID,
			&deliveryPlan.RecordStatus,
			&deliveryPlan.Name,
			&deliveryPlan.TransactionWindowMin,
			&deliveryPlan.TransactionWindowMax,
			&deliveryPlan.SameDay,
			&deliveryPlan.SameDayWindowMin,
			&deliveryPlan.SameDayWindowMax,
			&deliveryPlan.SameDayInterval,
			&deliveryPlan.NextDay,
			&deliveryPlan.NextDayWindowMin,
			&deliveryPlan.NextDayWindowMax,
			&deliveryPlan.NextDayInterval,
			&deliveryPlan.NextTwoDay,
			&deliveryPlan.NextTwoDayWindowMin,
			&deliveryPlan.NextTwoDayWindowMax,
			&deliveryPlan.NextTwoDayInterval,
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
		mapDeliveryPlanIDs[deliveryPlan.ID] = len(response)
		response = append(response, new(deliveryPlan).Populate())
		deliveryPlanIDs = append(deliveryPlanIDs, deliveryPlan.ID)
	}
	if err = q.findDeliveryPlanHours(ctx, &sb, &response, mapDeliveryPlanIDs, deliveryPlanIDs...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrFindDeliveryPlanHours")
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
			, lower(transaction_window)
			, upper(transaction_window)
			, same_day
			, lower(same_day_window)
			, upper(same_day_window)
			, same_day_interval
			, next_day
			, lower(next_day_window)
			, upper(next_day_window)
			, next_day_interval
			, next_two_day
			, lower(next_two_day_window)
			, upper(next_two_day_window)
			, next_two_day_interval
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
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.SameDay,
		&response.SameDayWindowMin,
		&response.SameDayWindowMax,
		&response.SameDayInterval,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.NextDayInterval,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.NextTwoDayInterval,
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
	new(response).Populate()
	deliveryPlans := []*models.DeliveryPlan{new(response)}
	mapDeliveryPlanIDs := map[uuid.UUID]int{response.ID: 0}
	if err = q.findDeliveryPlanHours(ctx, &sb, &deliveryPlans, mapDeliveryPlanIDs, response.ID); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrFindDeliveryPlanHours")
		return nil, err
	}
	return deliveryPlans[0], nil
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
	err := tx.QueryRow(
		ctx,
		helper.Sprintf(
			`INSERT INTO delivery_plans (
				record_status
				, name
				, transaction_window
				, same_day
				, same_day_window
				, same_day_interval
				, next_day
				, next_day_window
				, next_day_interval
				, next_two_day
				, next_two_day_window
				, next_two_day_interval
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES (
				$1
				, $2
				, tstzrange($3, $4, '[]')
				, $5
				, tstzrange($6, $7%s)
				, $8
				, $9
				, tstzrange($10, $11%s)
				, $12
				, $13
				, tstzrange($14, $15%s)
				, $16
				, $17
				, $18
				, $19
				, $20
			)
			RETURNING id
				, record_status
				, name
				, lower(transaction_window)
				, upper(transaction_window)
				, same_day
				, lower(same_day_window)
				, upper(same_day_window)
				, same_day_interval
				, next_day
				, lower(next_day_window)
				, upper(next_day_window)
				, next_day_interval
				, next_two_day
				, lower(next_two_day_window)
				, upper(next_two_day_window)
				, next_two_day_interval
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
		),
		request.RecordStatus,
		request.Name,
		request.TransactionWindowMin,
		request.TransactionWindowMax,
		request.SameDay,
		request.SameDayWindowMin,
		request.SameDayWindowMax,
		request.SameDayInterval,
		request.NextDay,
		request.NextDayWindowMin,
		request.NextDayWindowMax,
		request.NextDayInterval,
		request.NextTwoDay,
		request.NextTwoDayWindowMin,
		request.NextTwoDayWindowMax,
		request.NextTwoDayInterval,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.SameDay,
		&response.SameDayWindowMin,
		&response.SameDayWindowMax,
		&response.SameDayInterval,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.NextDayInterval,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.NextTwoDayInterval,
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
	if err = q.saveDeliveryPlanHours(ctx, tx, request, new(response).Populate()); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrSaveDeliveryPlanHours")
		return nil, err
	}
	return new(response), nil
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
		helper.Sprintf(
			`UPDATE delivery_plans SET
				record_status = $1
				, name = $2
				, transaction_window = tstzrange($3, $4, '[]')
				, same_day = $5
				, same_day_window = tstzrange($6, $7%s)
				, same_day_interval = $8
				, next_day = $9
				, next_day_window = tstzrange($10, $11%s)
				, next_day_interval = $12
				, next_two_day = $13
				, next_two_day_window = tstzrange($14, $15%s)
				, next_two_day_interval = $16
				, updated_by = $17
				, updated_at = $18
				, deleted_by = $19
				, deleted_at = $20
				, deletion_reason = $21
			WHERE id = $22
			RETURNING id
				, record_status
				, name
				, lower(transaction_window)
				, upper(transaction_window)
				, same_day
				, lower(same_day_window)
				, upper(same_day_window)
				, same_day_interval
				, next_day
				, lower(next_day_window)
				, upper(next_day_window)
				, next_day_interval
				, next_two_day
				, lower(next_two_day_window)
				, upper(next_two_day_window)
				, next_two_day_interval
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
		),
		request.RecordStatus,
		request.Name,
		request.TransactionWindowMin,
		request.TransactionWindowMax,
		request.SameDay,
		request.SameDayWindowMin,
		request.SameDayWindowMax,
		request.SameDayInterval,
		request.NextDay,
		request.NextDayWindowMin,
		request.NextDayWindowMax,
		request.NextDayInterval,
		request.NextTwoDay,
		request.NextTwoDayWindowMin,
		request.NextTwoDayWindowMax,
		request.NextTwoDayInterval,
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
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.SameDay,
		&response.SameDayWindowMin,
		&response.SameDayWindowMax,
		&response.SameDayInterval,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.NextDayInterval,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.NextTwoDayInterval,
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
	if err = q.saveDeliveryPlanHours(ctx, tx, request, new(response).Populate()); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrSaveDeliveryPlanHours")
		return nil, err
	}
	return new(response), nil
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
	if request.TransactionWindowMin != nil {
		params = append(params, request.TransactionWindowMin)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "lower(transaction_window) = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.TransactionWindowMax != nil {
		params = append(params, request.TransactionWindowMax)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "upper(transaction_window) = $%d", len(params))
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

func (q *deliveryPlanRepository) saveDeliveryPlanHours(
	ctx context.Context,
	tx pgx.Tx,
	request,
	response *models.DeliveryPlan,
) error {
	ctxt := "DeliveryPlanRepository-saveDeliveryPlanHours"
	_, err := tx.Exec(
		ctx,
		"DELETE FROM delivery_plan_hours WHERE delivery_plan_id = $1",
		response.ID,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return err
	}
	var (
		sb     strings.Builder
		params []any
		n      int
	)
	_, _ = sb.WriteString(
		`INSERT INTO delivery_plan_hours (
			delivery_plan_id
			, delivery_speed
			, hour
			, created_by
			, created_at
		) VALUES`,
	)
	if request.SameDay {
		for _, deliveryHour := range request.SameDayHours {
			params = append(params, response.ID, models.DeliverySpeedSameDay, deliveryHour, response.UpdatedBy, response.UpdatedAt)
			if n = len(params); n > 5 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " ($%d, $%d, $%d, $%d, $%d)", n-4, n-3, n-2, n-1, n)
		}
	}
	if request.NextDay {
		for _, deliveryHour := range request.NextDayHours {
			params = append(params, response.ID, models.DeliverySpeedNextDay, deliveryHour, response.UpdatedBy, response.UpdatedAt)
			if n = len(params); n > 5 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " ($%d, $%d, $%d, $%d, $%d)", n-4, n-3, n-2, n-1, n)
		}
	}
	if request.NextTwoDay {
		for _, deliveryHour := range request.NextTwoDayHours {
			params = append(params, response.ID, models.DeliverySpeedNextTwoDay, deliveryHour, response.UpdatedBy, response.UpdatedAt)
			if n = len(params); n > 5 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " ($%d, $%d, $%d, $%d, $%d)", n-4, n-3, n-2, n-1, n)
		}
	}
	_, _ = sb.WriteString(
		`RETURNING delivery_speed
			, hour`,
	)
	if len(params) == 0 {
		return nil
	}
	rows, err := tx.Query(
		ctx,
		sb.String(),
		params...,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return err
	}
	defer rows.Close()
	var (
		deliverySpeed models.DeliverySpeed
		deliveryHour  time.Time
	)
	for rows.Next() {
		if err = rows.Scan(
			&deliverySpeed,
			&deliveryHour,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return err
		}
		switch deliverySpeed {
		case models.DeliverySpeedNextTwoDay:
			response.NextTwoDayHours = append(response.NextTwoDayHours, deliveryHour)
		case models.DeliverySpeedNextDay:
			response.NextDayHours = append(response.NextDayHours, deliveryHour)
		case models.DeliverySpeedSameDay:
			fallthrough
		default:
			response.SameDayHours = append(response.SameDayHours, deliveryHour)
		}
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return err
	}
	return nil
}

func (q *deliveryPlanRepository) findDeliveryPlanHours(
	ctx context.Context,
	sb *strings.Builder,
	response *[]*models.DeliveryPlan,
	mapDeliveryPlanIDs map[uuid.UUID]int,
	deliveryPlanIDs ...any,
) error {
	ctxt := "DeliveryPlanRepository-saveDeliveryPlanHours"
	if len(deliveryPlanIDs) == 0 {
		return nil
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT
			delivery_plan_id
			, delivery_speed
			, hour
		FROM delivery_plan_hours
		WHERE delivery_plan_id IN (`,
	)
	for i := range deliveryPlanIDs {
		if i > 0 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, "$%d", i+1)
	}
	_, _ = sb.WriteString(") ORDER BY id")
	rows, err := q.dbRead.Query(ctx, sb.String(), deliveryPlanIDs...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			deliveryPlanID uuid.UUID
			deliverySpeed  models.DeliverySpeed
			deliveryHour   time.Time
		)
		if err = rows.Scan(
			&deliveryPlanID,
			&deliverySpeed,
			&deliveryHour,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return err
		}
		i := mapDeliveryPlanIDs[deliveryPlanID]
		switch deliverySpeed {
		case models.DeliverySpeedNextTwoDay:
			(*response)[i].NextTwoDayHours = append((*response)[i].NextTwoDayHours, deliveryHour)
		case models.DeliverySpeedNextDay:
			(*response)[i].NextDayHours = append((*response)[i].NextDayHours, deliveryHour)
		case models.DeliverySpeedSameDay:
			fallthrough
		default:
			(*response)[i].SameDayHours = append((*response)[i].SameDayHours, deliveryHour)
		}
	}
	return nil
}
