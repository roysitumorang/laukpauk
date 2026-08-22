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

type RecipeRepository interface {
	FindRecipes(
		ctx context.Context,
		request *models.RecipeFilter,
	) ([]*models.Recipe, int64, int64, error)
	FindRecipe(
		ctx context.Context,
		request *models.RecipeFilter,
	) (*models.Recipe, error)
	CreateRecipe(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Recipe,
	) (*models.Recipe, error)
	UpdateRecipe(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Recipe,
	) (*models.Recipe, error)
	SaveRecipe(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Recipe,
	) (*models.Recipe, error)
}

type recipeRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewRecipeRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) RecipeRepository {
	return &recipeRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *recipeRepository) FindRecipes(
	ctx context.Context,
	request *models.RecipeFilter,
) ([]*models.Recipe, int64, int64, error) {
	ctxt := "RecipeRepository-FindRecipes"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM recipes %s", condition)
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
		, photo
		, description
		, instagram_link
		, facebook_link
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
	var response []*models.Recipe
	for rows.Next() {
		var recipe models.Recipe
		if err = rows.Scan(
			&recipe.ID,
			&recipe.RecordStatus,
			&recipe.Name,
			&recipe.Photo,
			&recipe.Description,
			&recipe.InstagramLink,
			&recipe.FacebookLink,
			&recipe.CreatedBy,
			&recipe.CreatedAt,
			&recipe.UpdatedBy,
			&recipe.UpdatedAt,
			&recipe.DeletedBy,
			&recipe.DeletedAt,
			&recipe.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		recipe.RowNo = offset
		response = append(response, &recipe)
	}
	return response, total, pages, nil
}

func (q *recipeRepository) FindRecipe(
	ctx context.Context,
	request *models.RecipeFilter,
) (*models.Recipe, error) {
	ctxt := "RecipeRepository-FindRecipe"
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
			, photo
			, description
			, instagram_link
			, facebook_link
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM recipes %s`,
		condition,
	)
	var response models.Recipe
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.InstagramLink,
		&response.FacebookLink,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrRecipeNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *recipeRepository) CreateRecipe(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Recipe,
) (*models.Recipe, error) {
	ctxt := "RecipeRepository-CreateRecipe"
	var response models.Recipe
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO recipes (
			record_status
			, name
			, photo
			, description
			, instagram_link
			, facebook_link
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
			, record_status
			, name
			, photo
			, description
			, instagram_link
			, facebook_link
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.Photo,
		request.Description,
		request.InstagramLink,
		request.FacebookLink,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.InstagramLink,
		&response.FacebookLink,
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
			pgxErr.ConstraintName == "recipes_unique_name_idx" {
			err = models.ErrUniqueRecipeNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *recipeRepository) UpdateRecipe(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Recipe,
) (*models.Recipe, error) {
	ctxt := "RecipeRepository-UpdateRecipe"
	var response models.Recipe
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, name
			, photo
			, description
			, instagram_link
			, facebook_link
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM recipes
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.InstagramLink,
		&response.FacebookLink,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrRecipeNotFound
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
	_, _ = sb.WriteString("UPDATE recipes SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if (response.Name == nil && request.Name != nil) ||
		(response.Name != nil && request.Name == nil) ||
		(response.Name != nil && request.Name != nil &&
			*response.Name != *request.Name) {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " name = $%d", n)
	}
	if (response.DeletedName == nil && request.DeletedName != nil) ||
		(response.DeletedName != nil && request.DeletedName == nil) ||
		(response.DeletedName != nil && request.DeletedName != nil &&
			*response.DeletedName != *request.DeletedName) {
		params = append(params, request.DeletedName)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_name = $%d", n)
	}
	if response.Photo != request.Photo {
		params = append(params, request.Photo)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " photo = $%d", n)
	}
	if response.Description != request.Description {
		params = append(params, request.Description)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " description = $%d", n)
	}
	if response.InstagramLink != request.InstagramLink {
		params = append(params, request.InstagramLink)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " instagram_link = $%d", n)
	}
	if response.FacebookLink != request.FacebookLink {
		params = append(params, request.FacebookLink)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " facebook_link = $%d", n)
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
			response.DeletedBy.Compare(*request.DeletedBy) != 0) {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_by = $%d", n)
	}
	if (response.DeletedAt == nil && request.DeletedAt != nil) ||
		(response.DeletedAt != nil && request.DeletedAt == nil) ||
		(response.DeletedAt != nil && request.DeletedAt != nil &&
			!response.DeletedAt.Equal(*request.DeletedAt)) {
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
			, record_status
			, name
			, photo
			, description
			, instagram_link
			, facebook_link
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
		&response.RecordStatus,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.InstagramLink,
		&response.FacebookLink,
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
			pgxErr.ConstraintName == "recipes_unique_name_idx" {
			err = models.ErrUniqueRecipeNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *recipeRepository) SaveRecipe(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Recipe,
) (*models.Recipe, error) {
	if request.New() {
		return q.CreateRecipe(ctx, tx, request)
	}
	return q.UpdateRecipe(ctx, tx, request)
}

func (q *recipeRepository) buildCondition(request *models.RecipeFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.RecipeIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, recipeID := range request.RecipeIDs {
			params = append(params, recipeID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedRecipeIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedRecipeID := range request.ExcludedRecipeIDs {
			params = append(params, excludedRecipeID)
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
	if len(request.Names) > 0 {
		sb.Reset()
		for i, name := range request.Names {
			params = append(params, strings.ToLower(name))
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		placeholders := sb.String()
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "(LOWER(name) IN (%s) OR LOWER(deleted_name) IN (%s))", placeholders, placeholders)
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
