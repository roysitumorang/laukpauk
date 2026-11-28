package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type RecipeGuideRepository interface {
	FindRecipeGuides(
		ctx context.Context,
		request *models.RecipeGuideFilter,
	) ([]*models.RecipeGuide, int64, int64, error)
	FindRecipeGuide(
		ctx context.Context,
		request *models.RecipeGuideFilter,
	) (*models.RecipeGuide, error)
	CreateRecipeGuide(
		ctx context.Context,
		tx *sql.Tx,
		request *models.RecipeGuide,
	) (*models.RecipeGuide, error)
	UpdateRecipeGuide(
		ctx context.Context,
		tx *sql.Tx,
		request *models.RecipeGuide,
	) (*models.RecipeGuide, error)
}

type recipeGuideRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewRecipeGuideRepository(
	dbRead,
	dbWrite *sql.DB,
) RecipeGuideRepository {
	return &recipeGuideRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *recipeGuideRepository) FindRecipeGuides(
	ctx context.Context,
	request *models.RecipeGuideFilter,
) ([]*models.RecipeGuide, int64, int64, error) {
	ctxt := "RecipeGuideRepository-FindRecipeGuides"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM recipe_guides %s", condition)
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
		, recipe_id
		, instruction
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
	var response []*models.RecipeGuide
	for rows.Next() {
		var recipeGuide models.RecipeGuide
		if err = rows.Scan(
			&recipeGuide.ID,
			&recipeGuide.RecordStatus,
			&recipeGuide.RecipeID,
			&recipeGuide.Instruction,
			&recipeGuide.CreatedBy,
			&recipeGuide.CreatedAt,
			&recipeGuide.UpdatedBy,
			&recipeGuide.UpdatedAt,
			&recipeGuide.DeletedBy,
			&recipeGuide.DeletedAt,
			&recipeGuide.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		recipeGuide.RowNo = offset
		response = append(response, &recipeGuide)
	}
	return response, total, pages, nil
}

func (q *recipeGuideRepository) FindRecipeGuide(
	ctx context.Context,
	request *models.RecipeGuideFilter,
) (*models.RecipeGuide, error) {
	ctxt := "RecipeGuideRepository-FindRecipeGuide"
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
			, recipe_id
			, instruction
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM recipe_guides %s`,
		condition,
	)
	var response models.RecipeGuide
	err := q.dbRead.QueryRowContext(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RecipeID,
		&response.Instruction,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrRecipeGuideNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *recipeGuideRepository) CreateRecipeGuide(
	ctx context.Context,
	tx *sql.Tx,
	request *models.RecipeGuide,
) (*models.RecipeGuide, error) {
	ctxt := "RecipeGuideRepository-CreateRecipeGuide"
	var response models.RecipeGuide
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO recipe_guides (
			record_status
			, recipe_id
			, instruction
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id
			, record_status
			, recipe_id
			, instruction
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.RecipeID,
		request.Instruction,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RecipeID,
		&response.Instruction,
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

func (q *recipeGuideRepository) UpdateRecipeGuide(
	ctx context.Context,
	tx *sql.Tx,
	request *models.RecipeGuide,
) (*models.RecipeGuide, error) {
	ctxt := "RecipeGuideRepository-UpdateRecipeGuide"
	var response models.RecipeGuide
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, record_status
			, recipe_id
			, instruction
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM recipe_guides
		WHERE id = ?
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RecipeID,
		&response.Instruction,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrRecipeGuideNotFound
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
	_, _ = sb.WriteString("UPDATE recipe_guides SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" record_status = ?")
	}
	if response.Instruction != request.Instruction {
		params = append(params, request.Instruction)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" instruction = ?")
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
			!response.DeletedAt.Equal(*request.DeletedAt)) {
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
			, recipe_id
			, instruction
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
		&response.RecipeID,
		&response.Instruction,
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

func (q *recipeGuideRepository) buildCondition(request *models.RecipeGuideFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.RecipeIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("recipe_id IN (")
		for i, recipeID := range request.RecipeIDs {
			params = append(params, recipeID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecipeGuideIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, recipeGuideID := range request.RecipeGuideIDs {
			params = append(params, recipeGuideID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedRecipeGuideIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedRecipeGuideID := range request.ExcludedRecipeGuideIDs {
			params = append(params, excludedRecipeGuideID)
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
