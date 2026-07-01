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
	ProductVariantService interface {
		FindProductVariants(
			ctx context.Context,
			request *models.ProductVariantFilter,
		) ([]*models.ProductVariant, *models.Pagination, error)
		FindProductVariant(
			ctx context.Context,
			request *models.ProductVariantFilter,
		) (*models.ProductVariant, error)
		CreateProductVariant(
			ctx context.Context,
			tx pgx.Tx,
			request *models.ProductVariant,
		) (*models.ProductVariant, error)
		UpdateProductVariant(
			ctx context.Context,
			tx pgx.Tx,
			request *models.ProductVariant,
		) (*models.ProductVariant, error)
	}

	productVariantService struct {
		productVariantRepository repositories.ProductVariantRepository
		productRepository        repositories.ProductRepository
	}
)

func NewProductVariantService(
	productVariantRepository repositories.ProductVariantRepository,
	productRepository repositories.ProductRepository,
) ProductVariantService {
	return new(productVariantService{
		productVariantRepository: productVariantRepository,
		productRepository:        productRepository,
	})
}

func (q *productVariantService) FindProductVariants(
	ctx context.Context,
	request *models.ProductVariantFilter,
) ([]*models.ProductVariant, *models.Pagination, error) {
	ctxt := "ProductVariantService-FindProductVariants"
	productVariants, total, pages, err := q.productVariantRepository.FindProductVariants(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return nil, nil, err
	}
	n := len(productVariants)
	rows := make([]*models.ProductVariant, n)
	if n > 0 {
		copy(rows, productVariants)
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

func (q *productVariantService) FindProductVariant(
	ctx context.Context,
	request *models.ProductVariantFilter,
) (*models.ProductVariant, error) {
	ctxt := "ProductVariantService-FindProductVariant"
	response, err := q.productVariantRepository.FindProductVariant(ctx, request)
	if err != nil && !errors.Is(err, models.ErrProductVariantNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
	}
	return response, err
}

func (q *productVariantService) CreateProductVariant(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductVariant,
) (*models.ProductVariant, error) {
	ctxt := "ProductVariantService-CreateProductVariant"
	response, err := q.productVariantRepository.CreateProductVariant(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateProductVariant")
		return nil, err
	}
	return response, nil
}

func (q *productVariantService) UpdateProductVariant(
	ctx context.Context,
	tx pgx.Tx,
	request *models.ProductVariant,
) (*models.ProductVariant, error) {
	ctxt := "ProductVariantService-UpdateProductVariant"
	response, err := q.productVariantRepository.UpdateProductVariant(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
		return nil, err
	}
	return response, nil
}
