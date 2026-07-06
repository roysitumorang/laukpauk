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
	DeliveryPlanService interface {
		FindDeliveryPlans(
			ctx context.Context,
			request *models.DeliveryPlanFilter,
		) ([]*models.DeliveryPlan, *models.Pagination, error)
		FindDeliveryPlan(
			ctx context.Context,
			request *models.DeliveryPlanFilter,
		) (*models.DeliveryPlan, error)
		CreateDeliveryPlan(
			ctx context.Context,
			tx pgx.Tx,
			request *models.DeliveryPlan,
		) (*models.DeliveryPlan, error)
		UpdateDeliveryPlan(
			ctx context.Context,
			tx pgx.Tx,
			request *models.DeliveryPlan,
		) (*models.DeliveryPlan, error)
	}

	deliveryPlanService struct {
		deliveryPlanRepository repositories.DeliveryPlanRepository
	}
)

func NewDeliveryPlanService(
	deliveryPlanRepository repositories.DeliveryPlanRepository,
) DeliveryPlanService {
	return new(deliveryPlanService{
		deliveryPlanRepository: deliveryPlanRepository,
	})
}

func (q *deliveryPlanService) FindDeliveryPlans(
	ctx context.Context,
	request *models.DeliveryPlanFilter,
) ([]*models.DeliveryPlan, *models.Pagination, error) {
	ctxt := "DeliveryPlanService-FindDeliveryPlans"
	deliveryPlans, total, pages, err := q.deliveryPlanRepository.FindDeliveryPlans(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return nil, nil, err
	}
	n := len(deliveryPlans)
	rows := make([]*models.DeliveryPlan, n)
	if n > 0 {
		copy(rows, deliveryPlans)
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

func (q *deliveryPlanService) FindDeliveryPlan(
	ctx context.Context,
	request *models.DeliveryPlanFilter,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanService-FindDeliveryPlan"
	response, err := q.deliveryPlanRepository.FindDeliveryPlan(ctx, request)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
	}
	return response, err
}

func (q *deliveryPlanService) CreateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanService-CreateDeliveryPlan"
	response, err := q.deliveryPlanRepository.CreateDeliveryPlan(ctx, tx, request)
	if err != nil {
		if !errors.Is(err, models.ErrUniqueDeliveryPlanNameViolation) &&
			!errors.Is(err, models.ErrUniqueDeliveryTransactionWindowFrameViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateDeliveryPlan")
		}
		return nil, err
	}
	return response, nil
}

func (q *deliveryPlanService) UpdateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanService-UpdateDeliveryPlan"
	response, err := q.deliveryPlanRepository.UpdateDeliveryPlan(ctx, tx, request)
	if err != nil {
		if !errors.Is(err, models.ErrUniqueDeliveryPlanNameViolation) &&
			!errors.Is(err, models.ErrUniqueDeliveryTransactionWindowFrameViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
		}
		return nil, err
	}
	return response, nil
}
