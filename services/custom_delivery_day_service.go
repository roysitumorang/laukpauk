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
	CustomDeliveryDayService interface {
		FindCustomDeliveryDays(
			ctx context.Context,
			request *models.CustomDeliveryDayFilter,
		) ([]*models.CustomDeliveryDay, *models.Pagination, error)
		FindCustomDeliveryDay(
			ctx context.Context,
			request *models.CustomDeliveryDayFilter,
		) (*models.CustomDeliveryDay, error)
		CreateCustomDeliveryDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.CustomDeliveryDay,
		) (*models.CustomDeliveryDay, error)
		UpdateCustomDeliveryDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.CustomDeliveryDay,
		) (*models.CustomDeliveryDay, error)
	}

	customDeliveryDayService struct {
		customDeliveryDayRepository repositories.CustomDeliveryDayRepository
	}
)

func NewCustomDeliveryDayService(
	customDeliveryDayRepository repositories.CustomDeliveryDayRepository,
) CustomDeliveryDayService {
	return &customDeliveryDayService{
		customDeliveryDayRepository: customDeliveryDayRepository,
	}
}

func (q *customDeliveryDayService) FindCustomDeliveryDays(
	ctx context.Context,
	request *models.CustomDeliveryDayFilter,
) ([]*models.CustomDeliveryDay, *models.Pagination, error) {
	ctxt := "CustomDeliveryDayService-FindCustomDeliveryDays"
	customDeliveryDays, total, pages, err := q.customDeliveryDayRepository.FindCustomDeliveryDays(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCustomDeliveryDays")
		return nil, nil, err
	}
	n := len(customDeliveryDays)
	rows := make([]*models.CustomDeliveryDay, n)
	if n > 0 {
		copy(rows, customDeliveryDays)
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

func (q *customDeliveryDayService) FindCustomDeliveryDay(
	ctx context.Context,
	request *models.CustomDeliveryDayFilter,
) (*models.CustomDeliveryDay, error) {
	ctxt := "CustomDeliveryDayService-FindCustomDeliveryDay"
	response, err := q.customDeliveryDayRepository.FindCustomDeliveryDay(ctx, request)
	if err != nil && !errors.Is(err, models.ErrCustomDeliveryDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCustomDeliveryDay")
	}
	return response, err
}

func (q *customDeliveryDayService) CreateCustomDeliveryDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.CustomDeliveryDay,
) (*models.CustomDeliveryDay, error) {
	ctxt := "CustomDeliveryDayService-CreateCustomDeliveryDay"
	response, err := q.customDeliveryDayRepository.CreateCustomDeliveryDay(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrUniqueCustomDeliveryDayDateDeliveryPlanIdViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateCustomDeliveryDay")
	}
	return response, err
}

func (q *customDeliveryDayService) UpdateCustomDeliveryDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.CustomDeliveryDay,
) (*models.CustomDeliveryDay, error) {
	ctxt := "CustomDeliveryDayService-UpdateCustomDeliveryDay"
	response, err := q.customDeliveryDayRepository.UpdateCustomDeliveryDay(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrUniqueCustomDeliveryDayDateDeliveryPlanIdViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCustomDeliveryDay")
	}
	return response, err
}
