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

type ProductService interface {
	FindProducts(
		ctx context.Context,
		request *models.ProductFilter,
	) ([]*models.Product, *models.Pagination, error)
	InfiniteScrollProducts(
		ctx context.Context,
		request *models.ProductFilter,
	) (*models.InfiniteScroll, error)
	FindProduct(
		ctx context.Context,
		request *models.ProductFilter,
	) (*models.Product, error)
	CreateProduct(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Product,
	) (*models.Product, error)
	UpdateProduct(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Product,
	) (*models.Product, error)
	SaveProduct(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Product,
	) (*models.Product, error)
}

type productService struct {
	productRepository repositories.ProductRepository
}

func NewProductService(
	productRepository repositories.ProductRepository,
) ProductService {
	return &productService{
		productRepository: productRepository,
	}
}

func (q *productService) FindProducts(
	ctx context.Context,
	request *models.ProductFilter,
) ([]*models.Product, *models.Pagination, error) {
	ctxt := "ProductService-FindProducts"
	products, total, pages, err := q.productRepository.FindProducts(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return nil, nil, err
	}
	n := len(products)
	rows := make([]*models.Product, n)
	if n > 0 {
		copy(rows, products)
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

func (q *productService) InfiniteScrollProducts(
	ctx context.Context,
	request *models.ProductFilter,
) (*models.InfiniteScroll, error) {
	ctxt := "ProductService-InfiniteScrollProducts"
	products, nextRowNo, err := q.productRepository.InfiniteScrollProducts(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrInfiniteScrollProducts")
		return nil, err
	}
	rows := make([]any, len(products))
	for i := range products {
		rows[i] = products[i]
	}
	infiniteScroll, err := models.NewInfiniteScroll(
		request.Limit,
		nextRowNo,
		request.PaginationURL,
		request.UrlValues,
		rows,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewInfiniteScroll")
		return nil, err
	}
	return infiniteScroll, nil
}

func (q *productService) FindProduct(
	ctx context.Context,
	request *models.ProductFilter,
) (*models.Product, error) {
	ctxt := "ProductService-FindProduct"
	response, err := q.productRepository.FindProduct(ctx, request)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
	}
	return response, err
}

func (q *productService) CreateProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	ctxt := "ProductService-CreateProduct"
	response, err := q.productRepository.CreateProduct(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateProduct")
	}
	return response, err
}

func (q *productService) UpdateProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	ctxt := "ProductService-UpdateProduct"
	response, err := q.productRepository.UpdateProduct(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
	}
	return response, err
}

func (q *productService) SaveProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	ctxt := "ProductService-SaveProduct"
	response, err := q.productRepository.SaveProduct(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProduct")
	}
	return response, err
}
