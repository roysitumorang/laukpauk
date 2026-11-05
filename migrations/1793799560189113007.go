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
	Migrations[1793799560189113007] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793799560189113007"
		// point_spending_caps
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE point_spending_caps (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, record_status smallint NOT NULL
				, shopping_cost int8range NOT NULL
				, disposable_points bigint NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, CONSTRAINT point_spending_caps_no_overlapping_shopping_cost EXCLUDE USING gist (
					company_id WITH =,
					shopping_cost WITH &&
				)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON point_spending_caps (company_id);`,
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
		var companyID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`SELECT id
			FROM companies
			WHERE subdomain = $1`,
			"laukpauk",
		).Scan(&companyID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		now := time.Now().In(time.Local)
		pointSpendingCaps := [][]any{
			{
				companyID,
				models.RecordStatusActive,
				60_000,
				149_999,
				5_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				150_000,
				249_999,
				7_500,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				250_000,
				399_999,
				10_000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
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
					company_id
					, record_status
					, shopping_cost
					, disposable_points
					, created_by
					, created_at
					, updated_by
					, updated_at
				) VALUES ($1, $2, int8range($3, $4, '[]'), $5, $6, $7, $8, $9);`,
				pointSpendingCaps[i]...,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
		}
		return
	}
}
