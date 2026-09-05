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
	SellerService interface {
		FindSellers(
			ctx context.Context,
			request *models.SellerFilter,
		) ([]*models.Seller, *models.Pagination, error)
		FindSeller(
			ctx context.Context,
			request *models.SellerFilter,
		) (*models.Seller, error)
		CreateSeller(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Seller,
		) (*models.Seller, error)
		UpdateSeller(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Seller,
		) (*models.Seller, error)
	}

	sellerService struct {
		sellerRepository repositories.SellerRepository
	}
)

func NewSellerService(
	sellerRepository repositories.SellerRepository,
) SellerService {
	return &sellerService{
		sellerRepository: sellerRepository,
	}
}

func (q *sellerService) FindSellers(
	ctx context.Context,
	request *models.SellerFilter,
) ([]*models.Seller, *models.Pagination, error) {
	ctxt := "SellerService-FindSellers"
	sellers, total, pages, err := q.sellerRepository.FindSellers(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSellers")
		return nil, nil, err
	}
	n := len(sellers)
	rows := make([]*models.Seller, n)
	if n > 0 {
		copy(rows, sellers)
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

func (q *sellerService) FindSeller(
	ctx context.Context,
	request *models.SellerFilter,
) (*models.Seller, error) {
	ctxt := "SellerService-FindSeller"
	response, err := q.sellerRepository.FindSeller(ctx, request)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
	}
	return response, err
}

func (q *sellerService) CreateSeller(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Seller,
) (*models.Seller, error) {
	ctxt := "SellerService-CreateSeller"
	response, err := q.sellerRepository.CreateSeller(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSeller")
	}
	return response, err
}

func (q *sellerService) UpdateSeller(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Seller,
) (*models.Seller, error) {
	ctxt := "SellerService-UpdateSeller"
	response, err := q.sellerRepository.UpdateSeller(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSeller")
	}
	return response, err
}
