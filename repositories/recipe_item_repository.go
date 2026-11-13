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

type RecipeItemRepository interface {
	FindRecipeItems(
		ctx context.Context,
		request *models.RecipeItemFilter,
	) ([]*models.RecipeItem, int64, int64, error)
	FindRecipeItem(
		ctx context.Context,
		request *models.RecipeItemFilter,
	) (*models.RecipeItem, error)
	CreateRecipeItem(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeItem,
	) (*models.RecipeItem, error)
	UpdateRecipeItem(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeItem,
	) (*models.RecipeItem, error)
	SaveRecipeItem(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeItem,
	) (*models.RecipeItem, error)
}

type recipeItemRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewRecipeItemRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) RecipeItemRepository {
	return &recipeItemRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *recipeItemRepository) FindRecipeItems(
	ctx context.Context,
	request *models.RecipeItemFilter,
) ([]*models.RecipeItem, int64, int64, error) {
	ctxt := "RecipeItemRepository-FindRecipeItems"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM recipe_items %s", condition)
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
		, recipe_id
		, product_id
		, product_name
		, product_photo
		, product_ready_stock
		, product_variant_id
		, product_variant_stock_unit
		, quantity
		, note
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
	var response []*models.RecipeItem
	for rows.Next() {
		var recipeItem models.RecipeItem
		if err = rows.Scan(
			&recipeItem.ID,
			&recipeItem.RecordStatus,
			&recipeItem.RecipeID,
			&recipeItem.ProductID,
			&recipeItem.ProductName,
			&recipeItem.ProductPhoto,
			&recipeItem.ProductReadyStock,
			&recipeItem.ProductVariantID,
			&recipeItem.ProductVariantStockUnit,
			&recipeItem.Quantity,
			&recipeItem.Note,
			&recipeItem.CreatedBy,
			&recipeItem.CreatedAt,
			&recipeItem.UpdatedBy,
			&recipeItem.UpdatedAt,
			&recipeItem.DeletedBy,
			&recipeItem.DeletedAt,
			&recipeItem.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		recipeItem.RowNo = offset
		response = append(response, &recipeItem)
	}
	return response, total, pages, nil
}

func (q *recipeItemRepository) FindRecipeItem(
	ctx context.Context,
	request *models.RecipeItemFilter,
) (*models.RecipeItem, error) {
	ctxt := "RecipeItemRepository-FindRecipeItem"
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
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, product_variant_id
			, product_variant_stock_unit
			, quantity
			, note
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM recipe_items %s`,
		condition,
	)
	var response models.RecipeItem
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RecipeID,
		&response.ProductID,
		&response.ProductName,
		&response.ProductPhoto,
		&response.ProductReadyStock,
		&response.ProductVariantID,
		&response.ProductVariantStockUnit,
		&response.Quantity,
		&response.Note,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrRecipeItemNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *recipeItemRepository) CreateRecipeItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeItem,
) (*models.RecipeItem, error) {
	ctxt := "RecipeItemRepository-CreateRecipeItem"
	var response models.RecipeItem
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO recipe_items (
			record_status
			, recipe_id
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, product_variant_id
			, product_variant_stock_unit
			, quantity
			, note
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
			, record_status
			, recipe_id
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, product_variant_id
			, product_variant_stock_unit
			, quantity
			, note
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.RecipeID,
		request.ProductID,
		request.ProductName,
		request.ProductPhoto,
		request.ProductReadyStock,
		request.ProductVariantID,
		request.ProductVariantStockUnit,
		request.Quantity,
		request.Note,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RecipeID,
		&response.ProductID,
		&response.ProductName,
		&response.ProductPhoto,
		&response.ProductReadyStock,
		&response.ProductVariantID,
		&response.ProductVariantStockUnit,
		&response.Quantity,
		&response.Note,
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
			pgxErr.ConstraintName == "recipe_items_recipe_id_product_id_idx" {
			err = models.ErrUniqueRecipeItemProductIdViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *recipeItemRepository) UpdateRecipeItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeItem,
) (*models.RecipeItem, error) {
	ctxt := "RecipeItemRepository-UpdateRecipeItem"
	var response models.RecipeItem
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, recipe_id
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, product_variant_id
			, product_variant_stock_unit
			, quantity
			, note
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM recipe_items
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RecipeID,
		&response.ProductID,
		&response.ProductName,
		&response.ProductPhoto,
		&response.ProductReadyStock,
		&response.ProductVariantID,
		&response.ProductVariantStockUnit,
		&response.Quantity,
		&response.Note,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrRecipeItemNotFound
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
	_, _ = sb.WriteString("UPDATE recipe_items SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if response.ProductID.Compare(request.ProductID) != 0 {
		params = append(params, request.ProductID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_id = $%d", n)
	}
	if response.ProductName != request.ProductName {
		params = append(params, request.ProductName)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_name = $%d", n)
	}
	if response.ProductPhoto != request.ProductPhoto {
		params = append(params, request.ProductPhoto)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_photo = $%d", n)
	}
	if response.ProductReadyStock != request.ProductReadyStock {
		params = append(params, request.ProductReadyStock)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_ready_stock = $%d", n)
	}
	if response.ProductVariantID != request.ProductVariantID {
		params = append(params, request.ProductVariantID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_variant_id = $%d", n)
	}
	if response.ProductVariantStockUnit != request.ProductVariantStockUnit {
		params = append(params, request.ProductVariantStockUnit)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_variant_stock_unit = $%d", n)
	}
	if response.Quantity != request.Quantity {
		params = append(params, request.Quantity)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " quantity = $%d", n)
	}
	if response.Note != request.Note {
		params = append(params, request.Note)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " note = $%d", n)
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
			, recipe_id
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, product_variant_id
			, product_variant_stock_unit
			, quantity
			, note
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
		&response.RecipeID,
		&response.ProductID,
		&response.ProductName,
		&response.ProductPhoto,
		&response.ProductReadyStock,
		&response.ProductVariantID,
		&response.ProductVariantStockUnit,
		&response.Quantity,
		&response.Note,
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
			pgxErr.ConstraintName == "recipe_items_recipe_id_product_id_idx" {
			err = models.ErrUniqueRecipeItemProductIdViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *recipeItemRepository) SaveRecipeItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeItem,
) (*models.RecipeItem, error) {
	if request.New() {
		return q.CreateRecipeItem(ctx, tx, request)
	}
	return q.UpdateRecipeItem(ctx, tx, request)
}

func (q *recipeItemRepository) buildCondition(request *models.RecipeItemFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.RecipeIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("recipe_id IN (")
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
	if len(request.ProductIDs) > 0 {
		sb.Reset()
		for i, productID := range request.ProductIDs {
			params = append(params, productID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		placeholder := sb.String()
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "product_id IN (%s)", placeholder)
		conds = append(conds, sb.String())
	}
	if len(request.ProductVariantIDs) > 0 {
		sb.Reset()
		for i, productVariantID := range request.ProductVariantIDs {
			params = append(params, productVariantID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		placeholder := sb.String()
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "product_variant_id IN (%s)", placeholder)
		conds = append(conds, sb.String())
	}
	if len(request.RecipeItemIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, recipeItemID := range request.RecipeItemIDs {
			params = append(params, recipeItemID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedRecipeItemIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedRecipeItemID := range request.ExcludedRecipeItemIDs {
			params = append(params, excludedRecipeItemID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.ProductReadyStock != nil {
		if *request.ProductReadyStock {
			conds = append(conds, "product_ready_stock")
		} else {
			conds = append(conds, "NOT product_ready_stock")
		}
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
