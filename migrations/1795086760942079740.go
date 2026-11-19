package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1795086760942079740] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1795086760942079740"
		// addresses
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE addresses (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, record_status smallint NOT NULL
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, longitude character varying NOT NULL
				, latitude character varying NOT NULL
				, street_address character varying NOT NULL
				, delivery_note character varying NOT NULL
				, label character varying NOT NULL
				, recipient_name character varying NOT NULL
				, recipient_mobile_phone character varying NOT NULL
				, main boolean NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, UNIQUE (
					user_id
					, longitude
					, latitude
				)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON addresses (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON addresses (longitude);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON addresses (latitude);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON addresses (LOWER(label), user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON addresses (main);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
