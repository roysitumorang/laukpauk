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
	ProductCategoryRepository interface {
		FindProductCategories(
			ctx context.Context,
			request *models.ProductCategoryFilter,
		) ([]*models.ProductCategory, int64, int64, error)
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
	return new(productCategoryRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *productCategoryRepository) FindProductCategories(
	ctx context.Context,
	request *models.ProductCategoryFilter,
) ([]*models.ProductCategory, int64, int64, error) {
	ctxt := "ProductCategoryRepository-FindProductCategories"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.ProductCategoryIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
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
		FROM product_categories`,
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
	var response []*models.ProductCategory
	for rows.Next() {
		var productCategory models.ProductCategory
		if err = rows.Scan(
			&productCategory.RowNo,
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
		response = append(response, &productCategory)
	}
	return response, total, pages, nil
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
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
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
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "product_categories_unique_name_idx":
				err = models.ErrUniqueProductCategoryNameViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return new(response), nil
}
