package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800291028912766] = Migration{
		Version: 1793800291028912766,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// recipe_items
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE recipe_items (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, recipe_id UUID NOT NULL
					, product_id UUID NOT NULL
					, product_name TEXT NOT NULL
					, product_photo TEXT NOT NULL
					, product_ready_stock BOOLEAN NOT NULL
					, product_variant_id UUID NOT NULL
					, product_variant_stock_unit TEXT NOT NULL
					, quantity bigint NOT NULL
					, note TEXT NOT NULL
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, active_recipe_id UUID AS (IF(record_status = 0, recipe_id, NULL)) PERSISTENT INVISIBLE
					, CONSTRAINT recipe_items_recipe_id_fkey FOREIGN KEY (recipe_id) REFERENCES recipes (id) ON UPDATE RESTRICT ON DELETE RESTRICT
					-- RESTRICT, not CASCADE: MariaDB won't let a generated column read a column whose foreign key cascades
					, CONSTRAINT recipe_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT recipe_items_product_variant_id_fkey FOREIGN KEY (product_variant_id) REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT recipe_items_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT recipe_items_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT recipe_items_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT recipe_items_recipe_id_product_id_idx UNIQUE (active_recipe_id, product_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_ci;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_record_status_idx ON recipe_items (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_recipe_id_idx ON recipe_items (recipe_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_product_id_idx ON recipe_items (product_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_product_ready_stock_idx ON recipe_items (product_ready_stock);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_product_variant_id_idx ON recipe_items (product_variant_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_created_by_idx ON recipe_items (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_updated_by_idx ON recipe_items (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX recipe_items_deleted_by_idx ON recipe_items (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE recipe_items",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
