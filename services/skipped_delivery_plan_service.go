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
	SkippedDeliveryPlanService interface {
		FindSkippedDeliveryPlans(
			ctx context.Context,
			request *models.SkippedDeliveryPlanFilter,
		) ([]*models.SkippedDeliveryPlan, *models.Pagination, error)
		FindSkippedDeliveryPlan(
			ctx context.Context,
			request *models.SkippedDeliveryPlanFilter,
		) (*models.SkippedDeliveryPlan, error)
		CreateSkippedDeliveryPlan(
			ctx context.Context,
			tx pgx.Tx,
			request *models.SkippedDeliveryPlan,
		) (*models.SkippedDeliveryPlan, error)
		UpdateSkippedDeliveryPlan(
			ctx context.Context,
			tx pgx.Tx,
			request *models.SkippedDeliveryPlan,
		) (*models.SkippedDeliveryPlan, error)
		SaveSkippedDeliveryPlan(
			ctx context.Context,
			tx pgx.Tx,
			request *models.SkippedDeliveryPlan,
		) (*models.SkippedDeliveryPlan, error)
	}

	skippedDeliveryPlanService struct {
		skippedDeliveryPlanRepository repositories.SkippedDeliveryPlanRepository
	}
)

func NewSkippedDeliveryPlanService(
	skippedDeliveryPlanRepository repositories.SkippedDeliveryPlanRepository,
) SkippedDeliveryPlanService {
	return &skippedDeliveryPlanService{
		skippedDeliveryPlanRepository: skippedDeliveryPlanRepository,
	}
}

func (q *skippedDeliveryPlanService) FindSkippedDeliveryPlans(
	ctx context.Context,
	request *models.SkippedDeliveryPlanFilter,
) ([]*models.SkippedDeliveryPlan, *models.Pagination, error) {
	ctxt := "SkippedDeliveryPlanService-FindSkippedDeliveryPlans"
	skippedDeliveryPlans, total, pages, err := q.skippedDeliveryPlanRepository.FindSkippedDeliveryPlans(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSkippedDeliveryPlans")
		return nil, nil, err
	}
	n := len(skippedDeliveryPlans)
	rows := make([]*models.SkippedDeliveryPlan, n)
	if n > 0 {
		copy(rows, skippedDeliveryPlans)
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

func (q *skippedDeliveryPlanService) FindSkippedDeliveryPlan(
	ctx context.Context,
	request *models.SkippedDeliveryPlanFilter,
) (*models.SkippedDeliveryPlan, error) {
	ctxt := "SkippedDeliveryPlanService-FindSkippedDeliveryPlan"
	response, err := q.skippedDeliveryPlanRepository.FindSkippedDeliveryPlan(ctx, request)
	if err != nil && !errors.Is(err, models.ErrSkippedDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSkippedDeliveryPlan")
	}
	return response, err
}

func (q *skippedDeliveryPlanService) CreateSkippedDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.SkippedDeliveryPlan,
) (*models.SkippedDeliveryPlan, error) {
	ctxt := "SkippedDeliveryPlanService-CreateSkippedDeliveryPlan"
	response, err := q.skippedDeliveryPlanRepository.CreateSkippedDeliveryPlan(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrUniqueSkippedDeliveryPlanDateDeliveryPlanIdViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSkippedDeliveryPlan")
	}
	return response, err
}

func (q *skippedDeliveryPlanService) UpdateSkippedDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.SkippedDeliveryPlan,
) (*models.SkippedDeliveryPlan, error) {
	ctxt := "SkippedDeliveryPlanService-UpdateSkippedDeliveryPlan"
	response, err := q.skippedDeliveryPlanRepository.UpdateSkippedDeliveryPlan(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrUniqueSkippedDeliveryPlanDateDeliveryPlanIdViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSkippedDeliveryPlan")
	}
	return response, err
}

func (q *skippedDeliveryPlanService) SaveSkippedDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.SkippedDeliveryPlan,
) (response *models.SkippedDeliveryPlan, err error) {
	ctxt := "SkippedDeliveryPlanService-SaveSkippedDeliveryPlan"
	if request.New() {
		response, err = q.skippedDeliveryPlanRepository.CreateSkippedDeliveryPlan(ctx, tx, request)
	} else {
		response, err = q.skippedDeliveryPlanRepository.UpdateSkippedDeliveryPlan(ctx, tx, request)
	}
	if err != nil && !errors.Is(err, models.ErrUniqueSkippedDeliveryPlanDateDeliveryPlanIdViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveSkippedDeliveryPlan")
	}
	return response, err
}
