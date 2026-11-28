package migrations

import (
	"context"
	"log/slog"
	"time"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798928531125831] = Migration{
		Version: 1793798928531125831,
		Up: func(ctx context.Context, tx database.Tx) (err error) {
			ctxt := "Migration-Up"
			// payment_channels
			if _, err = tx.Exec(
				ctx,
				`CREATE TABLE payment_channels (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, business_id TEXT NOT NULL
					, name TEXT NOT NULL
					, channel_code VARCHAR(255) NOT NULL
					, country TEXT NOT NULL
					, currency TEXT NOT NULL
					, minimum_amount bigint NOT NULL
					, maximum_amount bigint NOT NULL
					, payment_request_expiry double precision NOT NULL
					, refund BOOLEAN NOT NULL
					, channel_type TEXT NOT NULL
					, channel_category TEXT NOT NULL
					, created_at DATETIME(6) NOT NULL
					, CONSTRAINT payment_channels_channel_code_key UNIQUE (channel_code)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX payment_channels_minimum_amount_idx ON payment_channels (minimum_amount);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			if _, err = tx.Exec(
				ctx,
				`CREATE INDEX payment_channels_maximum_amount_idx ON payment_channels (maximum_amount);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
			now := time.Now().In(time.Local)
			params := [][]any{
				{
					"5f9e7c05d72844407e5289a3",
					"BCA Virtual Account",
					"BCA_VIRTUAL_ACCOUNT",
					"ID",
					"IDR",
					10_000,
					50_000_000_000,
					48,
					false,
					"BANK TRANSFER",
					"VIRTUAL_ACCOUNT",
					now,
				}, {
					"5f9e7c05d72844407e5289a3",
					"BNI Virtual Account",
					"BNI_VIRTUAL_ACCOUNT",
					"ID",
					"IDR",
					1,
					50_000_000_000,
					48,
					false,
					"BANK TRANSFER",
					"VIRTUAL_ACCOUNT",
					now,
				}, {
					"5f9e7c05d72844407e5289a3",
					"BRI Virtual Account",
					"BRI_VIRTUAL_ACCOUNT",
					"ID",
					"IDR",
					1,
					50_000_000_000,
					48,
					false,
					"BANK TRANSFER",
					"VIRTUAL_ACCOUNT",
					now,
				}, {
					"5f9e7c05d72844407e5289a3",
					"Cards",
					"CARDS",
					"ID",
					"IDR",
					5_000,
					200_000_000,
					0.166,
					true,
					"CARDS",
					"PAYMENT",
					now,
				}, {
					"5f9e7c05d72844407e5289a3",
					"Mandiri Virtual Account",
					"MANDIRI_VIRTUAL_ACCOUNT",
					"ID",
					"IDR",
					1,
					50_000_000_000,
					48,
					false,
					"BANK TRANSFER",
					"VIRTUAL_ACCOUNT",
					now,
				}, {
					"5f9e7c05d72844407e5289a3",
					"QRIS",
					"QRIS",
					"ID",
					"IDR",
					1,
					10_000_000,
					48,
					false,
					"QR CODE",
					"QRIS",
					now,
				},
			}
			for i := range params {
				if _, err = tx.Exec(
					ctx,
					`INSERT INTO payment_channels (
						business_id
						, name
						, channel_code
						, country
						, currency
						, minimum_amount
						, maximum_amount
						, payment_request_expiry
						, refund
						, channel_type
						, channel_category
						, created_at
					) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
					params[i]...,
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
				"DROP TABLE payment_channels",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			}
			return
		},
	}
}
