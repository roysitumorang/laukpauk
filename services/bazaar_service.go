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
	BazaarService interface {
		FindBazaars(
			ctx context.Context,
			request *models.BazaarFilter,
		) ([]*models.Bazaar, *models.Pagination, error)
		CreateBazaar(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Bazaar,
		) (*models.Bazaar, error)
		UpdateBazaar(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Bazaar,
		) (*models.Bazaar, error)
	}

	bazaarService struct {
		bazaarRepository repositories.BazaarRepository
	}
)

func NewBazaarService(
	bazaarRepository repositories.BazaarRepository,
) BazaarService {
	return new(bazaarService{
		bazaarRepository: bazaarRepository,
	})
}

func (q *bazaarService) FindBazaars(
	ctx context.Context,
	request *models.BazaarFilter,
) ([]*models.Bazaar, *models.Pagination, error) {
	ctxt := "BazaarService-FindBazaars"
	bazaars, total, pages, err := q.bazaarRepository.FindBazaars(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBazaars")
		return nil, nil, err
	}
	n := len(bazaars)
	rows := make([]*models.Bazaar, n)
	if n > 0 {
		copy(rows, bazaars)
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

func (q *bazaarService) CreateBazaar(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Bazaar,
) (*models.Bazaar, error) {
	ctxt := "BazaarService-CreateBazaar"
	response, err := q.bazaarRepository.CreateBazaar(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBazaar")
		return nil, err
	}
	return response, nil
}

func (q *bazaarService) UpdateBazaar(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Bazaar,
) (*models.Bazaar, error) {
	ctxt := "BazaarService-UpdateBazaar"
	response, err := q.bazaarRepository.UpdateBazaar(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBazaar")
		return nil, err
	}
	return response, nil
}
