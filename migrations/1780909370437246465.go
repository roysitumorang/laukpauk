package migrations

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1780909370437246465] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1780909370437246465"
		if _, err = tx.Exec(
			ctx,
			"CREATE EXTENSION IF NOT EXISTS btree_gist;",
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE COLLATION IF NOT EXISTS natural_sort (
				PROVIDER = 'icu'
				, LOCALE = 'en-US-u-kn-true'
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// admins
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE admins (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, status integer NOT NULL
				, admin_level integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, unconfirmed_mobile_phone character varying
				, mobile_phone_confirmation_sent_at timestamp with time zone
				, mobile_phone_confirmation_token character varying
				, mobile_phone_confirmed_at timestamp with time zone
				, mobile_phone character varying NOT NULL UNIQUE
				, username character varying NOT NULL UNIQUE
				, unconfirmed_email character varying
				, email_confirmation_sent_at timestamp with time zone
				, email_confirmation_token character varying
				, email_confirmed_at timestamp with time zone
				, email character varying NOT NULL UNIQUE
				, encrypted_password character varying
				, password_reset_at timestamp with time zone
				, password_reset_token character varying
				, password_changed_at timestamp with time zone
				, photo character varying NOT NULL
				, unread_notifications bigint NOT NULL DEFAULT 0
				, created_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON admins (status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON admins (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON admins (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON admins (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON admins (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
		var adminID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`INSERT INTO admins (
				status
				, admin_level
				, name
				, mobile_phone_confirmed_at
				, mobile_phone
				, username
				, email_confirmed_at
				, email
				, encrypted_password
				, photo
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id`,
			models.AdminStatusActive,
			models.AdminLevelSuperAdmin,
			"Ben",
			"2024-12-11 07:41:55.000 +0700",
			"+6281265688889",
			"bigben",
			"2024-12-11 07:41:55.000 +0700",
			"benni.surbakti@gmail.com",
			encryptedPassword,
			"admin-default.png",
			"2021-12-16 17:26:51.000 +0700",
			"2024-12-11 07:41:55.000 +0700",
		).Scan(&adminID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		// companies
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE companies (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL COLLATE "natural_sort"
				, billed BOOLEAN NOT NULL
				, billing_interval integer
				, subscription_status integer NOT NULL
				, subscription_start_date timestamp with time zone
				, subscription_end_date timestamp with time zone
				, telegram_bot_token character varying NOT NULL UNIQUE
				, vacation_mode BOOLEAN NOT NULL
				, vacation_start_date timestamp with time zone
				, vacation_end_date timestamp with time zone
				, voucher_enabled BOOLEAN NOT NULL
				, referral_points_enabled BOOLEAN NOT NULL
				, referral_points_minimum_shopping_cost bigint NOT NULL
				, referral_points_per_referee bigint NOT NULL
				, referral_points_slots bigint[] NOT NULL
				, referral_points_expense bigint NOT NULL
				, reward_points_accumulation_window_in_days smallint NOT NULL
				, delivery_origin_longitude character varying NOT NULL
				, delivery_origin_latitude character varying NOT NULL
				, delivery_cost_per_km bigint NOT NULL
				, delivery_max_distance_km bigint NOT NULL
				, whatsapp_phone_number character varying NOT NULL
				, payment_terms interval NOT NULL
				, maximum_delivery_time interval NOT NULL
				, transaction_auto_withdrawal_enabled boolean NOT NULL
				, transaction_auto_withdrawal_window TSTZRANGE NOT NULL
				, owner_id UUID NOT NULL
				, created_by UUID NOT NULL REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
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
		now := time.Now().In(time.Local)
		company := models.Company{
			ReferralPointsSlots: []int64{},
		}
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO companies (
				id
				, name
				, billed
				, subscription_status
				, telegram_bot_token
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
				, whatsapp_phone_number
				, payment_terms
				, maximum_delivery_time
				, transaction_auto_withdrawal_enabled
				, transaction_auto_withdrawal_window
				, owner_id
				, created_by
				, created_at
				, updated_by
				, updated_at
				, deactivated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
				, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
				, $21, $22, $23, tstzrange($24, $25, '[]'), $26, $27, $28, $29, $30
				, $31
			)`,
			uuid.Nil(),
			company.Name,
			company.Billed,
			company.SubscriptionStatus,
			company.TelegramBotToken,
			company.VacationMode,
			company.VacationStartDate,
			company.VacationEndDate,
			company.VoucherEnabled,
			company.ReferralPointsEnabled,
			company.ReferralPointsMinimumShoppingCost,
			company.ReferralPointsPerReferee,
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
			uuid.Nil(),
			adminID,
			now,
			uuid.Nil(),
			now,
			now,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		company = models.Company{
			Name:                                 "Laukpauk",
			Billed:                               false,
			SubscriptionStatus:                   models.SubscriptionStatusActive,
			TelegramBotToken:                     "7895135905:AAGh8uCKeDniAQdJ0sh2PZh1YO2afkAto5Y",
			VoucherEnabled:                       true,
			ReferralPointsEnabled:                true,
			ReferralPointsMinimumShoppingCost:    60_000,
			ReferralPointsPerReferee:             25_000,
			ReferralPointsSlots:                  []int64{10_000, 15_000},
			RewardPointsAccumulationWindowInDays: 30,
			DeliveryOriginLongitude:              longitude,
			DeliveryOriginLatitude:               latitude,
			DeliveryCostPerKM:                    3_000,
			DeliveryMaxDistanceKM:                15,
			WhatsappPhoneNumber:                  "+628116099887",
			PaymentTerms:                         24 * time.Hour,
			MaximumDeliveryTime:                  4 * time.Hour,
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
		}
		var companyID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`INSERT INTO companies (
				name
				, billed
				, subscription_status
				, telegram_bot_token
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
				, whatsapp_phone_number
				, payment_terms
				, maximum_delivery_time
				, transaction_auto_withdrawal_enabled
				, transaction_auto_withdrawal_window
				, owner_id
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
				, $11, $12, $13, $14, $15, $16, $17, $18, $19
				, $20, $21, $22, tstzrange($23, $24, '[]'), $25, $26, $27, $28, $29
			)
			RETURNING id`,
			company.Name,
			company.Billed,
			company.SubscriptionStatus,
			company.TelegramBotToken,
			company.VacationMode,
			company.VacationStartDate,
			company.VacationEndDate,
			company.VoucherEnabled,
			company.ReferralPointsEnabled,
			company.ReferralPointsMinimumShoppingCost,
			company.ReferralPointsPerReferee,
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
			uuid.Nil(),
			adminID,
			now,
			uuid.Nil(),
			now,
		).Scan(&companyID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		// sellers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE sellers (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, status integer NOT NULL
				, seller_level integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, unconfirmed_mobile_phone character varying
				, mobile_phone_confirmation_sent_at timestamp with time zone
				, mobile_phone_confirmation_token character varying
				, mobile_phone_confirmed_at timestamp with time zone
				, mobile_phone character varying NOT NULL
				, otp_hash character varying NOT NULL
				, otp_expired_at timestamp with time zone NOT NULL
				, username character varying NOT NULL
				, unconfirmed_email character varying
				, email_confirmation_sent_at timestamp with time zone
				, email_confirmation_token character varying
				, email_confirmed_at timestamp with time zone
				, email character varying
				, photo character varying NOT NULL
				, unread_notifications bigint NOT NULL DEFAULT 0
				, telegram_chat_id bigint UNIQUE
				, telegram_username character varying
				, created_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
				, UNIQUE (mobile_phone, company_id)
				, UNIQUE (username, company_id)
				, UNIQUE (email, company_id)
				, UNIQUE (telegram_username, company_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO sellers (
				id
				, company_id
				, status
				, seller_level
				, name
				, mobile_phone_confirmed_at
				, mobile_phone
				, otp_hash
				, otp_expired_at
				, username
				, photo
				, created_at
				, updated_at
				, deactivated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
			uuid.Nil(),
			uuid.Nil(),
			models.SellerStatusDeactivated,
			models.SellerLevelOwner,
			"",
			nil,
			"",
			"",
			time.Unix(0, 0),
			"",
			"seller-default.png",
			now,
			now,
			now,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		var sellerID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`INSERT INTO sellers (
				company_id
				, status
				, seller_level
				, name
				, mobile_phone_confirmed_at
				, mobile_phone
				, otp_hash
				, otp_expired_at
				, username
				, photo
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id`,
			companyID,
			models.SellerStatusActive,
			models.SellerLevelOwner,
			company.Name,
			now,
			company.WhatsappPhoneNumber,
			"",
			time.Unix(0, 0),
			strings.ToLower(company.Name),
			"seller-default.png",
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
				, updated_by = $2
			WHERE id = $3`,
			sellerID,
			sellerID,
			companyID,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`ALTER TABLE companies
				ADD FOREIGN KEY (owner_id) REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, ADD FOREIGN KEY (updated_by) REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// buyers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE buyers (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, status integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, unconfirmed_mobile_phone character varying
				, mobile_phone_confirmation_sent_at timestamp with time zone
				, mobile_phone_confirmation_token character varying
				, mobile_phone_confirmed_at timestamp with time zone
				, mobile_phone character varying
				, otp_hash character varying NOT NULL
				, otp_expired_at timestamp with time zone NOT NULL
				, username character varying NOT NULL
				, unconfirmed_email character varying
				, email_confirmation_sent_at timestamp with time zone
				, email_confirmation_token character varying
				, email_confirmed_at timestamp with time zone
				, email character varying
				, photo character varying NOT NULL
				, unread_notifications bigint NOT NULL DEFAULT 0
				, telegram_chat_id bigint UNIQUE
				, telegram_username character varying
				, referral_code character varying NOT NULL
				, address character varying NOT NULL
				, referrer_id UUID REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, referral_points bigint NOT NULL DEFAULT 0
				, reward_points bigint NOT NULL DEFAULT 0
				, wallet bigint NOT NULL DEFAULT 0
				, transaction_volume bigint NOT NULL DEFAULT 0
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
				, UNIQUE (mobile_phone, company_id)
				, UNIQUE (username, company_id)
				, UNIQUE (email, company_id)
				, UNIQUE (telegram_username, company_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON buyers (LOWER(referral_code), company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (referrer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// product_categories
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE product_categories (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, record_status integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, description character varying NOT NULL
				, photo character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX product_categories_unique_name_idx ON product_categories (LOWER(name), company_id) WHERE record_status = 0;`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// products
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE products (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, product_category_id UUID REFERENCES product_categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_category_name character varying NOT NULL COLLATE "natural_sort"
				, product_type integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, photo character varying NOT NULL
				, description character varying NOT NULL
				, ready_stock BOOLEAN NOT NULL
				, pinned BOOLEAN NOT NULL DEFAULT FALSE
				, hits bigint NOT NULL
				, lowest_selling_price bigint NOT NULL DEFAULT 0
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (product_category_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (product_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX product_categories_unique_idx ON products (LOWER(name)) WHERE record_status = 0;`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (ready_stock);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (pinned);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (hits);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// product_variants
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE product_variants (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, product_category_id UUID NOT NULL REFERENCES product_categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_category_name character varying NOT NULL COLLATE "natural_sort"
				, product_id UUID NOT NULL REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_name character varying NOT NULL COLLATE "natural_sort"
				, product_ready_stock bool NOT NULL
				, stock_unit character varying NOT NULL COLLATE "natural_sort"
				, base_price bigint NOT NULL
				, selling_price bigint NOT NULL
				, discount_type integer NOT NULL
				, discount_rate bigint NOT NULL
				, discount_minimum_quantity bigint NOT NULL
				, discount_amount bigint NOT NULL
				, sale_price bigint NOT NULL
				, description character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (product_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (LOWER(stock_unit));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON product_variants (product_id, LOWER(stock_unit)) WHERE record_status = 0;`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// notification_targets
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE notification_targets (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL
				, firebase_installation_id character varying NOT NULL UNIQUE
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_targets (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// notification_templates
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE notification_templates (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, photo character varying
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (LOWER(subject));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (LOWER(body));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// notifications
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE notifications (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL
				, notification_type character varying NOT NULL
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, photo character varying
				, params JSONB NOT NULL
				, created_by UUID NOT NULL
				, created_at timestamp with time zone NOT NULL
				, sent_at timestamp with time zone
				, read_at timestamp with time zone
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notifications (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notifications (notification_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notifications (sent_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notifications (read_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// setting
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE setting (
				id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY CHECK (id = 1)
				, maintenance_mode BOOLEAN NOT NULL
				, subscription_trial_period integer NOT NULL
				, monthly_billing_amount bigint NOT NULL
				, annual_billing_amount bigint NOT NULL
				, billing_grace_period integer NOT NULL
				, created_by UUID NOT NULL REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
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
		setting := models.Setting{
			SubscriptionTrialPeriod: 7,
			MonthlyBillingAmount:    50_000,
			AnnualBillingAmount:     500_000,
			BillingGracePeriod:      7,
		}
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO setting (
				maintenance_mode
				, subscription_trial_period
				, monthly_billing_amount
				, annual_billing_amount
				, billing_grace_period
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
			setting.MaintenanceMode,
			setting.SubscriptionTrialPeriod,
			setting.MonthlyBillingAmount,
			setting.AnnualBillingAmount,
			setting.BillingGracePeriod,
			adminID,
			now,
			adminID,
			now,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// text_messages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE text_messages (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, mobile_phone character varying NOT NULL
				, message_type integer NOT NULL
				, body character varying NOT NULL
				, delivery_status character varying
				, message_id bigint
				, created_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, sent_at timestamp with time zone
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (message_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (delivery_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (message_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// delivery_plans
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE delivery_plans (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, name character varying COLLATE "natural_sort"
				, weekday bigint NOT NULL
				, transaction_window TSTZRANGE NOT NULL
				, delivery_time_interval interval NOT NULL
				, same_day BOOLEAN NOT NULL
				, same_day_window TSTZRANGE NOT NULL
				, next_day BOOLEAN NOT NULL
				, next_day_window TSTZRANGE NOT NULL
				, next_two_day BOOLEAN NOT NULL
				, next_two_day_window TSTZRANGE NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, CONSTRAINT no_overlapping_delivery_plans_transaction_window EXCLUDE USING gist (
					weekday WITH =,
					transaction_window WITH &&
				)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (weekday);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		deliveryTimeInterval := 2 * time.Hour
		today000000 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today095959 := time.Date(now.Year(), now.Month(), now.Day(), 9, 59, 59, 0, time.Local).Format(time.RFC3339Nano)
		today100000 := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today120000 := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today215959 := time.Date(now.Year(), now.Month(), now.Day(), 21, 59, 59, 0, time.Local).Format(time.RFC3339Nano)
		today220000 := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today235959 := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local).Format(time.RFC3339Nano)
		nextDay080000 := time.Date(now.Year(), now.Month(), now.Day()+1, 8, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextDay100000 := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextDay120000 := time.Date(now.Year(), now.Month(), now.Day()+1, 12, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextTwoDay080000 := time.Date(now.Year(), now.Month(), now.Day()+2, 8, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextTwoDay120000 := time.Date(now.Year(), now.Month(), now.Day()+2, 12, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		deliveryPlans := [][]any{
			// weekdays
			{
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 10:00",
				models.DeliveryWeekdayWeekdays,
				today000000,
				today095959,
				deliveryTimeInterval,
				true,
				today100000,
				today120000,
				true,
				nextDay080000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi jam 10:00 - 21:59",
				models.DeliveryWeekdayWeekdays,
				today100000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdayWeekdays,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay100000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			},
			// friday
			{
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 10:00",
				models.DeliveryWeekdayFriday,
				today000000,
				today095959,
				deliveryTimeInterval,
				true,
				today100000,
				today120000,
				true,
				nextDay080000,
				nextDay120000,
				false,
				nil,
				nil,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi jam 10:00 - 21:59",
				models.DeliveryWeekdayFriday,
				today100000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				false,
				nil,
				nil,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdayFriday,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				false,
				nil,
				nil,
				sellerID,
				now,
				sellerID,
				now,
			},
			// saturday
			{
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 10:00",
				models.DeliveryWeekdaySaturday,
				today000000,
				today095959,
				deliveryTimeInterval,
				true,
				today100000,
				today120000,
				false,
				nil,
				nil,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi jam 10:00 - 21:59",
				models.DeliveryWeekdaySaturday,
				today100000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				false,
				nil,
				nil,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdaySaturday,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				false,
				nil,
				nil,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			},
			// sunday
			{
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 22:00",
				models.DeliveryWeekdaySunday,
				today000000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdaySunday,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay100000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			},
		}
		for i := range deliveryPlans {
			if _, err = tx.Exec(
				ctx,
				`INSERT INTO delivery_plans (
					record_status
					, name
					, weekday
					, transaction_window
					, delivery_time_interval
					, same_day
					, same_day_window
					, next_day
					, next_day_window
					, next_two_day
					, next_two_day_window
					, created_by
					, created_at
					, updated_by
					, updated_at
				) VALUES (
					$1
					, $2
					, $3
					, tstzrange($4, $5, '[]')
					, $6
					, $7
					, tstzrange($8, $9)
					, $10
					, tstzrange($11, $12, '[]')
					, $13
					, tstzrange($14, $15, '[]')
					, $16
					, $17
					, $18
					, $19
				)`,
				deliveryPlans[i]...,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
		}
		// vouchers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE vouchers (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, voucher_type integer NOT NULL
				, target_audience integer NOT NULL
				, code character varying
				, deleted_code character varying
				, photo character varying NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, discount bigint NOT NULL
				, minimum_purchase bigint NOT NULL
				, quota_per_buyer bigint NOT NULL
				, effective_start_time timestamp with time zone
				, effective_end_time timestamp with time zone
				, effective_start_date timestamp with time zone NOT NULL
				, effective_end_date timestamp with time zone NOT NULL
				, description character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (voucher_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (target_audience);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX vouchers_unique_code_idx ON vouchers (LOWER(code));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (LOWER(deleted_code));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (effective_start_time);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (effective_end_time);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (effective_start_date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (effective_end_date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// vouchers_buyers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE vouchers_buyers (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, voucher_id UUID NOT NULL REFERENCES vouchers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, buyer_id UUID NOT NULL REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, UNIQUE (voucher_id, buyer_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers_buyers (voucher_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers_buyers (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// json_web_tokens
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE json_web_tokens (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL
				, created_at timestamp with time zone NOT NULL
				, expired_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON json_web_tokens (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON json_web_tokens (expired_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// carts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE carts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT UNIQUE
				, deleted_buyer_id UUID REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON carts (deleted_buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// cart_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE cart_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, cart_id UUID NOT NULL REFERENCES carts (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, quantity bigint NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (cart_id, product_variant_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (cart_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (product_variant_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// delivery_routes
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE delivery_routes (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, origin_longitude character varying NOT NULL
				, origin_latitude character varying NOT NULL
				, destination_longitude character varying NOT NULL
				, destination_latitude character varying NOT NULL
				, distance_meters bigint NOT NULL
				, distance_km bigint NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (
					origin_longitude
					, origin_latitude
					, destination_longitude
					, destination_latitude
				)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_routes (origin_longitude);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_routes (origin_latitude);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_routes (destination_longitude);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_routes (destination_latitude);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`ALTER TABLE buyers
				ADD COLUMN delivery_route_id UUID REFERENCES delivery_routes (id) ON UPDATE CASCADE ON DELETE RESTRICT;`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (delivery_route_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// payment_channels
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE payment_channels (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, business_id character varying NOT NULL
				, name character varying NOT NULL
				, channel_code character varying NOT NULL UNIQUE
				, country character varying NOT NULL
				, currency character varying NOT NULL
				, minimum_amount bigint NOT NULL
				, maximum_amount bigint NOT NULL
				, payment_request_expiry double precision NOT NULL
				, refund boolean NOT NULL
				, channel_type character varying NOT NULL
				, channel_category character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			"CREATE INDEX ON payment_channels (LOWER(name));",
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			"CREATE INDEX ON payment_channels (minimum_amount);",
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			"CREATE INDEX ON payment_channels (maximum_amount);",
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		params := [][]any{
			{
				"5f9e7c05d72844407e5289a3",
				"BCA Virtual Account",
				"BCA_VIRTUAL_ACCOUNT",
				"ID",
				"IDR",
				10_000,
				50_000_000_000,
				48,
				false,
				"BANK TRANSFER",
				"VIRTUAL_ACCOUNT",
				now,
			}, {
				"5f9e7c05d72844407e5289a3",
				"BNI Virtual Account",
				"BNI_VIRTUAL_ACCOUNT",
				"ID",
				"IDR",
				1,
				50_000_000_000,
				48,
				false,
				"BANK TRANSFER",
				"VIRTUAL_ACCOUNT",
				now,
			}, {
				"5f9e7c05d72844407e5289a3",
				"BRI Virtual Account",
				"BRI_VIRTUAL_ACCOUNT",
				"ID",
				"IDR",
				1,
				50_000_000_000,
				48,
				false,
				"BANK TRANSFER",
				"VIRTUAL_ACCOUNT",
				now,
			}, {
				"5f9e7c05d72844407e5289a3",
				"Cards",
				"CARDS",
				"ID",
				"IDR",
				5_000,
				200_000_000,
				0.166,
				true,
				"CARDS",
				"PAYMENT",
				now,
			}, {
				"5f9e7c05d72844407e5289a3",
				"Mandiri Virtual Account",
				"MANDIRI_VIRTUAL_ACCOUNT",
				"ID",
				"IDR",
				1,
				50_000_000_000,
				48,
				false,
				"BANK TRANSFER",
				"VIRTUAL_ACCOUNT",
				now,
			}, {
				"5f9e7c05d72844407e5289a3",
				"QRIS",
				"QRIS",
				"ID",
				"IDR",
				1,
				10_000_000,
				48,
				false,
				"QR CODE",
				"QRIS",
				now,
			},
		}
		for i := range params {
			if _, err = tx.Exec(
				ctx,
				`INSERT INTO payment_channels (
					business_id
					, name
					, channel_code
					, country
					, currency
					, minimum_amount
					, maximum_amount
					, payment_request_expiry
					, refund
					, channel_type
					, channel_category
					, created_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`,
				params[i]...,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
		}
		// transactions
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transactions (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, parent_id UUID REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, reordered BOOLEAN NOT NULL
				, buyer_id UUID NOT NULL REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, cart_id UUID NOT NULL REFERENCES carts (id) ON UPDATE CASCADE ON DELETE RESTRICT UNIQUE
				, reference_number character varying NOT NULL
				, status integer NOT NULL
				, recipient_name character varying NOT NULL
				, recipient_mobile_phone character varying NOT NULL
				, buyer_note character varying NOT NULL
				, unique_code bigint NOT NULL
				, shopping_cost bigint NOT NULL
				, referrer_id UUID REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, referral_code character varying
				, referral_points bigint NOT NULL
				, voucher_id UUID REFERENCES vouchers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, voucher_type integer
				, voucher_code character varying
				, voucher_discount bigint NOT NULL
				, reward_points_discounted BOOLEAN NOT NULL
				, reward_points_discount bigint NOT NULL
				, wallet_discounted BOOLEAN NOT NULL
				, wallet_amount bigint NOT NULL
				, wallet_discount bigint NOT NULL
				, discounted_shopping_cost bigint NOT NULL
				, delivery_estimated_cost bigint NOT NULL
				, discounted_delivery_estimated_cost bigint NOT NULL
				, delivery_route_id UUID NOT NULL REFERENCES delivery_routes (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, delivery_latitude character varying NOT NULL
				, delivery_longitude character varying NOT NULL
				, delivery_distance_km bigint NOT NULL
				, delivery_address character varying NOT NULL
				, delivery_instruction character varying NOT NULL
				, delivery_datetime timestamp with time zone NOT NULL
				, payment_amount bigint NOT NULL
				, payment_due_date timestamp with time zone
				, payment_channel_id UUID REFERENCES payment_channels (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, payment_channel_name character varying
				, payment_request_id character varying UNIQUE
				, payment_token_id character varying UNIQUE
				, payment_virtual_account_number character varying
				, payment_virtual_account_expired_at timestamp with time zone
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, cancelled_at timestamp with time zone
				, payment_status character varying
				, payment_failure_code character varying
				, payment_captured_at timestamp with time zone
				, payment_id character varying UNIQUE
				, paid_at timestamp with time zone
				, withdrawal_requested_at timestamp with time zone
				, withdrawal_reason character varying
				, withdrawn_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, withdrawn_at timestamp with time zone
				, processed_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, processed_at timestamp with time zone
				, revised_shopping_cost bigint NOT NULL
				, revised_discounted_shopping_cost bigint NOT NULL
				, revised_payment_amount bigint NOT NULL
				, refund_amount bigint NOT NULL
				, refunded_at timestamp with time zone
				, refund_id character varying UNIQUE
				, shipped_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, shipped_at timestamp with time zone
				, delivery_actual_cost bigint NOT NULL
				, complained_at timestamp with time zone
				, complaint character varying
				, completed_by UUID REFERENCES admins (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, completed_at timestamp with time zone
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (parent_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (reordered);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX transactions_unique_reference_number_idx ON transactions (LOWER(reference_number));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (referrer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (voucher_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (delivery_route_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_due_date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_channel_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_request_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_token_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (created_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (withdrawn_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (processed_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (shipped_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (shipped_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (completed_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// transaction_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transaction_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, transaction_id UUID NOT NULL REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_id UUID NOT NULL REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_name character varying NOT NULL
				, stock_unit character varying NOT NULL
				, base_price bigint NOT NULL
				, selling_price bigint NOT NULL
				, discount_minimum_quantity bigint NOT NULL
				, discount_amount bigint NOT NULL
				, quantity bigint NOT NULL
				, subtotal bigint NOT NULL
				, buyer_note character varying NOT NULL
				, revised_quantity bigint NOT NULL
				, revised_subtotal bigint NOT NULL
				, revision_note character varying NOT NULL
				, out_of_stock bool NOT NULL
				, UNIQUE (transaction_id, product_variant_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (transaction_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (product_variant_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// transaction_counters
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transaction_counters (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, transaction_window character varying NOT NULL UNIQUE
				, hits bigint NOT NULL CHECK (hits >= 0)
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// point_earning_caps
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE point_earning_caps (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, shopping_cost int8range NOT NULL
				, earned_points bigint NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, CONSTRAINT point_earning_caps_no_overlapping_shopping_cost EXCLUDE USING gist (
					shopping_cost WITH &&
				)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_earning_caps (shopping_cost);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_earning_caps (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_earning_caps (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_earning_caps (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		pointEarningCaps := [][]any{
			{
				models.RecordStatusActive,
				100_000,
				149_999,
				1_500,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				150_000,
				199_999,
				2_500,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				200_000,
				299_999,
				3_500,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				300_000,
				399_999,
				5_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				400_000,
				499_999,
				7_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				500_000,
				649_999,
				10_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				650_000,
				799_999,
				13_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				800_000,
				999_999,
				17_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				1_000_000,
				100_000_000,
				22_000,
				sellerID,
				now,
				sellerID,
				now,
			},
		}
		for i := range pointEarningCaps {
			if _, err = tx.Exec(
				ctx,
				`INSERT INTO point_earning_caps (
					record_status
					, shopping_cost
					, earned_points
					, created_by
					, created_at
					, updated_by
					, updated_at
				) VALUES ($1, int8range($2, $3, '[]'), $4, $5, $6, $7, $8);`,
				pointEarningCaps[i]...,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
		}
		// point_earnings
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE point_earnings (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, start_date timestamp with time zone NOT NULL
				, end_date timestamp with time zone NOT NULL
				, amount bigint NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (buyer_id, start_date, end_date)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_earnings (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_earnings (start_date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_earnings (end_date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// point_spending_caps
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE point_spending_caps (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, shopping_cost int8range NOT NULL
				, disposable_points bigint NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, CONSTRAINT point_spending_caps_no_overlapping_shopping_cost EXCLUDE USING gist (
					shopping_cost WITH &&
				)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_spending_caps (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_spending_caps (shopping_cost);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_spending_caps (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_spending_caps (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_spending_caps (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		pointSpendingCaps := [][]any{
			{
				models.RecordStatusActive,
				60_000,
				149_999,
				5_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				150_000,
				249_999,
				7_500,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				250_000,
				399_999,
				10_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				models.RecordStatusActive,
				400_000,
				99_999_999,
				15_000,
				sellerID,
				now,
				sellerID,
				now,
			},
		}
		for i := range pointSpendingCaps {
			if _, err = tx.Exec(
				ctx,
				`INSERT INTO point_spending_caps (
					record_status
					, shopping_cost
					, disposable_points
					, created_by
					, created_at
					, updated_by
					, updated_at
				) VALUES ($1, int8range($2, $3, '[]'), $4, $5, $6, $7, $8);`,
				pointSpendingCaps[i]...,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
		}
		// points
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE points (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, point_type integer NOT NULL
				, buyer_id UUID NOT NULL REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, amount bigint NOT NULL
				, transaction_id UUID REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, description character varying NOT NULL
				, created_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (point_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (transaction_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (LOWER(description));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// wallets
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE wallets (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, wallet_transfer_type integer NOT NULL
				, transaction_id UUID REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, amount bigint NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (wallet_transfer_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (transaction_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// articles
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE articles (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, subject character varying NOT NULL COLLATE "natural_sort"
				, body character varying NOT NULL
				, photo character varying NOT NULL
				, link character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (LOWER(subject));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// banners
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE banners (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, description character varying NOT NULL
				, photo character varying NOT NULL
				, visible boolean NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (visible);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// blogs
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE blogs (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, subject character varying NOT NULL COLLATE "natural_sort"
				, body character varying NOT NULL
				, slug character varying UNIQUE
				, deleted_slug character varying
				, photo character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (LOWER(subject));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// landing_pages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE landing_pages (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, name character varying NOT NULL UNIQUE
				, custom_attributes JSONB NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_pages (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_pages (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_pages (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_pages (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// landing_page_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE landing_page_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, landing_page_name character varying NOT NULL REFERENCES landing_pages (name) ON UPDATE CASCADE ON DELETE RESTRICT
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, link_text character varying
				, link_url character varying
				, photo character varying
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_page_items (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_page_items (landing_page_name);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_page_items (LOWER(subject));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_page_items (LOWER(body));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_page_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_page_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON landing_page_items (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// contents
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE contents (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, page character varying NOT NULL
				, division character varying NOT NULL
				, content_type character varying NOT NULL
				, route character varying
				, ordinal_number bigint NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// content_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE content_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, content_id UUID NOT NULL REFERENCES contents (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, item character varying NOT NULL
				, option character varying
				, description character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// withdrawals
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE withdrawals (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES buyers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, recipient_name character varying NOT NULL
				, bank_account character varying NOT NULL
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, status character varying NOT NULL
				, reviewed_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, reviewed_at timestamp with time zone
				, rejection_reason character varying
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON withdrawals (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON withdrawals (reviewed_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// recipes
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE recipes (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, name character varying
				, deleted_name character varying
				, photo character varying NOT NULL
				, description character varying NOT NULL
				, instagram_link character varying NOT NULL
				, facebook_link character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON recipes (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (LOWER(deleted_name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// recipe_guides
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE recipe_guides (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, recipe_id UUID NOT NULL REFERENCES recipes (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, instruction character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_guides (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_guides (recipe_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_guides (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_guides (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_guides (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// recipe_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE recipe_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, recipe_id UUID NOT NULL REFERENCES recipes (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_id UUID NOT NULL REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_name character varying NOT NULL
				, product_photo character varying NOT NULL
				, product_ready_stock BOOLEAN NOT NULL
				, dose character varying NOT NULL
				, note character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES sellers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (recipe_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON recipe_items (recipe_id, product_id) WHERE record_status = 0;`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (product_ready_stock);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// messages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE messages (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL
				, mobile_phone character varying NOT NULL
				, body character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON messages (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON messages (mobile_phone);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// message_receipts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE message_receipts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, message_id UUID NOT NULL REFERENCES messages (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, user_id UUID NOT NULL
				, read_at timestamp with time zone NOT NULL
				, UNIQUE (message_id, user_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON message_receipts (message_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON message_receipts (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
