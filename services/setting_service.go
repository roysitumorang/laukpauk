package services

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	SettingService interface {
		FindSetting(
			ctx context.Context,
		) (*models.Setting, error)
		UpdateSetting(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Setting,
		) (*models.Setting, error)
		AddReferralPointsExpense(
			ctx context.Context,
			tx pgx.Tx,
			addition int64,
			updatedBy uuid.UUID,
			updatedAt time.Time,
		) (*models.Setting, error)
	}

	settingService struct {
		settingRepository repositories.SettingRepository
	}
)

func NewSettingService(
	settingRepository repositories.SettingRepository,
) SettingService {
	return &settingService{
		settingRepository: settingRepository,
	}
}

func (q *settingService) FindSetting(
	ctx context.Context,
) (*models.Setting, error) {
	ctxt := "SettingService-FindSetting"
	response, err := q.settingRepository.FindSetting(ctx)
	if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
	}
	return response, err
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
	}
	return response, err
}

func (q *settingService) AddReferralPointsExpense(
	ctx context.Context,
	tx pgx.Tx,
	addition int64,
	updatedBy uuid.UUID,
	updatedAt time.Time,
) (*models.Setting, error) {
	ctxt := "SettingService-AddReferralPointsExpense"
	response, err := q.settingRepository.AddReferralPointsExpense(
		ctx,
		tx,
		addition,
		updatedBy,
		updatedAt,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrAddReferralPointsExpense")
	}
	return response, err
}
