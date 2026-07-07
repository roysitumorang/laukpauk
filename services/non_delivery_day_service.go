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
	NonDeliveryDayService interface {
		FindNonDeliveryDays(
			ctx context.Context,
			request *models.NonDeliveryDayFilter,
		) ([]*models.NonDeliveryDay, *models.Pagination, error)
		FindNonDeliveryDay(
			ctx context.Context,
			request *models.NonDeliveryDayFilter,
		) (*models.NonDeliveryDay, error)
		CreateNonDeliveryDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.NonDeliveryDay,
		) (*models.NonDeliveryDay, error)
		UpdateNonDeliveryDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.NonDeliveryDay,
		) (*models.NonDeliveryDay, error)
	}

	nonDeliveryDayService struct {
		nonDeliveryDayRepository repositories.NonDeliveryDayRepository
	}
)

func NewNonDeliveryDayService(
	nonDeliveryDayRepository repositories.NonDeliveryDayRepository,
) NonDeliveryDayService {
	return new(nonDeliveryDayService{
		nonDeliveryDayRepository: nonDeliveryDayRepository,
	})
}

func (q *nonDeliveryDayService) FindNonDeliveryDays(
	ctx context.Context,
	request *models.NonDeliveryDayFilter,
) ([]*models.NonDeliveryDay, *models.Pagination, error) {
	ctxt := "NonDeliveryDayService-FindNonDeliveryDays"
	nonDeliveryDays, total, pages, err := q.nonDeliveryDayRepository.FindNonDeliveryDays(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNonDeliveryDays")
		return nil, nil, err
	}
	n := len(nonDeliveryDays)
	rows := make([]*models.NonDeliveryDay, n)
	if n > 0 {
		copy(rows, nonDeliveryDays)
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

func (q *nonDeliveryDayService) FindNonDeliveryDay(
	ctx context.Context,
	request *models.NonDeliveryDayFilter,
) (*models.NonDeliveryDay, error) {
	ctxt := "NonDeliveryDayService-FindNonDeliveryDay"
	response, err := q.nonDeliveryDayRepository.FindNonDeliveryDay(ctx, request)
	if err != nil && !errors.Is(err, models.ErrNonDeliveryDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNonDeliveryDay")
	}
	return response, err
}

func (q *nonDeliveryDayService) CreateNonDeliveryDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NonDeliveryDay,
) (*models.NonDeliveryDay, error) {
	ctxt := "NonDeliveryDayService-CreateNonDeliveryDay"
	response, err := q.nonDeliveryDayRepository.CreateNonDeliveryDay(ctx, tx, request)
	if err != nil {
		if !errors.Is(err, models.ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateNonDeliveryDay")
		}
		return nil, err
	}
	return response, nil
}

func (q *nonDeliveryDayService) UpdateNonDeliveryDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NonDeliveryDay,
) (*models.NonDeliveryDay, error) {
	ctxt := "NonDeliveryDayService-UpdateNonDeliveryDay"
	response, err := q.nonDeliveryDayRepository.UpdateNonDeliveryDay(ctx, tx, request)
	if err != nil {
		if !errors.Is(err, models.ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateNonDeliveryDay")
		}
		return nil, err
	}
	return response, nil
}
