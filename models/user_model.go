package models

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/nyaruka/phonenumbers"
	"github.com/roysitumorang/laukpauk/helper"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUniqueUserUsernameViolation = errors.New("username: already exists")
var ErrUniqueUserMobilePhoneViolation = errors.New("mobile_phone: already exists")
var ErrUniqueUserEmailViolation = errors.New("email: already exists")
var ErrUserSelfDeletingProhibited = errors.New("self deleting prohibited")
var ErrUserLocked = errors.New("user locked")

const CurrentUser = "current_user"
const UserID = "user_id"

const RoleSeller = "seller"
const RoleBuyer = "buyer"

var Roles = []string{RoleSeller, RoleBuyer}
var MapRoles = map[string]struct{}{RoleSeller: {}, RoleBuyer: {}}

const UserStatusActive = "active"
const UserStatusLocked = "locked"
const UserStatusDeactivated = "deactivated"

const UserLevelSellerOwner = "owner"
const UserLevelSellerStaff = "staff"
const UserLevelBuyer = "buyer"

var UserLevels = []string{
	UserLevelSellerOwner,
	UserLevelSellerStaff,
	UserLevelBuyer,
}
var MapUserLevels = map[string]struct{}{
	UserLevelSellerOwner: {},
	UserLevelSellerStaff: {},
	UserLevelBuyer:       {},
}

type User struct {
	RowNo                         int64      `json:"-" form:"-"`
	ID                            uuid.UUID  `json:"id" form:"-"`
	Role                          string     `json:"role" form:"-"`
	Status                        string     `json:"-" form:"-"`
	Name                          string     `json:"name" form:"name"`
	MobilePhoneConfirmationSentAt *time.Time `json:"-" form:"-"`
	MobilePhoneConfirmationToken  *string    `json:"-" form:"-"`
	MobilePhoneConfirmedAt        *time.Time `json:"mobile_phone_confirmed_at" form:"-"`
	MobilePhone                   string     `json:"mobile_phone" form:"mobile_phone"`
	Username                      *string    `json:"username" form:"username"`
	EmailConfirmationSentAt       *time.Time `json:"-" form:"-"`
	EmailConfirmationToken        *string    `json:"-" form:"-"`
	EmailConfirmedAt              *time.Time `json:"-" form:"-"`
	Email                         *string    `json:"email" form:"email"`
	EncryptedPassword             *string    `json:"-" form:"-"`
	PasswordResetSentAt           *time.Time `json:"-" form:"-"`
	PasswordResetToken            *string    `json:"-" form:"-"`
	PasswordChangedAt             *time.Time `json:"password_changed_at" form:"-"`
	SignInCount                   int        `json:"sign_in_count" form:"-"`
	CurrentSignInAt               *time.Time `json:"current_sign_in_at" form:"-"`
	LastSignInAt                  *time.Time `json:"last_sign_in_at" form:"-"`
	CurrentSignInIP               *string    `json:"current_sign_in_ip" form:"-"`
	LastSignInIP                  *string    `json:"last_sign_in_ip" form:"-"`
	FailedAttempts                int        `json:"failed_attempts" form:"-"`
	UnlockToken                   *string    `json:"-" form:"-"`
	LockedAt                      *time.Time `json:"locked_at" form:"-"`
	OtpSecret                     *string    `json:"-" form:"-"`
	OtpExpiredAt                  *time.Time `json:"-" form:"-"`
	Photo                         string     `json:"photo" form:"photo"`
	UserLevel                     string     `json:"user_level" form:"user_level"`
	ReferralCode                  *string    `json:"referral_code" form:"referral_code"`
	Address                       *string    `json:"address" form:"address"`
	ReferrerID                    *uuid.UUID `json:"referrer_id" form:"referrer_id"`
	ReferralPoints                *int64     `json:"referral_points" form:"-"`
	RewardPoints                  *int64     `json:"reward_points" form:"-"`
	Wallet                        *int64     `json:"wallet" form:"-"`
	DeliveryRouteID               *uuid.UUID `json:"delivery_route_id" form:"-"`
	CreatedBy                     *uuid.UUID `json:"-" form:"-"`
	CreatedAt                     time.Time  `json:"created_at" form:"-"`
	UpdatedBy                     *uuid.UUID `json:"-" form:"-"`
	UpdatedAt                     time.Time  `json:"-" form:"-"`
	DeactivatedBy                 *uuid.UUID `json:"-" form:"-"`
	DeactivatedAt                 *time.Time `json:"-" form:"-"`
	DeactivationReason            *string    `json:"-" form:"-"`
}

func (q *User) Validate() error {
	if q.Role = strings.TrimSpace(q.Role); q.Role == "" {
		return errors.New("role: is required")
	}
	_, ok := MapRoles[q.Role]
	if !ok {
		return errors.New("role: not found")
	}
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.MobilePhone = strings.TrimSpace(q.MobilePhone); q.MobilePhone == "" {
		return errors.New("mobile_phone: is required")
	}
	mobilePhone, err := phonenumbers.Parse(q.MobilePhone, "ID")
	if err != nil {
		return fmt.Errorf("mobile_phone: %w", err)
	}
	q.MobilePhone = phonenumbers.Format(mobilePhone, phonenumbers.E164)
	if q.Email != nil {
		if _, err = mail.ParseAddress(*q.Email); err != nil {
			return fmt.Errorf("email: %w", err)
		}
		*q.Email = strings.ToLower(*q.Email)
	}
	if q.Role == RoleBuyer {
		q.UserLevel = UserLevelBuyer
	} else if q.UserLevel = strings.TrimSpace(q.UserLevel); q.UserLevel == "" {
		return errors.New("user_level: is required")
	} else if q.UserLevel != UserLevelSellerOwner &&
		q.UserLevel != UserLevelSellerStaff {
		return errors.New("user_level: not found")
	}
	if q.ReferralCode != nil {
		if *q.ReferralCode = strings.TrimSpace(*q.ReferralCode); *q.ReferralCode == "" {
			q.ReferralCode = nil
		} else {
			*q.ReferralCode = strings.ToUpper(*q.ReferralCode)
		}
	}
	if q.ReferrerID != nil && q.ReferrerID.Compare(uuid.Nil()) == 0 {
		q.ReferrerID = nil
	}
	return nil
}

