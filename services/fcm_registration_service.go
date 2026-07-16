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
	FcmRegistrationService interface {
		FindFcmRegistrations(
			ctx context.Context,
			request *models.FcmRegistrationFilter,
		) ([]*models.FcmRegistration, *models.Pagination, error)
		FindFcmRegistration(
			ctx context.Context,
			request *models.FcmRegistrationFilter,
		) (*models.FcmRegistration, error)
		SaveFcmRegistration(
			ctx context.Context,
			tx pgx.Tx,
			request *models.FcmRegistration,
		) (*models.FcmRegistration, error)
		DeleteFcmRegistration(
			ctx context.Context,
			tx pgx.Tx,
			request *models.FcmRegistrationFilter,
		) (int64, error)
	}

	fcmRegistrationService struct {
		fcmRegistrationRepository repositories.FcmRegistrationRepository
	}
)

func NewFcmRegistrationService(
	fcmRegistrationRepository repositories.FcmRegistrationRepository,
) FcmRegistrationService {
	return &fcmRegistrationService{
		fcmRegistrationRepository: fcmRegistrationRepository,
	}
}

func (q *fcmRegistrationService) FindFcmRegistrations(
	ctx context.Context,
	request *models.FcmRegistrationFilter,
) ([]*models.FcmRegistration, *models.Pagination, error) {
	ctxt := "FcmRegistrationService-FindFcmRegistrations"
	fcmRegistrations, total, pages, err := q.fcmRegistrationRepository.FindFcmRegistrations(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindFcmRegistrations")
		return nil, nil, err
	}
	n := len(fcmRegistrations)
	rows := make([]*models.FcmRegistration, n)
	if n > 0 {
		copy(rows, fcmRegistrations)
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

func (q *fcmRegistrationService) FindFcmRegistration(
	ctx context.Context,
	request *models.FcmRegistrationFilter,
) (*models.FcmRegistration, error) {
	ctxt := "FcmRegistrationService-FindFcmRegistration"
	response, err := q.fcmRegistrationRepository.FindFcmRegistration(ctx, request)
	if err != nil && !errors.Is(err, models.ErrFcmRegistrationNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindFcmRegistration")
	}
	return response, err
}

func (q *fcmRegistrationService) SaveFcmRegistration(
	ctx context.Context,
	tx pgx.Tx,
	request *models.FcmRegistration,
) (*models.FcmRegistration, error) {
	ctxt := "FcmRegistrationService-CreateFcmRegistration"
	response, err := q.fcmRegistrationRepository.SaveFcmRegistration(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateFcmRegistration")
	}
	return response, err
}

func (q *fcmRegistrationService) DeleteFcmRegistration(
	ctx context.Context,
	tx pgx.Tx,
	request *models.FcmRegistrationFilter,
) (int64, error) {
	ctxt := "FcmRegistrationService-DeleteFcmRegistration"
	response, err := q.fcmRegistrationRepository.DeleteFcmRegistration(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteFcmRegistration")
	}
	return response, nil
}
