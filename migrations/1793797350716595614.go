package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797350716595614] = Migration{
		Version: 1793797350716595614,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// products
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE products (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, product_category_id UUID
					, product_category_name TEXT NOT NULL
					, name VARCHAR(255) NOT NULL
					, photo TEXT NOT NULL
					, description TEXT NOT NULL
					, ready_stock BOOLEAN NOT NULL
					, pinned BOOLEAN NOT NULL DEFAULT FALSE
					, hits bigint NOT NULL
					, lowest_selling_price bigint NOT NULL DEFAULT 0
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, active_name_key VARCHAR(255) AS (IF(record_status = 0, LOWER(name), NULL)) PERSISTENT INVISIBLE
					, CONSTRAINT products_product_category_id_fkey FOREIGN KEY (product_category_id) REFERENCES product_categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT products_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT products_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT products_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT products_unique_name_idx UNIQUE (active_name_key)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_record_status_idx ON products (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_product_category_id_idx ON products (product_category_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_ready_stock_idx ON products (ready_stock);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_pinned_idx ON products (pinned);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_hits_idx ON products (hits);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_created_by_idx ON products (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_updated_by_idx ON products (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX products_deleted_by_idx ON products (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE products",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
