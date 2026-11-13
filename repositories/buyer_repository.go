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

type BuyerRepository interface {
	FindBuyers(
		ctx context.Context,
		request *models.BuyerFilter,
	) ([]*models.Buyer, int64, int64, error)
	FindBuyer(
		ctx context.Context,
		request *models.BuyerFilter,
	) (*models.Buyer, error)
	CreateBuyer(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Buyer,
	) (*models.Buyer, error)
	UpdateBuyer(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Buyer,
	) (*models.Buyer, error)
}

type buyerRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

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
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM buyers %s", condition)
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
		`id
		, status
		, name
		, unconfirmed_mobile_phone
		, mobile_phone_confirmation_sent_at
		, mobile_phone_confirmation_token
		, mobile_phone_confirmed_at
		, mobile_phone
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
		, referral_code
		, address
		, referrer_id
		, referral_points
		, reward_points
		, wallet
		, transaction_volume
		, delivery_route_id
		, telegram_chat_id
		, telegram_username
		, created_at
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
	var response []*models.Buyer
	for rows.Next() {
		var buyer models.Buyer
		if err = rows.Scan(
			&buyer.ID,
			&buyer.Status,
			&buyer.Name,
			&buyer.UnconfirmedMobilePhone,
			&buyer.MobilePhoneConfirmationSentAt,
			&buyer.MobilePhoneConfirmationToken,
			&buyer.MobilePhoneConfirmedAt,
			&buyer.MobilePhone,
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
			&buyer.ReferralCode,
			&buyer.Address,
			&buyer.ReferrerID,
			&buyer.ReferralPoints,
			&buyer.RewardPoints,
			&buyer.Wallet,
			&buyer.TransactionVolume,
			&buyer.DeliveryRouteID,
			&buyer.TelegramChatID,
			&buyer.TelegramUsername,
			&buyer.CreatedAt,
			&buyer.UpdatedAt,
			&buyer.DeactivatedBy,
			&buyer.DeactivatedAt,
			&buyer.DeactivationReason,
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
			, status
			, name
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
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
			, referral_code
			, address
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, transaction_volume
			, delivery_route_id
			, telegram_chat_id
			, telegram_username
			, created_at
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM buyers %s`,
		condition,
	)
	var response models.Buyer
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.Status,
		&response.Name,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
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
		&response.ReferralCode,
		&response.Address,
		&response.ReferrerID,
		&response.ReferralPoints,
		&response.RewardPoints,
		&response.Wallet,
		&response.TransactionVolume,
		&response.DeliveryRouteID,
		&response.TelegramChatID,
		&response.TelegramUsername,
		&response.CreatedAt,
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
	return &response, nil
}

