package migration

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/nyaruka/phonenumbers"
)

func init() {
	Migrations[1698986864428484021] = func(ctx context.Context, tx pgx.Tx) (err error) {
		rows, err := tx.Query(
			ctx,
			`SELECT
				id
				, mobile_phone
			FROM users`,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if err != nil {
			return err
		}
		defer rows.Close()
		mapPhoneNumbers := map[int64]string{}
		for rows.Next() {
			var (
				userID      int64
				phoneNumber string
			)
			if err = rows.Scan(&userID, &phoneNumber); err != nil {
				return err
			}
			num, err := phonenumbers.Parse(phoneNumber, "ID")
			if err != nil {
				return err
			}
			mapPhoneNumbers[userID] = phonenumbers.Format(num, phonenumbers.E164)
		}
		for userID, phoneNumber := range mapPhoneNumbers {
			if _, err = tx.Exec(
				ctx,
				`UPDATE users SET mobile_phone = $1 WHERE id = $2`,
				phoneNumber,
				userID,
			); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON users (role_id, mobile_phone);`,
		); err != nil {
			return err
		}
		_, err = tx.Exec(
			ctx,
			`CREATE INDEX ON users (mobile_phone);`,
		)
		return
	}
}
