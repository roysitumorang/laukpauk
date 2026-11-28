package migrations

import (
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793799281976984004] = Migration{
		Version: 1793799281976984004,
		Up: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Up"
			// point_earning_caps
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE point_earning_caps (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, record_status smallint NOT NULL
					, shopping_cost_min BIGINT NOT NULL
					, shopping_cost_max BIGINT NOT NULL
					, earned_points bigint NOT NULL
					, created_by UUID NOT NULL
					, created_at DATETIME(6) NOT NULL
					, updated_by UUID NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, deleted_by UUID
					, deleted_at DATETIME(6)
					, deletion_reason TEXT
					, CONSTRAINT point_earning_caps_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT point_earning_caps_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT point_earning_caps_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earning_caps_shopping_cost_idx ON point_earning_caps (shopping_cost_min, shopping_cost_max);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earning_caps_created_by_idx ON point_earning_caps (created_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earning_caps_updated_by_idx ON point_earning_caps (updated_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX point_earning_caps_deleted_by_idx ON point_earning_caps (deleted_by);`,
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
			now := time.Now().In(time.Local)
			pointEarningCaps := [][]any{
				{
					models.RecordStatusActive,
					100_000,
					149_999,
					1_500,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					150_000,
					199_999,
					2_500,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					200_000,
					299_999,
					3_500,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					300_000,
					399_999,
					5_000,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					400_000,
					499_999,
					7_000,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					500_000,
					649_999,
					10_000,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					650_000,
					799_999,
					13_000,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					800_000,
					999_999,
					17_000,
					sellerID,
					now,
					sellerID,
					now,
				}, {
					models.RecordStatusActive,
					1_000_000,
					100_000_000,
					22_000,
					sellerID,
					now,
					sellerID,
					now,
				},
			}
			for i := range pointEarningCaps {
				if _, err = tx.Exec(
					ctx,
					`INSERT INTO point_earning_caps (
						record_status
						, shopping_cost_min
						, shopping_cost_max
						, earned_points
						, created_by
						, created_at
						, updated_by
						, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`,
					pointEarningCaps[i]...,
				); err != nil {
					helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
					return
				}
			}
			return
		},
		Down: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.Exec(
				ctx,
				"DROP TABLE point_earning_caps",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
