package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800182721188179] = Migration{
		Version: 1793800182721188179,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// recipes
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE recipes (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, record_status smallint NOT NULL
					, name text
					, deleted_name text
					, photo text NOT NULL
					, description text NOT NULL
					, instagram_link text NOT NULL
					, facebook_link text NOT NULL
					, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, created_at timestamp with time zone NOT NULL
					, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, updated_at timestamp with time zone NOT NULL
					, deleted_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, deleted_at timestamp with time zone
					, deletion_reason text
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON recipes (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE UNIQUE INDEX ON recipes (LOWER(name)) WHERE record_status = 0;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON recipes (LOWER(deleted_name));`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON recipes (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON recipes (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON recipes (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE recipes",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
