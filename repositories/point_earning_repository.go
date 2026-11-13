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
	"github.com/roysitumorang/laukpauk/pools"
)

type PointEarningRepository interface {
	FindPointEarnings(
		ctx context.Context,
		request *models.PointEarningFilter,
	) ([]*models.PointEarning, int64, int64, error)
	FindPointEarning(
		ctx context.Context,
		request *models.PointEarningFilter,
	) (*models.PointEarning, error)
	CreatePointEarning(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointEarning,
	) (*models.PointEarning, error)
	UpdatePointEarning(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointEarning,
	) (*models.PointEarning, error)
}

type pointEarningRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewPointEarningRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) PointEarningRepository {
	return &pointEarningRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *pointEarningRepository) FindPointEarnings(
	ctx context.Context,
	request *models.PointEarningFilter,
) ([]*models.PointEarning, int64, int64, error) {
	ctxt := "PointEarningRepository-FindPointEarnings"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM point_earnings %s", condition)
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
		, user_id
		, start_date
		, end_date
		, amount
		, created_at
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
	var response []*models.PointEarning
	for rows.Next() {
		var pointEarning models.PointEarning
		if err = rows.Scan(
			&pointEarning.ID,
			&pointEarning.UserID,
			&pointEarning.StartDate,
			&pointEarning.EndDate,
			&pointEarning.Amount,
			&pointEarning.CreatedAt,
			&pointEarning.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		pointEarning.RowNo = offset
		response = append(response, &pointEarning)
	}
	return response, total, pages, nil
}

func (q *pointEarningRepository) FindPointEarning(
	ctx context.Context,
	request *models.PointEarningFilter,
) (*models.PointEarning, error) {
	ctxt := "PointEarningRepository-FindPointEarning"
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
			, user_id
			, start_date
			, end_date
			, amount
			, created_at
			, updated_at
		FROM point_earnings %s`,
		condition,
	)
	query := sb.String()
	var response models.PointEarning
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.UserID,
		&response.StartDate,
		&response.EndDate,
		&response.Amount,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPointEarningNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *pointEarningRepository) CreatePointEarning(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarning,
) (*models.PointEarning, error) {
	ctxt := "PointEarningRepository-CreatePointEarning"
	var response models.PointEarning
	err := tx.QueryRow(
		ctx,
		`INSERT INTO point_earnings (
			user_id
			, start_date
			, end_date
			, amount
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
			, user_id
			, start_date
			, end_date
			, amount
			, created_at
			, updated_at`,
		request.UserID,
		request.StartDate,
		request.EndDate,
		request.Amount,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.UserID,
		&response.StartDate,
		&response.EndDate,
		&response.Amount,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "point_earnings_user_id_start_date_end_date_key" {
			err = models.ErrPointEarningUniqueBuyerIdStartDateEndDate
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *pointEarningRepository) UpdatePointEarning(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarning,
) (*models.PointEarning, error) {
	ctxt := "PointEarningRepository-UpdatePointEarning"
	var response models.PointEarning
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, user_id
			, start_date
			, end_date
			, amount
			, created_at
			, updated_at
		FROM point_earnings
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.UserID,
		&response.StartDate,
		&response.EndDate,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPointEarningNotFound
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
	_, _ = sb.WriteString("UPDATE point_earnings SET")
	if response.Amount != request.Amount {
		params = append(params, request.Amount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " amount = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_at = $%d", n)
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, user_id
			, start_date
			, end_date
			, amount
			, created_at
			, updated_at`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.UserID,
		&response.StartDate,
		&response.EndDate,
		&response.Amount,
		&response.CreatedAt,
		&response.UpdatedAt,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "point_earnings_user_id_start_date_end_date_key" {
			err = models.ErrPointEarningUniqueBuyerIdStartDateEndDate
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *pointEarningRepository) buildCondition(request *models.PointEarningFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.PointEarningIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, pointEarningID := range request.PointEarningIDs {
			params = append(params, pointEarningID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("user_id IN (")
		for i, buyerID := range request.UserIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if !request.StartDate.IsZero() {
		params = append(params, request.StartDate)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "start_date = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if !request.EndDate.IsZero() {
		params = append(params, request.EndDate)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "end_date = $%d", len(params))
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
