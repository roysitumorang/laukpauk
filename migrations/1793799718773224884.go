package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799718773224884] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793799718773224884"
		// wallets
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE wallets (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, wallet_transfer_type smallint NOT NULL
				, transaction_id UUID REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, amount bigint NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (wallet_transfer_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON wallets (transaction_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
