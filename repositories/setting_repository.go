package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type SettingRepository interface {
	FindSetting(
		ctx context.Context,
	) (*models.Setting, error)
	UpdateSetting(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Setting,
	) (*models.Setting, error)
}

type settingRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewSettingRepository(
	dbRead,
	dbWrite *sql.DB,
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
	err := q.dbRead.QueryRowContext(
		ctx,
		`SELECT
			id
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, transaction_auto_withdrawal_window_min
			, transaction_auto_withdrawal_window_max
			, otp_required_for_login
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM setting`,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.VacationMode,
			&response.VacationEndDate,
			&response.VoucherEnabled,
			&response.ReferralPointsEnabled,
			&response.ReferralPointsMinimumShoppingCost,
			&response.ReferralPointsAmount,
			&response.ReferralPointsSlots,
			&response.ReferralPointsExpense,
			&response.RewardPointsAccumulationWindowInDays,
			&response.DeliveryOriginLongitude,
			&response.DeliveryOriginLatitude,
			&response.DeliveryCostPerKM,
			&response.DeliveryMaxDistanceKM,
			&response.WhatsappPhoneNumber,
			&response.PaymentTerms,
			&response.MaximumDeliveryTime,
			&response.TransactionAutoWithdrawalEnabled,
			&response.TransactionAutoWithdrawalWindowMin,
			&response.TransactionAutoWithdrawalWindowMax,
			&response.OtpRequiredForLogin,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrSettingNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *settingRepository) UpdateSetting(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Setting,
) (*models.Setting, error) {
	ctxt := "SettingRepository-UpdateSetting"
	var response models.Setting
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, transaction_auto_withdrawal_window_min
			, transaction_auto_withdrawal_window_max
			, otp_required_for_login
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM setting
		WHERE id = 1
		FOR UPDATE`,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.VacationMode,
			&response.VacationEndDate,
			&response.VoucherEnabled,
			&response.ReferralPointsEnabled,
			&response.ReferralPointsMinimumShoppingCost,
			&response.ReferralPointsAmount,
			&response.ReferralPointsSlots,
			&response.ReferralPointsExpense,
			&response.RewardPointsAccumulationWindowInDays,
			&response.DeliveryOriginLongitude,
			&response.DeliveryOriginLatitude,
			&response.DeliveryCostPerKM,
			&response.DeliveryMaxDistanceKM,
			&response.WhatsappPhoneNumber,
			&response.PaymentTerms,
			&response.MaximumDeliveryTime,
			&response.TransactionAutoWithdrawalEnabled,
			&response.TransactionAutoWithdrawalWindowMin,
			&response.TransactionAutoWithdrawalWindowMax,
			&response.OtpRequiredForLogin,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
		)...,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE setting SET")
	if response.VacationMode != request.VacationMode {
		params = append(params, request.VacationMode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" vacation_mode = ?")
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
		_, _ = sb.WriteString(" vacation_end_date = ?")
	}
	if response.VoucherEnabled != request.VoucherEnabled {
		params = append(params, request.VoucherEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" voucher_enabled = ?")
	}
	if response.ReferralPointsEnabled != request.ReferralPointsEnabled {
		params = append(params, request.ReferralPointsEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" referral_points_enabled = ?")
	}
	if response.ReferralPointsMinimumShoppingCost != request.ReferralPointsMinimumShoppingCost {
		params = append(params, request.ReferralPointsMinimumShoppingCost)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" referral_points_minimum_shopping_cost = ?")
	}
	if response.ReferralPointsAmount != request.ReferralPointsAmount {
		params = append(params, request.ReferralPointsAmount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" referral_points_amount = ?")
	}
	if slices.Compare(response.ReferralPointsSlots, request.ReferralPointsSlots) != 0 {
		if request.ReferralPointsSlots == nil {
			request.ReferralPointsSlots = []int64{}
		}
		params = append(params, request.ReferralPointsSlots)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" referral_points_slots = ?")
	}
	if response.RewardPointsAccumulationWindowInDays != request.RewardPointsAccumulationWindowInDays {
		params = append(params, request.RewardPointsAccumulationWindowInDays)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" reward_points_accumulation_window_in_days = ?")
	}
	if response.DeliveryOriginLongitude != request.DeliveryOriginLongitude {
		params = append(params, request.DeliveryOriginLongitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" delivery_origin_longitude = ?")
	}
	if response.DeliveryOriginLatitude != request.DeliveryOriginLatitude {
		params = append(params, request.DeliveryOriginLatitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" delivery_origin_latitude = ?")
	}
	if response.DeliveryCostPerKM != request.DeliveryCostPerKM {
		params = append(params, request.DeliveryCostPerKM)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" delivery_cost_per_km = ?")
	}
	if response.DeliveryMaxDistanceKM != request.DeliveryMaxDistanceKM {
		params = append(params, request.DeliveryMaxDistanceKM)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" delivery_max_distance_km = ?")
	}
	if response.WhatsappPhoneNumber != request.WhatsappPhoneNumber {
		params = append(params, request.WhatsappPhoneNumber)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" whatsapp_phone_number = ?")
	}
	if response.PaymentTerms != request.PaymentTerms {
		params = append(params, request.PaymentTerms)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" payment_terms = ?")
	}
	if response.MaximumDeliveryTime != request.MaximumDeliveryTime {
		params = append(params, request.MaximumDeliveryTime)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" maximum_delivery_time = ?")
	}
	if response.TransactionAutoWithdrawalEnabled != request.TransactionAutoWithdrawalEnabled {
		params = append(params, request.TransactionAutoWithdrawalEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" transaction_auto_withdrawal_enabled = ?")
	}
	var requestWindowMin,
		responseWindowMin,
		requestWindowMax,
		responseWindowMax time.Time
	if request.TransactionAutoWithdrawalWindowMin != nil {
		requestWindowMin = *request.TransactionAutoWithdrawalWindowMin
	}
	if response.TransactionAutoWithdrawalWindowMin != nil {
		responseWindowMin = *response.TransactionAutoWithdrawalWindowMin
	}
	if request.TransactionAutoWithdrawalWindowMax != nil {
		requestWindowMax = *request.TransactionAutoWithdrawalWindowMax
	}
	if response.TransactionAutoWithdrawalWindowMax != nil {
		responseWindowMax = *response.TransactionAutoWithdrawalWindowMax
	}
	if !responseWindowMin.Equal(requestWindowMin) ||
		!responseWindowMax.Equal(requestWindowMax) {
		params = append(
			params,
			request.TransactionAutoWithdrawalWindowMin,
			request.TransactionAutoWithdrawalWindowMax,
		)
		if n = len(params); n > 2 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" transaction_auto_withdrawal_window_min = ?, transaction_auto_withdrawal_window_max = ?")
	}
	if response.OtpRequiredForLogin != request.OtpRequiredForLogin {
		params = append(params, request.OtpRequiredForLogin)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" otp_required_for_login = ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.UpdatedBy, request.UpdatedAt, response.ID)
	_, _ = sb.WriteString(`, updated_by = ?
		, updated_at = ?
		WHERE id = ?
		RETURNING id
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, transaction_auto_withdrawal_window_min
			, transaction_auto_withdrawal_window_max
			, otp_required_for_login
			, created_by
			, created_at
			, updated_by
			, updated_at`)
	if err = tx.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.VacationMode,
			&response.VacationEndDate,
			&response.VoucherEnabled,
			&response.ReferralPointsEnabled,
			&response.ReferralPointsMinimumShoppingCost,
			&response.ReferralPointsAmount,
			&response.ReferralPointsSlots,
			&response.ReferralPointsExpense,
			&response.RewardPointsAccumulationWindowInDays,
			&response.DeliveryOriginLongitude,
			&response.DeliveryOriginLatitude,
			&response.DeliveryCostPerKM,
			&response.DeliveryMaxDistanceKM,
			&response.WhatsappPhoneNumber,
			&response.PaymentTerms,
			&response.MaximumDeliveryTime,
			&response.TransactionAutoWithdrawalEnabled,
			&response.TransactionAutoWithdrawalWindowMin,
			&response.TransactionAutoWithdrawalWindowMax,
			&response.OtpRequiredForLogin,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
		)...,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}
