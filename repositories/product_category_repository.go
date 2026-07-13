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
	ProductCategoryRepository interface {
		FindProductCategories(
			ctx context.Context,
			request *models.ProductCategoryFilter,
		) ([]*models.ProductCategory, int64, int64, error)
		FindProductCategory(
			ctx context.Context,
			request *models.ProductCategoryFilter,
		) (*models.ProductCategory, error)
		CreateProductCategory(
			ctx context.Context,
			tx pgx.Tx,
			request *models.ProductCategory,
		) (*models.ProductCategory, error)
		UpdateProductCategory(
			ctx context.Context,
			tx pgx.Tx,
			request *models.ProductCategory,
		) (*models.ProductCategory, error)
	}

	productCategoryRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewProductCategoryRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) ProductCategoryRepository {
	return &productCategoryRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *productCategoryRepository) FindProductCategories(
	ctx context.Context,
	request *models.ProductCategoryFilter,
) ([]*models.ProductCategory, int64, int64, error) {
	ctxt := "ProductCategoryRepository-FindProductCategories"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM product_categories %s", condition)
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
		, description
		, photo
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
	var response []*models.ProductCategory
	for rows.Next() {
		var productCategory models.ProductCategory
		if err = rows.Scan(
			&productCategory.ID,
			&productCategory.RecordStatus,
			&productCategory.Name,
			&productCategory.Description,
			&productCategory.Photo,
			&productCategory.CreatedBy,
			&productCategory.CreatedAt,
			&productCategory.UpdatedBy,
			&productCategory.UpdatedAt,
			&productCategory.DeletedBy,
			&productCategory.DeletedAt,
			&productCategory.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		productCategory.RowNo = offset
		response = append(response, &productCategory)
	}
	return response, total, pages, nil
}

func (q *productCategoryRepository) FindProductCategory(
	ctx context.Context,
	request *models.ProductCategoryFilter,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryRepository-FindProductCategory"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, record_status
			, name
			, description
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM product_categories %s`,
		condition,
	)
	query := sb.String()
	var response models.ProductCategory
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Description,
		&response.Photo,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrProductCategoryNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *productCategoryRepository) CreateProductCategory(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductCategory,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryRepository-CreateProductCategory"
	var response models.ProductCategory
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO product_categories (
			record_status
			, name
			, description
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
			, record_status
			, name
			, description
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.Description,
		request.Photo,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Description,
		&response.Photo,
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
			pgxErr.ConstraintName == "product_categories_unique_name_idx" {
			err = models.ErrUniqueProductCategoryNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *productCategoryRepository) UpdateProductCategory(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductCategory,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryRepository-UpdateProductCategory"
	var response models.ProductCategory
	if err := tx.QueryRow(
		ctx,
		`UPDATE product_categories SET
			record_status = $1
			, name = $2
			, description = $3
			, photo = $4
			, updated_by = $5
			, updated_at = $6
			, deleted_by = $7
			, deleted_at = $8
			, deletion_reason = $9
		WHERE id = $10
		RETURNING id
			, record_status
			, name
			, description
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.Description,
		request.Photo,
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
		&response.Description,
		&response.Photo,
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
			pgxErr.ConstraintName == "product_categories_unique_name_idx" {
			err = models.ErrUniqueProductCategoryNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *productCategoryRepository) buildCondition(request *models.ProductCategoryFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.ProductCategoryIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, productCategoryID := range request.ProductCategoryIDs {
			params = append(params, productCategoryID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedProductCategoryIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedProductCategoryID := range request.ExcludedProductCategoryIDs {
			params = append(params, excludedProductCategoryID)
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
		_, _ = sb.WriteString("LOWER(name) IN (")
		for i, name := range request.Names {
			params = append(params, strings.ToLower(name))
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
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
			_, _ = sb.WriteString(" ")
			_, _ = sb.WriteString(cond)
		}
		condition = sb.String()
	}
	return
}
