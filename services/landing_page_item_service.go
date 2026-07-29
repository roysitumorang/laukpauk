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
	LandingPageItemService interface {
		FindLandingPageItems(
			ctx context.Context,
			request *models.LandingPageItemFilter,
		) ([]*models.LandingPageItem, *models.Pagination, error)
		FindLandingPageItem(
			ctx context.Context,
			request *models.LandingPageItemFilter,
		) (*models.LandingPageItem, error)
		CreateLandingPageItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageItem,
		) (*models.LandingPageItem, error)
		UpdateLandingPageItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageItem,
		) (*models.LandingPageItem, error)
	}

	landingPageItemService struct {
		landingPageItemRepository repositories.LandingPageItemRepository
	}
)

func NewLandingPageItemService(
	landingPageItemRepository repositories.LandingPageItemRepository,
) LandingPageItemService {
	return &landingPageItemService{
		landingPageItemRepository: landingPageItemRepository,
	}
}

func (q *landingPageItemService) FindLandingPageItems(
	ctx context.Context,
	request *models.LandingPageItemFilter,
) ([]*models.LandingPageItem, *models.Pagination, error) {
	ctxt := "LandingPageItemService-FindLandingPageItems"
	landingPageItems, total, pages, err := q.landingPageItemRepository.FindLandingPageItems(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return nil, nil, err
	}
	n := len(landingPageItems)
	rows := make([]*models.LandingPageItem, n)
	if n > 0 {
		copy(rows, landingPageItems)
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

func (q *landingPageItemService) FindLandingPageItem(
	ctx context.Context,
	request *models.LandingPageItemFilter,
) (*models.LandingPageItem, error) {
	ctxt := "LandingPageItemService-FindLandingPageItem"
	response, err := q.landingPageItemRepository.FindLandingPageItem(ctx, request)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
	}
	return response, err
}

func (q *landingPageItemService) CreateLandingPageItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageItem,
) (*models.LandingPageItem, error) {
	ctxt := "LandingPageItemService-CreateLandingPageItem"
	response, err := q.landingPageItemRepository.CreateLandingPageItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageItem")
	}
	return response, err
}

func (q *landingPageItemService) UpdateLandingPageItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageItem,
) (*models.LandingPageItem, error) {
	ctxt := "LandingPageItemService-UpdateLandingPageItem"
	response, err := q.landingPageItemRepository.UpdateLandingPageItem(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
	}
	return response, err
}
