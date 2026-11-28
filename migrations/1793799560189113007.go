package migrations

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793799560189113007] = Migration{
		Version: 1793799560189113007,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// point_spending_caps
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE point_spending_caps (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, shopping_cost_min BIGINT NOT NULL
					, shopping_cost_max BIGINT NOT NULL
					, disposable_points bigint NOT NULL
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, CONSTRAINT point_spending_caps_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT point_spending_caps_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT point_spending_caps_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_ci;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX point_spending_caps_record_status_idx ON point_spending_caps (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX point_spending_caps_shopping_cost_idx ON point_spending_caps (shopping_cost_min, shopping_cost_max);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX point_spending_caps_created_by_idx ON point_spending_caps (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX point_spending_caps_updated_by_idx ON point_spending_caps (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX point_spending_caps_deleted_by_idx ON point_spending_caps (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			var sellerID uuid.UUID
			if err = tx.QueryRowContext(
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
				if _, err = tx.ExecContext(
					ctx,
					`INSERT INTO point_spending_caps (
						record_status
						, shopping_cost_min
						, shopping_cost_max
						, disposable_points
						, created_by
						, created_at
						, updated_by
						, updated_at
					) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`,
					pointSpendingCaps[i]...,
				); err != nil {
					helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
					return
				}
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE point_spending_caps",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
