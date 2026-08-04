package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrNotificationTargetNotFound = errors.New("notification target not found")

type NotificationTarget struct {
	RowNo                  int64     `json:"-" form:"-"`
	ID                     uuid.UUID `json:"-" form:"-"`
	UserID                 uuid.UUID `json:"-" form:"-"`
	FirebaseInstallationID string    `json:"-" form:"fid"`
	CreatedAt              time.Time `json:"-" form:"-"`
	UpdatedAt              time.Time `json:"-" form:"-"`
}

func (q *NotificationTarget) Validate() error {
	if q.FirebaseInstallationID = strings.TrimSpace(q.FirebaseInstallationID); q.FirebaseInstallationID == "" {
		return errors.New("fid: is required")
	}
	return nil
}

func (q *NotificationTarget) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type NotificationTargetFilter struct {
	NotificationTargetIDs,
	UserIDs []uuid.UUID
	FirebaseInstallationIDs []string
	PaginationURL           string
	Sortings                []Sorting
	MapSortings             Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewNotificationTargetFilter() *NotificationTargetFilter {
	return &NotificationTargetFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *NotificationTargetFilter) WithNotificationTargetIDs(notificationTargetIDs ...uuid.UUID) *NotificationTargetFilter {
	q.NotificationTargetIDs = notificationTargetIDs
	return q
}

func (q *NotificationTargetFilter) WithUserIDs(userIDs ...uuid.UUID) *NotificationTargetFilter {
	q.UserIDs = userIDs
	return q
}

func (q *NotificationTargetFilter) WithFirebaseInstallationIDs(firebaseInstallationIDs ...string) *NotificationTargetFilter {
	q.FirebaseInstallationIDs = firebaseInstallationIDs
	return q
}

func (q *NotificationTargetFilter) WithPaginationURL(paginationURL string) *NotificationTargetFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *NotificationTargetFilter) WithSortings(sortings ...Sorting) *NotificationTargetFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *NotificationTargetFilter) WithLimit(limit int64) *NotificationTargetFilter {
	q.Limit = limit
	return q
}

func (q *NotificationTargetFilter) WithPage(page int64) *NotificationTargetFilter {
	q.Page = page
	return q
}

func (q *NotificationTargetFilter) WithUrlValues(urlValues url.Values) *NotificationTargetFilter {
	q.UrlValues = urlValues
	return q
}