func (q *buyerRepository) buildCondition(request *models.BuyerFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.BuyerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, buyerID := range request.BuyerIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedBuyerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedBuyerID := range request.ExcludedBuyerIDs {
			params = append(params, excludedBuyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.TelegramChatIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("telegram_chat_id IN (")
		for i, telegramChatID := range request.TelegramChatIDs {
			params = append(params, telegramChatID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Login != "" {
		params = append(params, strings.ToLower(request.Login))
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "(username = $%d OR email = $%d OR mobile_phone = $%d)", n, n, n)
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
			`(LOWER(name) LIKE $%d
				OR username LIKE $%d
				OR email LIKE $%d
				OR mobile_phone LIKE $%d)`,
			n, n, n, n,
		)
		conds = append(conds, sb.String())
	}
	if request.Username != "" {
		params = append(params, strings.ToLower(request.Username))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "username = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.TelegramUsername != "" {
		params = append(params, request.TelegramUsername)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "telegram_username = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.Email != "" {
		params = append(params, strings.ToLower(request.Email))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "email = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.MobilePhone != "" {
		params = append(params, strings.ToLower(request.MobilePhone))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "mobile_phone = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if len(request.ReferralCodes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("LOWER(referral_code) IN (")
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

func (q *buyerRepository) CreateBuyer(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Buyer,
) (*models.Buyer, error) {
	ctxt := "BuyerRepository-CreateBuyer"
	var response models.Buyer
	err := tx.QueryRow(
		ctx,
		`INSERT INTO buyers (
			status
			, name
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
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
			, referral_code
			, address
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, transaction_volume
			, delivery_route_id
			, telegram_chat_id
			, telegram_username
			, created_at
			, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30
			, $31
		)
		RETURNING id
			, status
			, name
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
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
			, referral_code
			, address
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, transaction_volume
			, delivery_route_id
			, telegram_chat_id
			, telegram_username
			, created_at
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		request.Status,
		request.Name,
		request.UnconfirmedMobilePhone,
		request.MobilePhoneConfirmationSentAt,
		request.MobilePhoneConfirmationToken,
		request.MobilePhoneConfirmedAt,
		request.MobilePhone,
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
		request.ReferralCode,
		request.Address,
		request.ReferrerID,
		request.ReferralPoints,
		request.RewardPoints,
		request.Wallet,
		request.TransactionVolume,
		request.DeliveryRouteID,
		request.TelegramChatID,
		request.TelegramUsername,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.Status,
		&response.Name,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
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
		&response.ReferralCode,
		&response.Address,
		&response.ReferrerID,
		&response.ReferralPoints,
		&response.RewardPoints,
		&response.Wallet,
		&response.TransactionVolume,
		&response.DeliveryRouteID,
		&response.TelegramChatID,
		&response.TelegramUsername,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "buyers_username_key":
				err = models.ErrUniqueBuyerUsernameViolation
			case "buyers_mobile_phone_key":
				err = models.ErrUniqueBuyerMobilePhoneViolation
			case "buyers_email_key":
				err = models.ErrUniqueBuyerEmailViolation
			case "buyers_telegram_chat_id_key":
				err = models.ErrUniqueTelegramChatIdViolation
			case "buyers_telegram_username_key":
				err = models.ErrUniqueTelegramUsernameViolation
			case "buyers_lower_idx":
				err = models.ErrUniqueBuyerReferralCodeViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *buyerRepository) UpdateBuyer(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Buyer,
) (*models.Buyer, error) {
	ctxt := "BuyerRepository-UpdateBuyer"
	var response models.Buyer
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, status
			, name
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
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
			, referral_code
			, address
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, transaction_volume
			, delivery_route_id
			, telegram_chat_id
			, telegram_username
			, created_at
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM buyers
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.Status,
		&response.Name,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
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
		&response.ReferralCode,
		&response.Address,
		&response.ReferrerID,
		&response.ReferralPoints,
		&response.RewardPoints,
		&response.Wallet,
		&response.TransactionVolume,
		&response.DeliveryRouteID,
		&response.TelegramChatID,
		&response.TelegramUsername,
		&response.CreatedAt,
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
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE buyers SET")
	if response.Status != request.Status {
		params = append(params, request.Status)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " status = $%d", n)
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " name = $%d", n)
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
	if response.MobilePhone != request.MobilePhone {
		params = append(params, request.MobilePhone)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " mobile_phone = $%d", n)
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
	if response.ReferralCode != request.ReferralCode {
		params = append(params, request.ReferralCode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " referral_code = $%d", n)
	}
	if response.Address != request.Address {
		params = append(params, request.Address)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " address = $%d", n)
	}
	if response.ReferrerID == nil && request.ReferrerID != nil {
		params = append(params, request.ReferrerID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " referrer_id = $%d", n)
	}
	if request.ReferralPoints != 0 {
		params = append(params, request.ReferralPoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " referral_points = referral_points + $%d", n)
	}
	if request.RewardPoints != 0 {
		params = append(params, request.RewardPoints)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " reward_points = reward_points + $%d", n)
	}
	if request.Wallet != 0 {
		params = append(params, request.Wallet)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " wallet = wallet + $%d", n)
	}
	if request.TransactionVolume != 0 {
		params = append(params, request.TransactionVolume)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " transaction_volume = transaction_volume + $%d", n)
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
		_, _ = fmt.Fprintf(sb, " delivery_route_id = $%d", n)
	}
	if (response.TelegramChatID != nil && request.TelegramChatID == nil) ||
		(response.TelegramChatID == nil && request.TelegramChatID != nil) ||
		(response.TelegramChatID != nil && request.TelegramChatID != nil &&
			*response.TelegramChatID != *request.TelegramChatID) {
		params = append(params, request.TelegramChatID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " telegram_chat_id = $%d", n)
	}
	if (response.TelegramUsername != nil && request.TelegramUsername == nil) ||
		(response.TelegramUsername == nil && request.TelegramUsername != nil) ||
		(response.TelegramUsername != nil && request.TelegramUsername != nil &&
			*response.TelegramUsername != *request.TelegramUsername) {
		params = append(params, request.TelegramUsername)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " telegram_username = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
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
	params = append(params, response.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, status
			, name
			, unconfirmed_mobile_phone
			, mobile_phone_confirmation_sent_at
			, mobile_phone_confirmation_token
			, mobile_phone_confirmed_at
			, mobile_phone
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
			, referral_code
			, address
			, referrer_id
			, referral_points
			, reward_points
			, wallet
			, transaction_volume
			, delivery_route_id
			, telegram_chat_id
			, telegram_username
			, created_at
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.Status,
		&response.Name,
		&response.UnconfirmedMobilePhone,
		&response.MobilePhoneConfirmationSentAt,
		&response.MobilePhoneConfirmationToken,
		&response.MobilePhoneConfirmedAt,
		&response.MobilePhone,
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
		&response.ReferralCode,
		&response.Address,
		&response.ReferrerID,
		&response.ReferralPoints,
		&response.RewardPoints,
		&response.Wallet,
		&response.TransactionVolume,
		&response.DeliveryRouteID,
		&response.TelegramChatID,
		&response.TelegramUsername,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "buyers_username_key":
				err = models.ErrUniqueBuyerUsernameViolation
			case "buyers_mobile_phone_key":
				err = models.ErrUniqueBuyerMobilePhoneViolation
			case "buyers_email_key":
				err = models.ErrUniqueBuyerEmailViolation
			case "buyers_telegram_chat_id_key":
				err = models.ErrUniqueTelegramChatIdViolation
			case "buyers_telegram_username_key":
				err = models.ErrUniqueTelegramUsernameViolation
			case "buyers_lower_idx":
				err = models.ErrUniqueBuyerReferralCodeViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}
