package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1795652725013695980] = Migration{
		Version: 1795652725013695980,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// promotions
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE promotions (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, record_status smallint NOT NULL
					, name text COLLATE "case_insensitive" NOT NULL
					, code text COLLATE "case_insensitive" CONSTRAINT promotions_unique_code UNIQUE
					, deleted_code text COLLATE "case_insensitive"
					, description text NOT NULL
					, discount bigint NOT NULL
					, quota_per_buyer bigint NOT NULL
					, effective_start_date timestamp with time zone NOT NULL
					, effective_end_date timestamp with time zone NOT NULL
					, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, created_at timestamp with time zone NOT NULL
					, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, updated_at timestamp with time zone NOT NULL
					, deleted_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, deleted_at timestamp with time zone
					, deletion_reason text
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX promotions_record_status_idx ON promotions (record_status)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX promotions_name_idx ON promotions (name)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX promotions_effective_start_date_idx ON promotions (effective_start_date)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX promotions_effective_end_date_idx ON promotions (effective_end_date)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX promotions_created_by_idx ON promotions (created_by)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX promotions_updated_by_idx ON promotions (updated_by)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX promotions_deleted_by_idx ON promotions (deleted_by)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`ALTER TABLE transactions
					ADD FOREIGN KEY (promotion_id) REFERENCES promotions (id)
					ON UPDATE CASCADE ON DELETE RESTRICT`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"CREATE INDEX transactions_promotion_id_idx ON transactions (promotion_id)",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP INDEX transactions_promotion_id_idx",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`ALTER TABLE transactions
					DROP CONSTRAINT transactions_promotion_id_fkey`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE promotions",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
