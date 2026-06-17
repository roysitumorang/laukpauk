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
	WebsiteBlockService interface {
		FindWebsiteBlocks(
			ctx context.Context,
			request *models.WebsiteBlockFilter,
		) ([]*models.WebsiteBlock, *models.Pagination, error)
		CreateWebsiteBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteBlock,
		) (*models.WebsiteBlock, error)
		UpdateWebsiteBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteBlock,
		) (*models.WebsiteBlock, error)
	}

	websiteBlockService struct {
		websiteBlockRepository repositories.WebsiteBlockRepository
	}
)

func NewWebsiteBlockService(
	websiteBlockRepository repositories.WebsiteBlockRepository,
) WebsiteBlockService {
	return new(websiteBlockService{
		websiteBlockRepository: websiteBlockRepository,
	})
}

func (q *websiteBlockService) FindWebsiteBlocks(
	ctx context.Context,
	request *models.WebsiteBlockFilter,
) ([]*models.WebsiteBlock, *models.Pagination, error) {
	ctxt := "WebsiteBlockService-FindWebsiteBlocks"
	websiteBlocks, total, pages, err := q.websiteBlockRepository.FindWebsiteBlocks(ctx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlocks")
		return nil, nil, err
	}
	n := len(websiteBlocks)
	rows := make([]*models.WebsiteBlock, n)
	if n > 0 {
		copy(rows, websiteBlocks)
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

func (q *websiteBlockService) CreateWebsiteBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteBlock,
) (*models.WebsiteBlock, error) {
	ctxt := "WebsiteBlockService-CreateWebsiteBlock"
	response, err := q.websiteBlockRepository.CreateWebsiteBlock(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateWebsiteBlock")
		return nil, err
	}
	return response, nil
}
func (q *websiteBlockService) UpdateWebsiteBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteBlock,
) (*models.WebsiteBlock, error) {
	ctxt := "WebsiteBlockService-UpdateWebsiteBlock"
	response, err := q.websiteBlockRepository.UpdateWebsiteBlock(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateWebsiteBlock")
		return nil, err
	}
	return response, nil
}
