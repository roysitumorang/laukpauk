package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	ArticleRepository interface {
		FindArticles(
			ctx context.Context,
			request *models.ArticleFilter,
		) ([]*models.Article, int64, int64, error)
		FindArticle(
			ctx context.Context,
			request *models.ArticleFilter,
		) (*models.Article, error)
		CreateArticle(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Article,
		) (*models.Article, error)
		UpdateArticle(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Article,
		) (*models.Article, error)
	}

	articleRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewArticleRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) ArticleRepository {
	return new(articleRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *articleRepository) FindArticles(
	ctx context.Context,
	request *models.ArticleFilter,
) ([]*models.Article, int64, int64, error) {
	ctxt := "ArticleRepository-FindArticles"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM articles %s", condition)
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
		, photo
		, link
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
		_, _ = fmt.Fprintf(&sb, " %s %s", item.Field, item.Direction)
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
	var response []*models.Article
	for rows.Next() {
		var article models.Article
		if err = rows.Scan(
			&article.ID,
			&article.RecordStatus,
			&article.Subject,
			&article.Body,
			&article.Photo,
			&article.Link,
			&article.CreatedBy,
			&article.CreatedAt,
			&article.UpdatedBy,
			&article.UpdatedAt,
			&article.DeletedBy,
			&article.DeletedAt,
			&article.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		article.RowNo = offset
		response = append(response, &article)
	}
	return response, total, pages, nil
}

func (q *articleRepository) FindArticle(
	ctx context.Context,
	request *models.ArticleFilter,
) (*models.Article, error) {
	ctxt := "ArticleRepository-FindArticle"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, record_status
			, subject
			, body
			, photo
			, link
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM articles %s`,
		condition,
	)
	query := sb.String()
	var response models.Article
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.Link,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrArticleNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}

func (q *articleRepository) CreateArticle(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Article,
) (*models.Article, error) {
	ctxt := "ArticleRepository-CreateArticle"
	var response models.Article
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO articles (
			record_status
			, subject
			, body
			, photo
			, link
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
			, record_status
			, subject
			, body
			, photo
			, link
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
		request.Photo,
		request.Link,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.Link,
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
	return new(response), nil
}

func (q *articleRepository) UpdateArticle(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Article,
) (*models.Article, error) {
	ctxt := "ArticleRepository-UpdateArticle"
	var response models.Article
	if err := tx.QueryRow(
		ctx,
		`UPDATE articles SET
			record_status = $1
			, subject = $2
			, body = $3
			, photo = $4
			, link = $5
			, updated_by = $6
			, updated_at = $7
			, deleted_by = $8
			, deleted_at = $9
			, deletion_reason = $10
		WHERE id = $11
		RETURNING id
			, record_status
			, subject
			, body
			, photo
			, link
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
		request.Photo,
		request.Link,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.Link,
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
	return new(response), nil
}

func (q *articleRepository) buildCondition(request *models.ArticleFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.ArticleIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, articleID := range request.ArticleIDs {
			params = append(params, articleID)
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
