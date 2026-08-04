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
	NotificationTargetService interface {
		FindNotificationTargets(
			ctx context.Context,
			request *models.NotificationTargetFilter,
		) ([]*models.NotificationTarget, *models.Pagination, error)
		FindNotificationTarget(
			ctx context.Context,
			request *models.NotificationTargetFilter,
		) (*models.NotificationTarget, error)
		SaveNotificationTarget(
			ctx context.Context,
			tx pgx.Tx,
			request *models.NotificationTarget,
		) (*models.NotificationTarget, error)
		DeleteNotificationTarget(
			ctx context.Context,
			tx pgx.Tx,
			request *models.NotificationTargetFilter,
		) (int64, error)
	}

	notificationTargetService struct {
		notificationTargetRepository repositories.NotificationTargetRepository
	}
)

func NewNotificationTargetService(
	notificationTargetRepository repositories.NotificationTargetRepository,
) NotificationTargetService {
	return &notificationTargetService{
		notificationTargetRepository: notificationTargetRepository,
	}
}

func (q *notificationTargetService) FindNotificationTargets(
	ctx context.Context,
	request *models.NotificationTargetFilter,
) ([]*models.NotificationTarget, *models.Pagination, error) {
	ctxt := "NotificationTargetService-FindNotificationTargets"
	notificationTargets, total, pages, err := q.notificationTargetRepository.FindNotificationTargets(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotificationTargets")
		return nil, nil, err
	}
	n := len(notificationTargets)
	rows := make([]*models.NotificationTarget, n)
	if n > 0 {
		copy(rows, notificationTargets)
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

func (q *notificationTargetService) FindNotificationTarget(
	ctx context.Context,
	request *models.NotificationTargetFilter,
) (*models.NotificationTarget, error) {
	ctxt := "NotificationTargetService-FindNotificationTarget"
	response, err := q.notificationTargetRepository.FindNotificationTarget(ctx, request)
	if err != nil && !errors.Is(err, models.ErrNotificationTargetNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotificationTarget")
	}
	return response, err
}

func (q *notificationTargetService) SaveNotificationTarget(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NotificationTarget,
) (*models.NotificationTarget, error) {
	ctxt := "NotificationTargetService-CreateNotificationTarget"
	response, err := q.notificationTargetRepository.SaveNotificationTarget(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateNotificationTarget")
	}
	return response, err
}

func (q *notificationTargetService) DeleteNotificationTarget(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NotificationTargetFilter,
) (int64, error) {
	ctxt := "NotificationTargetService-DeleteNotificationTarget"
	response, err := q.notificationTargetRepository.DeleteNotificationTarget(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteNotificationTarget")
	}
	return response, nil
}
