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

type PointSpendingCapRepository interface {
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

type pointSpendingCapRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

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
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM point_spending_caps %s", condition)
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
		, company_id
		, record_status
		, LOWER(shopping_cost)
		, UPPER(shopping_cost) - 1
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
	var response []*models.PointSpendingCap
	for rows.Next() {
		var pointSpendingCap models.PointSpendingCap
		if err = rows.Scan(
			&pointSpendingCap.ID,
			&pointSpendingCap.CompanyID,
			&pointSpendingCap.RecordStatus,
			&pointSpendingCap.MinimumShoppingCost,
			&pointSpendingCap.MaximumShoppingCost,
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
			, company_id
			, record_status
			, LOWER(shopping_cost)
			, UPPER(shopping_cost) - 1
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
		&response.CompanyID,
		&response.RecordStatus,
		&response.MinimumShoppingCost,
		&response.MaximumShoppingCost,
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
			company_id
			, record_status
			, shopping_cost
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, int8range($3, $4, '[]'), $5, $6, $7, $8, $9)
		RETURNING id
			, company_id
			, record_status
			, LOWER(shopping_cost)
			, UPPER(shopping_cost) - 1
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.CompanyID,
		request.RecordStatus,
		request.MinimumShoppingCost,
		request.MaximumShoppingCost,
		request.DisposablePoints,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.CompanyID,
		&response.RecordStatus,
		&response.MinimumShoppingCost,
		&response.MaximumShoppingCost,
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
			pgxErr.Code == pgerrcode.ExclusionViolation &&
			pgxErr.ConstraintName == "point_spending_caps_no_overlapping_shopping_cost" {
			err = models.ErrPointSpendingCapShoppingCostOverlapping
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
		`SELECT
			id
			, company_id
			, record_status
			, LOWER(shopping_cost)
			, UPPER(shopping_cost) - 1
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM point_spending_caps
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.CompanyID,
		&response.RecordStatus,
		&response.MinimumShoppingCost,
		&response.MaximumShoppingCost,
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
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE point_spending_caps SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if (response.MinimumShoppingCost != request.MinimumShoppingCost) ||
		(response.MaximumShoppingCost != request.MaximumShoppingCost) {
		params = append(params, request.MinimumShoppingCost, request.MaximumShoppingCost)
		if n = len(params); n > 2 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " shopping_cost = int8range($%d, $%d, '[]')", n-1, n)
	}
	if response.DisposablePoints != request.DisposablePoints {
		params = append(params, request.DisposablePoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " disposable_points = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	if response.UpdatedBy.Compare(request.UpdatedBy) != 0 {
		params = append(params, request.UpdatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_by = $%d", n)
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_at = $%d", n)
	}
	if (response.DeletedBy == nil && request.DeletedBy != nil) ||
		(response.DeletedBy != nil && request.DeletedBy == nil) ||
		(response.DeletedBy != nil && request.DeletedBy != nil &&
			*response.DeletedBy != *request.DeletedBy) {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_by = $%d", n)
	}
	if (response.DeletedAt == nil && request.DeletedAt != nil) ||
		(response.DeletedAt != nil && request.DeletedAt == nil) ||
		(response.DeletedAt != nil && request.DeletedAt != nil &&
			*response.DeletedAt != *request.DeletedAt) {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_at = $%d", n)
	}
	if (response.DeletionReason == nil && request.DeletionReason != nil) ||
		(response.DeletionReason != nil && request.DeletionReason == nil) ||
		(response.DeletionReason != nil && request.DeletionReason != nil &&
			*response.DeletionReason != *request.DeletionReason) {
		params = append(params, request.DeletionReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deletion_reason = $%d", n)
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, company_id
			, record_status
			, LOWER(shopping_cost)
			, UPPER(shopping_cost) - 1
			, disposable_points
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
		&response.CompanyID,
		&response.RecordStatus,
		&response.MinimumShoppingCost,
		&response.MaximumShoppingCost,
		&response.DisposablePoints,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "point_spending_caps_no_overlapping_shopping_cost" {
			err = models.ErrPointSpendingCapShoppingCostOverlapping
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
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
	if len(request.CompanyIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("company_id IN (")
		for i, companyID := range request.CompanyIDs {
			params = append(params, companyID)
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
	if request.MinimumShoppingCost > 0 && request.MaximumShoppingCost > 0 {
		params = append(params, request.MinimumShoppingCost, request.MaximumShoppingCost)
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "shopping_cost && int8range($%d, $%d, '[]')", n-1, n)
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
