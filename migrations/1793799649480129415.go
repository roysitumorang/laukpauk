package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799649480129415] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793799649480129415"
		// points
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE points (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, record_status smallint NOT NULL
				, point_type smallint NOT NULL
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, amount bigint NOT NULL
				, transaction_id UUID REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, description character varying NOT NULL
				, created_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (point_type);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (transaction_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (LOWER(description));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON points (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
