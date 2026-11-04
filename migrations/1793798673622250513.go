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
				, quantity bigint NOT NULL
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
