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
					, origin_longitude text NOT NULL
					, origin_latitude text NOT NULL
					, destination_longitude text NOT NULL
					, destination_latitude text NOT NULL
					, distance_meters bigint NOT NULL
					, distance_km bigint NOT NULL
					, street_address text NOT NULL
					, street_address_note text NOT NULL
					, label text NOT NULL COLLATE "case_insensitive"
					, main boolean NOT NULL
					, created_at timestamp with time zone NOT NULL
					, updated_at timestamp with time zone NOT NULL
					, deleted_at timestamp with time zone
					, deletion_reason text
					, CONSTRAINT delivery_addresses_unique_user_destination UNIQUE (
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
				`CREATE INDEX delivery_addresses_record_status_idx ON delivery_addresses (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX delivery_addresses_origin_longitude_idx ON delivery_addresses (origin_longitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX delivery_addresses_origin_latitude_idx ON delivery_addresses (origin_latitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX delivery_addresses_destination_longitude_idx ON delivery_addresses (destination_longitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX delivery_addresses_destination_latitude_idx ON delivery_addresses (destination_latitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX delivery_addresses_label_user_id_idx ON delivery_addresses (label, user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX delivery_addresses_main_idx ON delivery_addresses (main);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE UNIQUE INDEX delivery_addresses_unique_main_per_user_idx ON delivery_addresses (user_id) WHERE main;`,
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
