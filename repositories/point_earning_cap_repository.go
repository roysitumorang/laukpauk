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
	PointEarningCapRepository interface {
		FindPointEarningCaps(
			ctx context.Context,
			request *models.PointEarningCapFilter,
		) ([]*models.PointEarningCap, int64, int64, error)
		FindPointEarningCap(
			ctx context.Context,
			request *models.PointEarningCapFilter,
		) (*models.PointEarningCap, error)
		CreatePointEarningCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointEarningCap,
		) (*models.PointEarningCap, error)
		UpdatePointEarningCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointEarningCap,
		) (*models.PointEarningCap, error)
	}

	pointEarningCapRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewPointEarningCapRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) PointEarningCapRepository {
	return new(pointEarningCapRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *pointEarningCapRepository) FindPointEarningCaps(
	ctx context.Context,
	request *models.PointEarningCapFilter,
) ([]*models.PointEarningCap, int64, int64, error) {
	ctxt := "PointEarningCapRepository-FindPointEarningCaps"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM point_earning_caps %s", condition)
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
		, minimum_purchase
		, deleted_minimum_purchase
		, earned_points
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
	var response []*models.PointEarningCap
	for rows.Next() {
		var pointEarningCap models.PointEarningCap
		if err = rows.Scan(
			&pointEarningCap.ID,
			&pointEarningCap.RecordStatus,
			&pointEarningCap.MinimumPurchase,
			&pointEarningCap.DeletedMinimumPurchase,
			&pointEarningCap.EarnedPoints,
			&pointEarningCap.CreatedBy,
			&pointEarningCap.CreatedAt,
			&pointEarningCap.UpdatedBy,
			&pointEarningCap.UpdatedAt,
			&pointEarningCap.DeletedBy,
			&pointEarningCap.DeletedAt,
			&pointEarningCap.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		pointEarningCap.RowNo = offset
		response = append(response, (&pointEarningCap).Populate())
	}
	return response, total, pages, nil
}

func (q *pointEarningCapRepository) FindPointEarningCap(
	ctx context.Context,
	request *models.PointEarningCapFilter,
) (*models.PointEarningCap, error) {
	ctxt := "PointEarningCapRepository-FindPointEarningCap"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, record_status
			, minimum_purchase
			, deleted_minimum_purchase
			, earned_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM point_earning_caps %s`,
		condition,
	)
	query := sb.String()
	var response models.PointEarningCap
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.MinimumPurchase,
		&response.DeletedMinimumPurchase,
		&response.EarnedPoints,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPointEarningCapNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *pointEarningCapRepository) CreatePointEarningCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarningCap,
) (*models.PointEarningCap, error) {
	ctxt := "PointEarningCapRepository-CreatePointEarningCap"
	var response models.PointEarningCap
	err := tx.QueryRow(
		ctx,
		`INSERT INTO point_earning_caps (
			record_status
			, minimum_purchase
			, earned_points
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
			, record_status
			, minimum_purchase
			, deleted_minimum_purchase
			, earned_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.MinimumPurchase,
		request.EarnedPoints,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.MinimumPurchase,
		&response.DeletedMinimumPurchase,
		&response.EarnedPoints,
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
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "point_earning_caps_minimum_purchase_key" {
			err = models.ErrUniquePointEarningCapMinimumPurchaseViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *pointEarningCapRepository) UpdatePointEarningCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarningCap,
) (*models.PointEarningCap, error) {
	ctxt := "PointEarningCapRepository-UpdatePointEarningCap"
	var response models.PointEarningCap
	err := tx.QueryRow(
		ctx,
		`UPDATE point_earning_caps SET
			record_status = $1
			, minimum_purchase = $2
			, deleted_minimum_purchase = $3
			, earned_points = $4
			, updated_by = $5
			, updated_at = $6
			, deleted_by = $7
			, deleted_at = $8
			, deletion_reason = $9
		WHERE id = $10
		RETURNING id
			, record_status
			, minimum_purchase
			, deleted_minimum_purchase
			, earned_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.MinimumPurchase,
		request.DeletedMinimumPurchase,
		request.EarnedPoints,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.MinimumPurchase,
		&response.DeletedMinimumPurchase,
		&response.EarnedPoints,
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
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "point_earning_caps_minimum_purchase_key" {
			err = models.ErrUniquePointEarningCapMinimumPurchaseViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *pointEarningCapRepository) buildCondition(request *models.PointEarningCapFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.PointEarningCapIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, pointEarningCapID := range request.PointEarningCapIDs {
			params = append(params, pointEarningCapID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedPointEarningCapIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedPointEarningCapID := range request.ExcludedPointEarningCapIDs {
			params = append(params, excludedPointEarningCapID)
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
	if len(request.MinimumPurchases) > 0 {
		sb.Reset()
		for i, minimumPurchase := range request.MinimumPurchases {
			params = append(params, minimumPurchase)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		placeholder := sb.String()
		sb.Reset()
		_, _ = fmt.Fprintf(
			sb,
			"(minimum_purchase IN (%s) OR deleted_minimum_purchase IN (%s))",
			placeholder,
			placeholder,
		)
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
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
