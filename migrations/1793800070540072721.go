package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800070540072721] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793800070540072721"
		// content_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE content_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, content_id UUID NOT NULL REFERENCES contents (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, item character varying NOT NULL
				, option character varying
				, description character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deactivated_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deactivated_at timestamp with time zone
				, deactivation_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON content_items (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
