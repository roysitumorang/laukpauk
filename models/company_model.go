package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/nyaruka/phonenumbers"
	"github.com/roysitumorang/laukpauk/helper"
)

var ErrCompanyNotFound = errors.New("company not found")
var ErrUniqueCompanySubdomainViolation = errors.New("subdomain: already exists")
var ErrUniqueCompanyTelegramBotTokenViolation = errors.New("telegram_bot_token: already exists")
var ErrUniqueCompanyWhatsappPhoneNumberViolation = errors.New("whatsapp_phone_number: already exists")
var MapCompanySortings = map[string]struct{}{
	SortingFieldID:          {},
	CompanySortingFieldName: {},
}

const CompanySortingFieldName = "name"

type SubscriptionStatus uint8

const (
	SubscriptionStatusTrial SubscriptionStatus = iota
	SubscriptionStatusExpired
	SubscriptionStatusActive
	SubscriptionStatusCancelled
)

func (q SubscriptionStatus) String() string {
	if q == SubscriptionStatusExpired {
		return "expired"
	}
	if q == SubscriptionStatusActive {
		return "active"
	}
	if q == SubscriptionStatusCancelled {
		return "cancelled"
	}
	return "trial"
}

type BillingInterval uint8

const (
	BillingIntervalMonthly BillingInterval = iota
	BillingIntervalAnnual
)

func (q *BillingInterval) String() *string {
	if *q == BillingIntervalAnnual {
		return new("annual")
	}
	return new("monthly")
}

type Company struct {
	RowNo                                 int64              `json:"-" form:"-"`
	ID                                    uuid.UUID          `json:"-" form:"-"`
	Name                                  string             `json:"-" form:"name"`
	Android                               bool               `json:"-" form:"android"`
	Subdomain                             string             `json:"-" form:"subdomain"`
	Billed                                bool               `json:"-" form:"billed"`
	BillingInterval                       *BillingInterval   `json:"-" form:"billing_interval"`
	RawBillingInterval                    *string            `json:"billing_interval" form:"-"`
	SubscriptionStatus                    SubscriptionStatus `json:"-" form:"subscription_status"`
	RawSubscriptionStatus                 string             `json:"subscription_status" form:"-"`
	SubscriptionStartDate                 *time.Time         `json:"-" form:"subscription_start_date"`
	SubscriptionEndDate                   *time.Time         `json:"-" form:"subscription_end_date"`
	TelegramBotToken                      *string            `json:"-" form:"telegram_bot_token"`
	VacationMode                          bool               `json:"vacation_mode" form:"vacation_mode"`
	VacationEndDate                       *time.Time         `json:"vacation_end_date" form:"-"`
	RawVacationEndDate                    string             `json:"-" form:"vacation_end_date"`
	VoucherEnabled                        bool               `json:"voucher_enabled" form:"voucher_enabled"`
	ReferralPointsEnabled                 bool               `json:"referral_points_enabled" form:"referral_points_enabled"`
	ReferralPointsMinimumShoppingCost     int64              `json:"referral_points_minimum_shopping_cost" form:"referral_points_minimum_shopping_cost"`
	ReferralPointsPerReferee              int64              `json:"referral_points_per_referee" form:"referral_points_per_referee"`
	ReferralPointsSlots                   []int64            `json:"referral_points_slots" form:"referral_points_slots"`
	ReferralPointsExpense                 int64              `json:"referral_points_expense" form:"referral_points_expense"`
	RewardPointsAccumulationWindowInDays  int16              `json:"reward_points_accumulation_window_in_days" form:"reward_points_accumulation_window_in_days"`
	DeliveryOriginLongitude               decimal.Decimal    `json:"delivery_origin_longitude" form:"delivery_origin_longitude"`
	DeliveryOriginLatitude                decimal.Decimal    `json:"delivery_origin_latitude" form:"delivery_origin_latitude"`
	DeliveryCostPerKM                     int64              `json:"delivery_cost_per_km" form:"delivery_cost_per_km"`
	DeliveryMaxDistanceKM                 int64              `json:"delivery_max_distance_km" form:"delivery_max_distance_km"`
	WhatsappPhoneNumber                   string             `json:"whatsapp_phone_number" form:"whatsapp_phone_number"`
	PaymentTerms                          Duration           `json:"payment_terms" form:""`
	RawPaymentTerms                       string             `json:"-" form:"payment_terms"`
	MaximumDeliveryTime                   Duration           `json:"maximum_delivery_time" form:"-"`
	RawMaximumDeliveryTime                string             `json:"-" form:"maximum_delivery_time"`
	TransactionAutoWithdrawalEnabled      bool               `json:"-" form:"transaction_auto_withdrawal_enabled"`
	TransactionAutoWithdrawalWindowMin    *time.Time         `json:"transaction_auto_withdrawal_window_min" form:"-"`
	RawTransactionAutoWithdrawalWindowMin *string            `json:"-" form:"transaction_auto_withdrawal_window_min"`
	TransactionAutoWithdrawalWindowMax    *time.Time         `json:"transaction_auto_withdrawal_window_max" form:"-"`
	RawTransactionAutoWithdrawalWindowMax *string            `json:"-" form:"transaction_auto_withdrawal_window_max"`
	OwnerID                               uuid.UUID          `json:"-" form:"-"`
	CreatedBy                             uuid.UUID          `json:"-" form:"-"`
	CreatedAt                             time.Time          `json:"-" form:"-"`
	UpdatedBy                             uuid.UUID          `json:"-" form:"-"`
	UpdatedAt                             time.Time          `json:"-" form:"-"`
	DeactivatedBy                         *uuid.UUID         `json:"-" form:"-"`
	DeactivatedAt                         *time.Time         `json:"-" form:"-"`
	DeactivationReason                    *string            `json:"-" form:"-"`
}

