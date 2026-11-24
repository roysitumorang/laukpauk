package models

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/nyaruka/phonenumbers"
)

var ErrSettingNotFound = errors.New("setting not found")

type Setting struct {
	ID                                   uint64          `json:"-"`
	VacationMode                         bool            `json:"vacation_mode"`
	VacationEndDate                      *time.Time      `json:"vacation_end_date"`
	VoucherEnabled                       bool            `json:"voucher_enabled"`
	ReferralPointsEnabled                bool            `json:"referral_points_enabled"`
	ReferralPointsMinimumShoppingCost    int64           `json:"referral_points_minimum_shopping_cost"`
	ReferralPointsAmount                 int64           `json:"referral_points_amount"`
	ReferralPointsSlots                  []int64         `json:"referral_points_slots"`
	ReferralPointsExpense                int64           `json:"referral_points_expense"`
	RewardPointsAccumulationWindowInDays int16           `json:"reward_points_accumulation_window_in_days"`
	DeliveryOriginLongitude              decimal.Decimal `json:"delivery_origin_longitude"`
	DeliveryOriginLatitude               decimal.Decimal `json:"delivery_origin_latitude"`
	DeliveryCostPerKM                    int64           `json:"delivery_cost_per_km"`
	DeliveryMaxDistanceKM                int64           `json:"delivery_max_distance_km"`
	WhatsappPhoneNumber                  string          `json:"whatsapp_phone_number"`
	PaymentTerms                         Duration        `json:"payment_terms"`
	MaximumDeliveryTime                  Duration        `json:"maximum_delivery_time"`
	TransactionAutoWithdrawalEnabled     bool            `json:"-"`
	TransactionAutoWithdrawalWindowMin   *time.Time      `json:"transaction_auto_withdrawal_window_min"`
	TransactionAutoWithdrawalWindowMax   *time.Time      `json:"transaction_auto_withdrawal_window_max"`
	OtpRequiredForLogin                  bool            `json:"otp_required_for_login"`
	CreatedBy                            uuid.UUID       `json:"-"`
	CreatedAt                            time.Time       `json:"-"`
	UpdatedBy                            uuid.UUID       `json:"-"`
	UpdatedAt                            time.Time       `json:"-"`
}

func (q *Setting) Validate() error {
	if q.WhatsappPhoneNumber = strings.TrimSpace(q.WhatsappPhoneNumber); q.WhatsappPhoneNumber == "" {
		return errors.New("whatsapp_phone_number: is required")
	}
	whatsappPhoneNumber, err := phonenumbers.Parse(q.WhatsappPhoneNumber, "ID")
	if err != nil {
		return fmt.Errorf("whatsapp_phone_number: %w", err)
	}
	q.WhatsappPhoneNumber = phonenumbers.Format(whatsappPhoneNumber, phonenumbers.E164)
	now := time.Now().In(time.Local)
	if q.VacationMode {
		if q.VacationEndDate == nil {
			return errors.New("vacation_end_date: is required")
		}
		if !q.VacationEndDate.After(now) {
			return errors.New("vacation_end_date: should equal or exceed current date")
		}
	} else {
		q.VacationEndDate = nil
	}
	if q.ReferralPointsEnabled {
		if q.ReferralPointsAmount > 0 && len(q.ReferralPointsSlots) == 0 {
			return errors.New("referral_points_slots: is required")
		}
		var referralPointsSlots int64
		for i, referralPointSlot := range q.ReferralPointsSlots {
			if referralPointSlot < 1 {
				return fmt.Errorf("referral_points_slots[%d]: should exceed 0", i)
			}
			referralPointsSlots += referralPointSlot
		}
		if q.ReferralPointsAmount != referralPointsSlots {
			return errors.New("referral_points_slots: should match referral_points_amount")
		}
	}
	if q.PaymentTerms.Duration < 1 {
		return errors.New("payment_terms: should exceed 0")
	}
	if q.MaximumDeliveryTime.Duration < 1 {
		return errors.New("maximum_delivery_time: should exceed 0")
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
	} else {
		q.TransactionAutoWithdrawalWindowMin = nil
		q.TransactionAutoWithdrawalWindowMax = nil
	}
	return nil
}

func (q *Setting) Populate() *Setting {
	return q
}
