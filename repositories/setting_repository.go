package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
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
		AddReferralPointsExpense(
			ctx context.Context,
			tx pgx.Tx,
			addition int64,
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
			, vacation_mode
			, vacation_start_date
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_per_referee
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, delivery_time_interval
			, whatsapp_phone_number
			, payment_terms
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM setting`,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.VacationMode,
		&response.VacationStartDate,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.DeliveryTimeInterval,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms,
		&response.CreatedBy,
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
		`SELECT
			id
			, maintenance_mode
			, vacation_mode
			, vacation_start_date
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_per_referee
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, delivery_time_interval
			, whatsapp_phone_number
			, payment_terms
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM setting
		WHERE id = 1
		FOR NO KEY UPDATE`,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.VacationMode,
		&response.VacationStartDate,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.DeliveryTimeInterval,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	_, _ = sb.WriteString("UPDATE setting SET")
	if response.MaintenanceMode != request.MaintenanceMode {
		params = append(params, request.MaintenanceMode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " maintenance_mode = $%d", n)
	}
	if response.VacationMode != request.VacationMode {
		params = append(params, request.VacationMode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " vacation_mode = $%d", n)
	}
	if (response.VacationStartDate == nil &&
		request.VacationStartDate != nil) ||
		(response.VacationStartDate != nil &&
			request.VacationStartDate == nil) ||
		(response.VacationStartDate != nil &&
			request.VacationStartDate != nil &&
			!response.VacationStartDate.Equal(*request.VacationStartDate)) {
		params = append(params, request.VacationStartDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " vacation_start_date = $%d", n)
	}
	if (response.VacationEndDate == nil &&
		request.VacationEndDate != nil) ||
		(response.VacationEndDate != nil &&
			request.VacationEndDate == nil) ||
		(response.VacationEndDate != nil &&
			request.VacationEndDate != nil &&
			!response.VacationEndDate.Equal(*request.VacationEndDate)) {
		params = append(params, request.VacationEndDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " vacation_end_date = $%d", n)
	}
	if response.VoucherEnabled != request.VoucherEnabled {
		params = append(params, request.VoucherEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " voucher_enabled = $%d", n)
	}
	if response.ReferralPointsEnabled != request.ReferralPointsEnabled {
		params = append(params, request.ReferralPointsEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " referral_points_enabled = $%d", n)
	}
	if response.ReferralPointsMinimumShoppingCost != request.ReferralPointsMinimumShoppingCost {
		params = append(params, request.ReferralPointsMinimumShoppingCost)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " referral_points_minimum_shopping_cost = $%d", n)
	}
	if response.ReferralPointsPerReferee != request.ReferralPointsPerReferee {
		params = append(params, request.ReferralPointsPerReferee)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " referral_points_per_referee = $%d", n)
	}
	if slices.Compare(response.ReferralPointsSlots, request.ReferralPointsSlots) != 0 {
		params = append(params, request.ReferralPointsSlots)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " referral_points_slots = $%d", n)
	}
	if response.RewardPointsAccumulationWindowInDays != request.RewardPointsAccumulationWindowInDays {
		params = append(params, request.RewardPointsAccumulationWindowInDays)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " reward_points_accumulation_window_in_days = $%d", n)
	}
	if response.DeliveryOriginLongitude != request.DeliveryOriginLongitude {
		params = append(params, request.DeliveryOriginLongitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " delivery_origin_longitude = $%d", n)
	}
	if response.DeliveryOriginLatitude != request.DeliveryOriginLatitude {
		params = append(params, request.DeliveryOriginLatitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " delivery_origin_latitude = $%d", n)
	}
	if response.DeliveryCostPerKM != request.DeliveryCostPerKM {
		params = append(params, request.DeliveryCostPerKM)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " delivery_cost_per_km = $%d", n)
	}
	if response.DeliveryMaxDistanceKM != request.DeliveryMaxDistanceKM {
		params = append(params, request.DeliveryMaxDistanceKM)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " delivery_max_distance_km = $%d", n)
	}
	if response.DeliveryTimeInterval != request.DeliveryTimeInterval {
		params = append(params, request.DeliveryTimeInterval)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " delivery_time_interval = $%d", n)
	}
	if response.WhatsappPhoneNumber != request.WhatsappPhoneNumber {
		params = append(params, request.WhatsappPhoneNumber)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " whatsapp_phone_number = $%d", n)
	}
	if response.PaymentTerms != request.PaymentTerms {
		params = append(params, request.PaymentTerms)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_terms = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.UpdatedBy, request.UpdatedAt, request.ID)
	n = len(params)
	_, _ = fmt.Fprintf(
		&sb,
		`, updated_by = $%d
		, updated_at = $%d
		WHERE id = $%d
		RETURNING id
			, maintenance_mode
			, vacation_mode
			, vacation_start_date
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_per_referee
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, delivery_time_interval
			, whatsapp_phone_number
			, payment_terms
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		n-2, n-1, n,
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.VacationMode,
		&response.VacationStartDate,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.DeliveryTimeInterval,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *settingRepository) AddReferralPointsExpense(
	ctx context.Context,
	tx pgx.Tx,
	addition int64,
	updatedBy uuid.UUID,
	updatedAt time.Time,
) (*models.Setting, error) {
	ctxt := "SettingRepository-AddReferralPointsExpense"
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
			referral_points_expense = referral_points_expense + $1
			, updated_by = $2
			, updated_at = $3
		WHERE id = 1
		RETURNING id
			, maintenance_mode
			, vacation_mode
			, vacation_start_date
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_per_referee
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, delivery_time_interval
			, whatsapp_phone_number
			, payment_terms
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		addition,
		updatedBy,
		updatedAt,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.VacationMode,
		&response.VacationStartDate,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsPerReferee,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.DeliveryTimeInterval,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
