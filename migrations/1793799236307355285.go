package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799236307355285] = Migration{
		Version: 1793799236307355285,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// transaction_counters
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE transaction_counters (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, transaction_window TEXT NOT NULL
					, hits bigint NOT NULL CHECK (hits >= 0)
					, created_at DATETIME(6) NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, CONSTRAINT transaction_counters_transaction_window_key UNIQUE (transaction_window)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE transaction_counters",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
