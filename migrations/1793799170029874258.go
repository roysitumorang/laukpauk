package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799170029874258] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793799170029874258"
		// transaction_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transaction_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, transaction_id UUID NOT NULL REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_id UUID NOT NULL REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_name character varying NOT NULL
				, product_photo character varying NOT NULL
				, stock_unit character varying NOT NULL
				, base_price bigint NOT NULL
				, selling_price bigint NOT NULL
				, discount_minimum_quantity bigint NOT NULL
				, discount_amount bigint NOT NULL
				, quantity bigint NOT NULL
				, subtotal bigint NOT NULL
				, buyer_note character varying NOT NULL
				, revised_quantity bigint NOT NULL
				, revised_subtotal bigint NOT NULL
				, revision_note character varying NOT NULL
				, out_of_stock bool NOT NULL
				, UNIQUE (transaction_id, product_variant_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (product_variant_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transaction_items (product_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
