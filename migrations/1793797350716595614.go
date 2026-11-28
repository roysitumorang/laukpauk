package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797350716595614] = Migration{
		Version: 1793797350716595614,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// products
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE products (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, record_status smallint NOT NULL
					, product_category_id UUID REFERENCES product_categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, product_category_name text NOT NULL COLLATE "natural_sort"
					, name text NOT NULL COLLATE "case_insensitive"
					, photo text NOT NULL
					, description text NOT NULL
					, ready_stock BOOLEAN NOT NULL
					, pinned BOOLEAN NOT NULL DEFAULT FALSE
					, hits bigint NOT NULL
					, lowest_selling_price bigint NOT NULL DEFAULT 0
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
				`CREATE INDEX ON products (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON products (product_category_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE UNIQUE INDEX products_unique_name_idx ON products (name) WHERE record_status = 0;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON products (ready_stock);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON products (pinned);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON products (hits);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON products (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON products (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON products (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE products",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
