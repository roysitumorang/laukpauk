package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798673622250513] = Migration{
		Version: 1793798673622250513,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// cart_items
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE cart_items (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, cart_id UUID NOT NULL
					, product_variant_id UUID NOT NULL
					, product_category_id UUID NOT NULL
					, product_category_name TEXT NOT NULL
					, product_id UUID NOT NULL
					, product_name TEXT NOT NULL
					, product_photo TEXT NOT NULL
					, product_ready_stock BOOLEAN NOT NULL
					, stock_unit TEXT NOT NULL
					, base_price bigint NOT NULL
					, selling_price bigint NOT NULL
					, discount_type smallint NOT NULL
					, discount_rate bigint NOT NULL
					, discount_minimum_quantity bigint NOT NULL
					, discount_amount bigint NOT NULL
					, sale_price bigint NOT NULL
					, description TEXT NOT NULL
					, quantity bigint NOT NULL
					, note TEXT NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, CONSTRAINT cart_items_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES carts (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT cart_items_product_variant_id_fkey FOREIGN KEY (product_variant_id) REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT cart_items_product_category_id_fkey FOREIGN KEY (product_category_id) REFERENCES product_categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT cart_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT cart_items_cart_id_product_variant_id_key UNIQUE (cart_id, product_variant_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_ci;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX cart_items_product_variant_id_idx ON cart_items (product_variant_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE cart_items",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
