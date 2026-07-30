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

type (
	BlogRepository interface {
		FindBlogs(
			ctx context.Context,
			request *models.BlogFilter,
		) ([]*models.Blog, int64, int64, error)
		FindBlog(
			ctx context.Context,
			request *models.BlogFilter,
		) (*models.Blog, error)
		CreateBlog(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Blog,
		) (*models.Blog, error)
		UpdateBlog(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Blog,
		) (*models.Blog, error)
	}

	blogRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewBlogRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) BlogRepository {
	return &blogRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *blogRepository) FindBlogs(
	ctx context.Context,
	request *models.BlogFilter,
) ([]*models.Blog, int64, int64, error) {
	ctxt := "BlogRepository-FindBlogs"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM blogs %s", condition)
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
		, subject
		, body
		, slug
		, deleted_slug
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
	var response []*models.Blog
	for rows.Next() {
		var blog models.Blog
		if err = rows.Scan(
			&blog.ID,
			&blog.RecordStatus,
			&blog.Subject,
			&blog.Body,
			&blog.Slug,
			&blog.DeletedSlug,
			&blog.Photo,
			&blog.CreatedBy,
			&blog.CreatedAt,
			&blog.UpdatedBy,
			&blog.UpdatedAt,
			&blog.DeletedBy,
			&blog.DeletedAt,
			&blog.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		blog.RowNo = offset
		response = append(response, &blog)
	}
	return response, total, pages, nil
}

func (q *blogRepository) FindBlog(
	ctx context.Context,
	request *models.BlogFilter,
) (*models.Blog, error) {
	ctxt := "BlogRepository-FindBlog"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`SELECT
			id
			, record_status
			, subject
			, body
			, slug
			, deleted_slug
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM blogs %s`,
		condition,
	)
	query := sb.String()
	var response models.Blog
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Subject,
		&response.Body,
		&response.Slug,
		&response.DeletedSlug,
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
		return nil, models.ErrBlogNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *blogRepository) CreateBlog(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Blog,
) (*models.Blog, error) {
	ctxt := "BlogRepository-CreateBlog"
	var response models.Blog
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO blogs (
			record_status
			, subject
			, body
			, slug
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
			, record_status
			, subject
			, body
			, slug
			, deleted_slug
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Subject,
		request.Body,
		request.Slug,
		request.Photo,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Subject,
		&response.Body,
		&response.Slug,
		&response.DeletedSlug,
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
			pgxErr.ConstraintName == "blogs_slug_key" {
			err = models.ErrUniqueBlogSlugViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *blogRepository) UpdateBlog(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Blog,
) (*models.Blog, error) {
	ctxt := "BlogRepository-UpdateBlog"
	var response models.Blog
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, subject
			, body
			, slug
			, deleted_slug
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM blogs
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Subject,
		&response.Body,
		&response.Slug,
		&response.DeletedSlug,
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
		return nil, models.ErrBlogNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	var params []any
	var n int
	_, _ = sb.WriteString("UPDATE blogs SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if response.Subject != request.Subject {
		params = append(params, request.Subject)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " subject = $%d", n)
	}
	if response.Body != request.Body {
		params = append(params, request.Body)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " body = $%d", n)
	}
	if (response.Slug == nil && request.Slug != nil) ||
		(response.Slug != nil && request.Slug == nil) ||
		(response.Slug != nil && request.Slug != nil && response.Slug != request.Slug) {
		params = append(params, request.Slug)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " slug = $%d", n)
	}
	if (response.DeletedSlug == nil && request.DeletedSlug != nil) ||
		(response.DeletedSlug != nil && request.DeletedSlug == nil) ||
		(response.DeletedSlug != nil && request.DeletedSlug != nil && response.DeletedSlug != request.DeletedSlug) {
		params = append(params, request.Slug)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_slug = $%d", n)
	}
	if response.Photo != request.Photo {
		params = append(params, request.Photo)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " photo = $%d", n)
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
	if response.DeletedBy == nil && request.DeletedBy != nil {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_by = $%d", n)
	}
	if response.DeletedAt == nil && request.DeletedAt != nil {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_at = $%d", n)
	}
	if response.DeletionReason == nil && request.DeletionReason != nil {
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
			, subject
			, body
			, slug
			, deleted_slug
			, photo
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
		&response.Subject,
		&response.Body,
		&response.Slug,
		&response.DeletedSlug,
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
			pgxErr.ConstraintName == "blogs_slug_key" {
			err = models.ErrUniqueBlogSlugViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *blogRepository) buildCondition(request *models.BlogFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.BlogIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, blogID := range request.BlogIDs {
			params = append(params, blogID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedBlogIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedBlogID := range request.ExcludedBlogIDs {
			params = append(params, excludedBlogID)
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
	if len(request.Slugs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("slug IN (")
		for i, slug := range request.Slugs {
			params = append(params, slug)
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
		_, _ = fmt.Fprintf(sb, "LOWER(subject) LIKE $%d", len(params))
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
