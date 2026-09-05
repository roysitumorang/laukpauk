package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
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

const (
	SellerLevelOwner SellerLevel = iota
	SellerLevelStaff
)

type SellerLevel int

var SellerLevels = []SellerLevel{SellerLevelOwner, SellerLevelStaff}
var MapSellerLevels = map[SellerLevel]struct{}{
	SellerLevelOwner: {},
	SellerLevelStaff: {},
}

type Seller struct {
	RowNo int64     `json:"-" form:"-"`
	ID    uuid.UUID `json:"id" form:"-"`
	user
	SellerLevel SellerLevel `json:"-" form:"seller_level"`
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
	if q.PasswordIncluded {
		if q.Base64Password = strings.TrimSpace(q.Base64Password); q.Base64Password == "" {
			return errors.New("password: is required")
		}
		if q.Base64PasswordConfirmation = strings.TrimSpace(q.Base64PasswordConfirmation); q.Base64PasswordConfirmation == "" {
			return errors.New("password_confirmation: is required")
		}
		if q.Base64Password != q.Base64PasswordConfirmation {
			return errors.New("password: does not match password_confirmation")
		}
		if q.PlaintextPassword, err = helper.Base64Decode(q.Base64Password); err != nil {
			return fmt.Errorf("password: %s", err.Error())
		}
	}
	if _, ok := MapSellerLevels[q.SellerLevel]; !ok {
		return errors.New("seller_type: not found")
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
	userFilter
}

func NewSellerFilter() *SellerFilter {
	return &SellerFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *SellerFilter) WithSellerIDs(sellerIDs ...uuid.UUID) *SellerFilter {
	q.UserIDs = sellerIDs
	return q
}

func (q *SellerFilter) WithExcludedSellerIDs(excludedSellerIDs ...uuid.UUID) *SellerFilter {
	q.ExcludedUserIDs = excludedSellerIDs
	return q
}

func (q *SellerFilter) WithUserStatuses(userStatuses ...UserStatus) *SellerFilter {
	q.UserStatuses = userStatuses
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

func (q *SellerFilter) WithPasswordResetToken(passwordResetToken string) *SellerFilter {
	q.PasswordResetToken = passwordResetToken
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
