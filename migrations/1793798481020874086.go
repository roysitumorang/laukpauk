package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798481020874086] = Migration{
		Version: 1793798481020874086,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// vouchers_users
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE vouchers_users (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, voucher_id UUID NOT NULL REFERENCES vouchers (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, created_at timestamp with time zone NOT NULL
					, CONSTRAINT vouchers_users_unique_voucher_id_user_id UNIQUE (voucher_id, user_id)
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_users_user_id_idx ON vouchers_users (user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE vouchers_users",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
