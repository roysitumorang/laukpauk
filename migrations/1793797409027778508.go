package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797409027778508] = Migration{
		Version: 1793797409027778508,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// product_variants
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE product_variants (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, product_category_id UUID NOT NULL
					, product_category_name TEXT NOT NULL
					, product_id UUID NOT NULL
					, product_name TEXT NOT NULL
					, product_photo TEXT NOT NULL
					, product_ready_stock BOOLEAN NOT NULL
					, stock_unit VARCHAR(255) NOT NULL
					, weight_in_grams bigint NOT NULL
					, base_price bigint NOT NULL
					, selling_price bigint NOT NULL
					, discount_type smallint NOT NULL
					, discount_rate bigint NOT NULL
					, discount_minimum_quantity bigint NOT NULL
					, discount_amount bigint NOT NULL
					, sale_price bigint NOT NULL
					, description TEXT NOT NULL
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, active_stock_unit_key VARCHAR(255) AS (IF(record_status = 0, LOWER(stock_unit), NULL)) PERSISTENT INVISIBLE
					, CONSTRAINT product_variants_product_category_id_fkey FOREIGN KEY (product_category_id) REFERENCES product_categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_variants_product_id_fkey FOREIGN KEY (product_id) REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_variants_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_variants_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_variants_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_variants_lower_product_id_idx UNIQUE (active_stock_unit_key, product_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_variants_record_status_idx ON product_variants (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_variants_product_id_idx ON product_variants (product_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_variants_created_by_idx ON product_variants (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_variants_updated_by_idx ON product_variants (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_variants_deleted_by_idx ON product_variants (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE product_variants",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
