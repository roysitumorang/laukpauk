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
	LandingPageService interface {
		FindLandingPages(
			ctx context.Context,
			request *models.LandingPageFilter,
		) ([]*models.LandingPage, *models.Pagination, error)
		FindLandingPage(
			ctx context.Context,
			request *models.LandingPageFilter,
		) (*models.LandingPage, error)
		UpdateLandingPage(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPage,
		) (*models.LandingPage, error)
	}

	landingPageService struct {
		landingPageRepository repositories.LandingPageRepository
	}
)

func NewLandingPageService(
	landingPageRepository repositories.LandingPageRepository,
) LandingPageService {
	return &landingPageService{
		landingPageRepository: landingPageRepository,
	}
}

func (q *landingPageService) FindLandingPages(
	ctx context.Context,
	request *models.LandingPageFilter,
) ([]*models.LandingPage, *models.Pagination, error) {
	ctxt := "LandingPageService-FindLandingPages"
	landingPages, total, pages, err := q.landingPageRepository.FindLandingPages(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPages")
		return nil, nil, err
	}
	n := len(landingPages)
	rows := make([]*models.LandingPage, n)
	if n > 0 {
		copy(rows, landingPages)
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

func (q *landingPageService) FindLandingPage(
	ctx context.Context,
	request *models.LandingPageFilter,
) (*models.LandingPage, error) {
	ctxt := "LandingPageService-FindLandingPage"
	response, err := q.landingPageRepository.FindLandingPage(ctx, request)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
	}
	return response, nil
}

func (q *landingPageService) UpdateLandingPage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPage,
) (*models.LandingPage, error) {
	ctxt := "LandingPageService-UpdateLandingPage"
	response, err := q.landingPageRepository.UpdateLandingPage(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPage")
	}
	return response, err
}
