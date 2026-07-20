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
		ID                                   int64           `json:"-" form:"id"`
		MaintenanceMode                      bool            `json:"maintenance_mode" form:"maintenance_mode"`
		VoucherEnabled                       bool            `json:"voucher_enabled" form:"voucher_enabled"`
		ReferralPointsEnabled                bool            `json:"referral_points_enabled" form:"referral_points_enabled"`
		ReferralPointsBudget                 int64           `json:"referral_points_budget" form:"referral_points_budget"`
		ReferralPointsPerReferee             int64           `json:"referral_points_per_referee" form:"referral_points_per_referee"`
		ReferralPointsSlots                  []int64         `json:"referral_points_slots" form:"referral_points_slots"`
		RewardPointsAccumulationWindowInDays int16           `json:"reward_points_accumulation_window_in_days" form:"reward_points_accumulation_window_in_days"`
		DeliveryOriginLongitude              decimal.Decimal `json:"delivery_origin_longitude" form:"delivery_origin_longitude"`
		DeliveryOriginLatitude               decimal.Decimal `json:"delivery_origin_latitude" form:"delivery_origin_latitude"`
		DeliveryCostPerKM                    int64           `json:"delivery_cost_per_km" form:"delivery_cost_per_km"`
		DeliveryMaxDistanceKM                int64           `json:"delivery_max_distance_km" form:"delivery_max_distance_km"`
		WhatsappPhoneNumber                  string          `json:"whatsapp_phone_number" form:"whatsapp_phone_number"`
		CreatedAt                            time.Time       `json:"-" form:"-"`
		UpdatedBy                            *uuid.UUID      `json:"-" form:"-"`
		UpdatedAt                            time.Time       `json:"-" form:"-"`
	}
)

func (q *Setting) Validate() error {
	return nil
}
