package services

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"go.uber.org/zap"
)

type (
	LandingPageBlockService interface {
		FindLandingPageBlocks(
			ctx context.Context,
			request *models.LandingPageBlockFilter,
		) ([]*models.LandingPageBlock, *models.Pagination, error)
		CreateLandingPageBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageBlock,
		) (*models.LandingPageBlock, error)
		UpdateLandingPageBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageBlock,
		) (*models.LandingPageBlock, error)
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
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLandingPageBlocks")
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
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrNewPagination")
		return nil, nil, err
	}
	return rows, pagination, nil
}

func (q *landingPageBlockService) CreateLandingPageBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageBlock,
) (*models.LandingPageBlock, error) {
	ctxt := "LandingPageBlockService-CreateLandingPageBlock"
	response, err := q.landingPageBlockRepository.CreateLandingPageBlock(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateLandingPageBlock")
		return nil, err
	}
	return response, nil
}
func (q *landingPageBlockService) UpdateLandingPageBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageBlock,
) (*models.LandingPageBlock, error) {
	ctxt := "LandingPageBlockService-UpdateLandingPageBlock"
	response, err := q.landingPageBlockRepository.UpdateLandingPageBlock(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return nil, err
	}
	return response, nil
}
