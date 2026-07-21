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
			, weekday_transaction_cut_off_time
			, weekday_same_day
			, weekday_same_day_transaction_window_max
			, weekday_next_day
			, weekday_until_cut_off_next_day_delivery_start_time
			, weekday_after_cut_off_next_day_delivery_start_time
			, weekday_next_two_day
			, weekday_next_two_day_delivery_start_time
			, weekday_delivery_end_time
			, weekday_delivery_interval
			, saturday_transaction_cut_off_time
			, saturday_same_day
			, saturday_same_day_transaction_window_max
			, saturday_next_day
			, saturday_until_cut_off_next_day_delivery_start_time
			, saturday_after_cut_off_next_day_delivery_start_time
			, saturday_next_two_day
			, saturday_next_two_day_delivery_start_time
			, saturday_delivery_end_time
			, saturday_delivery_interval
			, sunday_transaction_cut_off_time
			, sunday_same_day
			, sunday_same_day_transaction_window_max
			, sunday_next_day
			, sunday_until_cut_off_next_day_delivery_start_time
			, sunday_after_cut_off_next_day_delivery_start_time
			, sunday_next_two_day
			, sunday_next_two_day_delivery_start_time
			, sunday_delivery_end_time
			, sunday_delivery_interval
			, skip_sunday_delivery
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
		&response.WeekdayTransactionCutOffTime,
		&response.WeekdaySameDay,
		&response.WeekdaySameDayTransactionWindowMax,
		&response.WeekdayNextDay,
		&response.WeekdayUntilCutOffNextDayDeliveryStartTime,
		&response.WeekdayAfterCutOffNextDayDeliveryStartTime,
		&response.WeekdayNextTwoDay,
		&response.WeekdayNextTwoDayDeliveryStartTime,
		&response.WeekdayDeliveryEndTime,
		&response.WeekdayDeliveryInterval,
		&response.SaturdayTransactionCutOffTime,
		&response.SaturdaySameDay,
		&response.SaturdaySameDayTransactionWindowMax,
		&response.SaturdayNextDay,
		&response.SaturdayUntilCutOffNextDayDeliveryStartTime,
		&response.SaturdayAfterCutOffNextDayDeliveryStartTime,
		&response.SaturdayNextTwoDay,
		&response.SaturdayNextTwoDayDeliveryStartTime,
		&response.SaturdayDeliveryEndTime,
		&response.SaturdayDeliveryInterval,
		&response.SundayTransactionCutOffTime,
		&response.SundaySameDay,
		&response.SundaySameDayTransactionWindowMax,
		&response.SundayNextDay,
		&response.SundayUntilCutOffNextDayDeliveryStartTime,
		&response.SundayAfterCutOffNextDayDeliveryStartTime,
		&response.SundayNextTwoDay,
		&response.SundayNextTwoDayDeliveryStartTime,
		&response.SundayDeliveryEndTime,
		&response.SundayDeliveryInterval,
		&response.SkipSundayDelivery,
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
			, weekday_transaction_cut_off_time = $15
			, weekday_same_day = $16
			, weekday_same_day_transaction_window_max = $17
			, weekday_next_day = $18
			, weekday_until_cut_off_next_day_delivery_start_time = $19
			, weekday_after_cut_off_next_day_delivery_start_time = $20
			, weekday_next_two_day = $21
			, weekday_next_two_day_delivery_start_time = $22
			, weekday_delivery_end_time = $23
			, weekday_delivery_interval = $24
			, saturday_transaction_cut_off_time = $25
			, saturday_same_day = $26
			, saturday_same_day_transaction_window_max = $27
			, saturday_next_day = $28
			, saturday_until_cut_off_next_day_delivery_start_time = $29
			, saturday_after_cut_off_next_day_delivery_start_time = $30
			, saturday_next_two_day = $31
			, saturday_next_two_day_delivery_start_time = $32
			, saturday_delivery_end_time = $33
			, saturday_delivery_interval = $34
			, sunday_transaction_cut_off_time = $35
			, sunday_same_day = $36
			, sunday_same_day_transaction_window_max = $37
			, sunday_next_day = $38
			, sunday_until_cut_off_next_day_delivery_start_time = $39
			, sunday_after_cut_off_next_day_delivery_start_time = $40
			, sunday_next_two_day = $41
			, sunday_next_two_day_delivery_start_time = $42
			, sunday_delivery_end_time = $43
			, sunday_delivery_interval = $44
			, skip_sunday_delivery = $45
			, updated_by = $46
			, updated_at = $47
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
			, weekday_transaction_cut_off_time
			, weekday_same_day
			, weekday_same_day_transaction_window_max
			, weekday_next_day
			, weekday_until_cut_off_next_day_delivery_start_time
			, weekday_after_cut_off_next_day_delivery_start_time
			, weekday_next_two_day
			, weekday_next_two_day_delivery_start_time
			, weekday_delivery_end_time
			, weekday_delivery_interval
			, saturday_transaction_cut_off_time
			, saturday_same_day
			, saturday_same_day_transaction_window_max
			, saturday_next_day
			, saturday_until_cut_off_next_day_delivery_start_time
			, saturday_after_cut_off_next_day_delivery_start_time
			, saturday_next_two_day
			, saturday_next_two_day_delivery_start_time
			, saturday_delivery_end_time
			, saturday_delivery_interval
			, sunday_transaction_cut_off_time
			, sunday_same_day
			, sunday_same_day_transaction_window_max
			, sunday_next_day
			, sunday_until_cut_off_next_day_delivery_start_time
			, sunday_after_cut_off_next_day_delivery_start_time
			, sunday_next_two_day
			, sunday_next_two_day_delivery_start_time
			, sunday_delivery_end_time
			, sunday_delivery_interval
			, skip_sunday_delivery
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
		request.WeekdayTransactionCutOffTime,
		request.WeekdaySameDay,
		request.WeekdaySameDayTransactionWindowMax,
		request.WeekdayNextDay,
		request.WeekdayUntilCutOffNextDayDeliveryStartTime,
		request.WeekdayAfterCutOffNextDayDeliveryStartTime,
		request.WeekdayNextTwoDay,
		request.WeekdayNextTwoDayDeliveryStartTime,
		request.WeekdayDeliveryEndTime,
		request.WeekdayDeliveryInterval,
		request.SaturdayTransactionCutOffTime,
		request.SaturdaySameDay,
		request.SaturdaySameDayTransactionWindowMax,
		request.SaturdayNextDay,
		request.SaturdayUntilCutOffNextDayDeliveryStartTime,
		request.SaturdayAfterCutOffNextDayDeliveryStartTime,
		request.SaturdayNextTwoDay,
		request.SaturdayNextTwoDayDeliveryStartTime,
		request.SaturdayDeliveryEndTime,
		request.SaturdayDeliveryInterval,
		request.SundayTransactionCutOffTime,
		request.SundaySameDay,
		request.SundaySameDayTransactionWindowMax,
		request.SundayNextDay,
		request.SundayUntilCutOffNextDayDeliveryStartTime,
		request.SundayAfterCutOffNextDayDeliveryStartTime,
		request.SundayNextTwoDay,
		request.SundayNextTwoDayDeliveryStartTime,
		request.SundayDeliveryEndTime,
		request.SundayDeliveryInterval,
		request.SkipSundayDelivery,
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
		&response.WeekdayTransactionCutOffTime,
		&response.WeekdaySameDay,
		&response.WeekdaySameDayTransactionWindowMax,
		&response.WeekdayNextDay,
		&response.WeekdayUntilCutOffNextDayDeliveryStartTime,
		&response.WeekdayAfterCutOffNextDayDeliveryStartTime,
		&response.WeekdayNextTwoDay,
		&response.WeekdayNextTwoDayDeliveryStartTime,
		&response.WeekdayDeliveryEndTime,
		&response.WeekdayDeliveryInterval,
		&response.SaturdayTransactionCutOffTime,
		&response.SaturdaySameDay,
		&response.SaturdaySameDayTransactionWindowMax,
		&response.SaturdayNextDay,
		&response.SaturdayUntilCutOffNextDayDeliveryStartTime,
		&response.SaturdayAfterCutOffNextDayDeliveryStartTime,
		&response.SaturdayNextTwoDay,
		&response.SaturdayNextTwoDayDeliveryStartTime,
		&response.SaturdayDeliveryEndTime,
		&response.SaturdayDeliveryInterval,
		&response.SundayTransactionCutOffTime,
		&response.SundaySameDay,
		&response.SundaySameDayTransactionWindowMax,
		&response.SundayNextDay,
		&response.SundayUntilCutOffNextDayDeliveryStartTime,
		&response.SundayAfterCutOffNextDayDeliveryStartTime,
		&response.SundayNextTwoDay,
		&response.SundayNextTwoDayDeliveryStartTime,
		&response.SundayDeliveryEndTime,
		&response.SundayDeliveryInterval,
		&response.SkipSundayDelivery,
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
			, weekday_transaction_cut_off_time
			, weekday_same_day
			, weekday_same_day_transaction_window_max
			, weekday_next_day
			, weekday_until_cut_off_next_day_delivery_start_time
			, weekday_after_cut_off_next_day_delivery_start_time
			, weekday_next_two_day
			, weekday_next_two_day_delivery_start_time
			, weekday_delivery_end_time
			, weekday_delivery_interval
			, saturday_transaction_cut_off_time
			, saturday_same_day
			, saturday_same_day_transaction_window_max
			, saturday_next_day
			, saturday_until_cut_off_next_day_delivery_start_time
			, saturday_after_cut_off_next_day_delivery_start_time
			, saturday_next_two_day
			, saturday_next_two_day_delivery_start_time
			, saturday_delivery_end_time
			, saturday_delivery_interval
			, sunday_transaction_cut_off_time
			, sunday_same_day
			, sunday_same_day_transaction_window_max
			, sunday_next_day
			, sunday_until_cut_off_next_day_delivery_start_time
			, sunday_after_cut_off_next_day_delivery_start_time
			, sunday_next_two_day
			, sunday_next_two_day_delivery_start_time
			, sunday_delivery_end_time
			, sunday_delivery_interval
			, skip_sunday_delivery
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
		&response.WeekdayTransactionCutOffTime,
		&response.WeekdaySameDay,
		&response.WeekdaySameDayTransactionWindowMax,
		&response.WeekdayNextDay,
		&response.WeekdayUntilCutOffNextDayDeliveryStartTime,
		&response.WeekdayAfterCutOffNextDayDeliveryStartTime,
		&response.WeekdayNextTwoDay,
		&response.WeekdayNextTwoDayDeliveryStartTime,
		&response.WeekdayDeliveryEndTime,
		&response.WeekdayDeliveryInterval,
		&response.SaturdayTransactionCutOffTime,
		&response.SaturdaySameDay,
		&response.SaturdaySameDayTransactionWindowMax,
		&response.SaturdayNextDay,
		&response.SaturdayUntilCutOffNextDayDeliveryStartTime,
		&response.SaturdayAfterCutOffNextDayDeliveryStartTime,
		&response.SaturdayNextTwoDay,
		&response.SaturdayNextTwoDayDeliveryStartTime,
		&response.SaturdayDeliveryEndTime,
		&response.SaturdayDeliveryInterval,
		&response.SundayTransactionCutOffTime,
		&response.SundaySameDay,
		&response.SundaySameDayTransactionWindowMax,
		&response.SundayNextDay,
		&response.SundayUntilCutOffNextDayDeliveryStartTime,
		&response.SundayAfterCutOffNextDayDeliveryStartTime,
		&response.SundayNextTwoDay,
		&response.SundayNextTwoDayDeliveryStartTime,
		&response.SundayDeliveryEndTime,
		&response.SundayDeliveryInterval,
		&response.SkipSundayDelivery,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
