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
			// delivery_routes
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE delivery_routes (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, origin_longitude character varying NOT NULL
					, origin_latitude character varying NOT NULL
					, destination_longitude character varying NOT NULL
					, destination_latitude character varying NOT NULL
					, distance_meters bigint NOT NULL
					, distance_km bigint NOT NULL
					, created_at timestamp with time zone NOT NULL
					, updated_at timestamp with time zone NOT NULL
					, UNIQUE (
						origin_longitude
						, origin_latitude
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
				`CREATE INDEX ON delivery_routes (origin_longitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_routes (origin_latitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_routes (destination_longitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON delivery_routes (destination_latitude);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`ALTER TABLE users ADD FOREIGN KEY (delivery_route_id) REFERENCES delivery_routes (id) ON UPDATE CASCADE ON DELETE SET NULL;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX ON users (delivery_route_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				`DROP INDEX users_delivery_route_id_idx`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`ALTER TABLE users
					DROP CONSTRAINT users_delivery_route_id_fkey`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE delivery_routes",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
