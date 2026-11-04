package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798581424684143] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793798581424684143"
		// carts
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE carts (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT UNIQUE
				, deleted_user_id UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON carts (deleted_user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
