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
	params = []any{models.RecordStatusActive}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT
			id
			, record_status
			, delivery_plan_id
			, delivery_speed
			, "timestamp"
			, created_by
			, created_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM delivery_datetimes
		WHERE record_status = $1
			AND delivery_plan_id IN (`,
	)
	var (
		response []*models.DeliveryPlan
		n        int
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
			&deliveryPlan.RawSameDayInterval,
			&deliveryPlan.NextDay,
			&deliveryPlan.NextDayWindowMin,
			&deliveryPlan.NextDayWindowMax,
			&deliveryPlan.RawNextDayInterval,
			&deliveryPlan.NextTwoDay,
			&deliveryPlan.NextTwoDayWindowMin,
			&deliveryPlan.NextTwoDayWindowMax,
			&deliveryPlan.RawNextTwoDayInterval,
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
		response = append(response, &deliveryPlan)
		mapDeliveryPlanIDs[deliveryPlan.ID] = len(response) - 1
		params = append(params, deliveryPlan.ID)
		if n = len(params); n > 2 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, "$%d", n)
	}
	_, _ = sb.WriteString(`) ORDER BY delivery_speed, "timestamp"`)
	if len(params) == 1 {
		return response, total, pages, nil
	}
	rows, err = q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var deliveryDatetime models.DeliveryDatetime
		if err = rows.Scan(
			&deliveryDatetime.ID,
			&deliveryDatetime.RecordStatus,
			&deliveryDatetime.DeliveryPlanID,
			&deliveryDatetime.DeliverySpeed,
			&deliveryDatetime.Timestamp,
			&deliveryDatetime.CreatedBy,
			&deliveryDatetime.CreatedAt,
			&deliveryDatetime.DeletedBy,
			&deliveryDatetime.DeletedAt,
			&deliveryDatetime.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		i := mapDeliveryPlanIDs[deliveryDatetime.DeliveryPlanID]
		deliveryPlan := response[i]
		switch deliveryDatetime.DeliverySpeed {
		case models.DeliverySpeedSameDay:
			deliveryPlan.SameDayDeliveryDatetimes = append(deliveryPlan.SameDayDeliveryDatetimes, &deliveryDatetime)
		case models.DeliverySpeedNextDay:
			deliveryPlan.NextDayDeliveryDatetimes = append(deliveryPlan.NextDayDeliveryDatetimes, &deliveryDatetime)
		case models.DeliverySpeedNextTwoDay:
			deliveryPlan.NextTwoDayDeliveryDatetimes = append(deliveryPlan.NextTwoDayDeliveryDatetimes, &deliveryDatetime)
		}
		response[i] = deliveryPlan
	}
	if !request.TransactionDatetime.IsZero() {
		for i := range response {
			response[i].UpdateDate(request.TransactionDatetime)
		}
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
	var response models.DeliveryPlan
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.TransactionWindowMin,
		&response.TransactionWindowMax,
		&response.SameDay,
		&response.SameDayWindowMin,
		&response.SameDayWindowMax,
		&response.RawSameDayInterval,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.RawNextDayInterval,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.RawNextTwoDayInterval,
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
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			id
			, record_status
			, delivery_plan_id
			, delivery_speed
			, "timestamp"
			, created_by
			, created_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM delivery_datetimes
		WHERE record_status = $1
			AND delivery_plan_id = $2
		ORDER BY delivery_speed, "timestamp"`,
		models.RecordStatusActive,
		response.ID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var deliveryDatetime models.DeliveryDatetime
		if err = rows.Scan(
			&deliveryDatetime.ID,
			&deliveryDatetime.RecordStatus,
			&deliveryDatetime.DeliveryPlanID,
			&deliveryDatetime.DeliverySpeed,
			&deliveryDatetime.Timestamp,
			&deliveryDatetime.CreatedBy,
			&deliveryDatetime.CreatedAt,
			&deliveryDatetime.DeletedBy,
			&deliveryDatetime.DeletedAt,
			&deliveryDatetime.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		switch deliveryDatetime.DeliverySpeed {
		case models.DeliverySpeedSameDay:
			response.SameDayDeliveryDatetimes = append(response.SameDayDeliveryDatetimes, &deliveryDatetime)
		case models.DeliverySpeedNextDay:
			response.NextDayDeliveryDatetimes = append(response.NextDayDeliveryDatetimes, &deliveryDatetime)
		case models.DeliverySpeedNextTwoDay:
			response.NextTwoDayDeliveryDatetimes = append(response.NextTwoDayDeliveryDatetimes, &deliveryDatetime)
		}
	}
	if !request.TransactionDatetime.IsZero() {
		(&response).UpdateDate(request.TransactionDatetime)
	}
	return &response, nil
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
	)
	err := tx.QueryRow(
		ctx,
		sb.String(),
		request.RecordStatus,
		request.Name,
		request.TransactionWindowMin,
		request.TransactionWindowMax,
		request.SameDay,
		request.SameDayWindowMin,
		request.SameDayWindowMax,
		request.RawSameDayInterval,
		request.NextDay,
		request.NextDayWindowMin,
		request.NextDayWindowMax,
		request.RawNextDayInterval,
		request.NextTwoDay,
		request.NextTwoDayWindowMin,
		request.NextTwoDayWindowMax,
		request.RawNextTwoDayInterval,
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
		&response.RawSameDayInterval,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.RawNextDayInterval,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.RawNextTwoDayInterval,
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
			(pgxErr.Code == pgerrcode.UniqueViolation ||
				pgxErr.Code == pgerrcode.ExclusionViolation) {
			switch pgxErr.ConstraintName {
			case "delivery_plans_unique_name_idx":
				err = models.ErrUniqueDeliveryPlanNameViolation
			case "no_overlapping_transaction_windows":
				err = models.ErrUniqueDeliveryPlanTransactionWindowViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	deliveryDatetimes := make(
		[]*models.DeliveryDatetime,
		len(request.SameDayDeliveryDatetimes)+
			len(request.NextDayDeliveryDatetimes)+
			len(request.NextTwoDayDeliveryDatetimes),
	)
	var j int
	for i := range request.SameDayDeliveryDatetimes {
		deliveryDatetimes[j] = request.SameDayDeliveryDatetimes[i]
		j++
	}
	for i := range request.NextDayDeliveryDatetimes {
		deliveryDatetimes[j] = request.NextDayDeliveryDatetimes[i]
		j++
	}
	for i := range request.NextTwoDayDeliveryDatetimes {
		deliveryDatetimes[j] = request.NextTwoDayDeliveryDatetimes[i]
		j++
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`INSERT INTO delivery_datetimes (
			record_status
			, delivery_plan_id
			, delivery_speed
			, "timestamp"
			, created_by
			, created_at
		) VALUES`,
	)
	var (
		params []any
		n      int
	)
	for _, deliveryDatetime := range deliveryDatetimes {
		params = append(
			params,
			models.RecordStatusActive,
			response.ID,
			deliveryDatetime.DeliverySpeed,
			deliveryDatetime.Timestamp,
			response.CreatedBy,
			response.CreatedAt,
		)
		if n = len(params); n > 6 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " ($%d, $%d, $%d, $%d, $%d, $%d)", n-5, n-4, n-3, n-2, n-1, n)
	}
	_, _ = sb.WriteString(
		` RETURNING id
			, record_status
			, delivery_plan_id
			, delivery_speed
			, "timestamp"
			, created_by
			, created_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
	)
	if len(params) == 0 {
		return &response, nil
	}
	rows, err := tx.Query(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	for rows.Next() {
		var deliveryDatetime models.DeliveryDatetime
		if rows.Scan(
			&deliveryDatetime.ID,
			&deliveryDatetime.RecordStatus,
			&deliveryDatetime.DeliveryPlanID,
			&deliveryDatetime.DeliverySpeed,
			&deliveryDatetime.Timestamp,
			&deliveryDatetime.CreatedBy,
			&deliveryDatetime.CreatedAt,
			&deliveryDatetime.DeletedBy,
			&deliveryDatetime.DeletedAt,
			&deliveryDatetime.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		switch deliveryDatetime.DeliverySpeed {
		case models.DeliverySpeedSameDay:
			response.SameDayDeliveryDatetimes = append(response.SameDayDeliveryDatetimes, &deliveryDatetime)
		case models.DeliverySpeedNextDay:
			response.NextDayDeliveryDatetimes = append(response.NextDayDeliveryDatetimes, &deliveryDatetime)
		case models.DeliverySpeedNextTwoDay:
			response.NextTwoDayDeliveryDatetimes = append(response.NextTwoDayDeliveryDatetimes, &deliveryDatetime)
		}
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
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
	)
	if err = tx.QueryRow(
		ctx,
		sb.String(),
		request.RecordStatus,
		request.Name,
		request.TransactionWindowMin,
		request.TransactionWindowMax,
		request.SameDay,
		request.SameDayWindowMin,
		request.SameDayWindowMax,
		request.RawSameDayInterval,
		request.NextDay,
		request.NextDayWindowMin,
		request.NextDayWindowMax,
		request.RawNextDayInterval,
		request.NextTwoDay,
		request.NextTwoDayWindowMin,
		request.NextTwoDayWindowMax,
		request.RawNextTwoDayInterval,
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
		&response.RawSameDayInterval,
		&response.NextDay,
		&response.NextDayWindowMin,
		&response.NextDayWindowMax,
		&response.RawNextDayInterval,
		&response.NextTwoDay,
		&response.NextTwoDayWindowMin,
		&response.NextTwoDayWindowMax,
		&response.RawNextTwoDayInterval,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			(pgxErr.Code == pgerrcode.UniqueViolation ||
				pgxErr.Code == pgerrcode.ExclusionViolation) {
			switch pgxErr.ConstraintName {
			case "delivery_plans_unique_name_idx":
				err = models.ErrUniqueDeliveryPlanNameViolation
			case "no_overlapping_transaction_windows":
				err = models.ErrUniqueDeliveryPlanTransactionWindowViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	sameDaySize := len(request.SameDayDeliveryDatetimes)
	nextDaySize := len(request.NextDayDeliveryDatetimes)
	nextTwoDaySize := len(request.NextTwoDayDeliveryDatetimes)
	response.SameDayDeliveryDatetimes = make([]*models.DeliveryDatetime, sameDaySize)
	response.NextDayDeliveryDatetimes = make([]*models.DeliveryDatetime, nextDaySize)
	response.NextTwoDayDeliveryDatetimes = make([]*models.DeliveryDatetime, nextTwoDaySize)
	deliveryDatetimes := make(
		[]*models.DeliveryDatetime,
		sameDaySize+nextDaySize+nextTwoDaySize,
	)
	var j int
	for i := range request.SameDayDeliveryDatetimes {
		deliveryDatetimes[j] = request.SameDayDeliveryDatetimes[i]
		j++
	}
	for i := range request.NextDayDeliveryDatetimes {
		deliveryDatetimes[j] = request.NextDayDeliveryDatetimes[i]
		j++
	}
	for i := range request.NextTwoDayDeliveryDatetimes {
		deliveryDatetimes[j] = request.NextTwoDayDeliveryDatetimes[i]
		j++
	}
	params := make([]any, 5+len(deliveryDatetimes))
	sb.Reset()
	_, _ = sb.WriteString(
		`UPDATE delivery_datetimes SET
			record_status = $1
			, deleted_by = $2
			, deleted_at = $3
		WHERE delivery_plan_id = $4
			AND record_status = $5
			AND id NOT IN (`,
	)
	params[0] = models.RecordStatusDeleted
	params[1] = response.UpdatedBy
	params[2] = response.UpdatedAt
	params[3] = response.ID
	params[4] = models.RecordStatusActive
	var sameDayCounter,
		nextDayCounter,
		nextTwoDayCounter int
	for i, deliveryDatetimeRequest := range deliveryDatetimes {
		var deliveryDatetimeResponse models.DeliveryDatetime
		if err = tx.QueryRow(
			ctx,
			`INSERT INTO delivery_datetimes (
				record_status
				, delivery_plan_id
				, delivery_speed
				, "timestamp"
				, created_by
				, created_at
			) VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (
				delivery_plan_id
				, delivery_speed
				, "timestamp"
			) DO UPDATE SET
				record_status = $1
				, deleted_by = NULL
				, deleted_at = NULL
				, deletion_reason = NULL
			RETURNING id
				, record_status
				, delivery_plan_id
				, delivery_speed
				, "timestamp"
				, created_by
				, created_at
				, deleted_by
				, deleted_at
				, deletion_reason`,
			models.RecordStatusActive,
			response.ID,
			deliveryDatetimeRequest.DeliverySpeed,
			deliveryDatetimeRequest.Timestamp,
			response.UpdatedBy,
			response.UpdatedAt,
		).Scan(
			&deliveryDatetimeResponse.ID,
			&deliveryDatetimeResponse.RecordStatus,
			&deliveryDatetimeResponse.DeliveryPlanID,
			&deliveryDatetimeResponse.DeliverySpeed,
			&deliveryDatetimeResponse.Timestamp,
			&deliveryDatetimeResponse.CreatedBy,
			&deliveryDatetimeResponse.CreatedAt,
			&deliveryDatetimeResponse.DeletedBy,
			&deliveryDatetimeResponse.DeletedAt,
			&deliveryDatetimeResponse.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		switch deliveryDatetimeResponse.DeliverySpeed {
		case models.DeliverySpeedSameDay:
			response.SameDayDeliveryDatetimes[sameDayCounter] = &deliveryDatetimeResponse
			sameDayCounter++
		case models.DeliverySpeedNextDay:
			response.NextDayDeliveryDatetimes[nextDayCounter] = &deliveryDatetimeResponse
			nextDayCounter++
		case models.DeliverySpeedNextTwoDay:
			response.NextTwoDayDeliveryDatetimes[nextTwoDayCounter] = &deliveryDatetimeResponse
			nextTwoDayCounter++
		}
		j := i + 5
		params[j] = deliveryDatetimeResponse.ID
		if i > 0 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, "$%d", j+1)
	}
	_, _ = sb.WriteString(")")
	if len(params) == 5 {
		return &response, nil
	}
	if _, err = tx.Exec(ctx, sb.String(), params...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
		params = append(params, t)
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
