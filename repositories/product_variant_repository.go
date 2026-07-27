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
		, v.base_price
		, v.selling_price
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
			&productVariant.BasePrice,
			&productVariant.SellingPrice,
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
			, v.base_price
			, v.selling_price
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
		&response.BasePrice,
		&response.SellingPrice,
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
				, base_price
				, selling_price
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
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING *
		)
		SELECT
			v.id
			, v.record_status
			, v.product_id
			, p.name
			, p.ready_stock
			, v.stock_unit
			, v.base_price
			, v.selling_price
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
		request.BasePrice,
		request.SellingPrice,
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
		&response.BasePrice,
		&response.SellingPrice,
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
	err := q.dbRead.QueryRow(
		ctx,
		`WITH v AS (
			SELECT
				id
				, record_status
				, product_id
				, stock_unit
				, base_price
				, selling_price
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
				, deleted_by
				, deleted_at
				, deletion_reason
			FROM product_variants
			WHERE id = $1
			FOR UPDATE
		)
		SELECT
			v.id
			, v.record_status
			, v.product_id
			, p.name
			, p.ready_stock
			, v.stock_unit
			, v.base_price
			, v.selling_price
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
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductID,
		&response.ProductName,
		&response.ProductReadyStock,
		&response.StockUnit,
		&response.BasePrice,
		&response.SellingPrice,
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
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	_, _ = sb.WriteString("WITH v AS (UPDATE product_variants SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " record_status = $%d", n)
	}
	if response.StockUnit != request.StockUnit {
		params = append(params, request.StockUnit)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " stock_unit = $%d", n)
	}
	if response.BasePrice != request.BasePrice {
		params = append(params, request.BasePrice)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " base_price = $%d", n)
	}
	if response.SellingPrice != request.SellingPrice {
		params = append(params, request.SellingPrice)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " selling_price = $%d", n)
	}
	if response.DiscountType != request.DiscountType {
		params = append(params, request.DiscountType)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " discount_type = $%d", n)
	}
	if response.DiscountRate != request.DiscountRate {
		params = append(params, request.DiscountRate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " discount_rate = $%d", n)
	}
	if response.DiscountMinimumQuantity != request.DiscountMinimumQuantity {
		params = append(params, request.DiscountMinimumQuantity)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " discount_minimum_quantity = $%d", n)
	}
	if response.DiscountAmount != request.DiscountAmount {
		params = append(params, request.DiscountAmount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " discount_amount = $%d", n)
	}
	if response.SalePrice != request.SalePrice {
		params = append(params, request.SalePrice)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " sale_price = $%d", n)
	}
	if response.Description != request.Description {
		params = append(params, request.Description)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " description = $%d", n)
	}
	if len(params) == 0 {
		return (&response).Populate(), nil
	}
	if response.UpdatedBy.Compare(request.UpdatedBy) != 0 {
		params = append(params, request.UpdatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " updated_by = $%d", n)
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " updated_at = $%d", n)
	}
	if response.DeletedBy == nil && request.DeletedBy != nil {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " deleted_by = $%d", n)
	}
	if response.DeletedAt == nil && request.DeletedAt != nil {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " deleted_at = $%d", n)
	}
	if response.DeletionReason == nil && request.DeletionReason != nil {
		params = append(params, request.DeletionReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " deletion_reason = $%d", n)
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		&sb,
		` WHERE id = $%d RETURNING *)
 		SELECT
			v.id
			, v.record_status
			, v.product_id
			, p.name
			, p.ready_stock
			, v.stock_unit
			, v.base_price
			, v.selling_price
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
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductID,
		&response.ProductName,
		&response.ProductReadyStock,
		&response.StockUnit,
		&response.BasePrice,
		&response.SellingPrice,
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
