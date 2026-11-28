package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797079227835743] = Migration{
		Version: 1793797079227835743,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// product_categories
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE product_categories (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, name VARCHAR(255) NOT NULL
					, description TEXT NOT NULL
					, photo TEXT NOT NULL
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, active_name_key VARCHAR(255) AS (IF(record_status = 0, LOWER(name), NULL)) PERSISTENT INVISIBLE
					, CONSTRAINT product_categories_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_categories_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_categories_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT product_categories_unique_name_idx UNIQUE (active_name_key)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_categories_record_status_idx ON product_categories (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_categories_created_by_idx ON product_categories (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_categories_updated_by_idx ON product_categories (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX product_categories_deleted_by_idx ON product_categories (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE product_categories",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