func (q *Company) Validate(includeSetting ...bool) error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.Subdomain = strings.ToLower(strings.TrimSpace(q.Subdomain)); q.Subdomain == "" {
		return errors.New("slug: is required")
	}
	if q.WhatsappPhoneNumber = strings.TrimSpace(q.WhatsappPhoneNumber); q.WhatsappPhoneNumber == "" {
		return errors.New("whatsapp_phone_number: is required")
	}
	whatsappPhoneNumber, err := phonenumbers.Parse(q.WhatsappPhoneNumber, "ID")
	if err != nil {
		return fmt.Errorf("whatsapp_phone_number: %w", err)
	}
	q.WhatsappPhoneNumber = phonenumbers.Format(whatsappPhoneNumber, phonenumbers.E164)
	if q.TelegramBotToken != nil {
		if *q.TelegramBotToken = strings.TrimSpace(*q.TelegramBotToken); *q.TelegramBotToken == "" {
			q.TelegramBotToken = nil
		}
	}
	if len(includeSetting) == 0 || !includeSetting[0] {
		return nil
	}
	now := time.Now().In(time.Local)
	if q.VacationMode {
		if q.RawVacationEndDate = strings.TrimSpace(q.RawVacationEndDate); q.RawVacationEndDate == "" {
			return errors.New("vacation_end_date: is required")
		}
		vacationEndDate, err := time.Parse(time.RFC3339, helper.Sprintf("%sT23:59:59%s", q.RawVacationEndDate, now.Format("Z07:00")))
		if err != nil {
			return fmt.Errorf("vacation_end_date: should use format '%s'", time.DateOnly)
		}
		q.VacationEndDate = &vacationEndDate
		if q.VacationEndDate.Before(now) {
			return errors.New("vacation_end_date: should equal or exceed vacation_start_date")
		}
	} else {
		q.VacationEndDate = nil
	}
	if q.RawPaymentTerms = strings.TrimSpace(q.RawPaymentTerms); q.RawPaymentTerms == "" {
		return errors.New("payment_terms: is required")
	}
	if q.PaymentTerms.Duration, err = time.ParseDuration(q.RawPaymentTerms); err != nil || q.PaymentTerms.Duration < 1 {
		return errors.New("payment_terms: is invalid")
	}
	if q.RawMaximumDeliveryTime = strings.TrimSpace(q.RawMaximumDeliveryTime); q.RawMaximumDeliveryTime == "" {
		return errors.New("maximum_delivery_time: is required")
	}
	if q.MaximumDeliveryTime.Duration, err = time.ParseDuration(q.RawMaximumDeliveryTime); err != nil || q.MaximumDeliveryTime.Duration < 1 {
		return errors.New("maximum_delivery_time: is invalid")
	}
	if q.TransactionAutoWithdrawalEnabled {
		if q.TransactionAutoWithdrawalWindowMin == nil {
			return errors.New("transaction_auto_withdrawal_window_min: is required")
		}
		if q.TransactionAutoWithdrawalWindowMax == nil {
			return errors.New("transaction_auto_withdrawal_window_max: is required")
		}
		if !q.TransactionAutoWithdrawalWindowMax.After(*q.TransactionAutoWithdrawalWindowMin) {
			return errors.New("transaction_auto_withdrawal_window_max: should exceed transaction_auto_withdrawal_window_min")
		}
		q.TransactionAutoWithdrawalWindowMin = new(time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.TransactionAutoWithdrawalWindowMin.Hour(),
			q.TransactionAutoWithdrawalWindowMin.Minute(),
			q.TransactionAutoWithdrawalWindowMin.Second(),
			q.TransactionAutoWithdrawalWindowMin.Nanosecond(),
			q.TransactionAutoWithdrawalWindowMin.Location(),
		))
		q.TransactionAutoWithdrawalWindowMax = new(time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.TransactionAutoWithdrawalWindowMax.Hour(),
			q.TransactionAutoWithdrawalWindowMax.Minute(),
			q.TransactionAutoWithdrawalWindowMax.Second(),
			q.TransactionAutoWithdrawalWindowMax.Nanosecond(),
			q.TransactionAutoWithdrawalWindowMax.Location(),
		))
	} else {
		q.TransactionAutoWithdrawalWindowMin = nil
		q.TransactionAutoWithdrawalWindowMax = nil
	}
	return nil
}

