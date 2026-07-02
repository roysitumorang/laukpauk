package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrAdminNotFound = errors.New("admin not found")
	ErrBuyerNotFound = errors.New("buyer not found")
)

const (
	UserTypeAdmin UserType = "admin"
	UserTypeBuyer UserType = "buyer"

	AdminLevelAdmin      AdminLevel = "admin"
	AdminLevelSuperAdmin AdminLevel = "superadmin"
)

type (
	UserType   string
	AdminLevel string

	User struct {
		RowNo                         int64      `json:"-" form:"-"`
		ID                            uuid.UUID  `json:"-" form:"-"`
		UserType                      UserType   `json:"-" form:"-"`
		Name                          string     `json:"-" form:"-"`
		ConfirmationToken             *string    `json:"-" form:"-"`
		ConfirmedAt                   *time.Time `json:"-" form:"-"`
		UnconfirmedMobilePhone        *string    `json:"-" form:"-"`
		MobilePhoneConfirmationSentAt *time.Time `json:"-" form:"-"`
		MobilePhoneConfirmationToken  *string    `json:"-" form:"-"`
		MobilePhoneConfirmedAt        *time.Time `json:"-" form:"-"`
		MobilePhone                   *string    `json:"-" form:"-"`
		OtpToken                      *string    `json:"-" form:"-"`
		OtpCreatedAt                  *time.Time `json:"-" form:"-"`
		OtpExpiredAt                  *time.Time `json:"-" form:"-"`
		Username                      string     `json:"-" form:"-"`
		UnconfirmedEmail              *string    `json:"-" form:"-"`
		EmailConfirmationSentAt       *time.Time `json:"-" form:"-"`
		EmailConfirmationToken        *string    `json:"-" form:"-"`
		EmailConfirmedAt              *time.Time `json:"-" form:"-"`
		Email                         *string    `json:"-" form:"-"`
		EncryptedPassword             *string    `json:"-" form:"-"`
		PasswordResetAt               *time.Time `json:"-" form:"-"`
		PasswordResetToken            *string    `json:"-" form:"-"`
		PasswordChangedAt             *time.Time `json:"-" form:"-"`
		CreatedBy                     *uuid.UUID `json:"-" form:"-"`
		CreatedAt                     time.Time  `json:"-" form:"-"`
		UpdatedBy                     *uuid.UUID `json:"-" form:"-"`
		UpdatedAt                     time.Time  `json:"-" form:"-"`
		DeactivatedBy                 *uuid.UUID `json:"-" form:"-"`
		DeactivatedAt                 *time.Time `json:"-" form:"-"`
		DeactivationReason            *string    `json:"-" form:"-"`
	}

	Admin struct {
		User
		AdminLevel AdminLevel `json:"-" form:"-"`
		Photo      string     `json:"-" form:"-"`
	}

	Buyer struct {
		User
		ReferralCode string     `json:"-" form:"-"`
		Photo        string     `json:"-" form:"-"`
		Address      string     `json:"-" form:"-"`
		ReferencedBy *uuid.UUID `json:"-" form:"-"`
	}

	UserFilter struct {
		UserIDs   []uuid.UUID
		UserTypes []UserType
		Login,
		Keyword,
		Username,
		Email,
		MobilePhone,
		PasswordResetToken,
		PaginationURL string
		Sortings    []Sorting
		MapSortings *Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewUserFilter() *UserFilter {
	return new(UserFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *UserFilter) WithUserIDs(userIDs ...uuid.UUID) *UserFilter {
	q.UserIDs = userIDs
	return q
}

func (q *UserFilter) WithUserTypes(userTypes ...UserType) *UserFilter {
	q.UserTypes = userTypes
	return q
}

func (q *UserFilter) WithLogin(login string) *UserFilter {
	q.Login = login
	return q
}

func (q *UserFilter) WithKeyword(keyword string) *UserFilter {
	q.Keyword = keyword
	return q
}

func (q *UserFilter) WithUsername(username string) *UserFilter {
	q.Username = username
	return q
}

func (q *UserFilter) WithEmail(email string) *UserFilter {
	q.Email = email
	return q
}

func (q *UserFilter) WithMobilePhone(mobilePhone string) *UserFilter {
	q.MobilePhone = mobilePhone
	return q
}

func (q *UserFilter) WithPasswordResetToken(passwordResetToken string) *UserFilter {
	q.PasswordResetToken = passwordResetToken
	return q
}

func (q *UserFilter) WithPaginationURL(paginationURL string) *UserFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *UserFilter) WithSortings(sortings ...Sorting) *UserFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *UserFilter) WithLimit(limit int64) *UserFilter {
	q.Limit = limit
	return q
}

func (q *UserFilter) WithPage(page int64) *UserFilter {
	q.Page = page
	return q
}

func (q *UserFilter) WithUrlValues(urlValues url.Values) *UserFilter {
	q.UrlValues = urlValues
	return q
}
