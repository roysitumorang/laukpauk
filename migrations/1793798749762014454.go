package migrations

import (
	"context"
	"log/slog"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798749762014454] = Migration{
		Version: 1793798749762014454,
		Up: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Up"
			// delivery_addresses
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE delivery_addresses (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, user_id UUID NOT NULL
					, origin_longitude VARCHAR(255) NOT NULL
					, origin_latitude VARCHAR(255) NOT NULL
					, destination_longitude VARCHAR(255) NOT NULL
					, destination_latitude VARCHAR(255) NOT NULL
					, distance_meters bigint NOT NULL
					, distance_km bigint NOT NULL
					, street_address TEXT NOT NULL
					, street_address_note TEXT NOT NULL
					, label TEXT NOT NULL
					, main BOOLEAN NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, main_user_id UUID AS (IF(main, user_id, NULL)) PERSISTENT INVISIBLE
					, CONSTRAINT delivery_addresses_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE RESTRICT ON DELETE RESTRICT
					-- RESTRICT, not CASCADE: MariaDB won't let a generated column read a column whose foreign key cascades
					, CONSTRAINT delivery_addresses_user_id_destination_longitude_destination_lat UNIQUE (user_id, destination_longitude, destination_latitude)
					, CONSTRAINT one_main_address_per_user_idx UNIQUE (main_user_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
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
				`CREATE INDEX delivery_addresses_main_idx ON delivery_addresses (main);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			return
		},
		Down: func(ctx context.Context, tx database.Tx) (err error) {
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
