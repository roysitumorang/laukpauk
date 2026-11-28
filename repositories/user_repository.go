package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type UserRepository interface {
	FindUsers(
		ctx context.Context,
		request *models.UserFilter,
	) ([]*models.User, int64, int64, error)
	FindUser(
		ctx context.Context,
		request *models.UserFilter,
	) (*models.User, error)
	CreateUser(
		ctx context.Context,
		tx *sql.Tx,
		request *models.User,
	) (*models.User, error)
	UpdateUser(
		ctx context.Context,
		tx *sql.Tx,
		request *models.User,
	) (*models.User, error)
	AddUserBalances(
		ctx context.Context,
		tx *sql.Tx,
		userID uuid.UUID,
		referralPoints,
		rewardPoints,
		wallet int64,
		now time.Time,
	) (*models.User, error)
}

type userRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewUserRepository(
	dbRead,
	dbWrite *sql.DB,
) UserRepository {
	return &userRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *userRepository) FindUsers(
	ctx context.Context,
	request *models.UserFilter,
) ([]*models.User, int64, int64, error) {
	ctxt := "UserRepository-FindUsers"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM users %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRowContext(ctx, query, helper.SQLArgs(params...)...).Scan(helper.SQLScan(&total)...)
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
		`id
		, role
		, status
		, name
		, mobile_phone_confirmation_sent_at
		, mobile_phone_confirmation_token
		, mobile_phone_confirmed_at
		, mobile_phone
		, username
		, email_confirmation_sent_at
		, email_confirmation_token
		, email_confirmed_at
		, email
		, encrypted_password
		, password_reset_sent_at
		, password_reset_token
		, password_changed_at
		, sign_in_count
		, current_sign_in_at
		, last_sign_in_at
		, current_sign_in_ip
		, last_sign_in_ip
		, failed_attempts
		, unlock_token
		, locked_at
		, otp_secret
		, otp_expired_at
		, photo
		, user_level
		, referral_code
		, referrer_id
		, referral_points
		, reward_points
		, wallet
		, created_by
		, created_at
		, updated_by
		, updated_at
		, deactivated_by
		, deactivated_at
		, deactivation_reason`,
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
		_, _ = fmt.Fprintf(sb, " %s %s", sortExpression(item.Field), item.Direction)
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
	rows, err := q.dbRead.QueryContext(ctx, sb.String(), helper.SQLArgs(params...)...)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer func() { _ = rows.Close() }()
	var response []*models.User
	for rows.Next() {
		var user models.User
		if err = rows.Scan(
			helper.SQLScan(
				&user.ID,
				&user.Role,
				&user.Status,
				&user.Name,
				&user.MobilePhoneConfirmationSentAt,
				&user.MobilePhoneConfirmationToken,
				&user.MobilePhoneConfirmedAt,
				&user.MobilePhone,
				&user.Username,
				&user.EmailConfirmationSentAt,
				&user.EmailConfirmationToken,
				&user.EmailConfirmedAt,
				&user.Email,
				&user.EncryptedPassword,
				&user.PasswordResetSentAt,
				&user.PasswordResetToken,
				&user.PasswordChangedAt,
				&user.SignInCount,
				&user.CurrentSignInAt,
				&user.LastSignInAt,
				&user.CurrentSignInIP,
				&user.LastSignInIP,
				&user.FailedAttempts,
				&user.UnlockToken,
				&user.LockedAt,
				&user.OtpSecret,
				&user.OtpExpiredAt,
				&user.Photo,
				&user.UserLevel,
				&user.ReferralCode,
				&user.ReferrerID,
				&user.ReferralPoints,
				&user.RewardPoints,
				&user.Wallet,
				&user.CreatedBy,
				&user.CreatedAt,
				&user.UpdatedBy,
				&user.UpdatedAt,
				&user.DeactivatedBy,
				&user.DeactivatedAt,
				&user.DeactivationReason,
			)...,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		user.RowNo = offset
		response = append(response, &user)
	}
	return response, total, pages, nil
}

func (q *userRepository) FindUser(
	ctx context.Context,
	request *models.UserFilter,
) (*models.User, error) {
	ctxt := "UserRepository-FindUser"
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
			id
			, role
			, status
			, name
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, username
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_sent_at
			, password_reset_token
			, password_changed_at
			, sign_in_count
			, current_sign_in_at
			, last_sign_in_at
			, current_sign_in_ip
			, last_sign_in_ip
			, failed_attempts
			, unlock_token
			, locked_at
			, otp_secret
			, otp_expired_at
			, photo
			, user_level
			, referral_code
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM users %s`,
		condition,
	)
	var response models.User
	err := q.dbRead.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.Role,
			&response.Status,
			&response.Name,
			&response.MobilePhoneConfirmationSentAt,
			&response.MobilePhoneConfirmationToken,
			&response.MobilePhoneConfirmedAt,
			&response.MobilePhone,
			&response.Username,
			&response.EmailConfirmationSentAt,
			&response.EmailConfirmationToken,
			&response.EmailConfirmedAt,
			&response.Email,
			&response.EncryptedPassword,
			&response.PasswordResetSentAt,
			&response.PasswordResetToken,
			&response.PasswordChangedAt,
			&response.SignInCount,
			&response.CurrentSignInAt,
			&response.LastSignInAt,
			&response.CurrentSignInIP,
			&response.LastSignInIP,
			&response.FailedAttempts,
			&response.UnlockToken,
			&response.LockedAt,
			&response.OtpSecret,
			&response.OtpExpiredAt,
			&response.Photo,
			&response.UserLevel,
			&response.ReferralCode,
			&response.ReferrerID,
			&response.ReferralPoints,
			&response.RewardPoints,
			&response.Wallet,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeactivatedBy,
			&response.DeactivatedAt,
			&response.DeactivationReason,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *userRepository) buildCondition(request *models.UserFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, userID := range request.UserIDs {
			params = append(params, userID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedUserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedUserID := range request.ExcludedUserIDs {
			params = append(params, excludedUserID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.Roles) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("role IN (")
		for i, role := range request.Roles {
			params = append(params, role)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.Statuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("status IN (")
		for i, status := range request.Statuses {
			params = append(params, status)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ReferralCodes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("referral_code IN (")
		for i, referralCodes := range request.ReferralCodes {
			params = append(params, referralCodes)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Login != "" {
		login := strings.ToLower(request.Login)
		params = append(params, login, login, login)
		sb.Reset()
		_, _ = sb.WriteString("(username = ? OR email = ? OR mobile_phone = ?)")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		keyword := sb.String()
		params = append(params, keyword, keyword, keyword, keyword)
		sb.Reset()
		_, _ = sb.WriteString(`(LOWER(name) LIKE ?
				OR username LIKE ?
				OR email LIKE ?
				OR mobile_phone LIKE ?)`)
		conds = append(conds, sb.String())
	}
	if request.Username != "" {
		params = append(params, strings.ToLower(request.Username))
		sb.Reset()
		_, _ = sb.WriteString("username = ?")
		conds = append(conds, sb.String())
	}
	if request.Email != "" {
		params = append(params, strings.ToLower(request.Email))
		sb.Reset()
		_, _ = sb.WriteString("email = ?")
		conds = append(conds, sb.String())
	}
	if request.MobilePhone != "" {
		params = append(params, strings.ToLower(request.MobilePhone))
		sb.Reset()
		_, _ = sb.WriteString("mobile_phone = ?")
		conds = append(conds, sb.String())
	}
	if request.PasswordResetToken != "" {
		params = append(params, request.PasswordResetToken)
		sb.Reset()
		_, _ = sb.WriteString("password_reset_token = ?")
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

func (q *userRepository) CreateUser(
	ctx context.Context,
	tx *sql.Tx,
	request *models.User,
) (*models.User, error) {
	ctxt := "UserRepository-CreateUser"
	var response models.User
	err := tx.QueryRowContext(
		ctx,
		`INSERT INTO users (
			role
			, status
			, name
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, username
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_sent_at
			, password_reset_token
			, password_changed_at
			, sign_in_count
			, current_sign_in_at
			, last_sign_in_at
			, current_sign_in_ip
			, last_sign_in_ip
			, failed_attempts
			, unlock_token
			, locked_at
			, otp_secret
			, otp_expired_at
			, photo
			, user_level
			, referral_code
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
		RETURNING id
			, role
			, status
			, name
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, username
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_sent_at
			, password_reset_token
			, password_changed_at
			, sign_in_count
			, current_sign_in_at
			, last_sign_in_at
			, current_sign_in_ip
			, last_sign_in_ip
			, failed_attempts
			, unlock_token
			, locked_at
			, otp_secret
			, otp_expired_at
			, photo
			, user_level
			, referral_code
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		helper.SQLArgs(
			request.Role,
			request.Status,
			request.Name,
			request.MobilePhoneConfirmationSentAt,
			request.MobilePhoneConfirmationToken,
			request.MobilePhoneConfirmedAt,
			request.MobilePhone,
			request.Username,
			request.EmailConfirmationSentAt,
			request.EmailConfirmationToken,
			request.EmailConfirmedAt,
			request.Email,
			request.EncryptedPassword,
			request.PasswordResetSentAt,
			request.PasswordResetToken,
			request.PasswordChangedAt,
			request.SignInCount,
			request.CurrentSignInAt,
			request.LastSignInAt,
			request.CurrentSignInIP,
			request.LastSignInIP,
			request.FailedAttempts,
			request.UnlockToken,
			request.LockedAt,
			request.OtpSecret,
			request.OtpExpiredAt,
			request.Photo,
			request.UserLevel,
			request.ReferralCode,
			request.ReferrerID,
			request.ReferralPoints,
			request.RewardPoints,
			request.Wallet,
			request.CreatedBy,
			request.CreatedAt,
			request.UpdatedBy,
			request.UpdatedAt,
			request.DeactivatedBy,
			request.DeactivatedAt,
			request.DeactivationReason,
		)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.Role,
			&response.Status,
			&response.Name,
			&response.MobilePhoneConfirmationSentAt,
			&response.MobilePhoneConfirmationToken,
			&response.MobilePhoneConfirmedAt,
			&response.MobilePhone,
			&response.Username,
			&response.EmailConfirmationSentAt,
			&response.EmailConfirmationToken,
			&response.EmailConfirmedAt,
			&response.Email,
			&response.EncryptedPassword,
			&response.PasswordResetSentAt,
			&response.PasswordResetToken,
			&response.PasswordChangedAt,
			&response.SignInCount,
			&response.CurrentSignInAt,
			&response.LastSignInAt,
			&response.CurrentSignInIP,
			&response.LastSignInIP,
			&response.FailedAttempts,
			&response.UnlockToken,
			&response.LockedAt,
			&response.OtpSecret,
			&response.OtpExpiredAt,
			&response.Photo,
			&response.UserLevel,
			&response.ReferralCode,
			&response.ReferrerID,
			&response.ReferralPoints,
			&response.RewardPoints,
			&response.Wallet,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeactivatedBy,
			&response.DeactivatedAt,
			&response.DeactivationReason,
		)...,
	)
	if err != nil {
		if key, ok := uniqueViolationKey(err); ok {
			switch key {
			case "users_username_key":
				err = models.ErrUniqueUserUsernameViolation
			case "users_mobile_phone_key":
				err = models.ErrUniqueUserMobilePhoneViolation
			case "users_email_key":
				err = models.ErrUniqueUserEmailViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *userRepository) UpdateUser(
	ctx context.Context,
	tx *sql.Tx,
	request *models.User,
) (*models.User, error) {
	ctxt := "UserRepository-UpdateUser"
	var response models.User
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, role
			, status
			, name
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, username
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_sent_at
			, password_reset_token
			, password_changed_at
			, sign_in_count
			, current_sign_in_at
			, last_sign_in_at
			, current_sign_in_ip
			, last_sign_in_ip
			, failed_attempts
			, unlock_token
			, locked_at
			, otp_secret
			, otp_expired_at
			, photo
			, user_level
			, referral_code
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM users
		WHERE id = ?
		FOR UPDATE`,
		helper.SQLArgs(request.ID)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.Role,
			&response.Status,
			&response.Name,
			&response.MobilePhoneConfirmationSentAt,
			&response.MobilePhoneConfirmationToken,
			&response.MobilePhoneConfirmedAt,
			&response.MobilePhone,
			&response.Username,
			&response.EmailConfirmationSentAt,
			&response.EmailConfirmationToken,
			&response.EmailConfirmedAt,
			&response.Email,
			&response.EncryptedPassword,
			&response.PasswordResetSentAt,
			&response.PasswordResetToken,
			&response.PasswordChangedAt,
			&response.SignInCount,
			&response.CurrentSignInAt,
			&response.LastSignInAt,
			&response.CurrentSignInIP,
			&response.LastSignInIP,
			&response.FailedAttempts,
			&response.UnlockToken,
			&response.LockedAt,
			&response.OtpSecret,
			&response.OtpExpiredAt,
			&response.Photo,
			&response.UserLevel,
			&response.ReferralCode,
			&response.ReferrerID,
			&response.ReferralPoints,
			&response.RewardPoints,
			&response.Wallet,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeactivatedBy,
			&response.DeactivatedAt,
			&response.DeactivationReason,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE users SET")
	if response.Status != request.Status {
		params = append(params, request.Status)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" status = ?")
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" name = ?")
	}
	if (response.MobilePhoneConfirmationSentAt != nil && request.MobilePhoneConfirmationSentAt == nil) ||
		(response.MobilePhoneConfirmationSentAt == nil && request.MobilePhoneConfirmationSentAt != nil) ||
		(response.MobilePhoneConfirmationSentAt != nil && request.MobilePhoneConfirmationSentAt != nil &&
			!response.MobilePhoneConfirmationSentAt.Equal(*request.MobilePhoneConfirmationSentAt)) {
		params = append(params, request.MobilePhoneConfirmationSentAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" mobile_phone_confirmation_sent_at = ?")
	}
	if (response.MobilePhoneConfirmationToken != nil && request.MobilePhoneConfirmationToken == nil) ||
		(response.MobilePhoneConfirmationToken == nil && request.MobilePhoneConfirmationToken != nil) ||
		(response.MobilePhoneConfirmationToken != nil && request.MobilePhoneConfirmationToken != nil &&
			*response.MobilePhoneConfirmationToken != *request.MobilePhoneConfirmationToken) {
		params = append(params, request.MobilePhoneConfirmationToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" mobile_phone_confirmation_token = ?")
	}
	if (response.MobilePhoneConfirmedAt != nil && request.MobilePhoneConfirmedAt == nil) ||
		(response.MobilePhoneConfirmedAt == nil && request.MobilePhoneConfirmedAt != nil) ||
		(response.MobilePhoneConfirmedAt != nil && request.MobilePhoneConfirmedAt != nil &&
			!response.MobilePhoneConfirmedAt.Equal(*request.MobilePhoneConfirmedAt)) {
		params = append(params, request.MobilePhoneConfirmedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" mobile_phone_confirmed_at = ?")
	}
	if response.MobilePhone != request.MobilePhone {
		params = append(params, request.MobilePhone)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" mobile_phone = ?")
	}
	if response.Username != request.Username {
		params = append(params, request.Username)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" username = ?")
	}
	if (response.EmailConfirmationSentAt != nil && request.EmailConfirmationSentAt == nil) ||
		(response.EmailConfirmationSentAt == nil && request.EmailConfirmationSentAt != nil) ||
		(response.EmailConfirmationSentAt != nil && request.EmailConfirmationSentAt != nil &&
			!response.EmailConfirmationSentAt.Equal(*request.EmailConfirmationSentAt)) {
		params = append(params, request.EmailConfirmationSentAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" email_confirmation_sent_at = ?")
	}
	if (response.EmailConfirmationToken != nil && request.EmailConfirmationToken == nil) ||
		(response.EmailConfirmationToken == nil && request.EmailConfirmationToken != nil) ||
		(response.EmailConfirmationToken != nil && request.EmailConfirmationToken != nil &&
			*response.EmailConfirmationToken != *request.EmailConfirmationToken) {
		params = append(params, request.EmailConfirmationToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" email_confirmation_token = ?")
	}
	if (response.EmailConfirmedAt != nil && request.EmailConfirmedAt == nil) ||
		(response.EmailConfirmedAt == nil && request.EmailConfirmedAt != nil) ||
		(response.EmailConfirmedAt != nil && request.EmailConfirmedAt != nil &&
			!response.EmailConfirmedAt.Equal(*request.EmailConfirmedAt)) {
		params = append(params, request.EmailConfirmedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" email_confirmed_at = ?")
	}
	if response.Email != request.Email {
		params = append(params, request.Email)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" email = ?")
	}
	if (response.EncryptedPassword != nil && request.EncryptedPassword == nil) ||
		(response.EncryptedPassword == nil && request.EncryptedPassword != nil) ||
		(response.EncryptedPassword != nil && request.EncryptedPassword != nil &&
			*response.EncryptedPassword != *request.EncryptedPassword) {
		params = append(params, request.EncryptedPassword)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" encrypted_password = ?")
	}
	if (response.PasswordResetSentAt != nil && request.PasswordResetSentAt == nil) ||
		(response.PasswordResetSentAt == nil && request.PasswordResetSentAt != nil) ||
		(response.PasswordResetSentAt != nil && request.PasswordResetSentAt != nil &&
			!response.PasswordResetSentAt.Equal(*request.PasswordResetSentAt)) {
		params = append(params, request.PasswordResetSentAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" password_reset_sent_at = ?")
	}
	if (response.PasswordResetToken != nil && request.PasswordResetToken == nil) ||
		(response.PasswordResetToken == nil && request.PasswordResetToken != nil) ||
		(response.PasswordResetToken != nil && request.PasswordResetToken != nil &&
			*response.PasswordResetToken != *request.PasswordResetToken) {
		params = append(params, request.PasswordResetToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" password_reset_token = ?")
	}
	if (response.PasswordChangedAt != nil && request.PasswordChangedAt == nil) ||
		(response.PasswordChangedAt == nil && request.PasswordChangedAt != nil) ||
		(response.PasswordChangedAt != nil && request.PasswordChangedAt != nil &&
			!response.PasswordChangedAt.Equal(*request.PasswordChangedAt)) {
		params = append(params, request.PasswordChangedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" password_changed_at = ?")
	}
	if response.SignInCount != request.SignInCount {
		params = append(params, request.SignInCount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" sign_in_count = ?")
	}
	if (response.CurrentSignInAt != nil && request.CurrentSignInAt == nil) ||
		(response.CurrentSignInAt == nil && request.CurrentSignInAt != nil) ||
		(response.CurrentSignInAt != nil && request.CurrentSignInAt != nil &&
			!response.CurrentSignInAt.Equal(*request.CurrentSignInAt)) {
		params = append(params, request.CurrentSignInAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" current_sign_in_at = ?")
	}
	if (response.LastSignInAt != nil && request.LastSignInAt == nil) ||
		(response.LastSignInAt == nil && request.LastSignInAt != nil) ||
		(response.LastSignInAt != nil && request.LastSignInAt != nil &&
			!response.LastSignInAt.Equal(*request.LastSignInAt)) {
		params = append(params, request.LastSignInAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" last_sign_in_at = ?")
	}
	if (response.CurrentSignInIP != nil && request.CurrentSignInIP == nil) ||
		(response.CurrentSignInIP == nil && request.CurrentSignInIP != nil) ||
		(response.CurrentSignInIP != nil && request.CurrentSignInIP != nil &&
			*response.CurrentSignInIP != *request.CurrentSignInIP) {
		params = append(params, request.CurrentSignInIP)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" current_sign_in_ip = ?")
	}
	if (response.LastSignInIP != nil && request.LastSignInIP == nil) ||
		(response.LastSignInIP == nil && request.LastSignInIP != nil) ||
		(response.LastSignInIP != nil && request.LastSignInIP != nil &&
			*response.LastSignInIP != *request.LastSignInIP) {
		params = append(params, request.LastSignInIP)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" last_sign_in_ip = ?")
	}
	if response.FailedAttempts != request.FailedAttempts {
		params = append(params, request.FailedAttempts)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" failed_attempts = ?")
	}
	if (response.UnlockToken != nil && request.UnlockToken == nil) ||
		(response.UnlockToken == nil && request.UnlockToken != nil) ||
		(response.UnlockToken != nil && request.UnlockToken != nil &&
			*response.UnlockToken != *request.UnlockToken) {
		params = append(params, request.UnlockToken)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" unlock_token = ?")
	}
	if (response.LockedAt != nil && request.LockedAt == nil) ||
		(response.LockedAt == nil && request.LockedAt != nil) ||
		(response.LockedAt != nil && request.LockedAt != nil &&
			!response.LockedAt.Equal(*request.LockedAt)) {
		params = append(params, request.LockedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" locked_at = ?")
	}
	if (response.OtpSecret != nil && request.OtpSecret == nil) ||
		(response.OtpSecret == nil && request.OtpSecret != nil) ||
		(response.OtpSecret != nil && request.OtpSecret != nil &&
			*response.OtpSecret != *request.OtpSecret) {
		params = append(params, request.OtpSecret)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" otp_secret = ?")
	}
	if (response.OtpExpiredAt != nil && request.OtpExpiredAt == nil) ||
		(response.OtpExpiredAt == nil && request.OtpExpiredAt != nil) ||
		(response.OtpExpiredAt != nil && request.OtpExpiredAt != nil &&
			!response.OtpExpiredAt.Equal(*request.OtpExpiredAt)) {
		params = append(params, request.OtpExpiredAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" otp_expired_at = ?")
	}
	if response.Photo != request.Photo {
		params = append(params, request.Photo)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" photo = ?")
	}
	if (response.ReferrerID == nil && request.ReferrerID != nil) ||
		(response.ReferrerID != nil && request.ReferrerID == nil) ||
		(response.ReferrerID != nil && request.ReferrerID != nil &&
			response.ReferrerID.Compare(*request.ReferrerID) != 0) {
		params = append(params, request.ReferrerID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" referrer_id = ?")
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
		_, _ = sb.WriteString(" updated_by = ?")
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" updated_at = ?")
	}
	if (response.DeactivatedBy != nil && request.DeactivatedBy == nil) ||
		(response.DeactivatedBy == nil && request.DeactivatedBy != nil) ||
		(response.DeactivatedBy != nil && request.DeactivatedBy != nil &&
			*response.DeactivatedBy != *request.DeactivatedBy) {
		params = append(params, request.DeactivatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deactivated_by = ?")
	}
	if (response.DeactivatedAt != nil && request.DeactivatedAt == nil) ||
		(response.DeactivatedAt == nil && request.DeactivatedAt != nil) ||
		(response.DeactivatedAt != nil && request.DeactivatedAt != nil &&
			!response.DeactivatedAt.Equal(*request.DeactivatedAt)) {
		params = append(params, request.DeactivatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deactivated_at = ?")
	}
	if (response.DeactivationReason != nil && request.DeactivationReason == nil) ||
		(response.DeactivationReason == nil && request.DeactivationReason != nil) ||
		(response.DeactivationReason != nil && request.DeactivationReason != nil &&
			*response.DeactivationReason != *request.DeactivationReason) {
		params = append(params, request.DeactivationReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deactivation_reason = ?")
	}
	params = append(params, response.ID)
	_, _ = sb.WriteString(` WHERE id = ?
		RETURNING id
			, role
			, status
			, name
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, username
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_sent_at
			, password_reset_token
			, password_changed_at
			, sign_in_count
			, current_sign_in_at
			, last_sign_in_at
			, current_sign_in_ip
			, last_sign_in_ip
			, failed_attempts
			, unlock_token
			, locked_at
			, otp_secret
			, otp_expired_at
			, photo
			, user_level
			, referral_code
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`)
	err = tx.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.Role,
			&response.Status,
			&response.Name,
			&response.MobilePhoneConfirmationSentAt,
			&response.MobilePhoneConfirmationToken,
			&response.MobilePhoneConfirmedAt,
			&response.MobilePhone,
			&response.Username,
			&response.EmailConfirmationSentAt,
			&response.EmailConfirmationToken,
			&response.EmailConfirmedAt,
			&response.Email,
			&response.EncryptedPassword,
			&response.PasswordResetSentAt,
			&response.PasswordResetToken,
			&response.PasswordChangedAt,
			&response.SignInCount,
			&response.CurrentSignInAt,
			&response.LastSignInAt,
			&response.CurrentSignInIP,
			&response.LastSignInIP,
			&response.FailedAttempts,
			&response.UnlockToken,
			&response.LockedAt,
			&response.OtpSecret,
			&response.OtpExpiredAt,
			&response.Photo,
			&response.UserLevel,
			&response.ReferralCode,
			&response.ReferrerID,
			&response.ReferralPoints,
			&response.RewardPoints,
			&response.Wallet,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeactivatedBy,
			&response.DeactivatedAt,
			&response.DeactivationReason,
		)...,
	)
	if err != nil {
		if key, ok := uniqueViolationKey(err); ok {
			switch key {
			case "users_username_key":
				err = models.ErrUniqueUserUsernameViolation
			case "users_mobile_phone_key":
				err = models.ErrUniqueUserMobilePhoneViolation
			case "users_email_key":
				err = models.ErrUniqueUserEmailViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

// AddUserBalances adds the given amounts (which may be negative) to a user's
// referral points, reward points and wallet. Zero amounts are left alone.
func (q *userRepository) AddUserBalances(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	referralPoints,
	rewardPoints,
	wallet int64,
	now time.Time,
) (*models.User, error) {
	ctxt := "UserRepository-AddUserBalances"
	var response models.User
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, role
			, status
			, name
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, username
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_sent_at
			, password_reset_token
			, password_changed_at
			, sign_in_count
			, current_sign_in_at
			, last_sign_in_at
			, current_sign_in_ip
			, last_sign_in_ip
			, failed_attempts
			, unlock_token
			, locked_at
			, otp_secret
			, otp_expired_at
			, photo
			, user_level
			, referral_code
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM users
		WHERE id = ?
		FOR UPDATE`,
		helper.SQLArgs(userID)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.Role,
			&response.Status,
			&response.Name,
			&response.MobilePhoneConfirmationSentAt,
			&response.MobilePhoneConfirmationToken,
			&response.MobilePhoneConfirmedAt,
			&response.MobilePhone,
			&response.Username,
			&response.EmailConfirmationSentAt,
			&response.EmailConfirmationToken,
			&response.EmailConfirmedAt,
			&response.Email,
			&response.EncryptedPassword,
			&response.PasswordResetSentAt,
			&response.PasswordResetToken,
			&response.PasswordChangedAt,
			&response.SignInCount,
			&response.CurrentSignInAt,
			&response.LastSignInAt,
			&response.CurrentSignInIP,
			&response.LastSignInIP,
			&response.FailedAttempts,
			&response.UnlockToken,
			&response.LockedAt,
			&response.OtpSecret,
			&response.OtpExpiredAt,
			&response.Photo,
			&response.UserLevel,
			&response.ReferralCode,
			&response.ReferrerID,
			&response.ReferralPoints,
			&response.RewardPoints,
			&response.Wallet,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeactivatedBy,
			&response.DeactivatedAt,
			&response.DeactivationReason,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE users SET")
	if referralPoints != 0 {
		params = append(params, referralPoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" referral_points = COALESCE(referral_points, 0) + ?")
	}
	if rewardPoints != 0 {
		params = append(params, rewardPoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" reward_points = COALESCE(reward_points, 0) + ?")
	}
	if wallet != 0 {
		params = append(params, wallet)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" wallet = COALESCE(wallet, 0) + ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, now)
	if n = len(params); n > 1 {
		_, _ = sb.WriteString(",")
	}
	_, _ = sb.WriteString(" updated_at = ?")
	params = append(params, userID)
	_, _ = sb.WriteString(` WHERE id = ?
		RETURNING id
			, role
			, status
			, name
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
			, username
			, email_confirmation_sent_at
			, email_confirmation_token
			, email_confirmed_at
			, email
			, encrypted_password
			, password_reset_sent_at
			, password_reset_token
			, password_changed_at
			, sign_in_count
			, current_sign_in_at
			, last_sign_in_at
			, current_sign_in_ip
			, last_sign_in_ip
			, failed_attempts
			, unlock_token
			, locked_at
			, otp_secret
			, otp_expired_at
			, photo
			, user_level
			, referral_code
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`)
	err = tx.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.Role,
			&response.Status,
			&response.Name,
			&response.MobilePhoneConfirmationSentAt,
			&response.MobilePhoneConfirmationToken,
			&response.MobilePhoneConfirmedAt,
			&response.MobilePhone,
			&response.Username,
			&response.EmailConfirmationSentAt,
			&response.EmailConfirmationToken,
			&response.EmailConfirmedAt,
			&response.Email,
			&response.EncryptedPassword,
			&response.PasswordResetSentAt,
			&response.PasswordResetToken,
			&response.PasswordChangedAt,
			&response.SignInCount,
			&response.CurrentSignInAt,
			&response.LastSignInAt,
			&response.CurrentSignInIP,
			&response.LastSignInIP,
			&response.FailedAttempts,
			&response.UnlockToken,
			&response.LockedAt,
			&response.OtpSecret,
			&response.OtpExpiredAt,
			&response.Photo,
			&response.UserLevel,
			&response.ReferralCode,
			&response.ReferrerID,
			&response.ReferralPoints,
			&response.RewardPoints,
			&response.Wallet,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeactivatedBy,
			&response.DeactivatedAt,
			&response.DeactivationReason,
		)...,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
