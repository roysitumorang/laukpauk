package migration

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"go.uber.org/zap"
)

func init() {
	Migrations[1780909370437246465] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1780909370437246465"
		// users
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE users (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_type character varying NOT NULL
				, name character varying NOT NULL
				, uncofirmed_mobile_phone character varying UNIQUE
				, mobile_phone_unconfirmed_at timestamp with time zone
				, mobile_phone_confirmation_token character varying
				, mobile_phone_confirmed_at timestamp with time zone
				, mobile_phone character varying NOT NULL UNIQUE
				, otp_token character varying
				, otp_created_at timestamp with time zone
				, otp_expired_at timestamp with time zone
				, username character varying NOT NULL UNIQUE
				, uncofirmed_email character varying UNIQUE
				, email_unconfirmed_at timestamp with time zone
				, email_confirmation_token character varying
				, email_confirmed_at timestamp with time zone
				, email character varying NOT NULL UNIQUE
				, encrypted_password character varying NOT NULL
				, password_reset_at timestamp with time zone
				, reset_password_token character varying
				, password_changed_at timestamp with time zone
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (user_type);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// admins
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE admins (
				user_id UUID NOT NULL PRIMARY KEY REFERENCES users (id)
				, admin_level character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON admins (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON admins (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// buyers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE buyers (
				user_id UUID NOT NULL PRIMARY KEY REFERENCES users (id)
				, referral_code character varying NOT NULL
				, photo character varying NOT NULL
				, address character varying NOT NULL
				, referenced_by UUID REFERENCES users (id)
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (referenced_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// bazaars
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE bazaars (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL
				, address character varying NOT NULL
				, photo character varying NOT NULL
				, coordinates POINT NOT NULL
				, bazaar_type character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON bazaars USING GiST(coordinates);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON bazaars (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON bazaars (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON bazaars (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// sellers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE sellers (
				user_id UUID NOT NULL PRIMARY KEY REFERENCES users (id)
				, address character varying NOT NULL
				, banner character varying NOT NULL
				, bio character varying NOT NULL
				, bazaar_id UUID NOT NULL REFERENCES bazaars (id)
				, open_on_sunday BOOLEAN NOT NULL DEFAULT FALSE
				, open_on_monday BOOLEAN NOT NULL DEFAULT FALSE
				, open_on_tuesday BOOLEAN NOT NULL DEFAULT FALSE
				, open_on_wednesday BOOLEAN NOT NULL DEFAULT FALSE
				, open_on_thursday BOOLEAN NOT NULL DEFAULT FALSE
				, open_on_friday BOOLEAN NOT NULL DEFAULT FALSE
				, open_on_saturday BOOLEAN NOT NULL DEFAULT FALSE
				, status character varying NOT NULL
				, status_reviewed_by UUID REFERENCES users (id)
				, status_reviewed_at timestamp with time zone
				, status_rejection_reason character varying
				, commission_rate bigint NOT NULL
				, commission_type character varying NOT NULL
				, bank_account character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (bazaar_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (status_reviewed_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sellers (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// product_categories
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE product_categories (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL
				, photo character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON product_categories (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// product_category_discounts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE product_category_discounts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, product_category_id UUID NOT NULL REFERENCES product_categories (id)
				, minimum_transaction bigint NOT NULL
				, discount bigint NOT NULL
				, effective_date_start timestamp with time zone NOT NULL
				, effective_date_end timestamp with time zone NOT NULL
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_category_discounts (product_category_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_category_discounts (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_category_discounts (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_category_discounts (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// products
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE products (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, seller_id UUID NOT NULL REFERENCES users (id)
				, product_category_id UUID NOT NULL REFERENCES product_categories (id)
				, name character varying NOT NULL
				, photo character varying NOT NULL
				, short_description character varying NOT NULL
				, full_description character varying NOT NULL
				, stock_status character varying NOT NULL
				, pinned BOOLEAN NOT NULL DEFAULT FALSE
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (seller_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (product_category_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (stock_status);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (pinned);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// product_variants
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE product_variants (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, product_id UUID NOT NULL REFERENCES products (id)
				, unit_of_measurement character varying NOT NULL
				, price bigint NOT NULL
				, discount bigint NOT NULL
				, price_type character varying NOT NULL
				, price_variant_percentage NUMERIC(10, 2) NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
				, UNIQUE (unit_of_measurement, product_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (product_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_variants (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// provinces
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE provinces (
				id bigserial NOT NULL PRIMARY KEY
				, name character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON provinces (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON provinces (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON provinces (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON provinces (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// regencies
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE regencies (
				id bigserial NOT NULL PRIMARY KEY
				, province_id bigint NOT NULL REFERENCES provinces (id)
				, name character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (province_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON regencies (LOWER(name), province_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// districts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE districts (
				id bigserial NOT NULL PRIMARY KEY
				, regency_id bigint NOT NULL REFERENCES regencies (id)
				, name character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (regency_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON districts (LOWER(name), regency_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// subdistricts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE subdistricts (
				id bigserial NOT NULL PRIMARY KEY
				, district_id bigint NOT NULL REFERENCES districts (id)
				, name character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (district_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON subdistricts (LOWER(name), district_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
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
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (LOWER(subject));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (LOWER(body));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_templates (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// notifications
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE notifications (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id)
				, notification_type character varying NOT NULL
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, params JSONB NOT NULL
				, created_at timestamp with time zone NOT NULL
				, read_at timestamp with time zone
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notifications (user_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notifications (read_at);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// vouchers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE vouchers (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, code character varying NOT NULL
				, photo character varying NOT NULL
				, name character varying NOT NULL
				, amount bigint NOT NULL
				, voucher_type character varying NOT NULL
				, target_user character varying NOT NULL
				, description character varying NOT NULL
				, minimum_transaction bigint NOT NULL
				, maximum_usage bigint NOT NULL
				, effective_date_start timestamp with time zone NOT NULL
				, effective_date_end timestamp with time zone NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// vouchers_users
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE vouchers_users (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, voucher_id UUID NOT NULL REFERENCES vouchers (id)
				, user_id UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, UNIQUE (voucher_id, user_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers_users (voucher_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON vouchers_users (user_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// settings
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE settings (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, key character varying NOT NULL
				, value character varying NOT NULL
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON settings (LOWER(key));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON settings (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON settings (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// wallets
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE wallets (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id)
				, wallet_type character varying NOT NULL
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (user_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// points
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE points (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id)
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, point_type character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (user_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// sms
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE sms (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, body character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sms (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// sms_users
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE sms_users (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, sms_id UUID NOT NULL REFERENCES sms (id)
				, user_id UUID NOT NULL REFERENCES users (id)
				, mobile_phone character varying NOT NULL
				, UNIQUE (sms_id, user_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sms_users (sms_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sms_users (user_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// promos
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE promos (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL
				, code character varying NOT NULL
				, discount bigint NOT NULL
				, quota_per_user bigint NOT NULL
				, description character varying NOT NULL
				, budget bigint NOT NULL
				, effective_date_start timestamp with time zone NOT NULL
				, effective_date_end timestamp with time zone NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promos (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promos (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promos (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// promo_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE promo_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, qr_code character varying NOT NULL
				, link character varying NOT NULL
				, promo_id UUID NOT NULL REFERENCES promos (id)
				, bazaar_id UUID NOT NULL REFERENCES bazaars (id)
				, seller_id UUID NOT NULL REFERENCES users (id)
				, product_id UUID NOT NULL REFERENCES products (id)
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id)
				, max_transaction bigint NOT NULL
				, quota bigint NOT NULL
				, discount bigint NOT NULL
				, note character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (promo_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (bazaar_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (seller_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (product_variant_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON promo_items (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// shipping_plans
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE shipping_plans (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL
				, transaction_time_start TIME WITH TIME ZONE NOT NULL
				, transaction_time_end TIME WITH TIME ZONE NOT NULL
				, transaction_time_description character varying NOT NULL
				, shipping_time_start TIME WITH TIME ZONE NOT NULL
				, shipping_time_end TIME WITH TIME ZONE NOT NULL
				, shipping_time_description character varying NOT NULL
				, daily_quota bigint NOT NULL
				, ordinal_number bigint NOT NULL UNIQUE
				, shipping_type character varying NOT NULL
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON shipping_plans (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON shipping_plans (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON shipping_plans (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON shipping_plans (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// shipping_plans_sellers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE shipping_plans_sellers (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, shipping_plan_id UUID NOT NULL REFERENCES shipping_plans (id)
				, seller_id UUID NOT NULL REFERENCES users (id)
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, UNIQUE (shipping_plan_id, seller_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON shipping_plans_sellers (shipping_plan_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON shipping_plans_sellers (seller_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON shipping_plans_sellers (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// non_shipping_days
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE non_shipping_days (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, day timestamp with time zone NOT NULL
				, shipping_plan_id UUID NOT NULL REFERENCES shipping_plans (id)
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, UNIQUE (day, shipping_plan_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON non_shipping_days (shipping_plan_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON non_shipping_days (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// carts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE carts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES users (id)
				, shipping_type character varying NOT NULL
				, shipping_plan_id UUID NOT NULL REFERENCES shipping_plans (id)
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (buyer_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON carts (shipping_plan_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// cart_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE cart_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, cart_id UUID NOT NULL REFERENCES carts (id)
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id)
				, product_id UUID NOT NULL REFERENCES products (id)
				, promo_id UUID NOT NULL REFERENCES promos (id)
				, promo_discount bigint NOT NULL
				, seller_id UUID NOT NULL REFERENCES users (id)
				, bazaar_id UUID NOT NULL REFERENCES bazaars (id)
				, price bigint NOT NULL
				, discount bigint NOT NULL
				, quantity bigint NOT NULL
				, note character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (cart_id, product_variant_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (cart_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (product_variant_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (promo_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (seller_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (bazaar_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// payment_methods
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE payment_methods (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, channel_code character VARYING NOT NULL UNIQUE
				, ordinal_number bigint NOT NULL UNIQUE
				, business_id character varying NOT NULL
				, name character varying NOT NULL
				, channel_type character varying NOT NULL
				, minimum_transaction bigint NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON payment_methods (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON payment_methods (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON payment_methods (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// transactions
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transactions (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, bazaar_id UUID NOT NULL REFERENCES bazaars (id)
				, seller_id UUID NOT NULL REFERENCES users (id)
				, buyer_id UUID NOT NULL REFERENCES users (id)
				, transaction_no character varying NOT NULL UNIQUE
				, down_payment bigint NOT NULL
				, voucher_id UUID REFERENCES vouchers (id)
				, voucher_type character varying
				, voucher_amount bigint
				, recipient_name character varying NOT NULL
				, recipient_mobile_phone character varying NOT NULL
				, buyer_note character varying NOT NULL
				, payment_method_id UUID NOT NULL REFERENCES payment_methods (id)
				, payment_method_name character varying NOT NULL
				, payment_method_channel_code character varying NOT NULL
				, payment_method_extra_data character varying
				, payment_no character varying
				, payment_status character varying NOT NULL
				, shipping_status character varying NOT NULL
				, reviewed_by UUID REFERENCES users (id)
				, reviewed_at timestamp with time zone
				, rejection_reason character varying
				, shipping_cost bigint NOT NULL
				, shipping_cost_type character varying NOT NULL
				, wallet_deduction bigint NOT NULL
				, cashback bigint NOT NULL
				, shipping_coordinates POINT NOT NULL
				, shipping_address character varying NOT NULL
				, shipping_instruction character varying NOT NULL
				, shipping_date timestamp with time zone NOT NULL
				, shipping_plan_id UUID NOT NULL REFERENCES shipping_plans (id)
				, shipping_type character varying NOT NULL
				, shipping_timeline character varying NOT NULL
				, refunded_shipping_cost bigint NOT NULL
				, refunded_transaction_cost bigint NOT NULL
				, unique_code bigint NOT NULL UNIQUE
				, minimum_point bigint NOT NULL
				, point_deduction bigint NOT NULL
				, promo_id UUID REFERENCES promos (id)
				, promo_discount bigint
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (bazaar_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (seller_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (voucher_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_method_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (reviewed_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions USING GiST(shipping_coordinates);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (shipping_plan_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (promo_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// transaction_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transaction_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, transaction_id UUID NOT NULL REFERENCES transactions (id)
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id)
				, product_id UUID NOT NULL REFERENCES products (id)
				, promo_id UUID NOT NULL REFERENCES promos (id)
				, promo_discount bigint NOT NULL
				, price bigint NOT NULL
				, discount bigint NOT NULL
				, quantity bigint NOT NULL
				, status character varying NOT NULL
				, reviewed_by UUID REFERENCES users (id)
				, reviewed_at timestamp with time zone
				, rejection_reason character varying
				, buyer_note character varying NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (transaction_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (product_variant_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (promo_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (reviewed_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// articles
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE articles (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, photo character varying NOT NULL
				, link character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON articles (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// banners
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE banners (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL
				, description character varying NOT NULL
				, photo character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON banners (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// blogs
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE blogs (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, photo character varying NOT NULL
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, category character varying NOT NULL
				, slug character varying NOT NULL UNIQUE
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blogs (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// gifts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE gifts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, photo character varying NOT NULL
				, name character varying NOT NULL
				, minimum_transaction bigint NOT NULL
				, effective_date_start timestamp with time zone NOT NULL
				, effective_date_end timestamp with time zone NOT NULL
				, quota bigint NOT NULL
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON gifts (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON gifts (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON gifts (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// gifts_transactions
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE gifts_transactions (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, gift_id UUID NOT NULL REFERENCES gifts (id)
				, transaction_id UUID NOT NULL REFERENCES transactions (id)
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, UNIQUE (gift_id, transaction_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON gifts_transactions (gift_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON gifts_transactions (transaction_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON gifts_transactions (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// homepages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE homepages (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL UNIQUE
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON homepages (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON homepages (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON homepages (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// homepage_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE homepage_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, homepage_id UUID REFERENCES homepages (id)
				, parent character varying
				, key character varying NOT NULL
				, value character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON homepage_items (homepage_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON homepage_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON homepage_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON homepage_items (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
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
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// content_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE content_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, content_id UUID NOT NULL REFERENCES contents (id)
				, bazaar_id UUID REFERENCES bazaars (id)
				, seller_id UUID REFERENCES users (id)
				, item character varying NOT NULL
				, option character varying
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// withdrawals
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE withdrawals (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES users (id)
				, recipient_name character varying NOT NULL
				, bank_account character varying NOT NULL
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, status character varying NOT NULL
				, reviewed_by UUID REFERENCES users (id)
				, reviewed_at timestamp with time zone
				, rejection_reason character varying
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON withdrawals (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON withdrawals (reviewed_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// recipes
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE recipes (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, photo character varying NOT NULL
				, name character varying NOT NULL
				, description character varying NOT NULL
				, instagram_link character varying NOT NULL
				, facebook_link character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipes (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// recipe_instructions
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE recipe_instructions (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, recipe_id UUID NOT NULL REFERENCES recipes (id)
				, action character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (recipe_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// recipe_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE recipe_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, recipe_id UUID NOT NULL REFERENCES recipes (id)
				, product_id UUID NOT NULL REFERENCES products (id)
				, dose character varying NOT NULL
				, note character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone NOT NULL
				, deactivation_reason character varying
				, UNIQUE (recipe_id, product_id)
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (recipe_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (created_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (updated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// balances
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE balances (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES users (id)
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON balances (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// invoices
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE invoices (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, seller_id UUID NOT NULL REFERENCES users (id)
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON invoices (seller_id);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
		}
		return
	}
}
