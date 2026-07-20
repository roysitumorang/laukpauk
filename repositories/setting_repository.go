package repositories

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	SettingRepository interface {
		FindSetting(
			ctx context.Context,
		) (*models.Setting, error)
		UpdateSetting(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Setting,
		) (*models.Setting, error)
		DeductReferralPointsBudget(
			ctx context.Context,
			tx pgx.Tx,
			deduction int64,
			updatedBy uuid.UUID,
			updatedAt time.Time,
		) (*models.Setting, error)
	}

	settingRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewSettingRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) SettingRepository {
	return &settingRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *settingRepository) FindSetting(
	ctx context.Context,
) (*models.Setting, error) {
	ctxt := "SettingRepository-FindSetting"
	var response models.Setting
	err := q.dbRead.QueryRow(
		ctx,
		`SELECT
			id
			, maintenance_mode
			, voucher_enabled
			, referral_points_enabled
			, referral_points_budget
			, referral_points_per_referee
			, referral_points_slots
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, created_at
			, updated_by
			, updated_at
		FROM setting`,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsBudget,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrSettingNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *settingRepository) UpdateSetting(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Setting,
) (*models.Setting, error) {
	ctxt := "SettingRepository-UpdateSetting"
	var response models.Setting
	err := tx.QueryRow(
		ctx,
		`SELECT id FROM setting WHERE id = 1 FOR NO KEY UPDATE`,
	).Scan(&response.ID)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	if err = tx.QueryRow(
		ctx,
		`UPDATE setting SET
			maintenance_mode = $1
			, voucher_enabled = $2
			, referral_points_enabled = $3
			, referral_points_budget = $4
			, referral_points_per_referee = $5
			, referral_points_slots = $6
			, reward_points_accumulation_window_in_days = $7
			, delivery_origin_longitude = $8
			, delivery_origin_latitude = $9
			, delivery_cost_per_km = $10
			, delivery_max_distance_km = $11
			, whatsapp_phone_number = $12
			, updated_by = $13
			, updated_at = $14
		WHERE id = 1
		RETURNING id
			, maintenance_mode
			, voucher_enabled
			, referral_points_enabled
			, referral_points_budget
			, referral_points_per_referee
			, referral_points_slots
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, created_at
			, updated_by
			, updated_at`,
		request.MaintenanceMode,
		request.VoucherEnabled,
		request.ReferralPointsEnabled,
		request.ReferralPointsBudget,
		request.ReferralPointsPerReferee,
		request.ReferralPointsSlots,
		request.RewardPointsAccumulationWindowInDays,
		request.DeliveryOriginLongitude,
		request.DeliveryOriginLatitude,
		request.DeliveryCostPerKM,
		request.DeliveryMaxDistanceKM,
		request.WhatsappPhoneNumber,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsBudget,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *settingRepository) DeductReferralPointsBudget(
	ctx context.Context,
	tx pgx.Tx,
	deduction int64,
	updatedBy uuid.UUID,
	updatedAt time.Time,
) (*models.Setting, error) {
	ctxt := "SettingRepository-DeductReferralPointsBudget"
	var response models.Setting
	err := tx.QueryRow(
		ctx,
		`SELECT id FROM setting WHERE id = 1 FOR NO KEY UPDATE`,
	).Scan(&response.ID)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	if err = tx.QueryRow(
		ctx,
		`UPDATE setting SET
			referral_points_budget = referral_points_budget - $1
			, updated_by = $2
			, updated_at = $3
		WHERE id = 1
		RETURNING id
			, maintenance_mode
			, voucher_enabled
			, referral_points_enabled
			, referral_points_budget
			, referral_points_per_referee
			, referral_points_slots
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, created_at
			, updated_by
			, updated_at`,
		deduction,
		updatedBy,
		updatedAt,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsBudget,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
