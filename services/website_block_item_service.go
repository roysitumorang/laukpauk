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
	WebsiteBlockItemService interface {
		FindWebsiteBlockItems(
			ctx context.Context,
			request *models.WebsiteBlockItemFilter,
		) ([]*models.WebsiteBlockItem, *models.Pagination, error)
		CreateWebsiteBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteBlockItem,
		) (*models.WebsiteBlockItem, error)
		UpdateWebsiteBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteBlockItem,
		) (*models.WebsiteBlockItem, error)
	}

	websiteBlockItemService struct {
		websiteBlockItemRepository repositories.WebsiteBlockItemRepository
	}
)

func NewWebsiteBlockItemService(
	websiteBlockItemRepository repositories.WebsiteBlockItemRepository,
) WebsiteBlockItemService {
	return new(websiteBlockItemService{
		websiteBlockItemRepository: websiteBlockItemRepository,
	})
}

func (q *websiteBlockItemService) FindWebsiteBlockItems(
	ctx context.Context,
	request *models.WebsiteBlockItemFilter,
) ([]*models.WebsiteBlockItem, *models.Pagination, error) {
	ctxt := "WebsiteBlockItemService-FindWebsiteBlockItems"
	websiteBlockItems, total, pages, err := q.websiteBlockItemRepository.FindWebsiteBlockItems(ctx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlockItems")
		return nil, nil, err
	}
	n := len(websiteBlockItems)
	rows := make([]*models.WebsiteBlockItem, n)
	if n > 0 {
		copy(rows, websiteBlockItems)
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

func (q *websiteBlockItemService) CreateWebsiteBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteBlockItem,
) (*models.WebsiteBlockItem, error) {
	ctxt := "WebsiteBlockItemService-CreateWebsiteBlockItem"
	response, err := q.websiteBlockItemRepository.CreateWebsiteBlockItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateWebsiteBlockItem")
		return nil, err
	}
	return response, nil
}
func (q *websiteBlockItemService) UpdateWebsiteBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteBlockItem,
) (*models.WebsiteBlockItem, error) {
	ctxt := "WebsiteBlockItemService-UpdateWebsiteBlockItem"
	response, err := q.websiteBlockItemRepository.UpdateWebsiteBlockItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateWebsiteBlockItem")
		return nil, err
	}
	return response, nil
}
