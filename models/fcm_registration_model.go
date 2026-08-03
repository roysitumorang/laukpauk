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
	ErrFcmRegistrationNotFound = errors.New("FCM registration not found")
)

type (
	FcmRegistration struct {
		RowNo     int64     `json:"-" form:"-"`
		ID        uuid.UUID `json:"id" form:"-"`
		UserID    uuid.UUID `json:"user_id" form:"-"`
		Token     string    `json:"token" form:"-"`
		CreatedAt time.Time `json:"created_at" form:"-"`
		UpdatedAt time.Time `json:"updated_at" form:"-"`
		ExpiredAt time.Time `json:"expired_at" form:"-"`
	}

	FcmRegistrationFilter struct {
		FcmRegistrationIDs,
		UserIDs []uuid.UUID
		Tokens        []string
		Expired       *bool
		PaginationURL string
		Sortings      []Sorting
		MapSortings   Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewFcmRegistrationFilter() *FcmRegistrationFilter {
	return &FcmRegistrationFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *FcmRegistrationFilter) WithFcmRegistrationIDs(fcmRegistrationIDs ...uuid.UUID) *FcmRegistrationFilter {
	q.FcmRegistrationIDs = fcmRegistrationIDs
	return q
}

func (q *FcmRegistrationFilter) WithUserIDs(userIDs ...uuid.UUID) *FcmRegistrationFilter {
	q.UserIDs = userIDs
	return q
}

func (q *FcmRegistrationFilter) WithTokens(tokens ...string) *FcmRegistrationFilter {
	q.Tokens = tokens
	return q
}

func (q *FcmRegistrationFilter) WithExpired(expired *bool) *FcmRegistrationFilter {
	q.Expired = expired
	return q
}

func (q *FcmRegistrationFilter) WithPaginationURL(paginationURL string) *FcmRegistrationFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *FcmRegistrationFilter) WithSortings(sortings ...Sorting) *FcmRegistrationFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *FcmRegistrationFilter) WithLimit(limit int64) *FcmRegistrationFilter {
	q.Limit = limit
	return q
}

func (q *FcmRegistrationFilter) WithPage(page int64) *FcmRegistrationFilter {
	q.Page = page
	return q
}

func (q *FcmRegistrationFilter) WithUrlValues(urlValues url.Values) *FcmRegistrationFilter {
	q.UrlValues = urlValues
	return q
}

func (q *FcmRegistration) Validate() error {
	if q.Token = strings.TrimSpace(q.Token); q.Token == "" {
		return errors.New("token: is required")
	}
	return nil
}

func (q *FcmRegistration) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *FcmRegistration) Expired() bool {
	return !q.ExpiredAt.After(time.Now())
}
