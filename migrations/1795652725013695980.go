package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1795652725013695980] = Migration{
		Version: 1795652725013695980,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// promotions
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE promotions (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, name VARCHAR(255) COLLATE utf8mb4_uca1400_as_ci NOT NULL
					, code VARCHAR(255) COLLATE utf8mb4_uca1400_as_ci
					, deleted_code TEXT COLLATE utf8mb4_uca1400_as_ci
					, description TEXT NOT NULL
					, discount bigint NOT NULL
					, quota_per_buyer bigint NOT NULL
					, effective_start_date DATETIME(6) NOT NULL
					, effective_end_date DATETIME(6) NOT NULL
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, CONSTRAINT promotions_code_key UNIQUE (code)
					, CONSTRAINT promotions_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT promotions_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT promotions_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX promotions_record_status_idx ON promotions (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX promotions_name_idx ON promotions (name);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX promotions_effective_start_date_idx ON promotions (effective_start_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX promotions_effective_end_date_idx ON promotions (effective_end_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX promotions_created_by_idx ON promotions (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX promotions_updated_by_idx ON promotions (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX promotions_deleted_by_idx ON promotions (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`ALTER TABLE transactions
					ADD CONSTRAINT transactions_promotion_id_fkey FOREIGN KEY (promotion_id) REFERENCES promotions (id)
					ON UPDATE CASCADE ON DELETE RESTRICT`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_promotion_id_idx ON transactions (promotion_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				`ALTER TABLE transactions
					DROP FOREIGN KEY transactions_promotion_id_fkey`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				"DROP INDEX transactions_promotion_id_idx ON transactions",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE promotions",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
