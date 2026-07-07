package migration

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1780909370437246465] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1780909370437246465"
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
		// users
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE users (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_type character varying NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, confirmation_token character varying UNIQUE
				, confirmed_at timestamp with time zone
				, unconfirmed_mobile_phone character varying
				, mobile_phone_confirmation_sent_at timestamp with time zone
				, mobile_phone_confirmation_token character varying
				, mobile_phone_confirmed_at timestamp with time zone
				, mobile_phone character varying UNIQUE
				, otp_token character varying
				, otp_created_at timestamp with time zone
				, otp_expired_at timestamp with time zone
				, username character varying NOT NULL UNIQUE
				, unconfirmed_email character varying
				, email_confirmation_sent_at timestamp with time zone
				, email_confirmation_token character varying
				, email_confirmed_at timestamp with time zone
				, email character varying UNIQUE
				, encrypted_password character varying
				, password_reset_at timestamp with time zone
				, password_reset_token character varying
				, password_changed_at timestamp with time zone
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (user_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// admins
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE admins (
				user_id UUID NOT NULL PRIMARY KEY REFERENCES users (id)
				, admin_level character varying NOT NULL
				, photo character varying NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// buyers
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE buyers (
				user_id UUID NOT NULL PRIMARY KEY REFERENCES users (id)
				, referral_code character varying NOT NULL UNIQUE
				, photo character varying NOT NULL
				, address character varying NOT NULL
				, referenced_by UUID REFERENCES users (id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON buyers (referenced_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// product_categories
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE product_categories (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, description character varying NOT NULL
				, photo character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
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
			`CREATE UNIQUE INDEX product_categories_unique_name_idx ON product_categories (LOWER(name));`,
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
				, product_category_id UUID REFERENCES product_categories (id)
				, product_type integer NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, photo character varying NOT NULL
				, description character varying NOT NULL
				, stock_status integer NOT NULL
				, pinned BOOLEAN NOT NULL DEFAULT FALSE
				, base_price bigint NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
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
			`CREATE UNIQUE INDEX product_categories_unique_idx ON products (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON products (stock_status);`,
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
				, product_id UUID NOT NULL REFERENCES products (id)
				, measure_unit character varying NOT NULL COLLATE "natural_sort"
				, price bigint NOT NULL
				, price_type integer NOT NULL
				, price_variant_percentage numeric(5, 2) NOT NULL
				, discount_type integer NOT NULL
				, discount_rate bigint NOT NULL
				, discount_minimum_quantity bigint NOT NULL
				, discount_amount bigint NOT NULL
				, sale_price bigint NOT NULL
				, description character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, UNIQUE (product_id, measure_unit)
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
			`CREATE INDEX ON product_variants (LOWER(measure_unit));`,
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
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON provinces (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON provinces (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON provinces (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON provinces (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (province_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON regencies (LOWER(name), province_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON regencies (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (regency_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON districts (LOWER(name), regency_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON districts (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (district_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON subdistricts (LOWER(name), district_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON subdistricts (deactivated_by);`,
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
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
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
				, user_id UUID NOT NULL REFERENCES users (id)
				, notification_type character varying NOT NULL
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, params JSONB NOT NULL
				, created_at timestamp with time zone NOT NULL
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
			`CREATE INDEX ON notifications (read_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// settings
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE settings (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, active BOOLEAN NOT NULL
				, key character varying NOT NULL COLLATE "natural_sort"
				, value character varying NOT NULL
				, description character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON settings (active);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX settings_unique_key_idx ON settings (LOWER(key));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON settings (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON settings (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// sms_messages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE sms_messages (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id)
				, mobile_phone character varying NOT NULL
				, message_type integer NOT NULL
				, body character varying NOT NULL
				, delivery_status character varying
				, message_id bigint
				, created_by UUID REFERENCES users (id)
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
			`CREATE INDEX ON sms_messages (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sms_messages (message_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sms_messages (delivery_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sms_messages (message_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON sms_messages (created_by);`,
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
				, transaction_window TSTZRANGE NOT NULL
				, same_day BOOLEAN NOT NULL
				, same_day_window TSTZRANGE NOT NULL
				, same_day_interval character varying
				, next_day BOOLEAN NOT NULL
				, next_day_window TSTZRANGE NOT NULL
				, next_day_interval character varying
				, next_two_day BOOLEAN NOT NULL
				, next_two_day_window TSTZRANGE NOT NULL
				, next_two_day_interval character varying
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, CONSTRAINT no_overlapping_transaction_windows EXCLUDE USING gist (
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
			`CREATE UNIQUE INDEX delivery_plans_unique_name_idx ON delivery_plans (LOWER(name));`,
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
		// non_delivery_days
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE non_delivery_days (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, date timestamp with time zone NOT NULL
				, delivery_plan_id UUID NOT NULL REFERENCES delivery_plans (id)
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, UNIQUE (date, delivery_plan_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON non_delivery_days (date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON non_delivery_days (delivery_plan_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON non_delivery_days (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON non_delivery_days (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// points
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE points (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, point_type integer NOT NULL
				, buyer_id UUID NOT NULL REFERENCES users (id)
				, amount bigint NOT NULL
				, description character varying NOT NULL COLLATE "natural_sort"
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
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
		// carts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE carts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES users (id)
				, delivery_type character varying NOT NULL
				, delivery_plan_id UUID NOT NULL REFERENCES delivery_plans (id)
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (buyer_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON carts (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON carts (delivery_plan_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
				, price bigint NOT NULL
				, discount bigint NOT NULL
				, quantity bigint NOT NULL
				, note character varying NOT NULL
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
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// transactions
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transactions (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, buyer_id UUID NOT NULL REFERENCES users (id)
				, transaction_no character varying NOT NULL UNIQUE
				, payment_amount bigint NOT NULL
				, recipient_name character varying NOT NULL
				, recipient_mobile_phone character varying NOT NULL
				, buyer_note character varying NOT NULL
				, payment_no character varying
				, payment_status character varying NOT NULL
				, delivery_status character varying NOT NULL
				, reviewed_by UUID REFERENCES users (id)
				, reviewed_at timestamp with time zone
				, rejection_reason character varying
				, delivery_cost bigint NOT NULL
				, delivery_cost_type character varying NOT NULL
				, wallet_deduction bigint NOT NULL
				, cashback bigint NOT NULL
				, delivery_coordinates POINT NOT NULL
				, delivery_address character varying NOT NULL
				, delivery_instruction character varying NOT NULL
				, delivery_date timestamp with time zone NOT NULL
				, delivery_plan_id UUID NOT NULL REFERENCES delivery_plans (id)
				, delivery_type character varying NOT NULL
				, delivery_timeline character varying NOT NULL
				, refunded_delivery_cost bigint NOT NULL
				, refunded_transaction_cost bigint NOT NULL
				, unique_code character varying NOT NULL UNIQUE
				, minimum_point bigint NOT NULL
				, point_deduction bigint NOT NULL
				, promo_discount bigint NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
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
			`CREATE INDEX ON transactions (reviewed_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions USING GiST(delivery_coordinates);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (delivery_plan_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// transaction_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transaction_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, transaction_id UUID REFERENCES transactions (id)
				, product_variant_id UUID REFERENCES product_variants (id)
				, product_id UUID REFERENCES products (id)
				, product_type character varying NOT NULL
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
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (reviewed_by);`,
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
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
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
				, name character varying NOT NULL
				, description character varying NOT NULL
				, photo character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
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
			`CREATE INDEX ON banners (deactivated_by);`,
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
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
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
		// blocks
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE blocks (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, name character varying NOT NULL UNIQUE
				, custom_attributes JSONB NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blocks (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blocks (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blocks (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON blocks (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// block_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE block_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status integer NOT NULL
				, block_name character varying NOT NULL REFERENCES blocks (name)
				, subject character varying NOT NULL
				, body character varying NOT NULL
				, link_text character varying
				, link_url character varying
				, photo character varying
				, created_by UUID NOT NULL REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id)
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON block_items (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON block_items (block_name);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON block_items (LOWER(subject));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON block_items (LOWER(body));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON block_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON block_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON block_items (deleted_by);`,
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
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
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
				, content_id UUID NOT NULL REFERENCES contents (id)
				, item character varying NOT NULL
				, option character varying
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id)
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id)
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id)
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
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
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
			`CREATE INDEX ON recipes (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (recipe_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
				, UNIQUE (recipe_id, product_id)
			);`,
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
			`CREATE INDEX ON recipe_instructions (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON recipe_instructions (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON balances (buyer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		// invoices
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE invoices (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
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
				, message_id UUID NOT NULL REFERENCES messages (id)
				, user_id UUID NOT NULL REFERENCES users (id)
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
