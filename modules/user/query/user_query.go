package query

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	authModel "github.com/roysitumorang/laukpauk/modules/auth/model"
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
				, u.business_days
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
		var (
			user             model.User
			businessDaysByte []byte
		)
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
			&businessDaysByte,
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
		if businessDaysByte != nil {
			var businessDays model.BusinessDays
			if err = json.Unmarshal(businessDaysByte, &businessDays); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrUnmarshal")
				return
			}
			user.BusinessDays = &businessDays
		}
		response = append(response, user)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}

func (q *userQuery) ChangePassword(ctx context.Context, userID int64, encryptedPassword string) (err error) {
	ctxt := "UserQuery-ChangePassword"
	if _, err = q.dbWrite.Exec(
		ctx,
		`UPDATE users SET
			password = $1
		WHERE id = $2`,
		encryptedPassword,
		userID,
	); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
	}
	return
}

func (q *userQuery) Register(ctx context.Context, request authModel.RegisterRequest) (*authModel.RegisterResponse, error) {
	ctxt := "UserQuery-Register"
	var response authModel.RegisterResponse
	tx, err := q.dbWrite.Begin(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrBegin")
		return nil, err
	}
	for {
		userID, err := helper.GenerateSnowflakeUniqueID()
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGenerateSnowflakeUniqueID")
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
			}
			return nil, err
		}
		activationToken := helper.GenerateRandomString(32)
		now := time.Now().UTC()
		err = tx.QueryRow(
			ctx,
			`INSERT INTO users (
				id
				, role_id
				, name
				, password
				, mobile_phone
				, village_id
				, subdistrict_id
				, activation_token
				, minimum_purchase
				, admin_fee
				, accumulation_divisor
				, status
				, deposit
				, registration_ip
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15)
			RETURNING activation_token`,
			userID,
			request.RoleID,
			request.Name,
			request.Password,
			request.MobilePhone,
			request.VillageID,
			request.SubdistrictID,
			activationToken,
			0,
			0,
			0,
			model.StatusHold,
			0,
			request.IpAddress,
			now,
		).Scan(&response.ActivationToken)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
			}
			var pgxErr *pgconn.PgError
			if errors.As(err, &pgxErr) && pgxErr.Code == pgerrcode.UniqueViolation {
				if pgxErr.ConstraintName == "users_role_id_mobile_phone_idx" {
					return nil, fmt.Errorf("mobile phone %s already registered", request.MobilePhone)
				}
				continue
			}
			return nil, err
		}
		break
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrCommit")
		return nil, err
	}
	return &response, err
}

func (q *userQuery) Activate(ctx context.Context, roleID int64, activationToken string) (response int64, err error) {
	ctxt := "UserQuery-Activate"
	now := time.Now().UTC()
	err = q.dbWrite.QueryRow(
		ctx,
		`UPDATE users SET
			status = $1
			, activation_token = NULL
			, activated_at = $2
		WHERE role_id = $3
		AND status = $4
		AND activation_token = $5
		RETURNING id`,
		model.StatusActive,
		now,
		roleID,
		model.StatusHold,
		activationToken,
	).Scan(&response)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
	}
	return
}
