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
	SettingService interface {
		FindSettings(
			ctx context.Context,
			request *models.SettingFilter,
		) ([]*models.Setting, *models.Pagination, error)
		FindSetting(
			ctx context.Context,
			request *models.SettingFilter,
		) (*models.Setting, error)
		CreateSetting(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Setting,
		) (*models.Setting, error)
		UpdateSetting(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Setting,
		) (*models.Setting, error)
	}

	settingService struct {
		settingRepository repositories.SettingRepository
	}
)

func NewSettingService(
	settingRepository repositories.SettingRepository,
) SettingService {
	return new(settingService{
		settingRepository: settingRepository,
	})
}

func (q *settingService) FindSettings(
	ctx context.Context,
	request *models.SettingFilter,
) ([]*models.Setting, *models.Pagination, error) {
	ctxt := "SettingService-FindSettings"
	settings, total, pages, err := q.settingRepository.FindSettings(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSettings")
		return nil, nil, err
	}
	n := len(settings)
	rows := make([]*models.Setting, n)
	if n > 0 {
		copy(rows, settings)
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

func (q *settingService) FindSetting(
	ctx context.Context,
	request *models.SettingFilter,
) (*models.Setting, error) {
	ctxt := "SettingService-FindSetting"
	response, err := q.settingRepository.FindSetting(ctx, request)
	if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
	}
	return response, err
}

func (q *settingService) CreateSetting(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Setting,
) (*models.Setting, error) {
	ctxt := "SettingService-CreateSetting"
	response, err := q.settingRepository.CreateSetting(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSetting")
		return nil, err
	}
	return response, nil
}
func (q *settingService) UpdateSetting(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Setting,
) (*models.Setting, error) {
	ctxt := "SettingService-UpdateSetting"
	response, err := q.settingRepository.UpdateSetting(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSetting")
		return nil, err
	}
	return response, nil
}
