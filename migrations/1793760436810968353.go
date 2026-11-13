package migrations

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793760436810968353] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793760436810968353"
		// users
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE users (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, role character varying NOT NULL
				, status character varying NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, mobile_phone_confirmation_sent_at timestamp with time zone
				, mobile_phone_confirmation_token character varying
				, mobile_phone_confirmed_at timestamp with time zone
				, mobile_phone character varying NOT NULL UNIQUE
				, username character varying UNIQUE
				, email_confirmation_sent_at timestamp with time zone
				, email_confirmation_token character varying
				, email_confirmed_at timestamp with time zone
				, email character varying UNIQUE
				, encrypted_password character varying
				, password_reset_sent_at timestamp with time zone
				, password_reset_token character varying
				, password_changed_at timestamp with time zone
				, sign_in_count integer NOT NULL DEFAULT 0
				, current_sign_in_at timestamp with time zone
				, last_sign_in_at timestamp with time zone
				, current_sign_in_ip character varying
				, last_sign_in_ip character varying
				, failed_attempts integer NOT NULL DEFAULT 0
				, unlock_token character varying
				, locked_at timestamp with time zone
				, otp_secret character varying
				, otp_expired_at timestamp with time zone
				, photo character varying NOT NULL
				, user_level character varying NOT NULL
				, referral_code character varying UNIQUE
				, address character varying
				, referrer_id UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, referral_points bigint
				, reward_points bigint
				, wallet bigint
				, delivery_route_id UUID
				, created_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
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
			`CREATE INDEX ON users (role);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (status);`,
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
			`CREATE INDEX ON users (referrer_id);`,
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
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO users (
				role
				, status
				, name
				, mobile_phone_confirmed_at
				, mobile_phone
				, username
				, email_confirmed_at
				, email
				, encrypted_password
				, photo
				, user_level
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			models.RoleSeller,
			models.UserStatusActive,
			"Ben",
			"2024-12-11 07:41:55.000 +0700",
			"+6281265688889",
			"bigben",
			"2024-12-11 07:41:55.000 +0700",
			"benni.surbakti@gmail.com",
			encryptedPassword,
			"admin-default.png",
			models.UserLevelSellerOwner,
			"2021-12-16 17:26:51.000 +0700",
			"2024-12-11 07:41:55.000 +0700",
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
