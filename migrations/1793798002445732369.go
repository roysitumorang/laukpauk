package migrations

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793798002445732369] = Migration{
		Version: 1793798002445732369,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// delivery_plans
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE delivery_plans (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, name TEXT
					, weekday bigint NOT NULL
					, transaction_window_min DATETIME(6) NOT NULL
					, transaction_window_max DATETIME(6) NOT NULL
					, delivery_time_interval bigint NOT NULL
					, same_day BOOLEAN NOT NULL
					, same_day_window_min DATETIME(6)
					, same_day_window_max DATETIME(6)
					, next_day BOOLEAN NOT NULL
					, next_day_window_min DATETIME(6)
					, next_day_window_max DATETIME(6)
					, next_two_day BOOLEAN NOT NULL
					, next_two_day_window_min DATETIME(6)
					, next_two_day_window_max DATETIME(6)
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, CONSTRAINT delivery_plans_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT delivery_plans_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT delivery_plans_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX delivery_plans_record_status_idx ON delivery_plans (record_status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX delivery_plans_weekday_idx ON delivery_plans (weekday);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX delivery_plans_created_by_idx ON delivery_plans (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX delivery_plans_updated_by_idx ON delivery_plans (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX delivery_plans_deleted_by_idx ON delivery_plans (deleted_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			var sellerID uuid.UUID
			if err = tx.QueryRowContext(
				ctx,
				`SELECT id
				FROM users
				WHERE email = ?`,
				helper.SQLArgs("benni.surbakti@gmail.com")...,
			).Scan(helper.SQLScan(&sellerID)...); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
				return
			}
			now := time.Now().In(time.Local)
			deliveryTimeInterval := 2 * time.Hour
			today000000 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
			today095959 := time.Date(now.Year(), now.Month(), now.Day(), 9, 59, 59, 0, time.Local)
			today100000 := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local)
			today120000 := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
			today215959 := time.Date(now.Year(), now.Month(), now.Day(), 21, 59, 59, 0, time.Local)
			today220000 := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, time.Local)
			today235959 := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
			nextDay080000 := time.Date(now.Year(), now.Month(), now.Day()+1, 8, 0, 0, 0, time.Local)
			nextDay100000 := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, time.Local)
			nextDay120000 := time.Date(now.Year(), now.Month(), now.Day()+1, 12, 0, 0, 0, time.Local)
			nextTwoDay080000 := time.Date(now.Year(), now.Month(), now.Day()+2, 8, 0, 0, 0, time.Local)
			nextTwoDay120000 := time.Date(now.Year(), now.Month(), now.Day()+2, 12, 0, 0, 0, time.Local)
			deliveryPlans := [][]any{
				// weekdays
				{
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
				if _, err = tx.ExecContext(
					ctx,
					`INSERT INTO delivery_plans (
						record_status
						, name
						, weekday
						, transaction_window_min
						, transaction_window_max
						, delivery_time_interval
						, same_day
						, same_day_window_min
						, same_day_window_max
						, next_day
						, next_day_window_min
						, next_day_window_max
						, next_two_day
						, next_two_day_window_min
						, next_two_day_window_max
						, created_by
						, created_at
						, updated_by
						, updated_at
					) VALUES (
						?
						, ?
						, ?
						, ?, ?
						, ?
						, ?
						, ?, ?
						, ?
						, ?, ?
						, ?
						, ?, ?
						, ?
						, ?
						, ?
						, ?
					)`,
					helper.SQLArgs(deliveryPlans[i]...)...,
				); err != nil {
					helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
					return
				}
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE delivery_plans",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
