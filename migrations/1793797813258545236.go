package migrations

import (
	"context"
	"log/slog"
	"time"
	"uuid"

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
				, maintenance_mode BOOLEAN NOT NULL
				, subscription_trial_period smallint NOT NULL
				, monthly_billing_amount bigint NOT NULL
				, annual_billing_amount bigint NOT NULL
				, billing_grace_period smallint NOT NULL
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
		}
		return
	}
}
