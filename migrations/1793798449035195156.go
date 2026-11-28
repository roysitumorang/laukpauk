package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798449035195156] = Migration{
		Version: 1793798449035195156,
		Up: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Up"
			// vouchers
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE vouchers (
					id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
					, record_status smallint NOT NULL
					, voucher_type smallint NOT NULL
					, target_audience smallint NOT NULL
					, code text COLLATE "case_insensitive"
					, deleted_code text COLLATE "case_insensitive"
					, photo text NOT NULL
					, name text NOT NULL COLLATE "case_insensitive"
					, discount bigint NOT NULL
					, minimum_purchase bigint NOT NULL
					, quota_per_buyer bigint NOT NULL
					, effective_start_time timestamp with time zone NOT NULL
					, effective_end_time timestamp with time zone NOT NULL
					, effective_start_date timestamp with time zone NOT NULL
					, effective_end_date timestamp with time zone NOT NULL
					, description text NOT NULL
					, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, created_at timestamp with time zone NOT NULL
					, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, updated_at timestamp with time zone NOT NULL
					, deleted_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, deleted_at timestamp with time zone
					, deletion_reason text
				);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_record_status_idx ON vouchers (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_voucher_type_idx ON vouchers (voucher_type);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_target_audience_idx ON vouchers (target_audience);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE UNIQUE INDEX vouchers_unique_code_idx ON vouchers (code);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_deleted_code_idx ON vouchers (deleted_code);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_name_idx ON vouchers (name);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_effective_start_time_idx ON vouchers (effective_start_time);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_effective_end_time_idx ON vouchers (effective_end_time);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_effective_start_date_idx ON vouchers (effective_start_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_effective_end_date_idx ON vouchers (effective_end_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_created_by_idx ON vouchers (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_updated_by_idx ON vouchers (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX vouchers_deleted_by_idx ON vouchers (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
		Down: func(ctx context.Context, tx pgx.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE vouchers",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
