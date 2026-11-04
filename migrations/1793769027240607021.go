package migrations

import (
	"context"
	"log/slog"
	"os"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793769027240607021] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793769027240607021"
		// companies
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE companies (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL COLLATE "natural_sort"
				, subdomain character varying NOT NULL UNIQUE
				, billed BOOLEAN NOT NULL
				, billing_interval smallint
				, subscription_status smallint NOT NULL
				, subscription_start_date timestamp with time zone
				, subscription_end_date timestamp with time zone
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
				, owner_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON companies (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON companies (subscription_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON companies (subscription_end_date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON companies (owner_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON companies (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON companies (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON companies (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`ALTER TABLE users
				ADD FOREIGN KEY (company_id) REFERENCES companies (id)
				ON UPDATE CASCADE ON DELETE RESTRICT;`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		var adminID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`SELECT id
			FROM users
			WHERE email = $1`,
			"benni.surbakti@gmail.com",
		).Scan(&adminID); err != nil {
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
		company := models.Company{
			Name:                                 "Laukpauk",
			Android:                              true,
			Subdomain:                            "laukpauk",
			Billed:                               false,
			SubscriptionStatus:                   models.SubscriptionStatusActive,
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
		var companyID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`INSERT INTO companies (
				name
				, subdomain
				, billed
				, subscription_status
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
				, transaction_auto_withdrawal_window
				, otp_required_for_login
				, owner_id
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
				, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
				, $21, tstzrange($22, $23, '[]'), $24, $25, $26, $27, $28, $29
			)
			RETURNING id`,
			company.Name,
			company.Subdomain,
			company.Billed,
			company.SubscriptionStatus,
			company.VacationMode,
			company.VacationEndDate,
			company.VoucherEnabled,
			company.ReferralPointsEnabled,
			company.ReferralPointsMinimumShoppingCost,
			company.ReferralPointsAmount,
			company.ReferralPointsSlots,
			company.ReferralPointsExpense,
			company.RewardPointsAccumulationWindowInDays,
			company.DeliveryOriginLongitude,
			company.DeliveryOriginLatitude,
			company.DeliveryCostPerKM,
			company.DeliveryMaxDistanceKM,
			company.WhatsappPhoneNumber,
			company.PaymentTerms,
			company.MaximumDeliveryTime,
			company.TransactionAutoWithdrawalEnabled,
			company.TransactionAutoWithdrawalWindowMin,
			company.TransactionAutoWithdrawalWindowMax,
			company.OtpRequiredForLogin,
			adminID,
			adminID,
			company.CreatedAt,
			adminID,
			company.UpdatedAt,
		).Scan(&companyID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		password, err := helper.Base64Decode(os.Getenv("DEFAULT_PASSWORD"))
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrBase64Decode")
			return
		}
		encryptedPassword, err := helper.HashPassword(password)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrHashPassword")
			return
		}
		var sellerID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`INSERT INTO users (
				role
				, status
				, name
				, mobile_phone_confirmed_at
				, mobile_phone
				, encrypted_password
				, photo
				, user_level
				, company_id
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id`,
			models.RoleSeller,
			models.UserStatusActive,
			company.Name,
			now,
			company.WhatsappPhoneNumber,
			encryptedPassword,
			"seller-default.png",
			models.SellerLevelOwner,
			companyID,
			now,
			now,
		).Scan(&sellerID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`UPDATE companies SET
				owner_id = $1
			WHERE id = $2`,
			sellerID,
			companyID,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
