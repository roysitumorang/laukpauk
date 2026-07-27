package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	BuyerRepository interface {
		FindBuyers(
			ctx context.Context,
			request *models.BuyerFilter,
		) ([]*models.Buyer, int64, int64, error)
		FindBuyer(
			ctx context.Context,
			request *models.BuyerFilter,
		) (*models.Buyer, error)
		UpdateBuyer(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Buyer,
		) (*models.Buyer, error)
	}

	buyerRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewBuyerRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) BuyerRepository {
	return &buyerRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *buyerRepository) FindBuyers(
	ctx context.Context,
	request *models.BuyerFilter,
) ([]*models.Buyer, int64, int64, error) {
	ctxt := "BuyerRepository-FindBuyers"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM users u JOIN buyers b ON u.id = b.user_id %s", condition)
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
		`u.id
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
		, b.referral_code
		, b.address
		, b.referrer_id
		, b.referral_points
		, b.reward_points
		, b.wallet
		, b.delivery_route_id`,
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
		_, _ = fmt.Fprintf(&sb, " %s %s", item.Field, item.Direction)
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
		_, _ = fmt.Fprintf(&sb, " LIMIT %d OFFSET %d", request.Limit, offset)
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
	var response []*models.Buyer
	for rows.Next() {
		var buyer models.Buyer
		if err = rows.Scan(
			&buyer.ID,
			&buyer.UserType,
			&buyer.UserStatus,
			&buyer.Name,
			&buyer.ConfirmationToken,
			&buyer.ConfirmedAt,
			&buyer.UnconfirmedMobilePhone,
			&buyer.MobilePhoneConfirmationSentAt,
			&buyer.MobilePhoneConfirmationToken,
			&buyer.MobilePhoneConfirmedAt,
			&buyer.MobilePhone,
			&buyer.OtpToken,
			&buyer.OtpCreatedAt,
			&buyer.OtpExpiredAt,
			&buyer.Username,
			&buyer.UnconfirmedEmail,
			&buyer.EmailConfirmationSentAt,
			&buyer.EmailConfirmationToken,
			&buyer.EmailConfirmedAt,
			&buyer.Email,
			&buyer.EncryptedPassword,
			&buyer.PasswordResetAt,
			&buyer.PasswordResetToken,
			&buyer.PasswordChangedAt,
			&buyer.Photo,
			&buyer.UnreadNotifications,
			&buyer.CreatedBy,
			&buyer.CreatedAt,
			&buyer.UpdatedBy,
			&buyer.UpdatedAt,
			&buyer.DeactivatedBy,
			&buyer.DeactivatedAt,
			&buyer.DeactivationReason,
			&buyer.ReferralCode,
			&buyer.Address,
			&buyer.ReferrerID,
			&buyer.ReferralPoints,
			&buyer.RewardPoints,
			&buyer.Wallet,
			&buyer.DeliveryRouteID,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		buyer.RowNo = offset
		response = append(response, &buyer)
	}
	return response, total, pages, nil
}

func (q *buyerRepository) FindBuyer(
	ctx context.Context,
	request *models.BuyerFilter,
) (*models.Buyer, error) {
	ctxt := "BuyerRepository-FindBuyer"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			u.id
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
			, b.referral_code
			, b.address
			, b.referrer_id
			, b.referral_points
			, b.reward_points
			, b.wallet
			, b.delivery_route_id
		FROM users u
		JOIN buyers b ON u.id = b.user_id %s`,
		condition,
	)
	query := sb.String()
	var response models.Buyer
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
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
		&response.ReferralCode,
		&response.Address,
		&response.ReferrerID,
		&response.ReferralPoints,
		&response.RewardPoints,
		&response.Wallet,
		&response.DeliveryRouteID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrBuyerNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}

func (q *buyerRepository) buildCondition(request *models.BuyerFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("u.id IN (")
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
		_, _ = sb.WriteString("u.id NOT IN (")
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
	if len(request.ReferralCodes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("LOWER(b.referral_code) IN (")
		for i, referralCode := range request.ReferralCodes {
			params = append(params, strings.ToLower(referralCode))
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
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

func (q *buyerRepository) UpdateBuyer(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Buyer,
) (*models.Buyer, error) {
	ctxt := "BuyerRepository-UpdateBuyer"
	var response models.Buyer
	err := q.dbRead.QueryRow(
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
		request.ID,
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
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	if err = tx.QueryRow(
		ctx,
		`SELECT
			user_id
			, referral_code
			, address
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, delivery_route_id
		FROM buyers
		WHERE user_id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.ReferralCode,
		&response.Address,
		&response.ReferrerID,
		&response.ReferralPoints,
		&response.RewardPoints,
		&response.Wallet,
		&response.DeliveryRouteID,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	params = params[:0]
	sb.Reset()
	_, _ = sb.WriteString("UPDATE buyers SET")
	if response.ReferralCode != request.ReferralCode {
		params = append(params, request.ReferralCode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " referral_code = $%d", n)
	}
	if response.Address != request.Address {
		params = append(params, request.Address)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " address = $%d", n)
	}
	if response.ReferrerID == nil && request.ReferrerID != nil {
		params = append(params, request.ReferrerID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " referrer_id = $%d", n)
	}
	if request.ReferralPoints != 0 {
		params = append(params, request.ReferralPoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " referral_points = referral_points + $%d", n)
	}
	if request.RewardPoints != 0 {
		params = append(params, request.RewardPoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " reward_points = reward_points + $%d", n)
	}
	if request.Wallet != 0 {
		params = append(params, request.Wallet)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " wallet = wallet + $%d", n)
	}
	if (response.DeliveryRouteID == nil &&
		request.DeliveryRouteID != nil) ||
		(response.DeliveryRouteID != nil &&
			request.DeliveryRouteID != nil &&
			response.DeliveryRouteID.Compare(*request.DeliveryRouteID) != 0) {
		params = append(params, request.DeliveryRouteID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " delivery_route_id = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		&sb,
		` WHERE id = $%d
		RETURNING user_id
			, referral_code
			, address
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, delivery_route_id`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.ReferralCode,
		&response.Address,
		&response.ReferrerID,
		&response.ReferralPoints,
		&response.RewardPoints,
		&response.Wallet,
		&response.DeliveryRouteID,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
