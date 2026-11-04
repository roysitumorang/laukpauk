package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type CompanyRepository interface {
	FindCompanies(
		ctx context.Context,
		request *models.CompanyFilter,
	) ([]*models.Company, int64, int64, error)
	FindCompany(
		ctx context.Context,
		request *models.CompanyFilter,
	) (*models.Company, error)
	CreateCompany(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Company,
	) (*models.Company, error)
	UpdateCompany(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Company,
	) (*models.Company, error)
	AddReferralPointsExpense(
		ctx context.Context,
		tx pgx.Tx,
		companyID uuid.UUID,
		addition int64,
		updatedBy uuid.UUID,
		updatedAt time.Time,
	) (*models.Company, error)
}

type companyRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewCompanyRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) CompanyRepository {
	return &companyRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *companyRepository) FindCompanies(
	ctx context.Context,
	request *models.CompanyFilter,
) ([]*models.Company, int64, int64, error) {
	ctxt := "CompanyRepository-FindCompanies"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM companies %s", condition)
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
		, name
		, subdomain
		, billed
		, billing_interval
		, subscription_status
		, subscription_start_date
		, subscription_end_date
		, vacation_mode
		, vacation_end_date
		, voucher_enabled
		, referral_points_enabled
		, referral_points_minimum_shopping_cost
		, referral_points_amount
		, referral_points_slots
		, referral_points_expense
		, reward_points_accumulation_window_in_days
		, delivery_origin_longitude
		, delivery_origin_latitude
		, delivery_cost_per_km
		, delivery_max_distance_km
		, whatsapp_phone_number
		, payment_terms
		, maximum_delivery_time
		, transaction_auto_withdrawal_enabled
		, lower(transaction_auto_withdrawal_window)
		, upper(transaction_auto_withdrawal_window)
		, otp_required_for_login
		, owner_id
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
		field := item.Field
		if item.Field == models.CompanySortingFieldName {
			field = "lower(name)"
		}
		_, _ = fmt.Fprintf(sb, " %s %s", field, item.Direction)
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
	var response []*models.Company
	for rows.Next() {
		var company models.Company
		if err = rows.Scan(
			&company.ID,
			&company.Name,
			&company.Subdomain,
			&company.Billed,
			&company.BillingInterval,
			&company.SubscriptionStatus,
			&company.SubscriptionStartDate,
			&company.SubscriptionEndDate,
			&company.VacationMode,
			&company.VacationEndDate,
			&company.VoucherEnabled,
			&company.ReferralPointsEnabled,
			&company.ReferralPointsMinimumShoppingCost,
			&company.ReferralPointsAmount,
			&company.ReferralPointsSlots,
			&company.ReferralPointsExpense,
			&company.RewardPointsAccumulationWindowInDays,
			&company.DeliveryOriginLongitude,
			&company.DeliveryOriginLatitude,
			&company.DeliveryCostPerKM,
			&company.DeliveryMaxDistanceKM,
			&company.WhatsappPhoneNumber,
			&company.PaymentTerms.Duration,
			&company.MaximumDeliveryTime.Duration,
			&company.TransactionAutoWithdrawalEnabled,
			&company.TransactionAutoWithdrawalWindowMin,
			&company.TransactionAutoWithdrawalWindowMax,
			&company.OtpRequiredForLogin,
			&company.OwnerID,
			&company.CreatedBy,
			&company.CreatedAt,
			&company.UpdatedBy,
			&company.UpdatedAt,
			&company.DeactivatedBy,
			&company.DeactivatedAt,
			&company.DeactivationReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		company.RowNo = offset
		response = append(response, (&company).Populate())
	}
	return response, total, pages, nil
}

func (q *companyRepository) FindCompany(
	ctx context.Context,
	request *models.CompanyFilter,
) (*models.Company, error) {
	ctxt := "CompanyRepository-FindCompany"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`SELECT
			id
			, name
			, subdomain
			, billed
			, billing_interval
			, subscription_status
			, subscription_start_date
			, subscription_end_date
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, lower(transaction_auto_withdrawal_window)
			, upper(transaction_auto_withdrawal_window)
			, otp_required_for_login
			, owner_id
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM companies %s`,
		condition,
	)
	query := sb.String()
	var response models.Company
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.Name,
		&response.Subdomain,
		&response.Billed,
		&response.BillingInterval,
		&response.SubscriptionStatus,
		&response.SubscriptionStartDate,
		&response.SubscriptionEndDate,
		&response.VacationMode,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsAmount,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms.Duration,
		&response.MaximumDeliveryTime.Duration,
		&response.TransactionAutoWithdrawalEnabled,
		&response.TransactionAutoWithdrawalWindowMin,
		&response.TransactionAutoWithdrawalWindowMax,
		&response.OtpRequiredForLogin,
		&response.OwnerID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrCompanyNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *companyRepository) CreateCompany(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Company,
) (*models.Company, error) {
	ctxt := "CompanyRepository-CreateCompany"
	var response models.Company
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO companies (
			name
			, subdomain
			, billed
			, billing_interval
			, subscription_status
			, subscription_start_date
			, subscription_end_date
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, transaction_auto_withdrawal_window
			, otp_required_for_login
			, owner_id
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			, $21, $22, $23, $24, tstzrange($25, $26, '[]'), $27, $28, $29, $30
			, $31, $32
		)
		RETURNING id
			, name
			, subdomain
			, billed
			, billing_interval
			, subscription_status
			, subscription_start_date
			, subscription_end_date
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, lower(transaction_auto_withdrawal_window)
			, upper(transaction_auto_withdrawal_window)
			, otp_required_for_login
			, owner_id
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		request.Name,
		request.Subdomain,
		request.Billed,
		request.BillingInterval,
		request.SubscriptionStatus,
		request.SubscriptionStartDate,
		request.SubscriptionEndDate,
		request.VacationMode,
		request.VacationEndDate,
		request.VoucherEnabled,
		request.ReferralPointsEnabled,
		request.ReferralPointsMinimumShoppingCost,
		request.ReferralPointsAmount,
		request.ReferralPointsSlots,
		request.ReferralPointsExpense,
		request.RewardPointsAccumulationWindowInDays,
		request.DeliveryOriginLongitude,
		request.DeliveryOriginLatitude,
		request.DeliveryCostPerKM,
		request.DeliveryMaxDistanceKM,
		request.WhatsappPhoneNumber,
		request.PaymentTerms.Duration,
		request.MaximumDeliveryTime.Duration,
		request.TransactionAutoWithdrawalEnabled,
		request.TransactionAutoWithdrawalWindowMin,
		request.TransactionAutoWithdrawalWindowMax,
		request.OtpRequiredForLogin,
		request.OwnerID,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.Name,
		&response.Subdomain,
		&response.Billed,
		&response.BillingInterval,
		&response.SubscriptionStatus,
		&response.SubscriptionStartDate,
		&response.SubscriptionEndDate,
		&response.VacationMode,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsAmount,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms.Duration,
		&response.MaximumDeliveryTime.Duration,
		&response.TransactionAutoWithdrawalEnabled,
		&response.TransactionAutoWithdrawalWindowMin,
		&response.TransactionAutoWithdrawalWindowMax,
		&response.OtpRequiredForLogin,
		&response.OwnerID,
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
			case "companies_subdomain_key":
				err = models.ErrUniqueCompanySubdomainViolation
			case "companies_whatsapp_phone_number_key":
				err = models.ErrUniqueCompanyWhatsappPhoneNumberViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *companyRepository) UpdateCompany(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Company,
) (*models.Company, error) {
	ctxt := "CompanyRepository-UpdateCompany"
	var response models.Company
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, name
			, subdomain
			, billed
			, billing_interval
			, subscription_status
			, subscription_start_date
			, subscription_end_date
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, lower(transaction_auto_withdrawal_window)
			, upper(transaction_auto_withdrawal_window)
			, otp_required_for_login
			, owner_id
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM companies
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.Name,
		&response.Subdomain,
		&response.Billed,
		&response.BillingInterval,
		&response.SubscriptionStatus,
		&response.SubscriptionStartDate,
		&response.SubscriptionEndDate,
		&response.VacationMode,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsAmount,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms.Duration,
		&response.MaximumDeliveryTime.Duration,
		&response.TransactionAutoWithdrawalEnabled,
		&response.TransactionAutoWithdrawalWindowMin,
		&response.TransactionAutoWithdrawalWindowMax,
		&response.OtpRequiredForLogin,
		&response.OwnerID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrCompanyNotFound
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
	var params []any
	var n int
	_, _ = sb.WriteString("UPDATE companies SET")
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " name = $%d", n)
	}
	if response.Subdomain != request.Subdomain {
		params = append(params, request.Subdomain)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " subdomain = $%d", n)
	}
	if response.Billed != request.Billed {
		params = append(params, request.Billed)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " billed = $%d", n)
	}
	if response.BillingInterval != request.BillingInterval {
		params = append(params, request.BillingInterval)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " billing_interval = $%d", n)
	}
	if response.SubscriptionStatus != request.SubscriptionStatus {
		params = append(params, request.SubscriptionStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " subscription_status = $%d", n)
	}
	if response.SubscriptionStartDate != request.SubscriptionStartDate {
		params = append(params, request.SubscriptionStartDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " subscription_start_date = $%d", n)
	}
	if response.SubscriptionEndDate != request.SubscriptionEndDate {
		params = append(params, request.SubscriptionEndDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " subscription_end_date = $%d", n)
	}
	if response.VacationMode != request.VacationMode {
		params = append(params, request.VacationMode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " vacation_mode = $%d", n)
	}
	if (response.VacationEndDate == nil &&
		request.VacationEndDate != nil) ||
		(response.VacationEndDate != nil &&
			request.VacationEndDate == nil) ||
		(response.VacationEndDate != nil &&
			request.VacationEndDate != nil &&
			!response.VacationEndDate.Equal(*request.VacationEndDate)) {
		params = append(params, request.VacationEndDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " vacation_end_date = $%d", n)
	}
	if response.VoucherEnabled != request.VoucherEnabled {
		params = append(params, request.VoucherEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " voucher_enabled = $%d", n)
	}
	if response.ReferralPointsEnabled != request.ReferralPointsEnabled {
		params = append(params, request.ReferralPointsEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " referral_points_enabled = $%d", n)
	}
	if response.ReferralPointsMinimumShoppingCost != request.ReferralPointsMinimumShoppingCost {
		params = append(params, request.ReferralPointsMinimumShoppingCost)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " referral_points_minimum_shopping_cost = $%d", n)
	}
	if response.ReferralPointsAmount != request.ReferralPointsAmount {
		params = append(params, request.ReferralPointsAmount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " referral_points_amount = $%d", n)
	}
	if slices.Compare(response.ReferralPointsSlots, request.ReferralPointsSlots) != 0 {
		if request.ReferralPointsSlots == nil {
			request.ReferralPointsSlots = []int64{}
		}
		params = append(params, request.ReferralPointsSlots)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " referral_points_slots = $%d", n)
	}
	if response.RewardPointsAccumulationWindowInDays != request.RewardPointsAccumulationWindowInDays {
		params = append(params, request.RewardPointsAccumulationWindowInDays)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " reward_points_accumulation_window_in_days = $%d", n)
	}
	if response.DeliveryOriginLongitude != request.DeliveryOriginLongitude {
		params = append(params, request.DeliveryOriginLongitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " delivery_origin_longitude = $%d", n)
	}
	if response.DeliveryOriginLatitude != request.DeliveryOriginLatitude {
		params = append(params, request.DeliveryOriginLatitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " delivery_origin_latitude = $%d", n)
	}
	if response.DeliveryCostPerKM != request.DeliveryCostPerKM {
		params = append(params, request.DeliveryCostPerKM)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " delivery_cost_per_km = $%d", n)
	}
	if response.DeliveryMaxDistanceKM != request.DeliveryMaxDistanceKM {
		params = append(params, request.DeliveryMaxDistanceKM)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " delivery_max_distance_km = $%d", n)
	}
	if response.WhatsappPhoneNumber != request.WhatsappPhoneNumber {
		params = append(params, request.WhatsappPhoneNumber)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " whatsapp_phone_number = $%d", n)
	}
	if response.PaymentTerms != request.PaymentTerms {
		params = append(params, request.PaymentTerms.Duration)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " payment_terms = $%d", n)
	}
	if response.MaximumDeliveryTime != request.MaximumDeliveryTime {
		params = append(params, request.MaximumDeliveryTime.Duration)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " maximum_delivery_time = $%d", n)
	}
	if response.TransactionAutoWithdrawalEnabled != request.TransactionAutoWithdrawalEnabled {
		params = append(params, request.TransactionAutoWithdrawalEnabled)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " transaction_auto_withdrawal_enabled = $%d", n)
	}
	var requestWindowMin,
		responseWindowMin,
		requestWindowMax,
		responseWindowMax time.Time
	if request.TransactionAutoWithdrawalWindowMin != nil {
		requestWindowMin = *request.TransactionAutoWithdrawalWindowMin
	}
	if response.TransactionAutoWithdrawalWindowMin != nil {
		responseWindowMin = *response.TransactionAutoWithdrawalWindowMin
	}
	if request.TransactionAutoWithdrawalWindowMax != nil {
		requestWindowMax = *request.TransactionAutoWithdrawalWindowMax
	}
	if response.TransactionAutoWithdrawalWindowMax != nil {
		responseWindowMax = *response.TransactionAutoWithdrawalWindowMax
	}
	if !responseWindowMin.Equal(requestWindowMin) ||
		!responseWindowMax.Equal(requestWindowMax) {
		params = append(
			params,
			request.TransactionAutoWithdrawalWindowMin,
			request.TransactionAutoWithdrawalWindowMax,
		)
		if n = len(params); n > 2 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " transaction_auto_withdrawal_window = tstzrange($%d, $%d, '[]')", n-1, n)
	}
	if response.OtpRequiredForLogin != request.OtpRequiredForLogin {
		params = append(params, request.OtpRequiredForLogin)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " otp_required_for_login = $%d", n)
	}
	if response.OwnerID.Compare(request.OwnerID) != 0 {
		params = append(params, request.OwnerID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " owner_id = $%d", n)
	}
	if len(params) == 0 {
		return (&response).Populate(), nil
	}
	params = append(params, request.UpdatedBy, request.UpdatedAt, request.ID)
	n = len(params)
	_, _ = fmt.Fprintf(
		sb,
		`, updated_by = $%d
		, updated_at = $%d
		WHERE id = $%d
		RETURNING id
			, name
			, subdomain
			, billed
			, billing_interval
			, subscription_status
			, subscription_start_date
			, subscription_end_date
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, lower(transaction_auto_withdrawal_window)
			, upper(transaction_auto_withdrawal_window)
			, otp_required_for_login
			, owner_id
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		n-2, n-1, n,
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.Name,
		&response.Subdomain,
		&response.Billed,
		&response.BillingInterval,
		&response.SubscriptionStatus,
		&response.SubscriptionStartDate,
		&response.SubscriptionEndDate,
		&response.VacationMode,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsAmount,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms.Duration,
		&response.MaximumDeliveryTime.Duration,
		&response.TransactionAutoWithdrawalEnabled,
		&response.TransactionAutoWithdrawalWindowMin,
		&response.TransactionAutoWithdrawalWindowMax,
		&response.OtpRequiredForLogin,
		&response.OwnerID,
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
			case "companies_subdomain_key":
				err = models.ErrUniqueCompanySubdomainViolation
			case "companies_whatsapp_phone_number_key":
				err = models.ErrUniqueCompanyWhatsappPhoneNumberViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *companyRepository) AddReferralPointsExpense(
	ctx context.Context,
	tx pgx.Tx,
	companyID uuid.UUID,
	addition int64,
	updatedBy uuid.UUID,
	updatedAt time.Time,
) (*models.Company, error) {
	ctxt := "CompanyRepository-AddReferralPointsExpense"
	var response models.Company
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, name
			, subdomain
			, billed
			, billing_interval
			, subscription_status
			, subscription_start_date
			, subscription_end_date
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, lower(transaction_auto_withdrawal_window)
			, upper(transaction_auto_withdrawal_window)
			, otp_required_for_login
			, owner_id
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason
		FROM companies
		WHERE id = $1
		FOR NO KEY UPDATE`,
		companyID,
	).Scan(
		&response.ID,
		&response.Name,
		&response.Subdomain,
		&response.Billed,
		&response.BillingInterval,
		&response.SubscriptionStatus,
		&response.SubscriptionStartDate,
		&response.SubscriptionEndDate,
		&response.VacationMode,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsAmount,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms.Duration,
		&response.MaximumDeliveryTime.Duration,
		&response.TransactionAutoWithdrawalEnabled,
		&response.TransactionAutoWithdrawalWindowMin,
		&response.TransactionAutoWithdrawalWindowMax,
		&response.OtpRequiredForLogin,
		&response.OwnerID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrCompanyNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	if err = tx.QueryRow(
		ctx,
		`UPDATE companies SET
			referral_points_expense = referral_points_expense + $1
			, updated_by = $2
			, updated_at = $3
		WHERE id = $4
		RETURNING id
			, name
			, subdomain
			, billed
			, billing_interval
			, subscription_status
			, subscription_start_date
			, subscription_end_date
			, vacation_mode
			, vacation_end_date
			, voucher_enabled
			, referral_points_enabled
			, referral_points_minimum_shopping_cost
			, referral_points_amount
			, referral_points_slots
			, referral_points_expense
			, reward_points_accumulation_window_in_days
			, delivery_origin_longitude
			, delivery_origin_latitude
			, delivery_cost_per_km
			, delivery_max_distance_km
			, whatsapp_phone_number
			, payment_terms
			, maximum_delivery_time
			, transaction_auto_withdrawal_enabled
			, lower(transaction_auto_withdrawal_window)
			, upper(transaction_auto_withdrawal_window)
			, otp_required_for_login
			, owner_id
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deactivated_by
			, deactivated_at
			, deactivation_reason`,
		addition,
		updatedBy,
		updatedAt,
		companyID,
	).Scan(
		&response.ID,
		&response.Name,
		&response.Subdomain,
		&response.Billed,
		&response.BillingInterval,
		&response.SubscriptionStatus,
		&response.SubscriptionStartDate,
		&response.SubscriptionEndDate,
		&response.VacationMode,
		&response.VacationEndDate,
		&response.VoucherEnabled,
		&response.ReferralPointsEnabled,
		&response.ReferralPointsMinimumShoppingCost,
		&response.ReferralPointsAmount,
		&response.ReferralPointsSlots,
		&response.ReferralPointsExpense,
		&response.RewardPointsAccumulationWindowInDays,
		&response.DeliveryOriginLongitude,
		&response.DeliveryOriginLatitude,
		&response.DeliveryCostPerKM,
		&response.DeliveryMaxDistanceKM,
		&response.WhatsappPhoneNumber,
		&response.PaymentTerms.Duration,
		&response.MaximumDeliveryTime.Duration,
		&response.TransactionAutoWithdrawalEnabled,
		&response.TransactionAutoWithdrawalWindowMin,
		&response.TransactionAutoWithdrawalWindowMax,
		&response.OtpRequiredForLogin,
		&response.OwnerID,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeactivatedBy,
		&response.DeactivatedAt,
		&response.DeactivationReason,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *companyRepository) buildCondition(request *models.CompanyFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.CompanyIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
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
	if len(request.ExcludedCompanyIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedCompanyID := range request.ExcludedCompanyIDs {
			params = append(params, excludedCompanyID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.Subdomains) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("subdomain IN (")
		for i, subdomain := range request.Subdomains {
			params = append(params, subdomain)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(name) LIKE $%d", len(params))
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
