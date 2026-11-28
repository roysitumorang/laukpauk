package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798581424684143] = Migration{
		Version: 1793798581424684143,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// carts
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE carts (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, user_id UUID
					, deleted_user_id UUID
					, created_at DATETIME(6) NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, CONSTRAINT carts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT carts_user_id_key UNIQUE (user_id)
					, CONSTRAINT carts_deleted_user_id_fkey FOREIGN KEY (deleted_user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX carts_deleted_user_id_idx ON carts (deleted_user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE carts",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
