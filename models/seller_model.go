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

var ErrSellerNotFound = errors.New("seller not found")
var ErrUniqueSellerUsernameViolation = errors.New("username: already exists")
var ErrUniqueSellerMobilePhoneViolation = errors.New("mobile_phone: already exists")
var ErrUniqueSellerEmailViolation = errors.New("email: already exists")
var ErrSellerSelfDeletingProhibited = errors.New("self deleting prohibited")
var MapSellerSortings = map[string]struct{}{
	SortingFieldID:         {},
	SellerSortingFieldName: {},
}

const SellerSortingFieldName = "name"

type SellerStatus int

const (
	SellerStatusActive SellerStatus = iota
	SellerStatusDeactivated
)

type SellerLevel int

const (
	SellerLevelOwner SellerLevel = iota
	SellerLevelStaff
)

var SellerLevels = []SellerLevel{SellerLevelOwner, SellerLevelStaff}
var MapSellerLevels = map[SellerLevel]struct{}{
	SellerLevelOwner: {},
	SellerLevelStaff: {},
}

func (q SellerLevel) Equal(v SellerLevel) bool {
	return q.Int() == v.Int()
}

func (q SellerLevel) String() string {
	if q == SellerLevelStaff {
		return "staff"
	}
	return "owner"
}

func (q SellerLevel) Int() int {
	return int(q)
}

type Seller struct {
	RowNo                         int64        `json:"-" form:"-"`
	ID                            uuid.UUID    `json:"id" form:"-"`
	CompanyID                     *uuid.UUID   `json:"-" form:"-"`
	Status                        SellerStatus `json:"-" form:"-"`
	SellerLevel                   SellerLevel  `json:"-" form:"seller_level"`
	Name                          string       `json:"name" form:"name"`
	UnconfirmedMobilePhone        *string      `json:"-" form:"-"`
	MobilePhoneConfirmationSentAt *time.Time   `json:"-" form:"-"`
	MobilePhoneConfirmationToken  *string      `json:"-" form:"-"`
	MobilePhoneConfirmedAt        *time.Time   `json:"-" form:"-"`
	MobilePhone                   *string      `json:"mobile_phone" form:"mobile_phone"`
	OtpHash                       string       `json:"-" form:"-"`
	OtpExpiredAt                  time.Time    `json:"-" form:"-"`
	Username                      string       `json:"username" form:"username"`
	UnconfirmedEmail              *string      `json:"-" form:"-"`
	EmailConfirmationSentAt       *time.Time   `json:"-" form:"-"`
	EmailConfirmationToken        *string      `json:"-" form:"-"`
	EmailConfirmedAt              *time.Time   `json:"-" form:"-"`
	Email                         *string      `json:"email" form:"email"`
	Photo                         string       `json:"-" form:"photo"`
	UnreadNotifications           int64        `json:"-" form:"-"`
	TelegramChatID                *int64       `json:"telegram_chat_id" form:"telegram_chat_id"`
	TelegramUsername              *string      `json:"telegram_username" form:"telegram_username"`
	CreatedBy                     *uuid.UUID   `json:"-" form:"-"`
	CreatedAt                     time.Time    `json:"created_at" form:"-"`
	UpdatedBy                     *uuid.UUID   `json:"-" form:"-"`
	UpdatedAt                     time.Time    `json:"-" form:"-"`
	DeactivatedBy                 *uuid.UUID   `json:"-" form:"-"`
	DeactivatedAt                 *time.Time   `json:"-" form:"-"`
	DeactivationReason            *string      `json:"-" form:"-"`
}

func (q *Seller) Validate() error {
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
	if _, ok := MapSellerLevels[q.SellerLevel]; !ok {
		return errors.New("seller_level: not found")
	}
	return nil
}

func (q *Seller) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Seller) SetPhoto(photo string) *Seller {
	q.Photo = photo
	return q
}

func (q *Seller) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

func (q *Seller) Owner() bool {
	return q.SellerLevel == SellerLevelOwner
}

type SellerFilter struct {
	CompanyIDs,
	SellerIDs,
	ExcludedSellerIDs []uuid.UUID
	TelegramChatIDs []int64
	Statuses        []SellerStatus
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

func NewSellerFilter() *SellerFilter {
	return &SellerFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *SellerFilter) WithSellerIDs(sellerIDs ...uuid.UUID) *SellerFilter {
	q.SellerIDs = sellerIDs
	return q
}

func (q *SellerFilter) WithExcludedSellerIDs(excludedSellerIDs ...uuid.UUID) *SellerFilter {
	q.ExcludedSellerIDs = excludedSellerIDs
	return q
}

func (q *SellerFilter) WithStatuses(statuses ...SellerStatus) *SellerFilter {
	q.Statuses = statuses
	return q
}

func (q *SellerFilter) WithLogin(login string) *SellerFilter {
	q.Login = login
	return q
}

func (q *SellerFilter) WithKeyword(keyword string) *SellerFilter {
	q.Keyword = keyword
	return q
}

func (q *SellerFilter) WithUsername(username string) *SellerFilter {
	q.Username = username
	return q
}

func (q *SellerFilter) WithEmail(email string) *SellerFilter {
	q.Email = email
	return q
}

func (q *SellerFilter) WithMobilePhone(mobilePhone string) *SellerFilter {
	q.MobilePhone = mobilePhone
	return q
}

func (q *SellerFilter) WithPaginationURL(paginationURL string) *SellerFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *SellerFilter) WithSortings(sortings ...Sorting) *SellerFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *SellerFilter) WithLimit(limit int64) *SellerFilter {
	q.Limit = limit
	return q
}

func (q *SellerFilter) WithPage(page int64) *SellerFilter {
	q.Page = page
	return q
}

func (q *SellerFilter) WithUrlValues(urlValues url.Values) *SellerFilter {
	q.UrlValues = urlValues
	return q
}
