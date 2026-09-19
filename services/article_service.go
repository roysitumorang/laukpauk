package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type ArticleService interface {
	FindArticles(
		ctx context.Context,
		request *models.ArticleFilter,
	) ([]*models.Article, *models.Pagination, error)
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

type articleService struct {
	articleRepository repositories.ArticleRepository
}

func NewArticleService(
	articleRepository repositories.ArticleRepository,
) ArticleService {
	return &articleService{
		articleRepository: articleRepository,
	}
}

func (q *articleService) FindArticles(
	ctx context.Context,
	request *models.ArticleFilter,
) ([]*models.Article, *models.Pagination, error) {
	ctxt := "ArticleService-FindArticles"
	articles, total, pages, err := q.articleRepository.FindArticles(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticles")
		return nil, nil, err
	}
	n := len(articles)
	rows := make([]*models.Article, n)
	if n > 0 {
		copy(rows, articles)
	}
	pagination, err := models.NewPagination(
		total,
		pages,
		request.Limit,
		request.Page,
		request.PaginationURL,
		request.UrlValues,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewPagination")
		return nil, nil, err
	}
	return rows, pagination, nil
}

func (q *articleService) FindArticle(
	ctx context.Context,
	request *models.ArticleFilter,
) (*models.Article, error) {
	ctxt := "ArticleService-FindArticle"
	response, err := q.articleRepository.FindArticle(ctx, request)
	if err != nil && !errors.Is(err, models.ErrArticleNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticle")
	}
	return response, err
}

func (q *articleService) CreateArticle(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Article,
) (*models.Article, error) {
	ctxt := "ArticleService-CreateArticle"
	response, err := q.articleRepository.CreateArticle(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateArticle")
		return nil, err
	}
	return response, nil
}
func (q *articleService) UpdateArticle(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Article,
) (*models.Article, error) {
	ctxt := "ArticleService-UpdateArticle"
	response, err := q.articleRepository.UpdateArticle(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateArticle")
		return nil, err
	}
	return response, nil
}
