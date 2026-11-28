package migrations

import (
	"context"
	"log/slog"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793800358695041192] = Migration{
		Version: 1793800358695041192,
		Up: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Up"
			// message_receipts
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE message_receipts (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, message_id UUID NOT NULL
					, user_id UUID NOT NULL
					, read_at DATETIME(6) NOT NULL
					, CONSTRAINT message_receipts_message_id_fkey FOREIGN KEY (message_id) REFERENCES messages (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT message_receipts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT message_receipts_message_id_user_id_key UNIQUE (message_id, user_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX message_receipts_user_id_idx ON message_receipts (user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE message_receipts",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
