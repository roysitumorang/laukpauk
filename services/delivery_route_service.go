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
	DeliveryRouteService interface {
		FindDeliveryRoutes(
			ctx context.Context,
			request *models.DeliveryRouteFilter,
		) ([]*models.DeliveryRoute, *models.Pagination, error)
		FindDeliveryRoute(
			ctx context.Context,
			request *models.DeliveryRouteFilter,
		) (*models.DeliveryRoute, error)
		SaveDeliveryRoute(
			ctx context.Context,
			tx pgx.Tx,
			request *models.DeliveryRoute,
		) (*models.DeliveryRoute, error)
	}

	deliveryRouteService struct {
		deliveryRouteRepository repositories.DeliveryRouteRepository
	}
)

func NewDeliveryRouteService(
	deliveryRouteRepository repositories.DeliveryRouteRepository,
) DeliveryRouteService {
	return &deliveryRouteService{
		deliveryRouteRepository: deliveryRouteRepository,
	}
}

func (q *deliveryRouteService) FindDeliveryRoutes(
	ctx context.Context,
	request *models.DeliveryRouteFilter,
) ([]*models.DeliveryRoute, *models.Pagination, error) {
	ctxt := "DeliveryRouteService-FindDeliveryRoutes"
	deliveryRoutes, total, pages, err := q.deliveryRouteRepository.FindDeliveryRoutes(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryRoutes")
		return nil, nil, err
	}
	n := len(deliveryRoutes)
	rows := make([]*models.DeliveryRoute, n)
	if n > 0 {
		copy(rows, deliveryRoutes)
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

func (q *deliveryRouteService) FindDeliveryRoute(
	ctx context.Context,
	request *models.DeliveryRouteFilter,
) (*models.DeliveryRoute, error) {
	ctxt := "DeliveryRouteService-FindDeliveryRoute"
	response, err := q.deliveryRouteRepository.FindDeliveryRoute(ctx, request)
	if err != nil && !errors.Is(err, models.ErrDeliveryRouteNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryRoute")
	}
	return response, err
}

func (q *deliveryRouteService) SaveDeliveryRoute(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryRoute,
) (*models.DeliveryRoute, error) {
	ctxt := "DeliveryRouteService-SaveDeliveryRoute"
	response, err := q.deliveryRouteRepository.SaveDeliveryRoute(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveDeliveryRoute")
	}
	return response, err
}
