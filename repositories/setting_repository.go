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
			, referral_points_initial_budget
			, refferal_points_minimum_shopping_cost
			, referral_points_per_referee
			, referral_points_slots
			, referral_points_remaining_budget
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
		&response.ReferralPointsInitialBudget,
		&response.RefferalPointsMinimumShoppingCost,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.ReferralPointsRemainingBudget,
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
			, referral_points_initial_budget = $4
			, refferal_points_minimum_shopping_cost = $5
			, referral_points_per_referee = $6
			, referral_points_slots = $7
			, referral_points_remaining_budget = $8
			, reward_points_accumulation_window_in_days = $9
			, delivery_origin_longitude = $10
			, delivery_origin_latitude = $11
			, delivery_cost_per_km = $12
			, delivery_max_distance_km = $13
			, whatsapp_phone_number = $14
			, updated_by = $15
			, updated_at = $16
		WHERE id = 1
		RETURNING id
			, maintenance_mode
			, voucher_enabled
			, referral_points_enabled
			, referral_points_initial_budget
			, refferal_points_minimum_shopping_cost
			, referral_points_per_referee
			, referral_points_slots
			, referral_points_remaining_budget
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
		request.ReferralPointsInitialBudget,
		request.RefferalPointsMinimumShoppingCost,
		request.ReferralPointsPerReferee,
		request.ReferralPointsSlots,
		request.ReferralPointsRemainingBudget,
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
		&response.ReferralPointsInitialBudget,
		&response.RefferalPointsMinimumShoppingCost,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.ReferralPointsRemainingBudget,
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
			referral_points_remaining_budget = referral_points_remaining_budget - $1
			, updated_by = $2
			, updated_at = $3
		WHERE id = 1
		RETURNING id
			, maintenance_mode
			, voucher_enabled
			, referral_points_enabled
			, referral_points_initial_budget
			, refferal_points_minimum_shopping_cost
			, referral_points_per_referee
			, referral_points_slots
			, referral_points_remaining_budget
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
		&response.ReferralPointsInitialBudget,
		&response.RefferalPointsMinimumShoppingCost,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.ReferralPointsRemainingBudget,
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
