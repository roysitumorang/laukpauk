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
	ProductRepository interface {
		FindProducts(
			ctx context.Context,
			request *models.ProductFilter,
		) ([]*models.Product, int64, int64, error)
		FindProduct(
			ctx context.Context,
			request *models.ProductFilter,
		) (*models.Product, error)
		CreateProduct(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Product,
		) (*models.Product, error)
		UpdateProduct(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Product,
		) (*models.Product, error)
	}

	productRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewProductRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) ProductRepository {
	return new(productRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *productRepository) FindProducts(
	ctx context.Context,
	request *models.ProductFilter,
) ([]*models.Product, int64, int64, error) {
	ctxt := "ProductRepository-FindProducts"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM products p JOIN product_categories c ON p.product_category_id = c.id %s", condition)
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
		`p.id
		, p.record_status
		, p.product_category_id
		, c.name
		, p.product_type
		, p.name
		, p.photo
		, p.description
		, p.ready_stock
		, p.pinned
		, p.hits
		, p.created_by
		, p.created_at
		, p.updated_by
		, p.updated_at
		, p.deleted_by
		, p.deleted_at
		, p.deletion_reason`,
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
		_, _ = fmt.Fprintf(&sb, " p.%s %s", item.Field, item.Direction)
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
	var response []*models.Product
	for rows.Next() {
		var product models.Product
		if err = rows.Scan(
			&product.ID,
			&product.RecordStatus,
			&product.ProductCategoryID,
			&product.ProductCategoryName,
			&product.ProductType,
			&product.Name,
			&product.Photo,
			&product.Description,
			&product.ReadyStock,
			&product.Pinned,
			&product.Hits,
			&product.CreatedBy,
			&product.CreatedAt,
			&product.UpdatedBy,
			&product.UpdatedAt,
			&product.DeletedBy,
			&product.DeletedAt,
			&product.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		product.RowNo = offset
		response = append(response, (&product).Populate())
	}
	return response, total, pages, nil
}

func (q *productRepository) FindProduct(
	ctx context.Context,
	request *models.ProductFilter,
) (*models.Product, error) {
	ctxt := "ProductRepository-FindProduct"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			p.id
			, p.record_status
			, p.product_category_id
			, c.name
			, p.product_type
			, p.name
			, p.photo
			, p.description
			, p.ready_stock
			, p.pinned
			, p.hits
			, p.created_by
			, p.created_at
			, p.updated_by
			, p.updated_at
			, p.deleted_by
			, p.deleted_at
			, p.deletion_reason
		FROM products p
		JOIN product_categories c ON p.product_category_id = c.id %s`,
		condition,
	)
	query := sb.String()
	var response models.Product
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductCategoryID,
		&response.ProductCategoryName,
		&response.ProductType,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.ReadyStock,
		&response.Pinned,
		&response.Hits,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrProductNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *productRepository) CreateProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	ctxt := "ProductRepository-CreateProduct"
	var response models.Product
	if err := tx.QueryRow(
		ctx,
		`WITH p AS (
			INSERT INTO products (
				record_status
				, product_category_id
				, product_type
				, name
				, photo
				, description
				, ready_stock
				, pinned
				, hits
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING *
		)
		SELECT
			p.id
			, p.record_status
			, p.product_category_id
			, c.name
			, p.product_type
			, p.name
			, p.photo
			, p.description
			, p.ready_stock
			, p.pinned
			, p.hits
			, p.created_by
			, p.created_at
			, p.updated_by
			, p.updated_at
			, p.deleted_by
			, p.deleted_at
			, p.deletion_reason
		FROM p
		JOIN product_categories c ON p.product_category_id = c.id`,
		request.RecordStatus,
		request.ProductCategoryID,
		request.ProductType,
		request.Name,
		request.Photo,
		request.Description,
		request.ReadyStock,
		request.Pinned,
		request.Hits,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductCategoryID,
		&response.ProductCategoryName,
		&response.ProductType,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.ReadyStock,
		&response.Pinned,
		&response.Hits,
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
			pgxErr.ConstraintName == "products_unique_name_idx" {
			err = models.ErrUniqueProductNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *productRepository) UpdateProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	ctxt := "ProductRepository-UpdateProduct"
	var response models.Product
	if err := tx.QueryRow(
		ctx,
		`WITH p AS (
			UPDATE products SET
				record_status = $1
				, product_category_id = $2
				, product_type = $3
				, name = $4
				, photo = $5
				, description = $6
				, ready_stock = $7
				, pinned = $8
				, hits = $9
				, updated_by = $10
				, updated_at = $11
				, deleted_by = $12
				, deleted_at = $13
				, deletion_reason = $14
			WHERE id = $15
			RETURNING *
		)
		SELECT
			p.id
			, p.record_status
			, p.product_category_id
			, c.name
			, p.product_type
			, p.name
			, p.photo
			, p.description
			, p.ready_stock
			, p.pinned
			, p.hits
			, p.created_by
			, p.created_at
			, p.updated_by
			, p.updated_at
			, p.deleted_by
			, p.deleted_at
			, p.deletion_reason
		FROM p
		JOIN product_categories c ON p.product_category_id = c.id`,
		request.RecordStatus,
		request.ProductCategoryID,
		request.ProductType,
		request.Name,
		request.Photo,
		request.Description,
		request.ReadyStock,
		request.Pinned,
		request.Hits,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductCategoryID,
		&response.ProductCategoryName,
		&response.ProductType,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.ReadyStock,
		&response.Pinned,
		&response.Hits,
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
			pgxErr.ConstraintName == "products_unique_name_idx" {
			err = models.ErrUniqueProductNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *productRepository) buildCondition(request *models.ProductFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.ProductCategoryIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.product_category_id IN (")
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
	if len(request.ProductIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.id IN (")
		for i, productID := range request.ProductIDs {
			params = append(params, productID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedProductIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.id NOT IN (")
		for i, excludedProductID := range request.ExcludedProductIDs {
			params = append(params, excludedProductID)
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
		_, _ = sb.WriteString("p.record_status IN (")
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
	if len(request.ProductTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.product_type IN (")
		for i, productType := range request.ProductTypes {
			params = append(params, productType)
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
		_, _ = sb.WriteString("LOWER(p.name) IN (")
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
	if request.ReadyStock != nil {
		if *request.ReadyStock {
			conds = append(conds, "p.ready_stock")
		} else {
			conds = append(conds, "NOT p.ready_stock")
		}
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(p.name) LIKE $%d", len(params))
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
