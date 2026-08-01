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

var ErrAdminNotFound = errors.New("admin not found")
var ErrUniqueAdminUsernameViolation = errors.New("username: already exists")
var ErrUniqueAdminMobilePhoneViolation = errors.New("mobile_phone: already exists")
var ErrUniqueAdminEmailViolation = errors.New("email: already exists")
var MapAdminSortings = map[string]struct{}{
	SortingFieldID:        {},
	AdminSortingFieldName: {},
}

const AdminSortingFieldName = "name"

const (
	AdminTypeAdmin AdminType = iota
	AdminTypeSuperAdmin
)

type AdminType int

var AdminTypes = []AdminType{AdminTypeAdmin, AdminTypeSuperAdmin}
var MapAdminTypes = map[AdminType]struct{}{
	AdminTypeAdmin:      {},
	AdminTypeSuperAdmin: {},
}

type Admin struct {
	RowNo  int64     `json:"-" form:"-"`
	ID     uuid.UUID `json:"id" form:"-"`
	UserID uuid.UUID `json:"-" form:"-"`
	user
	AdminType AdminType `json:"-" form:"admin_type"`
}

func (q *Admin) Validate() error {
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
	if _, ok := MapAdminTypes[q.AdminType]; !ok {
		return errors.New("admin_type: not found")
	}
	return nil
}

func (q *Admin) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Admin) SetPhoto(photo string) *Admin {
	q.Photo = photo
	return q
}

func (q *Admin) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

type AdminFilter struct {
	userFilter
}

func NewAdminFilter() *AdminFilter {
	return &AdminFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *AdminFilter) WithAdminIDs(adminIDs ...uuid.UUID) *AdminFilter {
	q.UserIDs = adminIDs
	return q
}

func (q *AdminFilter) WithExcludedAdminIDs(excludedAdminIDs ...uuid.UUID) *AdminFilter {
	q.ExcludedUserIDs = excludedAdminIDs
	return q
}

func (q *AdminFilter) WithUserStatuses(userStatuses ...UserStatus) *AdminFilter {
	q.UserStatuses = userStatuses
	return q
}

func (q *AdminFilter) WithLogin(login string) *AdminFilter {
	q.Login = login
	return q
}

func (q *AdminFilter) WithKeyword(keyword string) *AdminFilter {
	q.Keyword = keyword
	return q
}

func (q *AdminFilter) WithUsername(username string) *AdminFilter {
	q.Username = username
	return q
}

func (q *AdminFilter) WithEmail(email string) *AdminFilter {
	q.Email = email
	return q
}

func (q *AdminFilter) WithMobilePhone(mobilePhone string) *AdminFilter {
	q.MobilePhone = mobilePhone
	return q
}

func (q *AdminFilter) WithPasswordResetToken(passwordResetToken string) *AdminFilter {
	q.PasswordResetToken = passwordResetToken
	return q
}

func (q *AdminFilter) WithPaginationURL(paginationURL string) *AdminFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *AdminFilter) WithSortings(sortings ...Sorting) *AdminFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *AdminFilter) WithLimit(limit int64) *AdminFilter {
	q.Limit = limit
	return q
}

func (q *AdminFilter) WithPage(page int64) *AdminFilter {
	q.Page = page
	return q
}

func (q *AdminFilter) WithUrlValues(urlValues url.Values) *AdminFilter {
	q.UrlValues = urlValues
	return q
}

func (q AdminType) Equal(v AdminType) bool {
	return q.Int() == v.Int()
}

func (q AdminType) String() string {
	if q == AdminTypeSuperAdmin {
		return "super admin"
	}
	return "admin"
}

func (q AdminType) Int() int {
	return int(q)
}
