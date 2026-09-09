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

type SellerRepository interface {
	FindSellers(
		ctx context.Context,
		request *models.SellerFilter,
	) ([]*models.Seller, int64, int64, error)
	FindSeller(
		ctx context.Context,
		request *models.SellerFilter,
	) (*models.Seller, error)
	CreateSeller(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Seller,
	) (*models.Seller, error)
	UpdateSeller(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Seller,
	) (*models.Seller, error)
}

type sellerRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewSellerRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) SellerRepository {
	return &sellerRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *sellerRepository) FindSellers(
	ctx context.Context,
	request *models.SellerFilter,
) ([]*models.Seller, int64, int64, error) {
	ctxt := "SellerRepository-FindSellers"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM sellers %s", condition)
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
		, company_id
		, status
		, seller_level
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
		, photo
		, unread_notifications
		, telegram_chat_id
		, telegram_username
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
	var response []*models.Seller
	for rows.Next() {
		var seller models.Seller
		if err = rows.Scan(
			&seller.ID,
			&seller.CompanyID,
			&seller.Status,
			&seller.SellerLevel,
			&seller.Name,
			&seller.UnconfirmedMobilePhone,
			&seller.MobilePhoneConfirmationSentAt,
			&seller.MobilePhoneConfirmationToken,
			&seller.MobilePhoneConfirmedAt,
			&seller.MobilePhone,
			&seller.Username,
			&seller.UnconfirmedEmail,
			&seller.EmailConfirmationSentAt,
			&seller.EmailConfirmationToken,
			&seller.EmailConfirmedAt,
			&seller.Email,
			&seller.Photo,
			&seller.UnreadNotifications,
			&seller.TelegramChatID,
			&seller.TelegramUsername,
			&seller.CreatedBy,
			&seller.CreatedAt,
			&seller.UpdatedBy,
			&seller.UpdatedAt,
			&seller.DeactivatedBy,
			&seller.DeactivatedAt,
			&seller.DeactivationReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		seller.RowNo = offset
		response = append(response, &seller)
	}
	return response, total, pages, nil
}

func (q *sellerRepository) FindSeller(
	ctx context.Context,
	request *models.SellerFilter,
) (*models.Seller, error) {
	ctxt := "SellerRepository-FindSeller"
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
			, company_id
			, status
			, seller_level
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
			, photo
			, unread_notifications
			, telegram_chat_id
			, telegram_username
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM sellers %s`,
		condition,
	)
	var response models.Seller
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.CompanyID,
		&response.Status,
		&response.SellerLevel,
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
		&response.Photo,
		&response.UnreadNotifications,
		&response.TelegramChatID,
		&response.TelegramUsername,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrSellerNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *sellerRepository) buildCondition(request *models.SellerFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.CompanyIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("company_id IN (")
		for i, companyID := range request.CompanyIDs {
			params = append(params, companyID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.SellerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, sellerID := range request.SellerIDs {
			params = append(params, sellerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedSellerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedSellerID := range request.ExcludedSellerIDs {
			params = append(params, excludedSellerID)
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

func (q *sellerRepository) CreateSeller(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Seller,
) (*models.Seller, error) {
	ctxt := "SellerRepository-CreateSeller"
	var response models.Seller
	err := tx.QueryRow(
		ctx,
		`INSERT INTO sellers (
			company_id
			, status
			, seller_level
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
			, photo
			, unread_notifications
			, telegram_chat_id
			, telegram_username
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			, $21, $22, $23
		)
		RETURNING id
			, company_id
			, status
			, seller_level
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
			, photo
			, unread_notifications
			, telegram_chat_id
			, telegram_username
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		request.CompanyID,
		request.Status,
		request.SellerLevel,
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
		request.Photo,
		request.UnreadNotifications,
		request.TelegramChatID,
		request.TelegramUsername,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.CompanyID,
		&response.Status,
		&response.SellerLevel,
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
		&response.Photo,
		&response.UnreadNotifications,
		&response.TelegramChatID,
		&response.TelegramUsername,
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
			case "sellers_username_company_id_key":
				err = models.ErrUniqueSellerUsernameViolation
			case "sellers_mobile_phone_company_id_key":
				err = models.ErrUniqueSellerMobilePhoneViolation
			case "sellers_email_company_id_key":
				err = models.ErrUniqueSellerEmailViolation
			case "sellers_telegram_chat_id_key":
				err = models.ErrUniqueTelegramChatIdViolation
			case "sellers_telegram_username_company_id_key":
				err = models.ErrUniqueTelegramUsernameViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *sellerRepository) UpdateSeller(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Seller,
) (*models.Seller, error) {
	ctxt := "SellerRepository-UpdateSeller"
	var response models.Seller
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, company_id
			, status
			, seller_level
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
			, photo
			, unread_notifications
			, telegram_chat_id
			, telegram_username
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM sellers
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.CompanyID,
		&response.Status,
		&response.SellerLevel,
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
		&response.Photo,
		&response.UnreadNotifications,
		&response.TelegramChatID,
		&response.TelegramUsername,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrSellerNotFound
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
	_, _ = sb.WriteString("UPDATE sellers SET")
	if response.Status != request.Status {
		params = append(params, request.Status)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " status = $%d", n)
	}
	if response.SellerLevel != request.SellerLevel {
		params = append(params, request.SellerLevel)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " seller_level = $%d", n)
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
	params = append(params, response.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, company_id
			, status
			, seller_level
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
			, photo
			, unread_notifications
			, telegram_chat_id
			, telegram_username
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.CompanyID,
		&response.Status,
		&response.SellerLevel,
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
		&response.Photo,
		&response.UnreadNotifications,
		&response.TelegramChatID,
		&response.TelegramUsername,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "sellers_username_company_id_key":
				err = models.ErrUniqueSellerUsernameViolation
			case "sellers_mobile_phone_company_id_key":
				err = models.ErrUniqueSellerMobilePhoneViolation
			case "sellers_email_company_id_key":
				err = models.ErrUniqueSellerEmailViolation
			case "sellers_telegram_chat_id_key":
				err = models.ErrUniqueTelegramChatIdViolation
			case "sellers_telegram_username_company_id_key":
				err = models.ErrUniqueTelegramUsernameViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}
