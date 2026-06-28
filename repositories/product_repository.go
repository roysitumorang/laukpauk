package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
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
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.ProductCategoryIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("product_category_id IN (")
		for i, productCategoryID := range request.ProductCategoryIDs {
			params = append(params, productCategoryID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ProductIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, productID := range request.ProductIDs {
			params = append(params, productID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedProductIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedProductID := range request.ExcludedProductIDs {
			params = append(params, excludedProductID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecordStatus) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("record_status IN (")
		for i, recordStatus := range request.RecordStatus {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ProductTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("product_type IN (")
		for i, productType := range request.ProductTypes {
			params = append(params, productType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
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
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.StockStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("stock_status IN (")
		for i, stockStatus := range request.StockStatuses {
			params = append(params, stockStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = sb.WriteString("%%")
		_, _ = sb.WriteString(strings.ToLower(request.Keyword))
		_, _ = sb.WriteString("%%")
		params = append(params, sb.String())
		n := strconv.Itoa(len(params))
		sb.Reset()
		_, _ = sb.WriteString("LOWER(name) LIKE $")
		_, _ = sb.WriteString(n)
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
		FROM products`,
	)
	if len(conds) > 0 {
		_, _ = sb.WriteString(" WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = sb.WriteString(" ")
			_, _ = sb.WriteString(cond)
		}
	}
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
		`ROW_NUMBER() OVER (ORDER BY id DESC) AS row_no
		, id
		, record_status
		, product_category_id
		, product_type
		, name
		, photo
		, description
		, stock_status
		, pinned
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
	pages := int64(1)
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
		offset := request.Page * request.Limit
		_, _ = sb.WriteString(" LIMIT ")
		_, _ = sb.WriteString(strconv.FormatInt(request.Limit, 10))
		_, _ = sb.WriteString(" OFFSET ")
		_, _ = sb.WriteString(strconv.FormatInt(offset, 10))
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
			&product.RowNo,
			&product.ID,
			&product.RecordStatus,
			&product.ProductCategoryID,
			&product.ProductType,
			&product.Name,
			&product.Photo,
			&product.Description,
			&product.StockStatus,
			&product.Pinned,
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
		response = append(response, &product)
	}
	return response, total, pages, nil
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
		`INSERT INTO products (
			record_status
			, product_category_id
			, product_type
			, name
			, photo
			, description
			, stock_status
			, pinned
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
			, record_status
			, product_category_id
			, product_type
			, name
			, photo
			, description
			, stock_status
			, pinned
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.ProductCategoryID,
		request.ProductType,
		request.Name,
		request.Photo,
		request.Description,
		request.StockStatus,
		request.Pinned,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductCategoryID,
		&response.ProductType,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.StockStatus,
		&response.Pinned,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
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
		`UPDATE products SET
			record_status = $1
			, product_category_id = $2
			, product_type = $3
			, name = $4
			, photo = $5
			, description = $6
			, stock_status = $7
			, pinned = $8
			, updated_by = $9
			, updated_at = $10
			, deleted_by = $11
			, deleted_at = $12
			, deletion_reason = $13
		WHERE id = $14
		RETURNING id
			, record_status
			, product_category_id
			, product_type
			, name
			, photo
			, description
			, stock_status
			, pinned
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.ProductCategoryID,
		request.ProductType,
		request.Name,
		request.Photo,
		request.Description,
		request.StockStatus,
		request.Pinned,
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
		&response.ProductType,
		&response.Name,
		&response.Photo,
		&response.Description,
		&response.StockStatus,
		&response.Pinned,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "products_unique_name_idx" {
			err = models.ErrUniqueProductNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return new(response), nil
}
