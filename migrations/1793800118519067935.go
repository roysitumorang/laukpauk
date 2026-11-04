package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800118519067935] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793800118519067935"
		// withdrawals
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE withdrawals (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, recipient_name character varying NOT NULL
				, bank_account character varying NOT NULL
				, amount bigint NOT NULL
				, description character varying NOT NULL
				, status character varying NOT NULL
				, reviewed_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, reviewed_at timestamp with time zone
				, rejection_reason character varying
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON withdrawals (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON withdrawals (reviewed_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
