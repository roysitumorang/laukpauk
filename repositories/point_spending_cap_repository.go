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
	PointSpendingCapRepository interface {
		FindPointSpendingCaps(
			ctx context.Context,
			request *models.PointSpendingCapFilter,
		) ([]*models.PointSpendingCap, int64, int64, error)
		FindPointSpendingCap(
			ctx context.Context,
			request *models.PointSpendingCapFilter,
		) (*models.PointSpendingCap, error)
		CreatePointSpendingCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointSpendingCap,
		) (*models.PointSpendingCap, error)
		UpdatePointSpendingCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointSpendingCap,
		) (*models.PointSpendingCap, error)
	}

	pointSpendingCapRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewPointSpendingCapRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) PointSpendingCapRepository {
	return &pointSpendingCapRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *pointSpendingCapRepository) FindPointSpendingCaps(
	ctx context.Context,
	request *models.PointSpendingCapFilter,
) ([]*models.PointSpendingCap, int64, int64, error) {
	ctxt := "PointSpendingCapRepository-FindPointSpendingCaps"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM point_spending_caps %s", condition)
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
		, disposable_points
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
	var response []*models.PointSpendingCap
	for rows.Next() {
		var pointSpendingCap models.PointSpendingCap
		if err = rows.Scan(
			&pointSpendingCap.ID,
			&pointSpendingCap.RecordStatus,
			&pointSpendingCap.MinimumPurchase,
			&pointSpendingCap.DeletedMinimumPurchase,
			&pointSpendingCap.DisposablePoints,
			&pointSpendingCap.CreatedBy,
			&pointSpendingCap.CreatedAt,
			&pointSpendingCap.UpdatedBy,
			&pointSpendingCap.UpdatedAt,
			&pointSpendingCap.DeletedBy,
			&pointSpendingCap.DeletedAt,
			&pointSpendingCap.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		pointSpendingCap.RowNo = offset
		response = append(response, (&pointSpendingCap).Populate())
	}
	return response, total, pages, nil
}

func (q *pointSpendingCapRepository) FindPointSpendingCap(
	ctx context.Context,
	request *models.PointSpendingCapFilter,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapRepository-FindPointSpendingCap"
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
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM point_spending_caps %s`,
		condition,
	)
	query := sb.String()
	var response models.PointSpendingCap
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.MinimumPurchase,
		&response.DeletedMinimumPurchase,
		&response.DisposablePoints,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPointSpendingCapNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *pointSpendingCapRepository) CreatePointSpendingCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointSpendingCap,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapRepository-CreatePointSpendingCap"
	var response models.PointSpendingCap
	err := tx.QueryRow(
		ctx,
		`INSERT INTO point_spending_caps (
			record_status
			, minimum_purchase
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
			, record_status
			, minimum_purchase
			, deleted_minimum_purchase
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.MinimumPurchase,
		request.DisposablePoints,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.MinimumPurchase,
		&response.DeletedMinimumPurchase,
		&response.DisposablePoints,
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
			pgxErr.ConstraintName == "point_spending_caps_minimum_purchase_key" {
			err = models.ErrUniquePointSpendingCapMinimumPurchaseViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *pointSpendingCapRepository) UpdatePointSpendingCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointSpendingCap,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapRepository-UpdatePointSpendingCap"
	var response models.PointSpendingCap
	err := tx.QueryRow(
		ctx,
		`UPDATE point_spending_caps SET
			record_status = $1
			, minimum_purchase = $2
			, deleted_minimum_purchase = $3
			, disposable_points = $4
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
			, disposable_points
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
		request.DisposablePoints,
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
		&response.DisposablePoints,
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
			pgxErr.ConstraintName == "point_spending_caps_minimum_purchase_key" {
			err = models.ErrUniquePointSpendingCapMinimumPurchaseViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *pointSpendingCapRepository) buildCondition(request *models.PointSpendingCapFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.PointSpendingCapIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, pointSpendingCapID := range request.PointSpendingCapIDs {
			params = append(params, pointSpendingCapID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedPointSpendingCapIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedPointSpendingCapID := range request.ExcludedPointSpendingCapIDs {
			params = append(params, excludedPointSpendingCapID)
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
