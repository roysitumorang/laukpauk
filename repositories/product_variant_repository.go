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
		, metric_unit_id
		, measure_min
		, measure_max
		, discount_type
		, discount_amount
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
			&productVariant.MetricUnitID,
			&productVariant.MeasureMin,
			&productVariant.MeasureMax,
			&productVariant.DiscountType,
			&productVariant.DiscountAmount,
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
		`INSERT INTO productVariant_variants (
			record_status
			, product_id
			, metric_unit_id
			, measure_min
			, measure_max
			, discount_type
			, discount_amount
			, discount_minimum_quantity
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
			, record_status
			, product_id
			, metric_unit_id
			, measure_min
			, measure_max
			, discount_type
			, discount_amount
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
		request.MetricUnitID,
		request.MeasureMin,
		request.MeasureMax,
		request.DiscountType,
		request.DiscountAmount,
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
		&response.MetricUnitID,
		&response.MeasureMin,
		&response.MeasureMax,
		&response.DiscountType,
		&response.DiscountAmount,
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
			, measure_min = $2
			, measure_max = $3
			, discount_type = $4
			, discount_amount = $5
			, discount_minimum_quantity = $6
			, description = $7
			, updated_by = $8
			, updated_at = $9
			, deleted_by = $10
			, deleted_at = $11
			, deletion_reason = $12
		WHERE id = $13
		RETURNING id
			, record_status
			, product_id
			, metric_unit_id
			, measure_min
			, measure_max
			, discount_type
			, discount_amount
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
		request.MeasureMin,
		request.MeasureMax,
		request.DiscountType,
		request.DiscountAmount,
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
		&response.MetricUnitID,
		&response.MeasureMin,
		&response.MeasureMax,
		&response.DiscountType,
		&response.DiscountAmount,
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
