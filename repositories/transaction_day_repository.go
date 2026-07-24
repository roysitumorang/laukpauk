package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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
	TransactionDayRepository interface {
		FindTransactionDays(
			ctx context.Context,
			request *models.TransactionDayFilter,
		) ([]*models.TransactionDay, int64, int64, error)
		FindTransactionDay(
			ctx context.Context,
			request *models.TransactionDayFilter,
		) (*models.TransactionDay, error)
		CreateTransactionDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.TransactionDay,
		) (*models.TransactionDay, error)
		UpdateTransactionDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.TransactionDay,
		) (*models.TransactionDay, error)
	}

	transactionDayRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewTransactionDayRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) TransactionDayRepository {
	return &transactionDayRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *transactionDayRepository) FindTransactionDays(
	ctx context.Context,
	request *models.TransactionDayFilter,
) ([]*models.TransactionDay, int64, int64, error) {
	ctxt := "TransactionDayRepository-FindTransactionDays"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM transaction_days %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(&total)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, 0, 0, err
	}
	if total == 0 {
		return nil, 0, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`id
		, date
		, public_holiday
		, created_by
		, created_at
		, updated_by
		, updated_at`,
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
	var response []*models.TransactionDay
	for rows.Next() {
		var transactionDay models.TransactionDay
		if err = rows.Scan(
			&transactionDay.ID,
			&transactionDay.Date,
			&transactionDay.PublicHoliday,
			&transactionDay.CreatedBy,
			&transactionDay.CreatedAt,
			&transactionDay.UpdatedBy,
			&transactionDay.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		transactionDay.RowNo = offset
		response = append(response, &transactionDay)
	}
	return response, total, pages, nil
}

func (q *transactionDayRepository) FindTransactionDay(
	ctx context.Context,
	request *models.TransactionDayFilter,
) (*models.TransactionDay, error) {
	ctxt := "TransactionDayRepository-FindTransactionDay"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, date
			, public_holiday
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM transaction_days %s`,
		condition,
	)
	query := sb.String()
	var response models.TransactionDay
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.Date,
		&response.PublicHoliday,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrTransactionDayNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *transactionDayRepository) CreateTransactionDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TransactionDay,
) (*models.TransactionDay, error) {
	ctxt := "TransactionDayRepository-CreateTransactionDay"
	var response models.TransactionDay
	if request.Today != nil && *request.Today {
		if _, err := tx.Exec(ctx, "UPDATE transaction_days SET today = $1 WHERE today", nil); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return nil, err
		}
	}
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO transaction_days (
			date
			, today
			, public_holiday
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
			, date
			, today
			, public_holiday
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		request.Date,
		request.Today,
		request.PublicHoliday,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.Date,
		&response.Today,
		&response.PublicHoliday,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "tranction_days_date_key" {
			err = models.ErrUniqueTransactionDayDateViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	_, _ = sb.WriteString(
		`INSERT INTO transaction_windows (
			transaction_day_id
			, delivery_plan_id
			, transaction_window
			, created_at
		) VALUES`,
	)
	for _, transactionWindow := range request.TransactionWindows {
		params = append(
			params,
			response.ID,
			transactionWindow.DeliveryPlanID,
			transactionWindow.TransactionWindowMin,
			transactionWindow.TransactionWindowMax,
			transactionWindow.CreatedAt,
		)
		if n = len(params); n > 5 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " ($%d, $%d, tstzrange($%d, $%d, '[]'), $%d)", n-4, n-3, n-2, n-1, n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	_, _ = sb.WriteString(
		` RETURNING id
			, transaction_day_id
			, delivery_plan_id
			, lower(transaction_window)
			, upper(transaction_window)
			, created_at`,
	)
	rows, err := tx.Query(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	defer rows.Close()
	mapTransactionWindowIDs := map[uuid.UUID]int{}
	for rows.Next() {
		var transactionWindow models.TransactionWindow
		if err = rows.Scan(
			&transactionWindow.ID,
			&transactionWindow.TransactionDayID,
			&transactionWindow.DeliveryPlanID,
			&transactionWindow.TransactionWindowMin,
			&transactionWindow.TransactionWindowMax,
			&transactionWindow.CreatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.TransactionWindows = append(response.TransactionWindows, &transactionWindow)
		mapTransactionWindowIDs[transactionWindow.ID] = len(response.TransactionWindows) - 1
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, err
	}
	params = []any{}
	sb.Reset()
	_, _ = sb.WriteString(
		`INSERT INTO delivery_times (
			transaction_window_id
			, delivery_speed
			, datetime
			, created_at
		) VALUES`,
	)
	for i, transactionWindow := range request.TransactionWindows {
		for _, deliverySchedule := range transactionWindow.DeliveryTimes {
			params = append(
				params,
				response.TransactionWindows[i].ID,
				deliverySchedule.DeliverySpeed,
				deliverySchedule.Datetime,
				deliverySchedule.CreatedAt,
			)
			if n = len(params); n > 4 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " ($%d, $%d, $%d, $%d)", n-3, n-2, n-1, n)
		}
	}
	if len(params) == 0 {
		return &response, nil
	}
	_, _ = sb.WriteString(
		` RETURNING id
			, transaction_window_id
			, delivery_speed
			, datetime
			, created_at`,
	)
	if rows, err = tx.Query(ctx, sb.String(), params...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var deliveryTime models.DeliveryTime
		if err = rows.Scan(
			&deliveryTime.ID,
			&deliveryTime.TransactionWindowID,
			&deliveryTime.DeliverySpeed,
			&deliveryTime.Datetime,
			&deliveryTime.CreatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		i := mapTransactionWindowIDs[deliveryTime.TransactionWindowID]
		response.TransactionWindows[i].DeliveryTimes = append(
			response.TransactionWindows[i].DeliveryTimes,
			&deliveryTime,
		)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, err
	}
	return &response, nil
}

func (q *transactionDayRepository) UpdateTransactionDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TransactionDay,
) (*models.TransactionDay, error) {
	ctxt := "TransactionDayRepository-UpdateTransactionDay"
	var response models.TransactionDay
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, date
			, today
			, public_holiday
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM transaction_days
		WHERE id = $1
		FOR NO KEY UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.Date,
		&response.Today,
		&response.PublicHoliday,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrTransactionDayNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	_, _ = sb.WriteString("UPDATE transaction_days SET")
	if response.Today == nil && request.Today != nil ||
		response.Today != nil && request.Today == nil {
		params = append(params, request.Today)
		_, _ = fmt.Fprintf(&sb, " today = $%d", len(params))
	}
	if response.PublicHoliday != request.PublicHoliday {
		params = append(params, request.PublicHoliday)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " public_holiday = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.UpdatedBy, request.UpdatedAt, request.ID)
	n = len(params)
	_, _ = fmt.Fprintf(
		&sb,
		`, updated_by = $%d
		, updated_at = $%d
		WHERE id = $%d
		RETURNING id
			, date
			, today
			, public_holiday
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		n-2, n-1, n,
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.Date,
		&response.Today,
		&response.PublicHoliday,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *transactionDayRepository) buildCondition(request *models.TransactionDayFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.TransactionDayIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, transactionDayID := range request.TransactionDayIDs {
			params = append(params, transactionDayID)
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
		_, _ = sb.WriteString("date IN (")
		for i, recordStatus := range request.Dates {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Today != nil {
		if *request.Today {
			conds = append(conds, "today")
		} else {
			conds = append(conds, "NOT today")
		}
	}
	if request.PublicHoliday != nil {
		if *request.PublicHoliday {
			conds = append(conds, "public_holiday")
		} else {
			conds = append(conds, "NOT public_holiday")
		}
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
