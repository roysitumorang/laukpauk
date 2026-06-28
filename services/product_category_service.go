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
	ProductCategoryService interface {
		FindProductCategories(
			ctx context.Context,
			request *models.ProductCategoryFilter,
		) ([]*models.ProductCategory, *models.Pagination, error)
		FindProductCategory(
			ctx context.Context,
			request *models.ProductCategoryFilter,
		) (*models.ProductCategory, error)
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

	productCategoryService struct {
		productCategoryRepository repositories.ProductCategoryRepository
	}
)

func NewProductCategoryService(
	productCategoryRepository repositories.ProductCategoryRepository,
) ProductCategoryService {
	return new(productCategoryService{
		productCategoryRepository: productCategoryRepository,
	})
}

func (q *productCategoryService) FindProductCategories(
	ctx context.Context,
	request *models.ProductCategoryFilter,
) ([]*models.ProductCategory, *models.Pagination, error) {
	ctxt := "ProductCategoryService-FindProductCategories"
	productCategories, total, pages, err := q.productCategoryRepository.FindProductCategories(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return nil, nil, err
	}
	n := len(productCategories)
	rows := make([]*models.ProductCategory, n)
	if n > 0 {
		copy(rows, productCategories)
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

func (q *productCategoryService) FindProductCategory(
	ctx context.Context,
	request *models.ProductCategoryFilter,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryService-FindProductCategory"
	response, err := q.productCategoryRepository.FindProductCategory(ctx, request)
	if err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
	}
	return response, err
}

func (q *productCategoryService) CreateProductCategory(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductCategory,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryService-CreateProductCategory"
	response, err := q.productCategoryRepository.CreateProductCategory(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateProductCategory")
		return nil, err
	}
	return response, nil
}
func (q *productCategoryService) UpdateProductCategory(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductCategory,
) (*models.ProductCategory, error) {
	ctxt := "ProductCategoryService-UpdateProductCategory"
	response, err := q.productCategoryRepository.UpdateProductCategory(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return nil, err
	}
	return response, nil
}
