package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type AdminRepository interface {
	FindAdmins(
		ctx context.Context,
		request *models.AdminFilter,
	) ([]*models.Admin, int64, int64, error)
	FindAdmin(
		ctx context.Context,
		request *models.AdminFilter,
	) (*models.Admin, error)
	CreateAdmin(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Admin,
	) (*models.Admin, error)
	UpdateAdmin(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Admin,
	) (*models.Admin, error)
}

type adminRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewAdminRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) AdminRepository {
	return &adminRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *adminRepository) FindAdmins(
	ctx context.Context,
	request *models.AdminFilter,
) ([]*models.Admin, int64, int64, error) {
	ctxt := "AdminRepository-FindAdmins"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM users u JOIN admins a ON u.id = a.user_id %s", condition)
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
		`a.id
		, a.user_id
		, u.user_type
		, u.user_status
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
		, u.photo
		, u.unread_notifications
		, u.created_by
		, u.created_at
		, u.updated_by
		, u.updated_at
		, u.deactivated_by
		, u.deactivated_at
		, u.deactivation_reason
		, a.admin_type`,
	)
	sb.Reset()
	_, _ = sb.WriteString(query)
	sortBy := request.Sortings
	if len(sortBy) == 0 {
		sortBy = append(sortBy, models.DefaultSortBy)
	}
	_, _ = sb.WriteString(" ORDER BY")
	for i, item := range sortBy {
		if i > 0 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " %s %s", item.Field, item.Direction)
	}
	pages := int64(1)
	var offset int64
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
		offset = request.Page * request.Limit
		_, _ = fmt.Fprintf(sb, " LIMIT %d OFFSET %d", request.Limit, offset)
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
			&admin.ID,
			&admin.UserID,
			&admin.UserType,
			&admin.UserStatus,
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
			&admin.Photo,
			&admin.UnreadNotifications,
			&admin.CreatedBy,
			&admin.CreatedAt,
			&admin.UpdatedBy,
			&admin.UpdatedAt,
			&admin.DeactivatedBy,
			&admin.DeactivatedAt,
			&admin.DeactivationReason,
			&admin.AdminType,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		admin.RowNo = offset
		response = append(response, &admin)
	}
	return response, total, pages, nil
}

