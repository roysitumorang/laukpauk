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
	LandingPageBlockService interface {
		FindLandingPageBlocks(
			ctx context.Context,
			request *models.LandingPageBlockFilter,
		) ([]*models.LandingPageBlock, *models.Pagination, error)
		UpdateLandingPageBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageBlock,
		) (*models.LandingPageBlock, error)
		FindLandingPageBlockItems(
			ctx context.Context,
			request *models.LandingPageBlockItemFilter,
		) ([]*models.LandingPageBlockItem, *models.Pagination, error)
		CreateLandingPageBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageBlockItem,
		) (*models.LandingPageBlockItem, error)
		UpdateLandingPageBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageBlockItem,
		) (*models.LandingPageBlockItem, error)
	}

	landingPageBlockService struct {
		landingPageBlockRepository repositories.LandingPageBlockRepository
	}
)

func NewLandingPageBlockService(
	landingPageBlockRepository repositories.LandingPageBlockRepository,
) LandingPageBlockService {
	return new(landingPageBlockService{
		landingPageBlockRepository: landingPageBlockRepository,
	})
}

func (q *landingPageBlockService) FindLandingPageBlocks(
	ctx context.Context,
	request *models.LandingPageBlockFilter,
) ([]*models.LandingPageBlock, *models.Pagination, error) {
	ctxt := "LandingPageBlockService-FindLandingPageBlocks"
	landingPageBlocks, total, pages, err := q.landingPageBlockRepository.FindLandingPageBlocks(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return nil, nil, err
	}
	n := len(landingPageBlocks)
	rows := make([]*models.LandingPageBlock, n)
	if n > 0 {
		copy(rows, landingPageBlocks)
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

func (q *landingPageBlockService) UpdateLandingPageBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageBlock,
) (*models.LandingPageBlock, error) {
	ctxt := "LandingPageBlockService-UpdateLandingPageBlock"
	response, err := q.landingPageBlockRepository.UpdateLandingPageBlock(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return nil, err
	}
	return response, nil
}

func (q *landingPageBlockService) FindLandingPageBlockItems(
	ctx context.Context,
	request *models.LandingPageBlockItemFilter,
) ([]*models.LandingPageBlockItem, *models.Pagination, error) {
	ctxt := "LandingPageBlockService-FindLandingPageBlockItems"
	landingPageBlockItems, total, pages, err := q.landingPageBlockRepository.FindLandingPageBlockItems(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return nil, nil, err
	}
	n := len(landingPageBlockItems)
	rows := make([]*models.LandingPageBlockItem, n)
	if n > 0 {
		copy(rows, landingPageBlockItems)
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

func (q *landingPageBlockService) CreateLandingPageBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageBlockItem,
) (*models.LandingPageBlockItem, error) {
	ctxt := "LandingPageBlockService-CreateLandingPageBlockItem"
	response, err := q.landingPageBlockRepository.CreateLandingPageBlockItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlockItem")
		return nil, err
	}
	return response, nil
}

func (q *landingPageBlockService) UpdateLandingPageBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageBlockItem,
) (*models.LandingPageBlockItem, error) {
	ctxt := "LandingPageBlockService-UpdateLandingPageBlockItem"
	response, err := q.landingPageBlockRepository.UpdateLandingPageBlockItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
		return nil, err
	}
	return response, nil
}
