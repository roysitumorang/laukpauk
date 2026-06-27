package services

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	ProductService interface {
		FindProducts(
			ctx context.Context,
			request *models.ProductFilter,
		) ([]*models.Product, *models.Pagination, error)
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
	}

	productService struct {
		productRepository repositories.ProductRepository
	}
)

func NewProductService(
	productRepository repositories.ProductRepository,
) ProductService {
	return new(productService{
		productRepository: productRepository,
	})
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

func (q *productService) CreateProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	ctxt := "ProductService-CreateProduct"
	response, err := q.productRepository.CreateProduct(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateProduct")
		return nil, err
	}
	return response, nil
}
func (q *productService) UpdateProduct(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Product,
) (*models.Product, error) {
	ctxt := "ProductService-UpdateProduct"
	response, err := q.productRepository.UpdateProduct(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return nil, err
	}
	return response, nil
}