func (q *adminRepository) FindAdmin(
	ctx context.Context,
	request *models.AdminFilter,
) (*models.Admin, error) {
	ctxt := "AdminRepository-FindAdmin"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`SELECT
			a.id
			, a.user_id
			, u.user_type
			, u.user_status
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
			, u.photo
			, u.unread_notifications
			, u.created_by
			, u.created_at
			, u.updated_by
			, u.updated_at
			, u.deactivated_by
			, u.deactivated_at
			, u.deactivation_reason
			, a.admin_type
		FROM users u
		JOIN admins a ON u.id = a.user_id %s`,
		condition,
	)
	var response models.Admin
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.UserID,
		&response.UserType,
		&response.UserStatus,
		&response.Name,
		&response.ConfirmationToken,
		&response.ConfirmedAt,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
		&response.OtpToken,
		&response.OtpCreatedAt,
		&response.OtpExpiredAt,
		&response.Username,
		&response.UnconfirmedEmail,
		&response.EmailConfirmationSentAt,
		&response.EmailConfirmationToken,
		&response.EmailConfirmedAt,
		&response.Email,
		&response.EncryptedPassword,
		&response.PasswordResetAt,
		&response.PasswordResetToken,
		&response.PasswordChangedAt,
		&response.Photo,
		&response.UnreadNotifications,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
		&response.AdminType,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAdminNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *adminRepository) buildCondition(request *models.AdminFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("a.id IN (")
		for i, userID := range request.UserIDs {
			params = append(params, userID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedUserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("a.id NOT IN (")
		for i, excludedUserID := range request.ExcludedUserIDs {
			params = append(params, excludedUserID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.UserStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("u.user_status IN (")
		for i, userStatus := range request.UserStatuses {
			params = append(params, userStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Login != "" {
		params = append(params, strings.ToLower(request.Login))
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "(u.username = $%d OR u.email = $%d OR u.mobile_phone = $%d)", n, n, n)
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(
			sb,
			`(LOWER(u.name) LIKE $%d
				OR u.username LIKE $%d
				OR u.email LIKE $%d
				OR u.mobile_phone LIKE $%d)`,
			n, n, n, n,
		)
		conds = append(conds, sb.String())
	}
	if request.Username != "" {
		params = append(params, strings.ToLower(request.Username))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "u.username = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.Email != "" {
		params = append(params, strings.ToLower(request.Email))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "u.email = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.MobilePhone != "" {
		params = append(params, strings.ToLower(request.MobilePhone))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "u.mobile_phone = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.PasswordResetToken != "" {
		params = append(params, request.PasswordResetToken)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "u.password_reset_token = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if len(conds) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = fmt.Fprintf(sb, " %s", cond)
		}
		condition = sb.String()
	}
	return
}

func (q *adminRepository) CreateAdmin(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Admin,
) (*models.Admin, error) {
	ctxt := "AdminRepository-CreateAdmin"
	var response models.Admin
	err := tx.QueryRow(
		ctx,
		`INSERT INTO users (
			user_type
			, user_status
			, name
			, confirmation_token
			, confirmed_at
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, otp_token
			, otp_created_at
			, otp_expired_at
			, username
			, unconfirmed_email
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_at
			, password_reset_token
			, password_changed_at
			, photo
			, unread_notifications
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			, $21, $22, $23, $24, $25, $26, $27, $28, $29
		)
		RETURNING id
			, user_type
			, user_status
			, name
			, confirmation_token
			, confirmed_at
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, otp_token
			, otp_created_at
			, otp_expired_at
			, username
			, unconfirmed_email
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_at
			, password_reset_token
			, password_changed_at
			, photo
			, unread_notifications
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		request.UserType,
		request.UserStatus,
		request.Name,
		request.ConfirmationToken,
		request.ConfirmedAt,
		request.UnconfirmedMobilePhone,
		request.MobilePhoneConfirmationSentAt,
		request.MobilePhoneConfirmationToken,
		request.MobilePhoneConfirmedAt,
		request.MobilePhone,
		request.OtpToken,
		request.OtpCreatedAt,
		request.OtpExpiredAt,
		request.Username,
		request.UnconfirmedEmail,
		request.EmailConfirmationSentAt,
		request.EmailConfirmationToken,
		request.EmailConfirmedAt,
		request.Email,
		request.EncryptedPassword,
		request.PasswordResetAt,
		request.PasswordResetToken,
		request.PasswordChangedAt,
		request.Photo,
		request.UnreadNotifications,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.UserID,
		&response.UserType,
		&response.UserStatus,
		&response.Name,
		&response.ConfirmationToken,
		&response.ConfirmedAt,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
		&response.OtpToken,
		&response.OtpCreatedAt,
		&response.OtpExpiredAt,
		&response.Username,
		&response.UnconfirmedEmail,
		&response.EmailConfirmationSentAt,
		&response.EmailConfirmationToken,
		&response.EmailConfirmedAt,
		&response.Email,
		&response.EncryptedPassword,
		&response.PasswordResetAt,
		&response.PasswordResetToken,
		&response.PasswordChangedAt,
		&response.Photo,
		&response.UnreadNotifications,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "users_username_key":
				err = models.ErrUniqueAdminUsernameViolation
			case "users_mobile_phone_key":
				err = models.ErrUniqueAdminMobilePhoneViolation
			case "users_email_key":
				err = models.ErrUniqueAdminEmailViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	if err = tx.QueryRow(
		ctx,
		`INSERT INTO admins (
			user_id
			, admin_type
		) VALUES ($1, $2)
		RETURNING id
			, user_id
			, admin_type`,
		response.UserID,
		request.AdminType,
	).Scan(
		&response.ID,
		&response.UserID,
		&response.AdminType,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *adminRepository) UpdateAdmin(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Admin,
) (*models.Admin, error) {
	ctxt := "AdminRepository-UpdateAdmin"
	var response models.Admin
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, user_id
			, admin_type
		FROM admins
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.UserID,
		&response.AdminType,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrBuyerNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	var params []any
	var n int
	_, _ = sb.WriteString("UPDATE admins SET")
	if response.AdminType != request.AdminType {
		params = append(params, request.AdminType)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " admin_type = $%d", n)
	}
	if len(params) > 0 {
		params = append(params, request.ID)
		_, _ = fmt.Fprintf(
			sb,
			` WHERE id = $%d
			RETURNING id
				, user_id
				, admin_type`,
			len(params),
		)
		if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
			&response.ID,
			&response.UserID,
			&response.AdminType,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
	}
	err = tx.QueryRow(
		ctx,
		`SELECT
			id
			, user_type
			, user_status
			, name
			, confirmation_token
			, confirmed_at
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, otp_token
			, otp_created_at
			, otp_expired_at
			, username
			, unconfirmed_email
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_at
			, password_reset_token
			, password_changed_at
			, photo
			, unread_notifications
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM users
		WHERE id = $1
		FOR UPDATE`,
		response.UserID,
	).Scan(
		&response.ID,
		&response.UserType,
		&response.UserStatus,
		&response.Name,
		&response.ConfirmationToken,
		&response.ConfirmedAt,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
		&response.OtpToken,
		&response.OtpCreatedAt,
		&response.OtpExpiredAt,
		&response.Username,
		&response.UnconfirmedEmail,
		&response.EmailConfirmationSentAt,
		&response.EmailConfirmationToken,
		&response.EmailConfirmedAt,
		&response.Email,
		&response.EncryptedPassword,
		&response.PasswordResetAt,
		&response.PasswordResetToken,
		&response.PasswordChangedAt,
		&response.Photo,
		&response.UnreadNotifications,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrBuyerNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	params = params[:0]
	sb.Reset()
	_, _ = sb.WriteString("UPDATE users SET")
	if response.UserStatus != request.UserStatus {
		params = append(params, request.UserStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " user_status = $%d", n)
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " name = $%d", n)
	}
	if (response.ConfirmationToken != nil && request.ConfirmationToken == nil) ||
		(response.ConfirmationToken == nil && request.ConfirmationToken != nil) ||
		(response.ConfirmationToken != nil && request.ConfirmationToken != nil &&
			*response.ConfirmationToken != *request.ConfirmationToken) {
		params = append(params, request.ConfirmationToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " confirmation_token = $%d", n)
	}
	if (response.ConfirmedAt != nil && request.ConfirmedAt == nil) ||
		(response.ConfirmedAt == nil && request.ConfirmedAt != nil) ||
		(response.ConfirmedAt != nil && request.ConfirmedAt != nil &&
			!response.ConfirmedAt.Equal(*request.ConfirmedAt)) {
		params = append(params, request.ConfirmedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " confirmed_at = $%d", n)
	}
	if (response.UnconfirmedMobilePhone != nil && request.UnconfirmedMobilePhone == nil) ||
		(response.UnconfirmedMobilePhone == nil && request.UnconfirmedMobilePhone != nil) ||
		(response.UnconfirmedMobilePhone != nil && request.UnconfirmedMobilePhone != nil &&
			*response.UnconfirmedMobilePhone != *request.UnconfirmedMobilePhone) {
		params = append(params, request.UnconfirmedMobilePhone)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " unconfirmed_mobile_phone = $%d", n)
	}
	if (response.MobilePhoneConfirmationSentAt != nil && request.MobilePhoneConfirmationSentAt == nil) ||
		(response.MobilePhoneConfirmationSentAt == nil && request.MobilePhoneConfirmationSentAt != nil) ||
		(response.MobilePhoneConfirmationSentAt != nil && request.MobilePhoneConfirmationSentAt != nil &&
			!response.MobilePhoneConfirmationSentAt.Equal(*request.MobilePhoneConfirmationSentAt)) {
		params = append(params, request.MobilePhoneConfirmationSentAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " mobile_phone_confirmation_sent_at = $%d", n)
	}
	if (response.MobilePhoneConfirmationToken != nil && request.MobilePhoneConfirmationToken == nil) ||
		(response.MobilePhoneConfirmationToken == nil && request.MobilePhoneConfirmationToken != nil) ||
		(response.MobilePhoneConfirmationToken != nil && request.MobilePhoneConfirmationToken != nil &&
			*response.MobilePhoneConfirmationToken != *request.MobilePhoneConfirmationToken) {
		params = append(params, request.MobilePhoneConfirmationToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " mobile_phone_confirmation_token = $%d", n)
	}
	if (response.MobilePhoneConfirmedAt != nil && request.MobilePhoneConfirmedAt == nil) ||
		(response.MobilePhoneConfirmedAt == nil && request.MobilePhoneConfirmedAt != nil) ||
		(response.MobilePhoneConfirmedAt != nil && request.MobilePhoneConfirmedAt != nil &&
			!response.MobilePhoneConfirmedAt.Equal(*request.MobilePhoneConfirmedAt)) {
		params = append(params, request.MobilePhoneConfirmedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " mobile_phone_confirmed_at = $%d", n)
	}
	if (response.MobilePhone != nil && request.MobilePhone == nil) ||
		(response.MobilePhone == nil && request.MobilePhone != nil) ||
		(response.MobilePhone != nil && request.MobilePhone != nil &&
			*response.MobilePhone != *request.MobilePhone) {
		params = append(params, request.MobilePhone)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " mobile_phone = $%d", n)
	}
	if (response.OtpToken != nil && request.OtpToken == nil) ||
		(response.OtpToken == nil && request.OtpToken != nil) ||
		(response.OtpToken != nil && request.OtpToken != nil &&
			*response.OtpToken != *request.OtpToken) {
		params = append(params, request.OtpToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " otp_token = $%d", n)
	}
	if (response.OtpCreatedAt != nil && request.OtpCreatedAt == nil) ||
		(response.OtpCreatedAt == nil && request.OtpCreatedAt != nil) ||
		(response.OtpCreatedAt != nil && request.OtpCreatedAt != nil &&
			!response.OtpCreatedAt.Equal(*request.OtpCreatedAt)) {
		params = append(params, request.OtpCreatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " otp_created_at = $%d", n)
	}
	if (response.OtpExpiredAt != nil && request.OtpExpiredAt == nil) ||
		(response.OtpExpiredAt == nil && request.OtpExpiredAt != nil) ||
		(response.OtpExpiredAt != nil && request.OtpExpiredAt != nil &&
			!response.OtpExpiredAt.Equal(*request.OtpExpiredAt)) {
		params = append(params, request.OtpExpiredAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " otp_expired_at = $%d", n)
	}
	if response.Username != request.Username {
		params = append(params, request.Username)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " username = $%d", n)
	}
	if (response.UnconfirmedEmail != nil && request.UnconfirmedEmail == nil) ||
		(response.UnconfirmedEmail == nil && request.UnconfirmedEmail != nil) ||
		(response.UnconfirmedEmail != nil && request.UnconfirmedEmail != nil &&
			*response.UnconfirmedEmail != *request.UnconfirmedEmail) {
		params = append(params, request.UnconfirmedEmail)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " unconfirmed_email = $%d", n)
	}
	if (response.EmailConfirmationSentAt != nil && request.EmailConfirmationSentAt == nil) ||
		(response.EmailConfirmationSentAt == nil && request.EmailConfirmationSentAt != nil) ||
		(response.EmailConfirmationSentAt != nil && request.EmailConfirmationSentAt != nil &&
			!response.EmailConfirmationSentAt.Equal(*request.EmailConfirmationSentAt)) {
		params = append(params, request.EmailConfirmationSentAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " email_confirmation_sent_at = $%d", n)
	}
	if (response.EmailConfirmationToken != nil && request.EmailConfirmationToken == nil) ||
		(response.EmailConfirmationToken == nil && request.EmailConfirmationToken != nil) ||
		(response.EmailConfirmationToken != nil && request.EmailConfirmationToken != nil &&
			*response.EmailConfirmationToken != *request.EmailConfirmationToken) {
		params = append(params, request.EmailConfirmationToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " email_confirmation_token = $%d", n)
	}
	if (response.EmailConfirmedAt != nil && request.EmailConfirmedAt == nil) ||
		(response.EmailConfirmedAt == nil && request.EmailConfirmedAt != nil) ||
		(response.EmailConfirmedAt != nil && request.EmailConfirmedAt != nil &&
			!response.EmailConfirmedAt.Equal(*request.EmailConfirmedAt)) {
		params = append(params, request.EmailConfirmedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " email_confirmed_at = $%d", n)
	}
	if (response.Email != nil && request.Email == nil) ||
		(response.Email == nil && request.Email != nil) ||
		(response.Email != nil && request.Email != nil &&
			*response.Email != *request.Email) {
		params = append(params, request.Email)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " email = $%d", n)
	}
	if (response.EncryptedPassword != nil && request.EncryptedPassword == nil) ||
		(response.EncryptedPassword == nil && request.EncryptedPassword != nil) ||
		(response.EncryptedPassword != nil && request.EncryptedPassword != nil &&
			*response.EncryptedPassword != *request.EncryptedPassword) {
		params = append(params, request.EncryptedPassword)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " encrypted_password = $%d", n)
	}
	if (response.PasswordResetAt != nil && request.PasswordResetAt == nil) ||
		(response.PasswordResetAt == nil && request.PasswordResetAt != nil) ||
		(response.PasswordResetAt != nil && request.PasswordResetAt != nil &&
			!response.PasswordResetAt.Equal(*request.PasswordResetAt)) {
		params = append(params, request.PasswordResetAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " password_reset_at = $%d", n)
	}
	if (response.PasswordResetToken != nil && request.PasswordResetToken == nil) ||
		(response.PasswordResetToken == nil && request.PasswordResetToken != nil) ||
		(response.PasswordResetToken != nil && request.PasswordResetToken != nil &&
			*response.PasswordResetToken != *request.PasswordResetToken) {
		params = append(params, request.PasswordResetToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " password_reset_token = $%d", n)
	}
	if (response.PasswordChangedAt != nil && request.PasswordChangedAt == nil) ||
		(response.PasswordChangedAt == nil && request.PasswordChangedAt != nil) ||
		(response.PasswordChangedAt != nil && request.PasswordChangedAt != nil &&
			!response.PasswordChangedAt.Equal(*request.PasswordChangedAt)) {
		params = append(params, request.PasswordChangedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " password_changed_at = $%d", n)
	}
	if response.Photo != request.Photo {
		params = append(params, request.Photo)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " photo = $%d", n)
	}
	if request.UnreadNotifications != 0 {
		params = append(params, request.UnreadNotifications)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " unread_notifications = unread_notifications + $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	if (response.UpdatedBy != nil && request.UpdatedBy == nil) ||
		(response.UpdatedBy == nil && request.UpdatedBy != nil) ||
		(response.UpdatedBy != nil && request.UpdatedBy != nil &&
			*response.UpdatedBy != *request.UpdatedBy) {
		params = append(params, request.UpdatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_by = $%d", n)
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_at = $%d", n)
	}
	if (response.DeactivatedBy != nil && request.DeactivatedBy == nil) ||
		(response.DeactivatedBy == nil && request.DeactivatedBy != nil) ||
		(response.DeactivatedBy != nil && request.DeactivatedBy != nil &&
			*response.DeactivatedBy != *request.DeactivatedBy) {
		params = append(params, request.DeactivatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deactivated_by = $%d", n)
	}
	if (response.DeactivatedAt != nil && request.DeactivatedAt == nil) ||
		(response.DeactivatedAt == nil && request.DeactivatedAt != nil) ||
		(response.DeactivatedAt != nil && request.DeactivatedAt != nil &&
			!response.DeactivatedAt.Equal(*request.DeactivatedAt)) {
		params = append(params, request.DeactivatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deactivated_at = $%d", n)
	}
	if (response.DeactivationReason != nil && request.DeactivationReason == nil) ||
		(response.DeactivationReason == nil && request.DeactivationReason != nil) ||
		(response.DeactivationReason != nil && request.DeactivationReason != nil &&
			*response.DeactivationReason != *request.DeactivationReason) {
		params = append(params, request.DeactivationReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deactivation_reason = $%d", n)
	}
	params = append(params, response.UserID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, user_type
			, user_status
			, name
			, confirmation_token
			, confirmed_at
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, otp_token
			, otp_created_at
			, otp_expired_at
			, username
			, unconfirmed_email
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_at
			, password_reset_token
			, password_changed_at
			, photo
			, unread_notifications
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		len(params),
	)
	err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.UserType,
		&response.UserStatus,
		&response.Name,
		&response.ConfirmationToken,
		&response.ConfirmedAt,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
		&response.OtpToken,
		&response.OtpCreatedAt,
		&response.OtpExpiredAt,
		&response.Username,
		&response.UnconfirmedEmail,
		&response.EmailConfirmationSentAt,
		&response.EmailConfirmationToken,
		&response.EmailConfirmedAt,
		&response.Email,
		&response.EncryptedPassword,
		&response.PasswordResetAt,
		&response.PasswordResetToken,
		&response.PasswordChangedAt,
		&response.Photo,
		&response.UnreadNotifications,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "users_username_key":
				err = models.ErrUniqueAdminUsernameViolation
			case "users_mobile_phone_key":
				err = models.ErrUniqueAdminMobilePhoneViolation
			case "users_email_key":
				err = models.ErrUniqueAdminEmailViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}
