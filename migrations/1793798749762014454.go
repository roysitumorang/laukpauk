package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798749762014454] = Migration{
		Version: 1793798749762014454,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// delivery_addresses
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE delivery_addresses (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, record_status smallint NOT NULL
					, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, origin_longitude character varying NOT NULL
					, origin_latitude character varying NOT NULL
					, destination_longitude character varying NOT NULL
					, destination_latitude character varying NOT NULL
					, distance_meters bigint NOT NULL
					, distance_km bigint NOT NULL
					, street_address character varying NOT NULL
					, street_address_note character varying NOT NULL
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
						, destination_longitude
						, destination_latitude
					)
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_addresses (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_addresses (origin_longitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_addresses (origin_latitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_addresses (destination_longitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_addresses (destination_latitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_addresses (LOWER(label), user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_addresses (main);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE UNIQUE INDEX one_main_address_per_user_idx ON delivery_addresses (user_id) WHERE main;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE delivery_addresses",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
