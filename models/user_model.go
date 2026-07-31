package models

import (
	"net/url"
	"time"
	"uuid"
)

const (
	UserTypeAdmin UserType = iota
	UserTypeBuyer
)

const (
	UserStatusActive UserStatus = iota
	UserStatusDeactivated
)

type (
	UserType   int
	UserStatus int

	user struct {
		UserType                      UserType   `json:"-" form:"-"`
		UserStatus                    UserStatus `json:"-" form:"-"`
		Name                          string     `json:"name" form:"-"`
		ConfirmationToken             *string    `json:"-" form:"-"`
		ConfirmedAt                   *time.Time `json:"-" form:"-"`
		UnconfirmedMobilePhone        *string    `json:"-" form:"-"`
		MobilePhoneConfirmationSentAt *time.Time `json:"-" form:"-"`
		MobilePhoneConfirmationToken  *string    `json:"-" form:"-"`
		MobilePhoneConfirmedAt        *time.Time `json:"-" form:"-"`
		MobilePhone                   *string    `json:"mobile_phone" form:"-"`
		OtpToken                      *string    `json:"-" form:"-"`
		OtpCreatedAt                  *time.Time `json:"-" form:"-"`
		OtpExpiredAt                  *time.Time `json:"-" form:"-"`
		Username                      string     `json:"username" form:"-"`
		UnconfirmedEmail              *string    `json:"-" form:"-"`
		EmailConfirmationSentAt       *time.Time `json:"-" form:"-"`
		EmailConfirmationToken        *string    `json:"-" form:"-"`
		EmailConfirmedAt              *time.Time `json:"-" form:"-"`
		Email                         *string    `json:"-" form:"-"`
		EncryptedPassword             *string    `json:"-" form:"-"`
		PasswordResetAt               *time.Time `json:"-" form:"-"`
		PasswordResetToken            *string    `json:"-" form:"-"`
		PasswordChangedAt             *time.Time `json:"-" form:"-"`
		Photo                         string     `json:"-" form:"-"`
		UnreadNotifications           int64      `json:"-" form:"-"`
		CreatedBy                     *uuid.UUID `json:"-" form:"-"`
		CreatedAt                     time.Time  `json:"created_at" form:"-"`
		UpdatedBy                     *uuid.UUID `json:"-" form:"-"`
		UpdatedAt                     time.Time  `json:"-" form:"-"`
		DeactivatedBy                 *uuid.UUID `json:"-" form:"-"`
		DeactivatedAt                 *time.Time `json:"-" form:"-"`
		DeactivationReason            *string    `json:"-" form:"-"`
	}

	userFilter struct {
		UserIDs,
		ExcludedUserIDs []uuid.UUID
		UserStatuses []UserStatus
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
)
