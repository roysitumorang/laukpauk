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

type ProductService interface {
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
	SaveProduct(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Product,
	) (*models.Product, error)
	IncreaseProductHits(
		ctx context.Context,
		tx pgx.Tx,
		request *models.ProductFilter,
	) (*models.Product, error)
}

type productService struct {
	productRepository         repositories.ProductRepository
	productCategoryRepository repositories.ProductCategoryRepository
}

func NewProductService(
	productRepository repositories.ProductRepository,
	productCategoryRepository repositories.ProductCategoryRepository,
) ProductService {
	return &productService{
		productRepository:         productRepository,
		productCategoryRepository: productCategoryRepository,
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
		mapProductCategoryNames := map[uuid.UUID]string{}
		for _, productCategory := range productCategories {
			mapProductCategoryNames[productCategory.ID] = productCategory.Name
		}
		for i, product := range products {
			product.ProductCategoryName = mapProductCategoryNames[product.ProductCategoryID]
			products[i] = product
		}
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
		return nil, err
	}
	productCategory, err := q.productCategoryRepository.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithProductCategoryIDs(response.ProductCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
		return nil, err
	}
	if productCategory != nil {
		response.ProductCategoryName = productCategory.Name
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
	productCategory, err := q.productCategoryRepository.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithProductCategoryIDs(response.ProductCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
		return nil, err
	}
	if productCategory != nil {
		response.ProductCategoryName = productCategory.Name
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
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		}
		return nil, err
	}
	productCategory, err := q.productCategoryRepository.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithProductCategoryIDs(response.ProductCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
		return nil, err
	}
	if productCategory != nil {
		response.ProductCategoryName = productCategory.Name
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
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProduct")
		}
		return nil, err
	}
	productCategory, err := q.productCategoryRepository.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithProductCategoryIDs(response.ProductCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
		return nil, err
	}
	if productCategory != nil {
		response.ProductCategoryName = productCategory.Name
	}
	return response, err
}

func (q *productService) IncreaseProductHits(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductFilter,
) (*models.Product, error) {
	ctxt := "ProductService-IncreaseProductHits"
	response, err := q.productRepository.IncreaseProductHits(ctx, tx, request)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrIncreaseProductHits")
		}
		return nil, err
	}
	productCategory, err := q.productCategoryRepository.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithProductCategoryIDs(response.ProductCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
		return nil, err
	}
	if productCategory != nil {
		response.ProductCategoryName = productCategory.Name
	}
	return response, err
}
