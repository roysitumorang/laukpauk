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
	ErrSmsMessageNotFound = errors.New("sms message not found")
)

const (
	SmsMessageTypeSms SmsMessageType = iota
	SmsMessageTypeOtp
)

const (
	TopicSms                 Topic      = "sms"
	TopicSmsEventSend        TopicEvent = "send"
	TopicSmsEventCheckReport TopicEvent = "check_report"
)

type (
	SmsMessageType int

	SmsMessage struct {
		RowNo          int64          `json:"-" form:"-"`
		ID             uuid.UUID      `json:"-" form:"-"`
		UserID         uuid.UUID      `json:"-" form:"-"`
		MobilePhone    string         `json:"-" form:"-"`
		MessageType    SmsMessageType `json:"-" form:"-"`
		Body           string         `json:"-" form:"-"`
		DeliveryStatus *string        `json:"-" form:"-"`
		MessageID      *int64         `json:"-" form:"-"`
		CreatedBy      *uuid.UUID     `json:"-" form:"-"`
		CreatedAt      time.Time      `json:"-" form:"-"`
		UpdatedAt      time.Time      `json:"-" form:"-"`
		SentAt         *time.Time     `json:"-" form:"-"`
	}

	SmsMessageFilter struct {
		SmsMessageIDs,
		UserIDs []uuid.UUID
		DeliveryStatusBlank bool
		DeliveryStatuses    []string
		MessageIDs          []int64
		PaginationURL       string
		Sortings            []Sorting
		MapSortings         *Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewSmsMessageFilter() *SmsMessageFilter {
	return new(SmsMessageFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *SmsMessageFilter) WithSmsMessageIDs(smsMessageIDs ...uuid.UUID) *SmsMessageFilter {
	q.SmsMessageIDs = smsMessageIDs
	return q
}

func (q *SmsMessageFilter) WithUserIDs(userIDs ...uuid.UUID) *SmsMessageFilter {
	q.UserIDs = userIDs
	return q
}

func (q *SmsMessageFilter) WithDeliveryStatusBlank() *SmsMessageFilter {
	q.DeliveryStatusBlank = true
	return q
}

func (q *SmsMessageFilter) WithDeliveryStatuses(deliveryStatuses ...string) *SmsMessageFilter {
	q.DeliveryStatuses = deliveryStatuses
	return q
}

func (q *SmsMessageFilter) WithMessageIDs(messageIDs ...int64) *SmsMessageFilter {
	q.MessageIDs = messageIDs
	return q
}

func (q *SmsMessageFilter) WithPaginationURL(paginationURL string) *SmsMessageFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *SmsMessageFilter) WithSortings(sortings ...Sorting) *SmsMessageFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *SmsMessageFilter) WithLimit(limit int64) *SmsMessageFilter {
	q.Limit = limit
	return q
}

func (q *SmsMessageFilter) WithPage(page int64) *SmsMessageFilter {
	q.Page = page
	return q
}

func (q *SmsMessageFilter) WithUrlValues(urlValues url.Values) *SmsMessageFilter {
	q.UrlValues = urlValues
	return q
}

func (q SmsMessageType) Int() int {
	return int(q)
}

func (q SmsMessageType) String() string {
	if q == SmsMessageTypeOtp {
		return "otp"
	}
	return "sms"
}
