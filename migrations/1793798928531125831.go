package migrations

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793798928531125831] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793798928531125831"
		// payment_channels
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE payment_channels (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, business_id character varying NOT NULL
				, name character varying NOT NULL
				, channel_code character varying NOT NULL UNIQUE
				, country character varying NOT NULL
				, currency character varying NOT NULL
				, minimum_amount bigint NOT NULL
				, maximum_amount bigint NOT NULL
				, payment_request_expiry double precision NOT NULL
				, refund boolean NOT NULL
				, channel_type character varying NOT NULL
				, channel_category character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			"CREATE INDEX ON payment_channels (LOWER(name));",
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			"CREATE INDEX ON payment_channels (minimum_amount);",
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			"CREATE INDEX ON payment_channels (maximum_amount);",
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
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`,
				params[i]...,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
				return
			}
		}
		return
	}
}
