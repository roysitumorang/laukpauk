package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797496585996818] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793797496585996818"
		// notification_targets
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE notification_targets (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, firebase_installation_id character varying NOT NULL UNIQUE
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON notification_targets (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
