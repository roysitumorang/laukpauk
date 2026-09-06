package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/nyaruka/phonenumbers"
	"github.com/roysitumorang/laukpauk/helper"
)

var ErrBuyerNotFound = errors.New("buyer not found")
var ErrUniqueBuyerUsernameViolation = errors.New("username: already exists")
var ErrUniqueBuyerMobilePhoneViolation = errors.New("mobile_phone: already exists")
var ErrUniqueBuyerEmailViolation = errors.New("email: already exists")
var ErrUniqueTelegramChatIdViolation = errors.New("telegram_chat_id: already exists")
var ErrUniqueTelegramUsernameViolation = errors.New("telegram_username: already exists")
var ErrUniqueBuyerReferralCodeViolation = errors.New("referral_code: already exists")
var MapBuyerSortings = map[string]struct{}{
	SortingFieldID:        {},
	BuyerSortingFieldName: {},
}

const BuyerSortingFieldName = "name"

type BuyerStatus int

const (
	BuyerStatusActive BuyerStatus = iota
	BuyerStatusDeactivated
)

type Buyer struct {
	RowNo                         int64       `json:"-" form:"-"`
	ID                            uuid.UUID   `json:"id" form:"id"`
	CompanyID                     uuid.UUID   `json:"-" form:"-"`
	Status                        BuyerStatus `json:"-" form:"-"`
	Name                          string      `json:"name" form:"name"`
	UnconfirmedMobilePhone        *string     `json:"-" form:"-"`
	MobilePhoneConfirmationSentAt *time.Time  `json:"-" form:"-"`
	MobilePhoneConfirmationToken  *string     `json:"-" form:"-"`
	MobilePhoneConfirmedAt        *time.Time  `json:"-" form:"-"`
	MobilePhone                   *string     `json:"mobile_phone" form:"mobile_phone"`
	OtpHash                       string      `json:"-" form:"-"`
	OtpExpiredAt                  time.Time   `json:"-" form:"-"`
	Username                      string      `json:"username" form:"username"`
	UnconfirmedEmail              *string     `json:"-" form:"-"`
	EmailConfirmationSentAt       *time.Time  `json:"-" form:"-"`
	EmailConfirmationToken        *string     `json:"-" form:"-"`
	EmailConfirmedAt              *time.Time  `json:"-" form:"-"`
	Email                         *string     `json:"email" form:"email"`
	Photo                         string      `json:"-" form:"photo"`
	UnreadNotifications           int64       `json:"-" form:"-"`
	ReferralCode                  string      `json:"referral_code" form:"referral_code"`
	Address                       string      `json:"address" form:"address"`
	ReferrerID                    *uuid.UUID  `json:"referrer_id" form:"referrer_id"`
	ReferralPoints                int64       `json:"referral_points" form:"-"`
	RewardPoints                  int64       `json:"reward_points" form:"-"`
	Wallet                        int64       `json:"wallet" form:"-"`
	TransactionVolume             int64       `json:"transaction_volume" form:"-"`
	DeliveryRouteID               *uuid.UUID  `json:"delivery_route_id" form:"-"`
	TelegramChatID                *int64      `json:"telegram_chat_id" form:"telegram_chat_id"`
	TelegramUsername              *string     `json:"telegram_username" form:"telegram_username"`
	CreatedAt                     time.Time   `json:"created_at" form:"-"`
	UpdatedAt                     time.Time   `json:"-" form:"-"`
	DeactivatedBy                 *uuid.UUID  `json:"-" form:"-"`
	DeactivatedAt                 *time.Time  `json:"-" form:"-"`
	DeactivationReason            *string     `json:"-" form:"-"`
}

func (q *Buyer) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.Username = strings.TrimSpace(strings.ToLower(q.Username)); q.Username == "" {
		return errors.New("username: is required")
	}
	if q.MobilePhone == nil || *q.MobilePhone == "" {
		return errors.New("mobile_phone: is required")
	}
	mobilePhone, err := phonenumbers.Parse(*q.MobilePhone, "ID")
	if err != nil {
		return fmt.Errorf("mobile_phone: %w", err)
	}
	q.MobilePhone = new(phonenumbers.Format(mobilePhone, phonenumbers.E164))
	if q.Address = strings.TrimSpace(q.Address); q.Address == "" {
		return errors.New("address: is required")
	}
	if q.ReferrerID != nil && q.ReferrerID.Compare(uuid.Nil()) == 0 {
		q.ReferrerID = nil
	}
	return nil
}

func (q *Buyer) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Buyer) SetPhoto(photo string) *Buyer {
	q.Photo = photo
	return q
}

func (q *Buyer) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

type BuyerFilter struct {
	CompanyIDs,
	BuyerIDs,
	ExcludedBuyerIDs []uuid.UUID
	TelegramChatIDs []int64
	ReferralCodes   []string
	Statuses        []BuyerStatus
	Login,
	Keyword,
	Username,
	TelegramUsername,
	Email,
	MobilePhone,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewBuyerFilter() *BuyerFilter {
	return &BuyerFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *BuyerFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *BuyerFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *BuyerFilter) WithExcludedBuyerIDs(excludedBuyerIDs ...uuid.UUID) *BuyerFilter {
	q.ExcludedBuyerIDs = excludedBuyerIDs
	return q
}

func (q *BuyerFilter) WithCompanyIDs(companyIDs ...uuid.UUID) *BuyerFilter {
	q.CompanyIDs = companyIDs
	return q
}

func (q *BuyerFilter) WithTelegramChatIDs(telegramChatIDs ...int64) *BuyerFilter {
	q.TelegramChatIDs = telegramChatIDs
	return q
}

func (q *BuyerFilter) WithStatuses(statuses ...BuyerStatus) *BuyerFilter {
	q.Statuses = statuses
	return q
}

func (q *BuyerFilter) WithLogin(login string) *BuyerFilter {
	q.Login = login
	return q
}

func (q *BuyerFilter) WithKeyword(keyword string) *BuyerFilter {
	q.Keyword = keyword
	return q
}

func (q *BuyerFilter) WithUsername(username string) *BuyerFilter {
	q.Username = username
	return q
}

func (q *BuyerFilter) WithTelegramUsername(telegramUsername string) *BuyerFilter {
	q.TelegramUsername = telegramUsername
	return q
}

func (q *BuyerFilter) WithEmail(email string) *BuyerFilter {
	q.Email = email
	return q
}

func (q *BuyerFilter) WithMobilePhone(mobilePhone string) *BuyerFilter {
	q.MobilePhone = mobilePhone
	return q
}

func (q *BuyerFilter) WithReferralCodes(referralCodes ...string) *BuyerFilter {
	q.ReferralCodes = referralCodes
	return q
}

func (q *BuyerFilter) WithPaginationURL(paginationURL string) *BuyerFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *BuyerFilter) WithSortings(sortings ...Sorting) *BuyerFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *BuyerFilter) WithLimit(limit int64) *BuyerFilter {
	q.Limit = limit
	return q
}

func (q *BuyerFilter) WithPage(page int64) *BuyerFilter {
	q.Page = page
	return q
}

func (q *BuyerFilter) WithUrlValues(urlValues url.Values) *BuyerFilter {
	q.UrlValues = urlValues
	return q
}
