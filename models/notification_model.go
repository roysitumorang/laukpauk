package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrNotificationNotFound = errors.New("notification not found")
var MapNotificationSortings = map[string]struct{}{
	SortingFieldID: {},
}

type NotificationType string

const (
	NotificationTypeOther       NotificationType = "other"
	NotificationTypeTransaction NotificationType = "transaction"
	NotificationTypeFeedback    NotificationType = "feedback"
	NotificationTypeVoucher     NotificationType = "voucher"
	NotificationTypeWallet      NotificationType = "wallet"
)

const TopicNotification Topic = "notification"
const TopicNotificationEventSend TopicEvent = "send"

type Notification struct {
	RowNo            int64             `json:"-" form:"-"`
	ID               uuid.UUID         `json:"id" form:"-"`
	UserID           uuid.UUID         `json:"user_id" form:"user_id"`
	NotificationType NotificationType  `json:"notification_type" form:"-"`
	Link             *string           `json:"link" form:"-"`
	LinkID           *uuid.UUID        `json:"link_id" form:"-"`
	Subject          string            `json:"subject" form:"subject"`
	Body             string            `json:"body" form:"body"`
	Photo            *string           `json:"photo" form:"photo"`
	Params           map[string]string `json:"params" form:"-"`
	CreatedBy        uuid.UUID         `json:"created_by" form:"-"`
	CreatedAt        time.Time         `json:"created_at" form:"-"`
	SentAt           *time.Time        `json:"sent_at" form:"-"`
	ReadAt           *time.Time        `json:"read_at" form:"-"`
}

func (q *Notification) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Notification) SetPhoto(photo *string) *Notification {
	q.Photo = photo
	return q
}

func (q *Notification) PhotoPath(baseURLs ...string) string {
	if q.Photo == nil {
		return ""
	}
	return helper.PhotoPath(*q.Photo, baseURLs...)
}

func (q *Notification) Validate() error {
	if q.Subject = strings.TrimSpace(q.Subject); q.Subject == "" {
		return errors.New("subject: is required")
	}
	q.Body = strings.TrimSpace(q.Body)
	return nil
}

type NotificationFilter struct {
	NotificationIDs,
	UserIDs []uuid.UUID
	NotificationTypes []NotificationType
	SentAtBlank,
	ReadAtBlank *bool
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewNotificationFilter() *NotificationFilter {
	return &NotificationFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *NotificationFilter) WithNotificationIDs(notificationIDs ...uuid.UUID) *NotificationFilter {
	q.NotificationIDs = notificationIDs
	return q
}

func (q *NotificationFilter) WithUserIDs(userIDs ...uuid.UUID) *NotificationFilter {
	q.UserIDs = userIDs
	return q
}

func (q *NotificationFilter) WithNotificationTypes(notificationTypes ...NotificationType) *NotificationFilter {
	q.NotificationTypes = notificationTypes
	return q
}

func (q *NotificationFilter) WithSentAtBlank(sentAtBlank *bool) *NotificationFilter {
	q.SentAtBlank = sentAtBlank
	return q
}

func (q *NotificationFilter) WithReadAtBlank(readAtBlank *bool) *NotificationFilter {
	q.ReadAtBlank = readAtBlank
	return q
}

func (q *NotificationFilter) WithKeyword(keyword string) *NotificationFilter {
	q.Keyword = keyword
	return q
}

func (q *NotificationFilter) WithPaginationURL(paginationURL string) *NotificationFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *NotificationFilter) WithSortings(sortings ...Sorting) *NotificationFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *NotificationFilter) WithLimit(limit int64) *NotificationFilter {
	q.Limit = limit
	return q
}

func (q *NotificationFilter) WithPage(page int64) *NotificationFilter {
	q.Page = page
	return q
}

func (q *NotificationFilter) WithUrlValues(urlValues url.Values) *NotificationFilter {
	q.UrlValues = urlValues
	return q
}
