package models

import (
	"errors"
	"time"
	"uuid"

	"github.com/govalues/decimal"
)

var (
	ErrSettingNotFound = errors.New("setting not found")
)

type (
	Setting struct {
		ID                                    int64           `json:"-" form:"id"`
		MaintenanceMode                       bool            `json:"maintenance_mode" form:"maintenance_mode"`
		VacationMode                          bool            `json:"vacation_mode" form:"vacation_mode"`
		VacationStartDate                     *time.Time      `json:"vacation_start_date" form:"vacation_start_date"`
		VacationEndDate                       *time.Time      `json:"vacation_end_date" form:"vacation_end_date"`
		VoucherEnabled                        bool            `json:"voucher_enabled" form:"voucher_enabled"`
		ReferralPointsEnabled                 bool            `json:"referral_points_enabled" form:"referral_points_enabled"`
		ReferralPointsMinimumShoppingCost     int64           `json:"referral_points_minimum_shopping_cost" form:"referral_points_minimum_shopping_cost"`
		ReferralPointsPerReferee              int64           `json:"referral_points_per_referee" form:"referral_points_per_referee"`
		ReferralPointsSlots                   []int64         `json:"referral_points_slots" form:"referral_points_slots"`
		ReferralPointsExpense                 int64           `json:"referral_points_expense" form:"referral_points_expense"`
		RewardPointsAccumulationWindowInDays  int16           `json:"reward_points_accumulation_window_in_days" form:"reward_points_accumulation_window_in_days"`
		DeliveryOriginLongitude               decimal.Decimal `json:"delivery_origin_longitude" form:"delivery_origin_longitude"`
		DeliveryOriginLatitude                decimal.Decimal `json:"delivery_origin_latitude" form:"delivery_origin_latitude"`
		DeliveryCostPerKM                     int64           `json:"delivery_cost_per_km" form:"delivery_cost_per_km"`
		DeliveryMaxDistanceKM                 int64           `json:"delivery_max_distance_km" form:"delivery_max_distance_km"`
		WhatsappPhoneNumber                   string          `json:"whatsapp_phone_number" form:"whatsapp_phone_number"`
		PaymentTerms                          time.Duration   `json:"-" form:"payment_terms"`
		RawPaymentTerms                       string          `json:"payment_terms" form:"-"`
		MaximumDeliveryTime                   time.Duration   `json:"-" form:"maximum_delivery_time"`
		RawMaximumDeliveryTime                string          `json:"maximum_delivery_time" form:"-"`
		TransactionAutoWithdrawalEnabled      bool            `json:"-" form:"transaction_auto_withdrawal_enabled"`
		TransactionAutoWithdrawalWindowMin    *time.Time      `json:"transaction_auto_withdrawal_window_min" form:"-"`
		RawTransactionAutoWithdrawalWindowMin *string         `json:"-" form:"transaction_auto_withdrawal_window_min"`
		TransactionAutoWithdrawalWindowMax    *time.Time      `json:"transaction_auto_withdrawal_window_max" form:"-"`
		RawTransactionAutoWithdrawalWindowMax *string         `json:"-" form:"transaction_auto_withdrawal_window_max"`
		CreatedBy                             uuid.UUID       `json:"-" form:"-"`
		CreatedAt                             time.Time       `json:"-" form:"-"`
		UpdatedBy                             uuid.UUID       `json:"-" form:"-"`
		UpdatedAt                             time.Time       `json:"-" form:"-"`
	}
)

func (q *Setting) Validate() error {
	now := time.Now().In(time.Local)
	if q.VacationMode {
		if q.VacationStartDate == nil {
			return errors.New("vacation_start_date: is required")
		}
		if q.VacationStartDate.Before(time.Now()) {
			return errors.New("vacation_start_date: should equal or exceed current date")
		}
		if q.VacationEndDate == nil {
			return errors.New("vacation_end_date: is required")
		}
		if q.VacationEndDate.Before(*q.VacationStartDate) {
			return errors.New("vacation_end_date: should equal or exceed vacation_start_date")
		}
	} else {
		q.VacationStartDate = nil
		q.VacationEndDate = nil
	}
	if q.PaymentTerms == 0 {
		return errors.New("payment_terms: is required")
	}
	if q.MaximumDeliveryTime == 0 {
		return errors.New("maximum_delivery_time: is required")
	}
	if q.TransactionAutoWithdrawalEnabled {
		if q.TransactionAutoWithdrawalWindowMin == nil {
			return errors.New("transaction_auto_withdrawal_window_min: is required")
		}
		if q.TransactionAutoWithdrawalWindowMax == nil {
			return errors.New("transaction_auto_withdrawal_window_max: is required")
		}
		if !q.TransactionAutoWithdrawalWindowMax.After(*q.TransactionAutoWithdrawalWindowMin) {
			return errors.New("transaction_auto_withdrawal_window_max: should exceed transaction_auto_withdrawal_window_min")
		}
		q.TransactionAutoWithdrawalWindowMin = new(time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.TransactionAutoWithdrawalWindowMin.Hour(),
			q.TransactionAutoWithdrawalWindowMin.Minute(),
			q.TransactionAutoWithdrawalWindowMin.Second(),
			q.TransactionAutoWithdrawalWindowMin.Nanosecond(),
			q.TransactionAutoWithdrawalWindowMin.Location(),
		))
		q.TransactionAutoWithdrawalWindowMax = new(time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.TransactionAutoWithdrawalWindowMax.Hour(),
			q.TransactionAutoWithdrawalWindowMax.Minute(),
			q.TransactionAutoWithdrawalWindowMax.Second(),
			q.TransactionAutoWithdrawalWindowMax.Nanosecond(),
			q.TransactionAutoWithdrawalWindowMax.Location(),
		))
	} else {
		q.TransactionAutoWithdrawalWindowMin = nil
		q.TransactionAutoWithdrawalWindowMax = nil
	}
	return nil
}

func (q *Setting) Populate() *Setting {
	q.RawPaymentTerms = q.PaymentTerms.String()
	q.RawMaximumDeliveryTime = q.MaximumDeliveryTime.String()
	if q.TransactionAutoWithdrawalWindowMin != nil {
		q.RawTransactionAutoWithdrawalWindowMin = new(q.TransactionAutoWithdrawalWindowMin.
			In(time.Local).
			Format(time.TimeOnly))
	}
	if q.TransactionAutoWithdrawalWindowMax != nil {
		q.RawTransactionAutoWithdrawalWindowMax = new(q.TransactionAutoWithdrawalWindowMax.
			In(time.Local).
			Format(time.TimeOnly))
	}
	return q
}