func (q *User) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *User) SetPhoto(photo string) *User {
	q.Photo = photo
	return q
}

func (q *User) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

func (q *User) SellerOwner() bool {
	return q.UserLevel == UserLevelSellerOwner
}

type UserFilter struct {
	UserIDs,
	ExcludedUserIDs []uuid.UUID
	Roles,
	Statuses,
	ReferralCodes []string
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

func NewUserFilter() *UserFilter {
	return &UserFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *UserFilter) WithUserIDs(userIDs ...uuid.UUID) *UserFilter {
	q.UserIDs = userIDs
	return q
}

func (q *UserFilter) WithExcludedUserIDs(excludedUserIDs ...uuid.UUID) *UserFilter {
	q.ExcludedUserIDs = excludedUserIDs
	return q
}

func (q *UserFilter) WithRoles(roles ...string) *UserFilter {
	q.Roles = roles
	return q
}

func (q *UserFilter) WithStatuses(statuses ...string) *UserFilter {
	q.Statuses = statuses
	return q
}

func (q *UserFilter) WithReferralCodes(referralCodes ...string) *UserFilter {
	q.ReferralCodes = referralCodes
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

type SignUp struct {
	Name           string `json:"name" form:"-"`
	Username       string `json:"username" form:"-"`
	MobilePhone    string `json:"mobile_phone" form:"-"`
	Base64Password string `json:"password" form:"-"`
	Password       string `json:"-" form:"-"`
	ReferralCode   string `json:"referral_code" form:"-"`
}

func (q *SignUp) Validate() error {
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
	if q.Base64Password = strings.TrimSpace(q.Base64Password); q.Base64Password == "" {
		return errors.New("password: is required")
	}
	password, err := helper.Base64Decode(q.Base64Password)
	if err != nil {
		return fmt.Errorf("password: %s", err.Error())
	}
	q.Password = password
	q.ReferralCode = strings.TrimSpace(q.ReferralCode)
	return nil
}

type SignInRequest struct {
	Login          string `json:"login" form:"login"`
	Base64Password string `json:"password" form:"password"`
	Password       string `json:"-" form:"-"`
}

func (q *SignInRequest) Validate() error {
	if q.Login = strings.TrimSpace(q.Login); q.Login == "" {
		return errors.New("login: is required")
	}
	if q.Base64Password = strings.TrimSpace(q.Base64Password); q.Base64Password == "" {
		return errors.New("password: is required")
	}
	password, err := helper.Base64Decode(q.Base64Password)
	if err != nil {
		return fmt.Errorf("password: %s", err.Error())
	}
	q.Password = password
	return nil
}

type SignInResponse struct {
	IDToken   string    `json:"id_token"`
	ExpiredAt time.Time `json:"expired_at"`
	User      *User     `json:"user"`
}

var MapUserAccountSortings = map[string]struct{}{
	SortingFieldID:   {},
	SortingFieldName: {},
}

// UserAccount is the request body a seller sends to create or update an
// account: a staff seller (owner only) or a buyer. Address and referrer only
// apply to buyers.
type UserAccount struct {
	Name           string     `json:"name" form:"name"`
	Username       string     `json:"username" form:"username"`
	MobilePhone    string     `json:"mobile_phone" form:"mobile_phone"`
	Email          *string    `json:"email" form:"email"`
	Address        *string    `json:"address" form:"address"`
	ReferrerID     *uuid.UUID `json:"referrer_id" form:"referrer_id"`
	Base64Password string     `json:"password" form:"password"`
	Password       string     `json:"-" form:"-"`
	Photo          string     `json:"-" form:"-"`
}

func (q *UserAccount) Validate(passwordRequired, addressRequired bool) error {
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
	if q.Email != nil {
		if *q.Email = strings.ToLower(strings.TrimSpace(*q.Email)); *q.Email == "" {
			q.Email = nil
		} else if _, err = mail.ParseAddress(*q.Email); err != nil {
			return fmt.Errorf("email: %w", err)
		}
	}
	if q.Address != nil {
		if *q.Address = strings.TrimSpace(*q.Address); *q.Address == "" {
			q.Address = nil
		}
	}
	if q.Address == nil && addressRequired {
		return errors.New("address: is required")
	}
	if q.ReferrerID != nil && q.ReferrerID.Compare(uuid.Nil()) == 0 {
		q.ReferrerID = nil
	}
	if q.Base64Password = strings.TrimSpace(q.Base64Password); q.Base64Password == "" {
		if passwordRequired {
			return errors.New("password: is required")
		}
		return nil
	}
	if q.Password, err = helper.Base64Decode(q.Base64Password); err != nil {
		return fmt.Errorf("password: %s", err.Error())
	}
	return nil
}
