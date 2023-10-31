package migration

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	roleModel "github.com/roysitumorang/laukpauk/modules/role/model"
	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
)

func init() {
	Migrations[1698311848047021856] = func(ctx context.Context, tx pgx.Tx) (err error) {
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON users (mobile_phone);`,
		); err != nil {
			return
		}
		if _, err = tx.Exec(
			ctx,
			`ALTER TABLE users
				ADD COLUMN business_days jsonb
				, ALTER COLUMN activated_at TYPE timestamp with time zone USING activated_at::timestamp with time zone
				, ALTER COLUMN date_of_birth TYPE timestamp with time zone USING date_of_birth::timestamp with time zone
				, ALTER COLUMN business_opening_hour TYPE smallint USING business_opening_hour::smallint
				, ALTER COLUMN business_closing_hour TYPE smallint USING business_closing_hour::smallint
				, ALTER COLUMN registration_ip TYPE inet USING registration_ip::inet
				, ALTER created_at SET NOT NULL
				, ALTER COLUMN created_at TYPE timestamp with time zone USING created_at::timestamp with time zone
				, ALTER updated_at SET NOT NULL
				, ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
		); err != nil {
			return
		}
		rows, err := tx.Query(
			ctx,
			`SELECT
				id
				, open_on_sunday = 1
				, open_on_monday = 1
				, open_on_tuesday = 1
				, open_on_wednesday = 1
				, open_on_thursday = 1
				, open_on_friday = 1
				, open_on_saturday = 1
			FROM users
			WHERE role_id = $1;`,
			roleModel.RoleSeller,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if err != nil {
			return
		}
		defer rows.Close()
		mapUserBusinessDays := map[int64]userModel.BusinessDays{}
		for rows.Next() {
			var (
				userID       int64
				businessDays userModel.BusinessDays
			)
			if err = rows.Scan(
				&userID,
				&businessDays.Sunday,
				&businessDays.Monday,
				&businessDays.Tuesday,
				&businessDays.Wednesday,
				&businessDays.Thursday,
				&businessDays.Friday,
				&businessDays.Saturday,
			); err != nil {
				return
			}
			mapUserBusinessDays[userID] = businessDays
		}
		for userID, businessDays := range mapUserBusinessDays {
			businessDaysByte, err := json.Marshal(businessDays)
			if err != nil {
				return err
			}
			if _, err = tx.Exec(
				ctx,
				`UPDATE users SET
					business_days = $1
				WHERE id = $2;`,
				helper.ByteSlice2String(businessDaysByte),
				userID,
			); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(
			ctx,
			`ALTER TABLE users
				ADD COLUMN business_delivery_hours integer[]
				, DROP COLUMN open_on_sunday
				, DROP COLUMN open_on_monday
				, DROP COLUMN open_on_tuesday
				, DROP COLUMN open_on_wednesday
				, DROP COLUMN open_on_thursday
				, DROP COLUMN open_on_friday
				, DROP COLUMN open_on_saturday;`,
		); err != nil {
			return
		}
		rows, err = tx.Query(
			ctx,
			`SELECT
				id
				, delivery_hours
			FROM users
			WHERE delivery_hours IS NOT NULL;`,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if err != nil {
			return
		}
		defer rows.Close()
		mapUserDeliveryHours := map[int64][]int{}
		for rows.Next() {
			var (
				userID           int64
				deliveryHoursStr string
				deliveryHours    []int
			)
			if err = rows.Scan(
				&userID,
				&deliveryHoursStr,
			); err != nil {
				return
			}
			items := strings.Split(deliveryHoursStr, ",")
			for _, item := range items {
				deliveryHour, err := strconv.Atoi(item)
				if err != nil {
					return err
				}
				deliveryHours = append(deliveryHours, deliveryHour)
			}
			mapUserDeliveryHours[userID] = deliveryHours
		}
		for userID, deliveryHours := range mapUserDeliveryHours {
			if _, err = tx.Exec(
				ctx,
				`UPDATE users SET
					business_delivery_hours = $1
				WHERE id = $2;`,
				deliveryHours,
				userID,
			); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(
			ctx,
			`ALTER TABLE users
				DROP COLUMN delivery_hours;`,
		); err != nil {
			return
		}
		_, err = tx.Exec(
			ctx,
			`ALTER TABLE users
				RENAME COLUMN business_delivery_hours TO delivery_hours;`,
		)
		return
	}
}
