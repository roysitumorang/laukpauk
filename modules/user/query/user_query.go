package query

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/user/model"
	"go.uber.org/zap"
)

type (
	userQuery struct {
		dbRead, dbWrite *pgxpool.Pool
	}
)

func NewUserQuery(
	dbRead,
	dbWrite *pgxpool.Pool,
) UserQuery {
	return &userQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *userQuery) FindUsers(ctx context.Context, filter model.UserFilter) (response []model.User, err error) {
	ctxt := "UserQuery-FindUsers"
	var (
		params     []interface{}
		conditions []string
	)
	if n := len(filter.UserIDs); n > 0 {
		placeholders := make([]string, n)
		for i, userID := range filter.UserIDs {
			params = append(params, userID)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		conditions = append(conditions, fmt.Sprintf("u.id IN (%s)", strings.Join(placeholders, ",")))
	}
	if n := len(filter.RoleIDs); n > 0 {
		placeholders := make([]string, n)
		for i, roleID := range filter.RoleIDs {
			params = append(params, roleID)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		conditions = append(conditions, fmt.Sprintf("u.role_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if n := len(filter.MobilePhones); n > 0 {
		placeholders := make([]string, n)
		for i, mobilePhone := range filter.MobilePhones {
			params = append(params, mobilePhone)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		joinedPlaceholders := strings.Join(placeholders, ",")
		conditions = append(conditions, fmt.Sprintf("(u.mobile_phone IN (%s) OR u.email IN (%s))", joinedPlaceholders, joinedPlaceholders))
	}
	if n := len(filter.Status); n > 0 {
		placeholders := make([]string, n)
		for i, status := range filter.Status {
			params = append(params, status)
			placeholders[i] = fmt.Sprintf("$%d", len(params))
		}
		conditions = append(conditions, fmt.Sprintf("u.status IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(params) == 0 {
		return
	}
	rows, err := q.dbRead.Query(
		ctx,
		fmt.Sprintf(
			`SELECT
				u.id
				, u.role_id
				, r.name
				, u.api_key
				, u.merchant_note
				, u.minimum_purchase
				, u.admin_fee
				, u.accumulation_divisor
				, u.name
				, u.email
				, u.password
				, u.address
				, u.village_id
				, v.name
				, u.subdistrict_id
				, s.name
				, c.id
				, c.type || ' ' || c.name
				, p.id
				, p.name
				, u.mobile_phone
				, u.device_token
				, u.status
				, u.activated_at
				, u.activation_token
				, u.password_reset_token
				, u.deposit
				, u.company
				, u.registration_ip
				, u.gender
				, u.date_of_birth
				, u.avatar
				, u.thumbnails
				, u.open_on_sunday
				, u.open_on_monday
				, u.open_on_tuesday
				, u.open_on_wednesday
				, u.open_on_thursday
				, u.open_on_friday
				, u.open_on_saturday
				, u.business_opening_hour
				, u.business_closing_hour
				, u.delivery_hours
				, u.latitude
				, u.longitude
				, u.keywords
				, u.delivery_max_distance
				, u.delivery_free_distance
				, u.delivery_rate
				, u.created_by
				, u.created_at
				, u.updated_by
				, u.updated_at
			FROM users u
			JOIN roles r ON u.role_id = r.id
			JOIN villages v ON u.village_id = v.id
			JOIN subdistricts s ON u.subdistrict_id = s.id
			JOIN cities c ON s.city_id = c.id
			JOIN provinces p ON c.province_id = p.id
			WHERE (%s)`,
			strings.Join(conditions, " AND "),
		),
		params...,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	for rows.Next() {
		var user model.User
		if err = rows.Scan(
			&user.ID,
			&user.Role.ID,
			&user.Role.Name,
			&user.ApiKey,
			&user.MerchantNote,
			&user.MinimumPurchase,
			&user.AdminFee,
			&user.AccumulationDivisor,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.Address,
			&user.Village.ID,
			&user.Village.Name,
			&user.Subdistrict.ID,
			&user.Subdistrict.Name,
			&user.City.ID,
			&user.City.Name,
			&user.Province.ID,
			&user.Province.Name,
			&user.MobilePhone,
			&user.DeviceToken,
			&user.Status,
			&user.ActivatedAt,
			&user.ActivationToken,
			&user.PasswordResetToken,
			&user.Deposit,
			&user.Company,
			&user.RegistrationIP,
			&user.Gender,
			&user.DateOfBirth,
			&user.Avatar,
			&user.Thumbnails,
			&user.OpenOnSunday,
			&user.OpenOnMonday,
			&user.OpenOnTuesday,
			&user.OpenOnWednesday,
			&user.OpenOnThursday,
			&user.OpenOnFriday,
			&user.OpenOnSaturday,
			&user.BusinessOpeningHour,
			&user.BusinessClosingHour,
			&user.DeliveryHours,
			&user.Latitude,
			&user.Longitude,
			&user.Keywords,
			&user.DeliveryMaxDistance,
			&user.DeliveryFreeDistance,
			&user.DeliveryRate,
			&user.CreatedBy,
			&user.CreatedAt,
			&user.UpdatedBy,
			&user.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response = append(response, user)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}
