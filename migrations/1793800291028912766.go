package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800291028912766] = Migration{
		Version: 1793800291028912766,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// recipe_items
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE recipe_items (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, record_status smallint NOT NULL
					, recipe_id UUID NOT NULL REFERENCES recipes (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, product_id UUID NOT NULL REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, product_name text NOT NULL
					, product_photo text NOT NULL
					, product_ready_stock BOOLEAN NOT NULL
					, product_variant_id UUID NOT NULL REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, product_variant_stock_unit text NOT NULL
					, quantity bigint NOT NULL
					, note text NOT NULL
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
				`CREATE INDEX recipe_items_record_status_idx ON recipe_items (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX recipe_items_recipe_id_idx ON recipe_items (recipe_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX recipe_items_product_id_idx ON recipe_items (product_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE UNIQUE INDEX recipe_items_unique_recipe_product_idx ON recipe_items (recipe_id, product_id) WHERE record_status = 0;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX recipe_items_product_ready_stock_idx ON recipe_items (product_ready_stock);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX recipe_items_product_variant_id_idx ON recipe_items (product_variant_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX recipe_items_created_by_idx ON recipe_items (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX recipe_items_updated_by_idx ON recipe_items (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX recipe_items_deleted_by_idx ON recipe_items (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE recipe_items",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
