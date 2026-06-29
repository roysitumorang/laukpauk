package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
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
	return new(productVariantRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *productVariantRepository) FindProductVariants(
	ctx context.Context,
	request *models.ProductVariantFilter,
) ([]*models.ProductVariant, int64, int64, error) {
	ctxt := "ProductVariantRepository-FindProductVariants"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.ProductIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("product_id IN (")
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
	if len(request.ProductVariantIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, productVariantID := range request.ProductVariantIDs {
			params = append(params, productVariantID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
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
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
		FROM product_variants`,
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
		, product_id
		, measure_unit
		, price
		, price_type
		, price_variant_percentage
		, discount_type
		, discount_rate
		, discount_minimum_quantity
		, description
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
	var response []*models.ProductVariant
	for rows.Next() {
		var productVariant models.ProductVariant
		if err = rows.Scan(
			&productVariant.RowNo,
			&productVariant.ID,
			&productVariant.RecordStatus,
			&productVariant.ProductID,
			&productVariant.MeasureUnit,
			&productVariant.Price,
			&productVariant.PriceType,
			&productVariant.PriceVariantPercentage,
			&productVariant.DiscountType,
			&productVariant.DiscountRate,
			&productVariant.DiscountMinimumQuantity,
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
		response = append(response, &productVariant)
	}
	return response, total, pages, nil
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
		`INSERT INTO product_variants (
			record_status
			, product_id
			, measure_unit
			, price
			, price_type
			, price_variant_percentage
			, discount_type
			, discount_rate
			, discount_minimum_quantity
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
			, record_status
			, product_id
			, measure_unit
			, price
			, price_type
			, price_variant_percentage
			, discount_type
			, discount_rate
			, discount_minimum_quantity
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.ProductID,
		request.MeasureUnit,
		request.Price,
		request.PriceType,
		request.PriceVariantPercentage,
		request.DiscountType,
		request.DiscountRate,
		request.DiscountMinimumQuantity,
		request.Description,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ProductID,
		&response.MeasureUnit,
		&response.Price,
		&response.PriceType,
		&response.PriceVariantPercentage,
		&response.DiscountType,
		&response.DiscountRate,
		&response.DiscountMinimumQuantity,
		&response.Description,
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

func (q *productVariantRepository) UpdateProductVariant(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductVariant,
) (*models.ProductVariant, error) {
	ctxt := "ProductVariantRepository-UpdateProductVariant"
	var response models.ProductVariant
	if err := tx.QueryRow(
		ctx,
		`UPDATE product_variants SET
			record_status = $1
			, measure_unit = $2
			, price = $3
			, price_type = $4
			, price_variant_percentage = $5
			, discount_type = $6
			, discount_rate = $7
			, discount_minimum_quantity = $8
			, description = $9
			, updated_by = $10
			, updated_at = $11
			, deleted_by = $12
			, deleted_at = $13
			, deletion_reason = $14
		WHERE id = $15
		RETURNING id
			, record_status
			, product_id
			, measure_unit
			, price
			, price_type
			, price_variant_percentage
			, discount_type
			, discount_rate
			, discount_minimum_quantity
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.MeasureUnit,
		request.Price,
		request.PriceType,
		request.PriceVariantPercentage,
		request.DiscountType,
		request.DiscountRate,
		request.DiscountMinimumQuantity,
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
		&response.MeasureUnit,
		&response.Price,
		&response.PriceType,
		&response.PriceVariantPercentage,
		&response.DiscountType,
		&response.DiscountRate,
		&response.DiscountMinimumQuantity,
		&response.Description,
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
