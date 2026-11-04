package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793796754063673434] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793796754063673434"
		// json_web_tokens
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE json_web_tokens (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, expired_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON json_web_tokens (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON json_web_tokens (expired_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
