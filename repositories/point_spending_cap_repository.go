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
		tx *sql.Tx,
		request *models.PointSpendingCap,
	) (*models.PointSpendingCap, error)
	UpdatePointSpendingCap(
		ctx context.Context,
		tx *sql.Tx,
		request *models.PointSpendingCap,
	) (*models.PointSpendingCap, error)
}

type pointSpendingCapRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewPointSpendingCapRepository(
	dbRead,
	dbWrite *sql.DB,
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
	err := q.dbRead.QueryRowContext(ctx, query, params...).Scan(&total)
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
		, record_status
		, shopping_cost_min
		, shopping_cost_max
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQueryContext")
		return nil, 0, 0, err
	}
	defer func() { _ = rows.Close() }()
	var response []*models.PointSpendingCap
	for rows.Next() {
		var pointSpendingCap models.PointSpendingCap
		if err = rows.Scan(
			&pointSpendingCap.ID,
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
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, 0, 0, err
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
			, record_status
			, shopping_cost_min
			, shopping_cost_max
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
	err := q.dbRead.QueryRowContext(ctx, query, params...).Scan(
		&response.ID,
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
	if errors.Is(err, sql.ErrNoRows) {
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
	tx *sql.Tx,
	request *models.PointSpendingCap,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapRepository-CreatePointSpendingCap"
	var response models.PointSpendingCap
	if err := q.checkShoppingCost(ctx, tx, nil, request.MinimumShoppingCost, request.MaximumShoppingCost); err != nil {
		return nil, err
	}
	err := tx.QueryRowContext(
		ctx,
		`INSERT INTO point_spending_caps (
			record_status
			, shopping_cost_min
			, shopping_cost_max
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
			, record_status
			, shopping_cost_min
			, shopping_cost_max
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *pointSpendingCapRepository) UpdatePointSpendingCap(
	ctx context.Context,
	tx *sql.Tx,
	request *models.PointSpendingCap,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapRepository-UpdatePointSpendingCap"
	var response models.PointSpendingCap
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, record_status
			, shopping_cost_min
			, shopping_cost_max
			, disposable_points
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM point_spending_caps
		WHERE id = ?
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
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
	if errors.Is(err, sql.ErrNoRows) {
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
		_, _ = sb.WriteString(" record_status = ?")
	}
	if (response.MinimumShoppingCost != request.MinimumShoppingCost) ||
		(response.MaximumShoppingCost != request.MaximumShoppingCost) {
		if err = q.checkShoppingCost(
			ctx,
			tx,
			&request.ID,
			request.MinimumShoppingCost,
			request.MaximumShoppingCost,
		); err != nil {
			return nil, err
		}
		params = append(params, request.MinimumShoppingCost, request.MaximumShoppingCost)
		if n = len(params); n > 2 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" shopping_cost_min = ?, shopping_cost_max = ?")
	}
	if response.DisposablePoints != request.DisposablePoints {
		params = append(params, request.DisposablePoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" disposable_points = ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	if response.UpdatedBy != request.UpdatedBy {
		params = append(params, request.UpdatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" updated_by = ?")
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" updated_at = ?")
	}
	if (response.DeletedBy == nil && request.DeletedBy != nil) ||
		(response.DeletedBy != nil && request.DeletedBy == nil) ||
		(response.DeletedBy != nil && request.DeletedBy != nil &&
			*response.DeletedBy != *request.DeletedBy) {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deleted_by = ?")
	}
	if (response.DeletedAt == nil && request.DeletedAt != nil) ||
		(response.DeletedAt != nil && request.DeletedAt == nil) ||
		(response.DeletedAt != nil && request.DeletedAt != nil &&
			*response.DeletedAt != *request.DeletedAt) {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deleted_at = ?")
	}
	if (response.DeletionReason == nil && request.DeletionReason != nil) ||
		(response.DeletionReason != nil && request.DeletionReason == nil) ||
		(response.DeletionReason != nil && request.DeletionReason != nil &&
			*response.DeletionReason != *request.DeletionReason) {
		params = append(params, request.DeletionReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deletion_reason = ?")
	}
	params = append(params, request.ID)
	_, _ = sb.WriteString(` WHERE id = ?
		RETURNING id
			, record_status
			, shopping_cost_min
			, shopping_cost_max
			, disposable_points
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
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
			_, _ = sb.WriteString("?")
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
	if request.MinimumShoppingCost > 0 && request.MaximumShoppingCost > 0 {
		params = append(params, request.MaximumShoppingCost, request.MinimumShoppingCost)
		sb.Reset()
		_, _ = sb.WriteString("shopping_cost_min <= ? AND shopping_cost_max >= ?")
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

// checkShoppingCost returns models.ErrPointSpendingCapShoppingCostOverlapping when another point spending cap's shopping cost range overlaps
// [minimum, maximum] (both ends included); excludeID is the cap being updated. MariaDB has no exclusion constraints,
// so this check takes their place: FOR UPDATE locks the overlapping caps (and, through the shopping cost index, the
// range around them) until the transaction ends, so two transactions can't add overlapping caps at the same time.
func (q *pointSpendingCapRepository) checkShoppingCost(
	ctx context.Context,
	tx *sql.Tx,
	excludeID *uuid.UUID,
	minimum,
	maximum int64,
) error {
	ctxt := "PointSpendingCapRepository-checkShoppingCost"
	params := []any{maximum, minimum}
	query := `SELECT id
		FROM point_spending_caps
		WHERE shopping_cost_min <= ?
			AND shopping_cost_max >= ?`
	if excludeID != nil {
		params = append(params, *excludeID)
		query += " AND id != ?"
	}
	rows, err := tx.QueryContext(ctx, query+" FOR UPDATE", params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQueryContext")
		return err
	}
	defer func() { _ = rows.Close() }()
	overlapping := rows.Next()
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return err
	}
	if overlapping {
		return models.ErrPointSpendingCapShoppingCostOverlapping
	}
	return nil
}
