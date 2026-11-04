package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800034903509882] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793800034903509882"
		// contents
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE contents (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, page character varying NOT NULL
				, division character varying NOT NULL
				, content_type character varying NOT NULL
				, route character varying
				, ordinal_number bigint NOT NULL
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
			`CREATE INDEX ON contents (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON contents (deactivated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
