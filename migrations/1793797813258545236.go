package migrations

import (
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793797813258545236] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793797813258545236"
		// setting
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE setting (
				id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY CHECK (id = 1)
				, vacation_mode BOOLEAN NOT NULL
				, vacation_end_date timestamp with time zone
				, voucher_enabled BOOLEAN NOT NULL
				, referral_points_enabled BOOLEAN NOT NULL
				, referral_points_minimum_shopping_cost bigint NOT NULL
				, referral_points_amount bigint NOT NULL
				, referral_points_slots bigint[] NOT NULL
				, referral_points_expense bigint NOT NULL
				, reward_points_accumulation_window_in_days smallint NOT NULL
				, delivery_origin_longitude character varying NOT NULL
				, delivery_origin_latitude character varying NOT NULL
				, delivery_cost_per_km bigint NOT NULL
				, delivery_max_distance_km bigint NOT NULL
				, whatsapp_phone_number character varying NOT NULL UNIQUE
				, payment_terms interval NOT NULL
				, maximum_delivery_time interval NOT NULL
				, transaction_auto_withdrawal_enabled boolean NOT NULL
				, transaction_auto_withdrawal_window TSTZRANGE NOT NULL
				, otp_required_for_login boolean NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON setting (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		var sellerID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`SELECT id
			FROM users
			WHERE email = $1`,
			"benni.surbakti@gmail.com",
		).Scan(&sellerID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		now := time.Now().In(time.Local)
		longitude, err := decimal.Parse("98.64455685385735")
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrParse")
			return
		}
		latitude, err := decimal.Parse("3.5394011973709434")
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrParse")
			return
		}
		setting := models.Setting{
			VoucherEnabled:                       true,
			ReferralPointsEnabled:                true,
			ReferralPointsMinimumShoppingCost:    60_000,
			ReferralPointsAmount:                 25_000,
			ReferralPointsSlots:                  []int64{10_000, 15_000},
			RewardPointsAccumulationWindowInDays: 30,
			DeliveryOriginLongitude:              longitude,
			DeliveryOriginLatitude:               latitude,
			DeliveryCostPerKM:                    3_000,
			DeliveryMaxDistanceKM:                15,
			WhatsappPhoneNumber:                  "+628116099887",
			PaymentTerms:                         models.Duration{Duration: 24 * time.Hour},
			MaximumDeliveryTime:                  models.Duration{Duration: 4 * time.Hour},
			TransactionAutoWithdrawalEnabled:     true,
			TransactionAutoWithdrawalWindowMin: new(time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				10,
				0,
				0,
				0,
				now.Location(),
			)),
			TransactionAutoWithdrawalWindowMax: new(time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				21,
				59,
				59,
				0,
				now.Location(),
			)),
			CreatedAt: time.Date(
				2021,
				12,
				16,
				17,
				26,
				51,
				0,
				now.Location(),
			),
			UpdatedAt: time.Date(
				2021,
				12,
				16,
				17,
				26,
				51,
				0,
				now.Location(),
			),
		}
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO setting (
				vacation_mode
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
				, transaction_auto_withdrawal_window
				, otp_required_for_login
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
				, $11, $12, $13, $14, $15, $16, $17, tstzrange($18, $19), $20
				, $21, $22, $23, $24
			);`,
			setting.VacationMode,
			setting.VacationEndDate,
			setting.VoucherEnabled,
			setting.ReferralPointsEnabled,
			setting.ReferralPointsMinimumShoppingCost,
			setting.ReferralPointsAmount,
			setting.ReferralPointsSlots,
			setting.ReferralPointsExpense,
			setting.RewardPointsAccumulationWindowInDays,
			setting.DeliveryOriginLongitude,
			setting.DeliveryOriginLatitude,
			setting.DeliveryCostPerKM,
			setting.DeliveryMaxDistanceKM,
			setting.WhatsappPhoneNumber,
			setting.PaymentTerms,
			setting.MaximumDeliveryTime,
			setting.TransactionAutoWithdrawalEnabled,
			setting.TransactionAutoWithdrawalWindowMin,
			setting.TransactionAutoWithdrawalWindowMax,
			setting.OtpRequiredForLogin,
			sellerID,
			now,
			sellerID,
			now,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
