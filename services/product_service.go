package services

import (
	"context"
	"errors"
	"log/slog"
	"uuid"

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
	}

	productService struct {
		productRepository         repositories.ProductRepository
		productCategoryRepository repositories.ProductCategoryRepository
	}
)

func NewProductService(
	productRepository repositories.ProductRepository,
	productCategoryRepository repositories.ProductCategoryRepository,
) ProductService {
	return new(productService{
		productRepository:         productRepository,
		productCategoryRepository: productCategoryRepository,
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
		var productCategoryIDs []uuid.UUID
		mapProductCategoryIDs := map[uuid.UUID]struct{}{}
		for _, product := range products {
			if _, ok := mapProductCategoryIDs[product.ProductCategoryID]; !ok {
				productCategoryIDs = append(productCategoryIDs, product.ProductCategoryID)
				mapProductCategoryIDs[product.ProductCategoryID] = struct{}{}
			}
		}
		productCategories, _, _, err := q.productCategoryRepository.FindProductCategories(
			ctx,
			models.NewProductCategoryFilter().
				WithProductCategoryIDs(productCategoryIDs...),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
			return nil, nil, err
		}
		mapProductCategories := map[uuid.UUID]*models.ProductCategory{}
		for _, productCategory := range productCategories {
			mapProductCategories[productCategory.ID] = productCategory
		}
		for i, row := range rows {
			if productCategory, ok := mapProductCategories[row.ProductCategoryID]; ok {
				row.ProductCategory = productCategory
			}
			rows[i] = row
		}
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

func (q *productService) FindProduct(
	ctx context.Context,
	request *models.ProductFilter,
) (*models.Product, error) {
	ctxt := "ProductService-FindProduct"
	response, err := q.productRepository.FindProduct(ctx, request)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		}
		return response, err
	}
	productCategory, err := q.productCategoryRepository.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithProductCategoryIDs(response.ProductCategoryID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		return response, err
	}
	if productCategory != nil {
		response.ProductCategory = productCategory
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
