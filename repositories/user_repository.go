package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	UserRepository interface {
		FindAdmins(
			ctx context.Context,
			request *models.UserFilter,
		) ([]*models.Admin, int64, int64, error)
	}

	userRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewUserRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) UserRepository {
	return new(userRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *userRepository) FindAdmins(
	ctx context.Context,
	request *models.UserFilter,
) ([]*models.Admin, int64, int64, error) {
	ctxt := "UserRepository-FindAdmins"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	params = append(params, models.UserTypeAdmin)
	_, _ = sb.WriteString("u.user_type = $")
	_, _ = sb.WriteString(strconv.Itoa(len(params)))
	conds = append(conds, sb.String())
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("u.id IN (")
		for i, userID := range request.UserIDs {
			params = append(params, userID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Login != "" {
		params = append(params, strings.ToLower(request.Login))
		n := strconv.Itoa(len(params))
		sb.Reset()
		_, _ = sb.WriteString("(u.username = $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(" OR u.email = $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(" OR u.mobile_phone = $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = sb.WriteString("%%")
		_, _ = sb.WriteString(strings.ToLower(request.Keyword))
		_, _ = sb.WriteString("%%")
		params = append(params, sb.String())
		n := strconv.Itoa(len(params))
		sb.Reset()
		_, _ = sb.WriteString("(LOWER(u.name) LIKE $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(" OR u.username LIKE $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(" OR u.email LIKE $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(" OR u.mobile_phone LIKE $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Username != "" {
		params = append(params, strings.ToLower(request.Username))
		sb.Reset()
		_, _ = sb.WriteString("u.username = $")
		_, _ = sb.WriteString(strconv.Itoa(len(params)))
		conds = append(conds, sb.String())
	}
	if request.Email != "" {
		params = append(params, strings.ToLower(request.Email))
		sb.Reset()
		_, _ = sb.WriteString("u.email = $")
		_, _ = sb.WriteString(strconv.Itoa(len(params)))
		conds = append(conds, sb.String())
	}
	if request.MobilePhone != "" {
		params = append(params, strings.ToLower(request.MobilePhone))
		sb.Reset()
		_, _ = sb.WriteString("u.mobile_phone = $")
		_, _ = sb.WriteString(strconv.Itoa(len(params)))
		conds = append(conds, sb.String())
	}
	if request.PasswordResetToken != "" {
		params = append(params, request.PasswordResetToken)
		sb.Reset()
		_, _ = sb.WriteString("u.password_reset_token = $")
		_, _ = sb.WriteString(strconv.Itoa(len(params)))
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
        FROM users u
        JOIN admins a ON u.id = a.user_id`,
	)
	if len(conds) > 0 {
		_, _ = sb.WriteString(" WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = sb.WriteString(" ")
			_, _ = sb.WriteString(cond)
		}
	}
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(&total)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return nil, 0, 0, err
	}
	if total == 0 {
		return nil, 0, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`ROW_NUMBER() OVER (ORDER BY u.id DESC) AS row_no
		, u.id
		, u.user_type
		, u.name
		, u.confirmation_token
		, u.confirmed_at
		, u.unconfirmed_mobile_phone
		, u.mobile_phone_confirmation_sent_at
		, u.mobile_phone_confirmation_token
		, u.mobile_phone_confirmed_at
		, u.mobile_phone
		, u.otp_token
		, u.otp_created_at
		, u.otp_expired_at
		, u.username
		, u.unconfirmed_email
		, u.email_confirmation_sent_at
		, u.email_confirmation_token
		, u.email_confirmed_at
		, u.email
		, u.encrypted_password
		, u.password_reset_at
		, u.password_reset_token
		, u.password_changed_at
		, u.created_by
		, u.created_at
		, u.updated_by
		, u.updated_at
		, u.deactivated_by
		, u.deactivated_at
		, u.deactivation_reason
		, a.admin_level
		, a.photo`,
	)
	sb.Reset()
	_, _ = sb.WriteString(query)
	pages := int64(1)
	if request.Limit > 0 {
		totalDecimal, err := decimal.New(total, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		perPageDecimal, err := decimal.New(request.Limit, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		pagesDecimal, err := totalDecimal.Quo(perPageDecimal)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuo")
			return nil, 0, 0, err
		}
		pages, _, _ = pagesDecimal.Ceil(0).Int64(0)
		offset := request.Page * request.Limit
		_, _ = sb.WriteString(" LIMIT ")
		_, _ = sb.WriteString(strconv.FormatInt(request.Limit, 10))
		_, _ = sb.WriteString(" OFFSET ")
		_, _ = sb.WriteString(strconv.FormatInt(offset, 10))
	}
	rows, err := q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.Admin
	for rows.Next() {
		var admin models.Admin
		if err = rows.Scan(
			&admin.RowNo,
			&admin.ID,
			&admin.UserType,
			&admin.Name,
			&admin.ConfirmationToken,
			&admin.ConfirmedAt,
			&admin.UnconfirmedMobilePhone,
			&admin.MobilePhoneConfirmationSentAt,
			&admin.MobilePhoneConfirmationToken,
			&admin.MobilePhoneConfirmedAt,
			&admin.MobilePhone,
			&admin.OtpToken,
			&admin.OtpCreatedAt,
			&admin.OtpExpiredAt,
			&admin.Username,
			&admin.UnconfirmedEmail,
			&admin.EmailConfirmationSentAt,
			&admin.EmailConfirmationToken,
			&admin.EmailConfirmedAt,
			&admin.Email,
			&admin.EncryptedPassword,
			&admin.PasswordResetAt,
			&admin.PasswordResetToken,
			&admin.PasswordChangedAt,
			&admin.CreatedBy,
			&admin.CreatedAt,
			&admin.UpdatedBy,
			&admin.UpdatedAt,
			&admin.DeactivatedBy,
			&admin.DeactivatedAt,
			&admin.DeactivationReason,
			&admin.AdminLevel,
			&admin.Photo,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &admin)
	}
	return response, total, pages, nil
}
