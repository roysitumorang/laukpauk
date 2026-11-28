package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799478525813455] = Migration{
		Version: 1793799478525813455,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// point_earnings
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE point_earnings (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, start_date timestamp with time zone NOT NULL
					, end_date timestamp with time zone NOT NULL
					, amount bigint NOT NULL
					, created_at timestamp with time zone NOT NULL
					, updated_at timestamp with time zone NOT NULL
					, CONSTRAINT point_earnings_unique_user_id_start_date_end_date UNIQUE (user_id, start_date, end_date)
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earnings_start_date_idx ON point_earnings (start_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earnings_end_date_idx ON point_earnings (end_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE point_earnings",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
