package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800324931593827] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793800324931593827"
		// messages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE messages (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, name character varying NOT NULL
				, mobile_phone character varying NOT NULL
				, body character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON messages (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON messages (mobile_phone);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
