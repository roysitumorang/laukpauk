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

var ErrAdminNotFound = errors.New("admin not found")
var ErrUniqueAdminUsernameViolation = errors.New("username: already exists")
var ErrUniqueAdminMobilePhoneViolation = errors.New("mobile_phone: already exists")
var ErrUniqueAdminEmailViolation = errors.New("email: already exists")
var ErrAdminSelfDeletingProhibited = errors.New("self deleting prohibited")
var MapAdminSortings = map[string]struct{}{
	SortingFieldID:        {},
	AdminSortingFieldName: {},
}

const AdminSortingFieldName = "name"

type AdminLevel int

const (
	AdminLevelSuperAdmin AdminLevel = iota
	AdminLevelAdmin
)

func (q AdminLevel) Equal(v AdminLevel) bool {
	return q.Int() == v.Int()
}

func (q AdminLevel) String() string {
	if q == AdminLevelSuperAdmin {
		return "super admin"
	}
	return "admin"
}

func (q AdminLevel) Int() int {
	return int(q)
}

var AdminLevels = []AdminLevel{AdminLevelAdmin, AdminLevelSuperAdmin}
var MapAdminLevels = map[AdminLevel]struct{}{
	AdminLevelAdmin:      {},
	AdminLevelSuperAdmin: {},
}

type AdminStatus int

const (
	AdminStatusActive AdminStatus = iota
	AdminStatusDeactivated
)

type Admin struct {
	RowNo                         int64       `json:"-" form:"-"`
	ID                            uuid.UUID   `json:"id" form:"-"`
	Status                        AdminStatus `json:"-" form:"-"`
	AdminLevel                    AdminLevel  `json:"-" form:"admin_level"`
	Name                          string      `json:"name" form:"name"`
	UnconfirmedMobilePhone        *string     `json:"-" form:"-"`
	MobilePhoneConfirmationSentAt *time.Time  `json:"-" form:"-"`
	MobilePhoneConfirmationToken  *string     `json:"-" form:"-"`
	MobilePhoneConfirmedAt        *time.Time  `json:"-" form:"-"`
	MobilePhone                   string      `json:"mobile_phone" form:"mobile_phone"`
	Username                      string      `json:"username" form:"username"`
	UnconfirmedEmail              *string     `json:"-" form:"-"`
	EmailConfirmationSentAt       *time.Time  `json:"-" form:"-"`
	EmailConfirmationToken        *string     `json:"-" form:"-"`
	EmailConfirmedAt              *time.Time  `json:"-" form:"-"`
	Email                         string      `json:"email" form:"email"`
	PasswordIncluded              bool        `json:"-" form:"password_included"`
	Base64Password                string      `json:"-" form:"password"`
	Base64PasswordConfirmation    string      `json:"-" form:"password_confirmation"`
	PlaintextPassword             string      `json:"-" form:"-"`
	EncryptedPassword             *string     `json:"-" form:"-"`
	PasswordResetAt               *time.Time  `json:"-" form:"-"`
	PasswordResetToken            *string     `json:"-" form:"-"`
	PasswordChangedAt             *time.Time  `json:"-" form:"-"`
	Photo                         string      `json:"-" form:"photo"`
	UnreadNotifications           int64       `json:"-" form:"-"`
	CreatedBy                     *uuid.UUID  `json:"-" form:"-"`
	CreatedAt                     time.Time   `json:"created_at" form:"-"`
	UpdatedBy                     *uuid.UUID  `json:"-" form:"-"`
	UpdatedAt                     time.Time   `json:"-" form:"-"`
	DeactivatedBy                 *uuid.UUID  `json:"-" form:"-"`
	DeactivatedAt                 *time.Time  `json:"-" form:"-"`
	DeactivationReason            *string     `json:"-" form:"-"`
}

func (q *Admin) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.Username = strings.TrimSpace(strings.ToLower(q.Username)); q.Username == "" {
		return errors.New("username: is required")
	}
	if q.MobilePhone = strings.TrimSpace(q.MobilePhone); q.MobilePhone == "" {
		return errors.New("mobile_phone: is required")
	}
	mobilePhone, err := phonenumbers.Parse(q.MobilePhone, "ID")
	if err != nil {
		return fmt.Errorf("mobile_phone: %w", err)
	}
	q.MobilePhone = phonenumbers.Format(mobilePhone, phonenumbers.E164)
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
	if _, ok := MapAdminLevels[q.AdminLevel]; !ok {
		return errors.New("admin_level: not found")
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

func (q *Admin) SuperAdmin() bool {
	return q.AdminLevel == AdminLevelSuperAdmin
}

type AdminFilter struct {
	AdminIDs,
	ExcludedAdminIDs []uuid.UUID
	Statuses []AdminStatus
	Login,
	Keyword,
	Username,
	Email,
	MobilePhone,
	PasswordResetToken,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewAdminFilter() *AdminFilter {
	return &AdminFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *AdminFilter) WithAdminIDs(adminIDs ...uuid.UUID) *AdminFilter {
	q.AdminIDs = adminIDs
	return q
}

func (q *AdminFilter) WithExcludedAdminIDs(excludedAdminIDs ...uuid.UUID) *AdminFilter {
	q.ExcludedAdminIDs = excludedAdminIDs
	return q
}

func (q *AdminFilter) WithStatuses(statuses ...AdminStatus) *AdminFilter {
	q.Statuses = statuses
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
