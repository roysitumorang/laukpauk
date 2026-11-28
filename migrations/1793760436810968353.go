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
	Migrations[1793760436810968353] = Migration{
		Version: 1793760436810968353,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// users
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE users (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, role text NOT NULL
					, status text NOT NULL
					, name text NOT NULL COLLATE "case_insensitive"
					, mobile_phone_confirmation_sent_at timestamp with time zone
					, mobile_phone_confirmation_token text
					, mobile_phone_confirmed_at timestamp with time zone
					, mobile_phone text NOT NULL CONSTRAINT users_unique_mobile_phone UNIQUE
					, username text CONSTRAINT users_unique_username UNIQUE COLLATE "case_insensitive"
					, email_confirmation_sent_at timestamp with time zone
					, email_confirmation_token text
					, email_confirmed_at timestamp with time zone
					, email text CONSTRAINT users_unique_email UNIQUE COLLATE "case_insensitive"
					, encrypted_password text
					, password_reset_sent_at timestamp with time zone
					, password_reset_token text
					, password_changed_at timestamp with time zone
					, sign_in_count integer NOT NULL DEFAULT 0
					, current_sign_in_at timestamp with time zone
					, last_sign_in_at timestamp with time zone
					, current_sign_in_ip text
					, last_sign_in_ip text
					, failed_attempts integer NOT NULL DEFAULT 0
					, unlock_token text
					, locked_at timestamp with time zone
					, otp_secret text
					, otp_expired_at timestamp with time zone
					, photo text NOT NULL
					, user_level text NOT NULL
					, referral_code text CONSTRAINT users_unique_referral_code UNIQUE
					, referrer_id UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, referral_points bigint
					, reward_points bigint
					, wallet bigint
					, created_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, created_at timestamp with time zone NOT NULL
					, updated_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, updated_at timestamp with time zone NOT NULL
					, deactivated_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, deactivated_at timestamp with time zone
					, deactivation_reason text
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX users_role_idx ON users (role);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX users_status_idx ON users (status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX users_name_idx ON users (name);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX users_referrer_id_idx ON users (referrer_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX users_created_by_idx ON users (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX users_updated_by_idx ON users (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX users_deactivated_by_idx ON users (deactivated_by);`,
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
				"seller-default.png",
				models.UserLevelSellerOwner,
				"2021-12-16 17:26:51.000 +0700",
				"2024-12-11 07:41:55.000 +0700",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE users",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
