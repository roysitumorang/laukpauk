package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799170029874258] = Migration{
		Version: 1793799170029874258,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// transaction_items
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE transaction_items (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, transaction_id UUID NOT NULL
					, product_variant_id UUID NOT NULL
					, product_id UUID NOT NULL
					, product_name TEXT NOT NULL
					, product_photo TEXT NOT NULL
					, stock_unit TEXT NOT NULL
					, base_price bigint NOT NULL
					, selling_price bigint NOT NULL
					, discount_minimum_quantity bigint NOT NULL
					, discount_amount bigint NOT NULL
					, quantity bigint NOT NULL
					, subtotal bigint NOT NULL
					, buyer_note TEXT NOT NULL
					, revised_quantity bigint NOT NULL
					, revised_subtotal bigint NOT NULL
					, revision_note TEXT NOT NULL
					, out_of_stock BOOLEAN NOT NULL
					, CONSTRAINT transaction_items_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transaction_items_product_variant_id_fkey FOREIGN KEY (product_variant_id) REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transaction_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transaction_items_transaction_id_product_variant_id_key UNIQUE (transaction_id, product_variant_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transaction_items_product_variant_id_idx ON transaction_items (product_variant_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transaction_items_product_id_idx ON transaction_items (product_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE transaction_items",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
