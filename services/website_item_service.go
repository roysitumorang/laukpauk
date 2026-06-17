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
	WebsiteItemService interface {
		FindWebsiteItems(
			ctx context.Context,
			request *models.WebsiteItemFilter,
		) ([]*models.WebsiteItem, *models.Pagination, error)
		CreateWebsiteItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteItem,
		) (*models.WebsiteItem, error)
		UpdateWebsiteItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteItem,
		) (*models.WebsiteItem, error)
	}

	websiteItemService struct {
		websiteItemRepository repositories.WebsiteItemRepository
	}
)

func NewWebsiteItemService(
	websiteItemRepository repositories.WebsiteItemRepository,
) WebsiteItemService {
	return new(websiteItemService{
		websiteItemRepository: websiteItemRepository,
	})
}

func (q *websiteItemService) FindWebsiteItems(
	ctx context.Context,
	request *models.WebsiteItemFilter,
) ([]*models.WebsiteItem, *models.Pagination, error) {
	ctxt := "WebsiteItemService-FindWebsiteItems"
	websiteItems, total, pages, err := q.websiteItemRepository.FindWebsiteItems(ctx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteItems")
		return nil, nil, err
	}
	n := len(websiteItems)
	rows := make([]*models.WebsiteItem, n)
	if n > 0 {
		copy(rows, websiteItems)
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

func (q *websiteItemService) CreateWebsiteItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteItem,
) (*models.WebsiteItem, error) {
	ctxt := "WebsiteItemService-CreateWebsiteItem"
	response, err := q.websiteItemRepository.CreateWebsiteItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateWebsiteItem")
		return nil, err
	}
	return response, nil
}
func (q *websiteItemService) UpdateWebsiteItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteItem,
) (*models.WebsiteItem, error) {
	ctxt := "WebsiteItemService-UpdateWebsiteItem"
	response, err := q.websiteItemRepository.UpdateWebsiteItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateWebsiteItem")
		return nil, err
	}
	return response, nil
}
