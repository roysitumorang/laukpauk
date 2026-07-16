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
	ProductVariantRepository interface {
		FindProductVariants(
			ctx context.Context,
			request *models.ProductVariantFilter,
		) ([]*models.ProductVariant, int64, int64, error)
		FindProductVariant(
			ctx context.Context,
			request *models.ProductVariantFilter,
		) (*models.ProductVariant, error)
		CreateProductVariant(
			ctx context.Context,
			tx pgx.Tx,
			request *models.ProductVariant,
		) (*models.ProductVariant, error)
		UpdateProductVariant(
			ctx context.Context,
			tx pgx.Tx,
			request *models.ProductVariant,
		) (*models.ProductVariant, error)
	}

	productVariantRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewProductVariantRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) ProductVariantRepository {
	return &productVariantRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *productVariantRepository) FindProductVariants(
	ctx context.Context,
	request *models.ProductVariantFilter,
) ([]*models.ProductVariant, int64, int64, error) {
	ctxt := "ProductVariantRepository-FindProductVariants"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM product_variants v JOIN products p ON v.product_id = p.id %s", condition)
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
		`v.id
		, v.record_status
		, v.product_id
		, p.name
		, p.ready_stock
		, v.stock_unit
		, v.price
		, v.discount_type
		, v.discount_rate
		, v.discount_minimum_quantity
		, v.discount_amount
		, v.sale_price
		, v.description
		, v.created_by
		, v.created_at
		, v.updated_by
		, v.updated_at
		, v.deleted_by
		, v.deleted_at
		, v.deletion_reason`,
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
		_, _ = fmt.Fprintf(&sb, " v.%s %s", item.Field, item.Direction)
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
	var response []*models.ProductVariant
	for rows.Next() {
		var productVariant models.ProductVariant
		if err = rows.Scan(
			&productVariant.ID,
			&productVariant.RecordStatus,
			&productVariant.ProductID,
			&productVariant.ProductName,
			&productVariant.ProductReadyStock,
			&productVariant.StockUnit,
			&productVariant.Price,
			&productVariant.DiscountType,
			&productVariant.DiscountRate,
			&productVariant.DiscountMinimumQuantity,
			&productVariant.DiscountAmount,
			&productVariant.SalePrice,
			&productVariant.Description,
			&productVariant.CreatedBy,
			&productVariant.CreatedAt,
			&productVariant.UpdatedBy,
			&productVariant.UpdatedAt,
			&productVariant.DeletedBy,
			&productVariant.DeletedAt,
			&productVariant.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		productVariant.RowNo = offset
		response = append(response, (&productVariant).Populate())
	}
	return response, total, pages, nil
}

func (q *productVariantRepository) FindProductVariant(
	ctx context.Context,
	request *models.ProductVariantFilter,
) (*models.ProductVariant, error) {
	ctxt := "ProductVariantRepository-FindProductVariant"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			v.id
			, v.record_status
			, v.product_id
			, p.name
			, p.ready_stock
			, v.stock_unit
			, v.price
			, v.discount_type
			, v.discount_rate
			, v.discount_minimum_quantity
			, v.discount_amount
			, v.sale_price
			, v.description
			, v.created_by
			, v.created_at
			, v.updated_by
			, v.updated_at
			, v.deleted_by
			, v.deleted_at
			, v.deletion_reason
		FROM product_variants v
		JOIN products p ON v.product_id = p.id %s`,
		condition,
	)
	query := sb.String()
	var response models.ProductVariant
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductID,
		&response.ProductName,
		&response.ProductReadyStock,
		&response.StockUnit,
		&response.Price,
		&response.DiscountType,
		&response.DiscountRate,
		&response.DiscountMinimumQuantity,
		&response.DiscountAmount,
		&response.SalePrice,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrProductVariantNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *productVariantRepository) CreateProductVariant(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductVariant,
) (*models.ProductVariant, error) {
	ctxt := "ProductVariantRepository-CreateProductVariant"
	var response models.ProductVariant
	if err := tx.QueryRow(
		ctx,
		`WITH v AS (
			INSERT INTO product_variants (
				record_status
				, product_id
				, stock_unit
				, price
				, discount_type
				, discount_rate
				, discount_minimum_quantity
				, discount_amount
				, sale_price
				, description
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING *
		)
		SELECT
			v.id
			, v.record_status
			, v.product_id
			, p.name
			, p.ready_stock
			, v.stock_unit
			, v.price
			, v.discount_type
			, v.discount_rate
			, v.discount_minimum_quantity
			, v.discount_amount
			, v.sale_price
			, v.description
			, v.created_by
			, v.created_at
			, v.updated_by
			, v.updated_at
			, v.deleted_by
			, v.deleted_at
			, v.deletion_reason
		FROM v
		JOIN products p ON v.product_id = p.id`,
		request.RecordStatus,
		request.ProductID,
		request.StockUnit,
		request.Price,
		request.DiscountType,
		request.DiscountRate,
		request.DiscountMinimumQuantity,
		request.DiscountAmount,
		request.SalePrice,
		request.Description,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductID,
		&response.ProductName,
		&response.ProductReadyStock,
		&response.StockUnit,
		&response.Price,
		&response.DiscountType,
		&response.DiscountRate,
		&response.DiscountMinimumQuantity,
		&response.DiscountAmount,
		&response.SalePrice,
		&response.Description,
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
			pgxErr.ConstraintName == "product_variants_product_id_stock_unit_key" {
			err = models.ErrUniqueProductVariantStockUnitViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *productVariantRepository) UpdateProductVariant(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductVariant,
) (*models.ProductVariant, error) {
	ctxt := "ProductVariantRepository-UpdateProductVariant"
	var response models.ProductVariant
	if err := tx.QueryRow(
		ctx,
		`WITH v AS (
			UPDATE product_variants SET
				record_status = $1
				, stock_unit = $2
				, price = $3
				, discount_type = $4
				, discount_rate = $5
				, discount_minimum_quantity = $6
				, discount_amount = $7
				, sale_price = $8
				, description = $9
				, updated_by = $10
				, updated_at = $11
				, deleted_by = $12
				, deleted_at = $13
				, deletion_reason = $14
			WHERE id = $15
			RETURNING *
		)
 		SELECT
			v.id
			, v.record_status
			, v.product_id
			, p.name
			, p.ready_stock
			, v.stock_unit
			, v.price
			, v.discount_type
			, v.discount_rate
			, v.discount_minimum_quantity
			, v.discount_amount
			, v.sale_price
			, v.description
			, v.created_by
			, v.created_at
			, v.updated_by
			, v.updated_at
			, v.deleted_by
			, v.deleted_at
			, v.deletion_reason
		FROM v
		JOIN products p ON v.product_id = p.id`,
		request.RecordStatus,
		request.StockUnit,
		request.Price,
		request.DiscountType,
		request.DiscountRate,
		request.DiscountMinimumQuantity,
		request.DiscountAmount,
		request.SalePrice,
		request.Description,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductID,
		&response.ProductName,
		&response.ProductReadyStock,
		&response.StockUnit,
		&response.Price,
		&response.DiscountType,
		&response.DiscountRate,
		&response.DiscountMinimumQuantity,
		&response.DiscountAmount,
		&response.SalePrice,
		&response.Description,
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
			pgxErr.ConstraintName == "product_variants_product_id_stock_unit_key" {
			err = models.ErrUniqueProductVariantStockUnitViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *productVariantRepository) buildCondition(request *models.ProductVariantFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.ProductIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("v.product_id IN (")
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
	if len(request.ProductVariantIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("v.id IN (")
		for i, productVariantID := range request.ProductVariantIDs {
			params = append(params, productVariantID)
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
		_, _ = sb.WriteString("v.record_status IN (")
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
	if len(request.StockUnits) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("LOWER(v.stock_unit) IN (")
		for i, stockUnit := range request.StockUnits {
			params = append(params, strings.ToLower(stockUnit))
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
