package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799236307355285] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793799236307355285"
		// transaction_counters
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transaction_counters (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE CASCADE
				, transaction_window character varying NOT NULL
				, hits bigint NOT NULL CHECK (hits >= 0)
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (company_id, transaction_window)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
