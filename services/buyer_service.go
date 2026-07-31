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
	BuyerService interface {
		FindBuyers(
			ctx context.Context,
			request *models.BuyerFilter,
		) ([]*models.Buyer, *models.Pagination, error)
		FindBuyer(
			ctx context.Context,
			request *models.BuyerFilter,
		) (*models.Buyer, error)
		CreateBuyer(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Buyer,
		) (*models.Buyer, error)
		UpdateBuyer(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Buyer,
		) (*models.Buyer, error)
	}

	buyerService struct {
		buyerRepository repositories.BuyerRepository
	}
)

func NewBuyerService(
	buyerRepository repositories.BuyerRepository,
) BuyerService {
	return &buyerService{
		buyerRepository: buyerRepository,
	}
}

func (q *buyerService) FindBuyers(
	ctx context.Context,
	request *models.BuyerFilter,
) ([]*models.Buyer, *models.Pagination, error) {
	ctxt := "BuyerService-FindBuyers"
	buyers, total, pages, err := q.buyerRepository.FindBuyers(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return nil, nil, err
	}
	n := len(buyers)
	rows := make([]*models.Buyer, n)
	if n > 0 {
		copy(rows, buyers)
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

func (q *buyerService) FindBuyer(
	ctx context.Context,
	request *models.BuyerFilter,
) (*models.Buyer, error) {
	ctxt := "BuyerService-FindBuyer"
	response, err := q.buyerRepository.FindBuyer(ctx, request)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
	}
	return response, err
}

func (q *buyerService) CreateBuyer(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Buyer,
) (*models.Buyer, error) {
	ctxt := "BuyerService-CreateBuyer"
	response, err := q.buyerRepository.CreateBuyer(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBuyer")
	}
	return response, err
}

func (q *buyerService) UpdateBuyer(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Buyer,
) (*models.Buyer, error) {
	ctxt := "BuyerService-UpdateBuyer"
	response, err := q.buyerRepository.UpdateBuyer(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
	}
	return response, err
}
