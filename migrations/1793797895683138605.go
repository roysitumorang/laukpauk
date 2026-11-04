package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793797895683138605] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793797895683138605"
		// text_messages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE text_messages (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, mobile_phone character varying NOT NULL
				, message_type smallint NOT NULL
				, body character varying NOT NULL
				, delivery_status character varying
				, message_id bigint
				, created_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, sent_at timestamp with time zone
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (message_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (delivery_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (message_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON text_messages (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
