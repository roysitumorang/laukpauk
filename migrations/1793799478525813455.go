package migrations

import (
	"context"
	"log/slog"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799478525813455] = Migration{
		Version: 1793799478525813455,
		Up: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Up"
			// point_earnings
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE point_earnings (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, user_id UUID NOT NULL
					, start_date DATETIME(6) NOT NULL
					, end_date DATETIME(6) NOT NULL
					, amount bigint NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, CONSTRAINT point_earnings_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT point_earnings_user_id_start_date_end_date_key UNIQUE (user_id, start_date, end_date)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earnings_start_date_idx ON point_earnings (start_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earnings_end_date_idx ON point_earnings (end_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE point_earnings",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
