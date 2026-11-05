package migrations

import (
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793798002445732369] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793798002445732369"
		// delivery_plans
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE delivery_plans (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, record_status smallint NOT NULL
				, name character varying COLLATE "natural_sort"
				, weekday bigint NOT NULL
				, transaction_window TSTZRANGE NOT NULL
				, delivery_time_interval interval NOT NULL
				, same_day BOOLEAN NOT NULL
				, same_day_window TSTZRANGE NOT NULL
				, next_day BOOLEAN NOT NULL
				, next_day_window TSTZRANGE NOT NULL
				, next_two_day BOOLEAN NOT NULL
				, next_two_day_window TSTZRANGE NOT NULL
				, created_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, created_at timestamp with time zone NOT NULL
				, updated_by UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, updated_at timestamp with time zone NOT NULL
				, deleted_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, deleted_at timestamp with time zone
				, deletion_reason character varying
				, CONSTRAINT no_overlapping_delivery_plans_transaction_window EXCLUDE USING gist (
					company_id WITH =,
					weekday WITH =,
					transaction_window WITH &&
				)
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (record_status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (LOWER(name));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (weekday);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (created_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (updated_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON delivery_plans (deleted_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		var sellerID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`SELECT id
			FROM users
			WHERE email = $1`,
			"benni.surbakti@gmail.com",
		).Scan(&sellerID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		var companyID uuid.UUID
		if err = tx.QueryRow(
			ctx,
			`SELECT id
			FROM companies
			WHERE subdomain = $1`,
			"laukpauk",
		).Scan(&companyID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return
		}
		now := time.Now().In(time.Local)
		deliveryTimeInterval := 2 * time.Hour
		today000000 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today095959 := time.Date(now.Year(), now.Month(), now.Day(), 9, 59, 59, 0, time.Local).Format(time.RFC3339Nano)
		today100000 := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today120000 := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today215959 := time.Date(now.Year(), now.Month(), now.Day(), 21, 59, 59, 0, time.Local).Format(time.RFC3339Nano)
		today220000 := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		today235959 := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local).Format(time.RFC3339Nano)
		nextDay080000 := time.Date(now.Year(), now.Month(), now.Day()+1, 8, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextDay100000 := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextDay120000 := time.Date(now.Year(), now.Month(), now.Day()+1, 12, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextTwoDay080000 := time.Date(now.Year(), now.Month(), now.Day()+2, 8, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		nextTwoDay120000 := time.Date(now.Year(), now.Month(), now.Day()+2, 12, 0, 0, 0, time.Local).Format(time.RFC3339Nano)
		deliveryPlans := [][]any{
			// weekdays
			{
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 10:00",
				models.DeliveryWeekdayWeekdays,
				today000000,
				today095959,
				deliveryTimeInterval,
				true,
				today100000,
				today120000,
				true,
				nextDay080000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi jam 10:00 - 21:59",
				models.DeliveryWeekdayWeekdays,
				today100000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdayWeekdays,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay100000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			},
			// friday
			{
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 10:00",
				models.DeliveryWeekdayFriday,
				today000000,
				today095959,
				deliveryTimeInterval,
				true,
				today100000,
				today120000,
				true,
				nextDay080000,
				nextDay120000,
				false,
				nil,
				nil,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi jam 10:00 - 21:59",
				models.DeliveryWeekdayFriday,
				today100000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				false,
				nil,
				nil,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdayFriday,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				false,
				nil,
				nil,
				sellerID,
				now,
				sellerID,
				now,
			},
			// saturday
			{
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 10:00",
				models.DeliveryWeekdaySaturday,
				today000000,
				today095959,
				deliveryTimeInterval,
				true,
				today100000,
				today120000,
				false,
				nil,
				nil,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi jam 10:00 - 21:59",
				models.DeliveryWeekdaySaturday,
				today100000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				false,
				nil,
				nil,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdaySaturday,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				false,
				nil,
				nil,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			},
			// sunday
			{
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi sebelum jam 22:00",
				models.DeliveryWeekdaySunday,
				today000000,
				today215959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay080000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			}, {
				companyID,
				models.RecordStatusActive,
				"Estimasi pengantaran untuk transaksi mulai jam 22:00",
				models.DeliveryWeekdaySunday,
				today220000,
				today235959,
				deliveryTimeInterval,
				false,
				nil,
				nil,
				true,
				nextDay100000,
				nextDay120000,
				true,
				nextTwoDay080000,
				nextTwoDay120000,
				sellerID,
				now,
				sellerID,
				now,
			},
		}
		for i := range deliveryPlans {
			if _, err = tx.Exec(
				ctx,
				`INSERT INTO delivery_plans (
					company_id
					, record_status
					, name
					, weekday
					, transaction_window
					, delivery_time_interval
					, same_day
					, same_day_window
					, next_day
					, next_day_window
					, next_two_day
					, next_two_day_window
					, created_by
					, created_at
					, updated_by
					, updated_at
				) VALUES (
					$1
					, $2
					, $3
					, $4
					, tstzrange($5, $6, '[]')
					, $7
					, $8
					, tstzrange($9, $10)
					, $11
					, tstzrange($12, $13, '[]')
					, $14
					, tstzrange($15, $16, '[]')
					, $17
					, $18
					, $19
					, $20
				)`,
				deliveryPlans[i]...,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
		}
		return
	}
}
