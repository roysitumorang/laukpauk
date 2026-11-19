package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798673622250513] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793798673622250513"
		// cart_items
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE cart_items (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, cart_id UUID NOT NULL REFERENCES carts (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_variant_id UUID NOT NULL REFERENCES product_variants (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_category_id UUID NOT NULL REFERENCES product_categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_category_name character varying NOT NULL COLLATE "natural_sort"
				, product_id UUID NOT NULL REFERENCES products (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, product_name character varying NOT NULL COLLATE "natural_sort"
				, product_photo character varying NOT NULL
				, product_ready_stock bool NOT NULL
				, stock_unit character varying NOT NULL COLLATE "natural_sort"
				, base_price bigint NOT NULL
				, selling_price bigint NOT NULL
				, discount_type smallint NOT NULL
				, discount_rate bigint NOT NULL
				, discount_minimum_quantity bigint NOT NULL
				, discount_amount bigint NOT NULL
				, sale_price bigint NOT NULL
				, description character varying NOT NULL
				, quantity bigint NOT NULL
				, note character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, UNIQUE (cart_id, product_variant_id)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON cart_items (product_variant_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
