package migrations

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func init() {
	Migrations[1793769027240607021] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793769027240607021"
		now := time.Now().In(time.Local)
		password, err := helper.Base64Decode(os.Getenv("DEFAULT_PASSWORD"))
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrBase64Decode")
			return
		}
		encryptedPassword, err := helper.HashPassword(password)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrHashPassword")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO users (
				role
				, status
				, name
				, mobile_phone_confirmed_at
				, mobile_phone
				, encrypted_password
				, photo
				, user_level
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id`,
			models.RoleSeller,
			models.UserStatusActive,
			"Laukpauk",
			now,
			"+628116099887",
			encryptedPassword,
			"seller-default.png",
			models.SellerLevelOwner,
			now,
			now,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
