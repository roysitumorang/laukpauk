package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799170029874258] = Migration{
		Version: 1793799170029874258,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// transaction_items
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE transaction_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, transaction_id UUID NOT NULL REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_id UUID NOT NULL REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_name text NOT NULL
				, product_photo text NOT NULL
				, stock_unit text NOT NULL
				, base_price bigint NOT NULL
				, selling_price bigint NOT NULL
				, discount_minimum_quantity bigint NOT NULL
				, discount_amount bigint NOT NULL
				, quantity bigint NOT NULL
				, subtotal bigint NOT NULL
				, buyer_note text NOT NULL
				, revised_quantity bigint NOT NULL
				, revised_subtotal bigint NOT NULL
				, revision_note text NOT NULL
				, out_of_stock bool NOT NULL
				, CONSTRAINT transaction_items_unique_transaction_id_product_variant_id UNIQUE (transaction_id, product_variant_id)
			);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX transaction_items_product_variant_id_idx ON transaction_items (product_variant_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX transaction_items_product_id_idx ON transaction_items (product_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE transaction_items",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
