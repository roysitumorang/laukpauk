package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrTextMessageNotFound = errors.New("sms message not found")
var MapTextMessageSortings = map[string]struct{}{
	SortingFieldID: {},
}

const (
	TextMessageTypeSms TextMessageType = iota
	TextMessageTypeOtp
)

const (
	TopicTextMessage                 Topic      = "text_message"
	TopicTextMessageEventSend        TopicEvent = "send"
	TopicTextMessageEventCheckReport TopicEvent = "check_report"
)

type TextMessageType uint8

func (q TextMessageType) Int() uint8 {
	return uint8(q)
}

func (q TextMessageType) String() string {
	if q == TextMessageTypeOtp {
		return "otp"
	}
	return "sms"
}

type TextMessage struct {
	RowNo          int64           `json:"-" form:"-"`
	ID             uuid.UUID       `json:"-" form:"-"`
	UserID         uuid.UUID       `json:"-" form:"user_id"`
	MobilePhone    string          `json:"-" form:"-"`
	MessageType    TextMessageType `json:"-" form:"-"`
	Body           string          `json:"-" form:"body"`
	DeliveryStatus *string         `json:"-" form:"-"`
	MessageID      *int64          `json:"-" form:"-"`
	CreatedBy      *uuid.UUID      `json:"-" form:"-"`
	CreatedAt      time.Time       `json:"-" form:"-"`
	UpdatedAt      time.Time       `json:"-" form:"-"`
	SentAt         *time.Time      `json:"-" form:"-"`
}

func (q *TextMessage) Validate() error {
	if q.UserID.Compare(uuid.Nil()) == 0 {
		return errors.New("user_id: is required")
	}
	if q.Body = strings.TrimSpace(q.Body); q.Body == "" {
		return errors.New("body: is required")
	}
	return nil
}

type TextMessageFilter struct {
	TextMessageIDs,
	UserIDs []uuid.UUID
	MessageTypes        []TextMessageType
	DeliveryStatusBlank bool
	DeliveryStatuses    []string
	MessageIDs          []int64
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewTextMessageFilter() *TextMessageFilter {
	return &TextMessageFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *TextMessageFilter) WithTextMessageIDs(textMessageIDs ...uuid.UUID) *TextMessageFilter {
	q.TextMessageIDs = textMessageIDs
	return q
}

func (q *TextMessageFilter) WithUserIDs(userIDs ...uuid.UUID) *TextMessageFilter {
	q.UserIDs = userIDs
	return q
}

func (q *TextMessageFilter) WithMessageTypes(messageTypes ...TextMessageType) *TextMessageFilter {
	q.MessageTypes = messageTypes
	return q
}

func (q *TextMessageFilter) WithDeliveryStatusBlank() *TextMessageFilter {
	q.DeliveryStatusBlank = true
	return q
}

func (q *TextMessageFilter) WithDeliveryStatuses(deliveryStatuses ...string) *TextMessageFilter {
	q.DeliveryStatuses = deliveryStatuses
	return q
}

func (q *TextMessageFilter) WithMessageIDs(messageIDs ...int64) *TextMessageFilter {
	q.MessageIDs = messageIDs
	return q
}

func (q *TextMessageFilter) WithKeyword(keyword string) *TextMessageFilter {
	q.Keyword = keyword
	return q
}

func (q *TextMessageFilter) WithPaginationURL(paginationURL string) *TextMessageFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *TextMessageFilter) WithSortings(sortings ...Sorting) *TextMessageFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *TextMessageFilter) WithLimit(limit int64) *TextMessageFilter {
	q.Limit = limit
	return q
}

func (q *TextMessageFilter) WithPage(page int64) *TextMessageFilter {
	q.Page = page
	return q
}

func (q *TextMessageFilter) WithUrlValues(urlValues url.Values) *TextMessageFilter {
	q.UrlValues = urlValues
	return q
}