func (q *Company) Populate() *Company {
	if q.BillingInterval != nil {
		q.RawBillingInterval = q.BillingInterval.String()
	}
	q.RawSubscriptionStatus = q.SubscriptionStatus.String()
	if q.VacationEndDate != nil {
		q.RawVacationEndDate = q.VacationEndDate.In(time.Local).Format(time.DateOnly)
	}
	q.RawPaymentTerms = q.PaymentTerms.String()
	q.RawMaximumDeliveryTime = q.MaximumDeliveryTime.String()
	if q.TransactionAutoWithdrawalWindowMin != nil {
		q.RawTransactionAutoWithdrawalWindowMin = new(q.TransactionAutoWithdrawalWindowMin.
			In(time.Local).
			Format(time.TimeOnly))
	}
	if q.TransactionAutoWithdrawalWindowMax != nil {
		q.RawTransactionAutoWithdrawalWindowMax = new(q.TransactionAutoWithdrawalWindowMax.
			In(time.Local).
			Format(time.TimeOnly))
	}
	return q
}

func (q *Company) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type CompanyFilter struct {
	CompanyIDs,
	ExcludedCompanyIDs []uuid.UUID
	Subdomains []string
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewCompanyFilter() *CompanyFilter {
	return &CompanyFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *CompanyFilter) WithCompanyIDs(companyIDs ...uuid.UUID) *CompanyFilter {
	q.CompanyIDs = companyIDs
	return q
}

func (q *CompanyFilter) WithExcludedCompanyIDs(excludedCompanyIDs ...uuid.UUID) *CompanyFilter {
	q.ExcludedCompanyIDs = excludedCompanyIDs
	return q
}

func (q *CompanyFilter) WithSubdomains(subdomains ...string) *CompanyFilter {
	q.Subdomains = subdomains
	return q
}

func (q *CompanyFilter) WithKeyword(keyword string) *CompanyFilter {
	q.Keyword = keyword
	return q
}

func (q *CompanyFilter) WithPaginationURL(paginationURL string) *CompanyFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *CompanyFilter) WithSortings(sortings ...Sorting) *CompanyFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *CompanyFilter) WithLimit(limit int64) *CompanyFilter {
	q.Limit = limit
	return q
}

func (q *CompanyFilter) WithPage(page int64) *CompanyFilter {
	q.Page = page
	return q
}

func (q *CompanyFilter) WithUrlValues(urlValues url.Values) *CompanyFilter {
	q.UrlValues = urlValues
	return q
}
