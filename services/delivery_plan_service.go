package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

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
		AutoUpdateDate(ctx context.Context) error
	}

	deliveryPlanService struct {
		deliveryPlanRepository repositories.DeliveryPlanRepository
	}
)

func NewDeliveryPlanService(
	deliveryPlanRepository repositories.DeliveryPlanRepository,
) DeliveryPlanService {
	return &deliveryPlanService{
		deliveryPlanRepository: deliveryPlanRepository,
	}
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
	if err != nil && !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateDeliveryPlan")
	}
	return response, err
}

func (q *deliveryPlanService) UpdateDeliveryPlan(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryPlan,
) (*models.DeliveryPlan, error) {
	ctxt := "DeliveryPlanService-UpdateDeliveryPlan"
	response, err := q.deliveryPlanRepository.UpdateDeliveryPlan(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
	}
	return response, err
}

func (q *deliveryPlanService) AutoUpdateDate(ctx context.Context) error {
	ctxt := "DeliveryPlanService-AutoUpdateDate"
	now := time.Now().In(time.Local)
	weekday := models.DeliveryWeekdayWeekdays
	if wd := now.Weekday(); wd == time.Sunday ||
		wd == time.Friday ||
		wd == time.Saturday {
		weekday = models.DeliveryWeekday(wd)
	}
	deliveryPlans, _, _, err := q.deliveryPlanRepository.FindDeliveryPlans(
		ctx,
		models.NewDeliveryPlanFilter().
			WithDeliveryWeekdays(weekday),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
		return err
	}
	var request []*models.DeliveryPlan
	for _, deliveryPlan := range deliveryPlans {
		if deliveryPlan.TransactionWindowMin.In(time.Local).Format(time.DateOnly) == now.Format(time.DateOnly) {
			continue
		}
		if err = deliveryPlan.Validate(); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
			return err
		}
		deliveryPlan.UpdatedAt = now
		request = append(request, deliveryPlan)
	}
	if len(request) == 0 {
		return nil
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	for i := range request {
		if _, err = q.deliveryPlanRepository.UpdateDeliveryPlan(ctx, tx, request[i]); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
			return err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return err
	}
	return nil
}
