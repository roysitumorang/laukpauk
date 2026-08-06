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

type ProductRepository interface {
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
	SaveProduct(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Product,
	) (*models.Product, error)
	IncreaseProductHits(
		ctx context.Context,
		tx pgx.Tx,
		request *models.ProductFilter,
	) (*models.Product, error)
}

type productRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewProductRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) ProductRepository {
	return &productRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *productRepository) FindProducts(
	ctx context.Context,
	request *models.ProductFilter,
) ([]*models.Product, int64, int64, error) {
	ctxt := "ProductRepository-FindProducts"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM products %s", condition)
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
		, product_category_id
		, product_category_name
		, product_type
		, name
		, deleted_name
		, photo
		, description
		, ready_stock
		, pinned
		, hits
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
			&product.DeletedName,
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
			, product_category_id
			, product_category_name
			, product_type
			, name
			, deleted_name
			, photo
			, description
			, ready_stock
			, pinned
			, hits
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM products %s`,
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
		&response.DeletedName,
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
		`INSERT INTO products (
			record_status
			, product_category_id
			, product_category_name
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
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
			, record_status
			, product_category_id
			, product_category_name
			, product_type
			, name
			, deleted_name
			, photo
			, description
			, ready_stock
			, pinned
			, hits
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.ProductCategoryID,
		request.ProductCategoryName,
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
		&response.DeletedName,
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
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, product_category_id
			, product_category_name
			, product_type
			, name
			, deleted_name
			, photo
			, description
			, ready_stock
			, pinned
			, hits
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM products
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductCategoryID,
		&response.ProductCategoryName,
		&response.ProductType,
		&response.Name,
		&response.DeletedName,
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
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE products SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if response.ProductCategoryID.Compare(request.ProductCategoryID) != 0 {
		params = append(params, request.ProductCategoryID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_category_id = $%d", n)
	}
	if response.ProductCategoryName != request.ProductCategoryName {
		params = append(params, request.ProductCategoryName)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_category_name = $%d", n)
	}
	if response.ProductType != request.ProductType {
		params = append(params, request.ProductType)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " product_type = $%d", n)
	}
	if response.Name != request.Name {
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
	if response.ReadyStock != request.ReadyStock {
		params = append(params, request.ReadyStock)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " ready_stock = $%d", n)
	}
	if response.Pinned != request.Pinned {
		params = append(params, request.Pinned)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " pinned = $%d", n)
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
			, product_category_id
			, product_category_name
			, product_type
			, name
			, deleted_name
			, photo
			, description
			, ready_stock
			, pinned
			, hits
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
		&response.ProductCategoryID,
		&response.ProductCategoryName,
		&response.ProductType,
		&response.Name,
		&response.DeletedName,
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

func (q *productRepository) SaveProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	if request.New() {
		return q.CreateProduct(ctx, tx, request)
	}
	return q.UpdateProduct(ctx, tx, request)
}

func (q *productRepository) buildCondition(request *models.ProductFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.ProductCategoryIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("product_category_id IN (")
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
		_, _ = sb.WriteString("id IN (")
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
		_, _ = sb.WriteString("id NOT IN (")
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
	if len(request.ProductTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("product_type IN (")
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
	if request.ReadyStock != nil {
		if *request.ReadyStock {
			conds = append(conds, "ready_stock")
		} else {
			conds = append(conds, "NOT ready_stock")
		}
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

func (q *productRepository) IncreaseProductHits(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductFilter,
) (*models.Product, error) {
	ctxt := "ProductRepository-IncreaseProductHits"
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
			, product_category_id
			, product_category_name
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
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM products
		%s
		FOR NO KEY UPDATE`,
		condition,
	)
	var response models.Product
	err := tx.QueryRow(ctx, sb.String(), params...).Scan(
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
	if err = tx.QueryRow(
		ctx,
		`UPDATE products p SET
			hits = hits + 1
		WHERE id = $1
		RETURNING id
			, record_status
			, product_category_id
			, product_category_name
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
			, deleted_by
			, deleted_at
			, deletion_reason`,
		condition,
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}
