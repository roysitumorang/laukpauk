package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799236307355285] = Migration{
		Version: 1793799236307355285,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// transaction_counters
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE transaction_counters (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, transaction_window text NOT NULL
					, hits bigint NOT NULL CHECK (hits >= 0)
					, created_at timestamp with time zone NOT NULL
					, updated_at timestamp with time zone NOT NULL
					, UNIQUE (transaction_window)
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE transaction_counters",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
