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
		ID                                          int64           `json:"-" form:"id"`
		MaintenanceMode                             bool            `json:"maintenance_mode" form:"maintenance_mode"`
		VoucherEnabled                              bool            `json:"voucher_enabled" form:"voucher_enabled"`
		ReferralPointsEnabled                       bool            `json:"referral_points_enabled" form:"referral_points_enabled"`
		ReferralPointsInitialBudget                 int64           `json:"referral_points_initial_budget" form:"referral_points_initial_budget"`
		RefferalPointsMinimumShoppingCost           int64           `json:"refferal_points_minimum_shopping_cost" form:"refferal_points_minimum_shopping_cost"`
		ReferralPointsPerReferee                    int64           `json:"referral_points_per_referee" form:"referral_points_per_referee"`
		ReferralPointsSlots                         []int64         `json:"referral_points_slots" form:"referral_points_slots"`
		ReferralPointsRemainingBudget               int64           `json:"referral_points_remaining_budget" form:"referral_points_remaining_budget"`
		RewardPointsAccumulationWindowInDays        int16           `json:"reward_points_accumulation_window_in_days" form:"reward_points_accumulation_window_in_days"`
		DeliveryOriginLongitude                     decimal.Decimal `json:"delivery_origin_longitude" form:"delivery_origin_longitude"`
		DeliveryOriginLatitude                      decimal.Decimal `json:"delivery_origin_latitude" form:"delivery_origin_latitude"`
		DeliveryCostPerKM                           int64           `json:"delivery_cost_per_km" form:"delivery_cost_per_km"`
		DeliveryMaxDistanceKM                       int64           `json:"delivery_max_distance_km" form:"delivery_max_distance_km"`
		WhatsappPhoneNumber                         string          `json:"whatsapp_phone_number" form:"whatsapp_phone_number"`
		WeekdayTransactionCutOffTime                time.Time       `json:"-" form:"-"`
		WeekdaySameDay                              bool            `json:"-" form:"-"`
		WeekdaySameDayTransactionWindowMax          time.Time       `json:"-" form:"-"`
		WeekdayNextDay                              bool            `json:"-" form:"-"`
		WeekdayUntilCutOffNextDayDeliveryStartTime  time.Time       `json:"-" form:"-"`
		WeekdayAfterCutOffNextDayDeliveryStartTime  time.Time       `json:"-" form:"-"`
		WeekdayNextTwoDay                           bool            `json:"-" form:"-"`
		WeekdayNextTwoDayDeliveryStartTime          time.Time       `json:"-" form:"-"`
		WeekdayDeliveryEndTime                      time.Time       `json:"-" form:"-"`
		WeekdayDeliveryInterval                     string          `json:"-" form:"-"`
		SaturdayTransactionCutOffTime               time.Time       `json:"-" form:"-"`
		SaturdaySameDay                             bool            `json:"-" form:"-"`
		SaturdaySameDayTransactionWindowMax         time.Time       `json:"-" form:"-"`
		SaturdayNextDay                             bool            `json:"-" form:"-"`
		SaturdayUntilCutOffNextDayDeliveryStartTime time.Time       `json:"-" form:"-"`
		SaturdayAfterCutOffNextDayDeliveryStartTime time.Time       `json:"-" form:"-"`
		SaturdayNextTwoDay                          bool            `json:"-" form:"-"`
		SaturdayNextTwoDayDeliveryStartTime         time.Time       `json:"-" form:"-"`
		SaturdayDeliveryEndTime                     time.Time       `json:"-" form:"-"`
		SaturdayDeliveryInterval                    string          `json:"-" form:"-"`
		SundayTransactionCutOffTime                 time.Time       `json:"-" form:"-"`
		SundaySameDay                               bool            `json:"-" form:"-"`
		SundaySameDayTransactionWindowMax           time.Time       `json:"-" form:"-"`
		SundayNextDay                               bool            `json:"-" form:"-"`
		SundayUntilCutOffNextDayDeliveryStartTime   time.Time       `json:"-" form:"-"`
		SundayAfterCutOffNextDayDeliveryStartTime   time.Time       `json:"-" form:"-"`
		SundayNextTwoDay                            bool            `json:"-" form:"-"`
		SundayNextTwoDayDeliveryStartTime           time.Time       `json:"-" form:"-"`
		SundayDeliveryEndTime                       time.Time       `json:"-" form:"-"`
		SundayDeliveryInterval                      string          `json:"-" form:"-"`
		SkipSundayDelivery                          bool            `json:"-" form:"-"`
		CreatedAt                                   time.Time       `json:"-" form:"-"`
		UpdatedBy                                   *uuid.UUID      `json:"-" form:"-"`
		UpdatedAt                                   time.Time       `json:"-" form:"-"`
	}
)

func (q *Setting) Validate() error {
	return nil
}
