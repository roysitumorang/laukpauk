package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797079227835743] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793797079227835743"
		// product_categories
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE product_categories (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status smallint NOT NULL
				, name character varying NOT NULL COLLATE "natural_sort"
				, description character varying NOT NULL
				, photo character varying NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX product_categories_unique_name_idx ON product_categories (LOWER(name)) WHERE record_status = 0;`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON product_categories (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
