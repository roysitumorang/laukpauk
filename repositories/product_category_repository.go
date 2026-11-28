package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type ProductCategoryRepository interface {
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
		tx *sql.Tx,
		request *models.ProductCategory,
	) (*models.ProductCategory, error)
	UpdateProductCategory(
		ctx context.Context,
		tx *sql.Tx,
		request *models.ProductCategory,
	) (*models.ProductCategory, error)
	SaveProductCategory(
		ctx context.Context,
		tx *sql.Tx,
		request *models.ProductCategory,
	) (*models.ProductCategory, error)
}

type productCategoryRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewProductCategoryRepository(
	dbRead,
	dbWrite *sql.DB,
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
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM product_categories %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRowContext(ctx, query, helper.SQLArgs(params...)...).Scan(helper.SQLScan(&total)...)
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
		_, _ = fmt.Fprintf(sb, " %s %s", sortExpression(item.Field), item.Direction)
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
	rows, err := q.dbRead.QueryContext(ctx, sb.String(), helper.SQLArgs(params...)...)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer func() { _ = rows.Close() }()
	var response []*models.ProductCategory
	for rows.Next() {
		var productCategory models.ProductCategory
		if err = rows.Scan(
			helper.SQLScan(
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
			)...,
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
	err := q.dbRead.QueryRowContext(ctx, query, helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
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
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
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
	tx *sql.Tx,
	request *models.ProductCategory,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryRepository-CreateProductCategory"
	var response models.ProductCategory
	if err := tx.QueryRowContext(
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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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
		helper.SQLArgs(
			request.RecordStatus,
			request.Name,
			request.Description,
			request.Photo,
			request.CreatedBy,
			request.CreatedAt,
			request.UpdatedBy,
			request.UpdatedAt,
		)...,
	).Scan(
		helper.SQLScan(
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
		)...,
	); err != nil {
		if isUniqueViolation(err, "product_categories_unique_name_idx") {
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
	tx *sql.Tx,
	request *models.ProductCategory,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryRepository-UpdateProductCategory"
	var response models.ProductCategory
	err := tx.QueryRowContext(
		ctx,
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
		FROM product_categories
		WHERE id = ?
		FOR UPDATE`,
		helper.SQLArgs(request.ID)...,
	).Scan(
		helper.SQLScan(
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
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrProductCategoryNotFound
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
	_, _ = sb.WriteString("UPDATE product_categories SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" record_status = ?")
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" name = ?")
	}
	if response.Description != request.Description {
		params = append(params, request.Description)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" description = ?")
	}
	if response.Photo != request.Photo {
		params = append(params, request.Photo)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" photo = ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	if response.UpdatedBy.Compare(request.UpdatedBy) != 0 {
		params = append(params, request.UpdatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" updated_by = ?")
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" updated_at = ?")
	}
	if (response.DeletedBy == nil && request.DeletedBy != nil) ||
		(response.DeletedBy != nil && request.DeletedBy == nil) ||
		(response.DeletedBy != nil && request.DeletedBy != nil &&
			response.DeletedBy.Compare(*request.DeletedBy) != 0) {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deleted_by = ?")
	}
	if (response.DeletedAt == nil && request.DeletedAt != nil) ||
		(response.DeletedAt != nil && request.DeletedAt == nil) ||
		(response.DeletedAt != nil && request.DeletedAt != nil &&
			!response.DeletedAt.Equal(*request.DeletedAt)) {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deleted_at = ?")
	}
	if (response.DeletionReason == nil && request.DeletionReason != nil) ||
		(response.DeletionReason != nil && request.DeletionReason == nil) ||
		(response.DeletionReason != nil && request.DeletionReason != nil &&
			*response.DeletionReason != *request.DeletionReason) {
		params = append(params, request.DeletionReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deletion_reason = ?")
	}
	params = append(params, request.ID)
	_, _ = sb.WriteString(` WHERE id = ?
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
			, deletion_reason`)
	if err = tx.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
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
		)...,
	); err != nil {
		if isUniqueViolation(err, "product_categories_unique_name_idx") {
			err = models.ErrUniqueProductCategoryNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *productCategoryRepository) SaveProductCategory(
	ctx context.Context,
	tx *sql.Tx,
	request *models.ProductCategory,
) (*models.ProductCategory, error) {
	if request.New() {
		return q.CreateProductCategory(ctx, tx, request)
	}
	return q.UpdateProductCategory(ctx, tx, request)
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
			_, _ = sb.WriteString("?")
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
			_, _ = sb.WriteString("?")
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
			_, _ = sb.WriteString("?")
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
			_, _ = sb.WriteString("?")
		}
		placeholders := sb.String()
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(name) IN (%s)", placeholders)
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		sb.Reset()
		_, _ = sb.WriteString("LOWER(name) LIKE ?")
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
