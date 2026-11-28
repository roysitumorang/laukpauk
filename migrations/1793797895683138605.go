package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797895683138605] = Migration{
		Version: 1793797895683138605,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// text_messages
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE text_messages (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, user_id UUID NOT NULL
					, mobile_phone TEXT NOT NULL
					, message_type smallint NOT NULL
					, body TEXT NOT NULL
					, delivery_status VARCHAR(255)
					, message_id bigint
					, created_by UUID
					, created_at DATETIME(6) NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, sent_at DATETIME(6)
					, CONSTRAINT text_messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT text_messages_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX text_messages_user_id_idx ON text_messages (user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX text_messages_message_type_idx ON text_messages (message_type);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX text_messages_delivery_status_idx ON text_messages (delivery_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX text_messages_message_id_idx ON text_messages (message_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX text_messages_created_by_idx ON text_messages (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE text_messages",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
