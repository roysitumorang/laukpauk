package migrations

import (
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793797813258545236] = Migration{
		Version: 1793797813258545236,
		Up: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Up"
			// setting
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE setting (
					id bigint NOT NULL PRIMARY KEY DEFAULT 1 CHECK (id = 1)
					, vacation_mode BOOLEAN NOT NULL
					, vacation_end_date DATETIME(6)
					, voucher_enabled BOOLEAN NOT NULL
					, referral_points_enabled BOOLEAN NOT NULL
					, referral_points_minimum_shopping_cost bigint NOT NULL
					, referral_points_amount bigint NOT NULL
					, referral_points_slots JSON NOT NULL
					, referral_points_expense bigint NOT NULL
					, reward_points_accumulation_window_in_days smallint NOT NULL
					, delivery_origin_longitude TEXT NOT NULL
					, delivery_origin_latitude TEXT NOT NULL
					, delivery_cost_per_km bigint NOT NULL
					, delivery_max_distance_km bigint NOT NULL
					, whatsapp_phone_number VARCHAR(255) NOT NULL
					, payment_terms bigint NOT NULL
					, maximum_delivery_time bigint NOT NULL
					, transaction_auto_withdrawal_enabled BOOLEAN NOT NULL
					, transaction_auto_withdrawal_window_min DATETIME(6)
					, transaction_auto_withdrawal_window_max DATETIME(6)
					, otp_required_for_login BOOLEAN NOT NULL
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, CONSTRAINT setting_whatsapp_phone_number_key UNIQUE (whatsapp_phone_number)
					, CONSTRAINT setting_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT setting_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX setting_updated_by_idx ON setting (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			var sellerID uuid.UUID
			if err = tx.QueryRow(
				ctx,
				`SELECT id
				FROM users
				WHERE email = ?`,
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
				PaymentTerms:                         models.Duration(24 * time.Hour),
				MaximumDeliveryTime:                  models.Duration(4 * time.Hour),
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
					, transaction_auto_withdrawal_window_min
					, transaction_auto_withdrawal_window_max
					, otp_required_for_login
					, created_by
					, created_at
					, updated_by
					, updated_at
				) VALUES (
					?, ?, ?, ?, ?, ?, ?, ?, ?, ?
					, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
					, ?, ?, ?, ?
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
		},
		Down: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE setting",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
