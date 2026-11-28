package migrations

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793760436810968353] = Migration{
		Version: 1793760436810968353,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// users
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE users (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, role VARCHAR(255) NOT NULL
					, status VARCHAR(255) NOT NULL
					, name TEXT NOT NULL
					, mobile_phone_confirmation_sent_at DATETIME(6)
					, mobile_phone_confirmation_token TEXT
					, mobile_phone_confirmed_at DATETIME(6)
					, mobile_phone VARCHAR(255) NOT NULL
					, username VARCHAR(255)
					, email_confirmation_sent_at DATETIME(6)
					, email_confirmation_token TEXT
					, email_confirmed_at DATETIME(6)
					, email VARCHAR(255)
					, encrypted_password TEXT
					, password_reset_sent_at DATETIME(6)
					, password_reset_token TEXT
					, password_changed_at DATETIME(6)
					, sign_in_count integer NOT NULL DEFAULT 0
					, current_sign_in_at DATETIME(6)
					, last_sign_in_at DATETIME(6)
					, current_sign_in_ip TEXT
					, last_sign_in_ip TEXT
					, failed_attempts integer NOT NULL DEFAULT 0
					, unlock_token TEXT
					, locked_at DATETIME(6)
					, otp_secret TEXT
					, otp_expired_at DATETIME(6)
					, photo TEXT NOT NULL
					, user_level TEXT NOT NULL
					, referral_code VARCHAR(255)
					, referrer_id UUID
					, referral_points bigint
					, reward_points bigint
					, wallet bigint
					, created_by UUID
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID
					, updated_at DATETIME(6) NOT NULL
					, deactivated_by UUID
					, deactivated_at DATETIME(6)
					, deactivation_reason TEXT
					, CONSTRAINT users_mobile_phone_key UNIQUE (mobile_phone)
					, CONSTRAINT users_username_key UNIQUE (username)
					, CONSTRAINT users_email_key UNIQUE (email)
					, CONSTRAINT users_referral_code_key UNIQUE (referral_code)
					, CONSTRAINT users_referrer_id_fkey FOREIGN KEY (referrer_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT users_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT users_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT users_deactivated_by_fkey FOREIGN KEY (deactivated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_ci;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX users_role_idx ON users (role);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX users_status_idx ON users (status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX users_referrer_id_idx ON users (referrer_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX users_created_by_idx ON users (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX users_updated_by_idx ON users (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX users_deactivated_by_idx ON users (deactivated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
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
			if _, err = tx.ExecContext(
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
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				models.RoleSeller,
				models.UserStatusActive,
				"Ben",
				seedTime("2024-12-11 07:41:55.000 +0700"),
				"+6281265688889",
				"bigben",
				seedTime("2024-12-11 07:41:55.000 +0700"),
				"benni.surbakti@gmail.com",
				encryptedPassword,
				"seller-default.png",
				models.UserLevelSellerOwner,
				seedTime("2021-12-16 17:26:51.000 +0700"),
				seedTime("2024-12-11 07:41:55.000 +0700"),
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE users",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
