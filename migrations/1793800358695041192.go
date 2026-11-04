package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800358695041192] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793800358695041192"
		// message_receipts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE message_receipts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, message_id UUID NOT NULL REFERENCES messages (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, read_at timestamp with time zone NOT NULL
				, UNIQUE (message_id, user_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON message_receipts (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
