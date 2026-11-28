package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798481020874086] = Migration{
		Version: 1793798481020874086,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// vouchers_users
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE vouchers_users (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, voucher_id UUID NOT NULL
					, user_id UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, CONSTRAINT vouchers_users_voucher_id_fkey FOREIGN KEY (voucher_id) REFERENCES vouchers (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT vouchers_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT vouchers_users_voucher_id_user_id_key UNIQUE (voucher_id, user_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX vouchers_users_user_id_idx ON vouchers_users (user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE vouchers_users",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
