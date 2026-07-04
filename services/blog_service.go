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

type (
	BlogService interface {
		FindBlogs(
			ctx context.Context,
			request *models.BlogFilter,
		) ([]*models.Blog, *models.Pagination, error)
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

	blogService struct {
		blogRepository repositories.BlogRepository
	}
)

func NewBlogService(
	blogRepository repositories.BlogRepository,
) BlogService {
	return new(blogService{
		blogRepository: blogRepository,
	})
}

func (q *blogService) FindBlogs(
	ctx context.Context,
	request *models.BlogFilter,
) ([]*models.Blog, *models.Pagination, error) {
	ctxt := "BlogService-FindBlogs"
	blogs, total, pages, err := q.blogRepository.FindBlogs(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return nil, nil, err
	}
	n := len(blogs)
	rows := make([]*models.Blog, n)
	if n > 0 {
		copy(rows, blogs)
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

func (q *blogService) FindBlog(
	ctx context.Context,
	request *models.BlogFilter,
) (*models.Blog, error) {
	ctxt := "BlogService-FindBlog"
	response, err := q.blogRepository.FindBlog(ctx, request)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
	}
	return response, err
}

func (q *blogService) CreateBlog(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Blog,
) (*models.Blog, error) {
	ctxt := "BlogService-CreateBlog"
	response, err := q.blogRepository.CreateBlog(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlog")
		return nil, err
	}
	return response, nil
}

func (q *blogService) UpdateBlog(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Blog,
) (*models.Blog, error) {
	ctxt := "BlogService-UpdateBlog"
	response, err := q.blogRepository.UpdateBlog(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlog")
		return nil, err
	}
	return response, nil
}
