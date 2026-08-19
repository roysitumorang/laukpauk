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

const (
	TextMessageTypeSms TextMessageType = iota
	TextMessageTypeOtp
)

const (
	TopicTextMessage                 Topic      = "text_message"
	TopicTextMessageEventSend        TopicEvent = "send"
	TopicTextMessageEventCheckReport TopicEvent = "check_report"
)

type (
	TextMessageType int

	TextMessage struct {
		RowNo          int64           `json:"-" form:"-"`
		ID             uuid.UUID       `json:"-" form:"-"`
		BuyerID        uuid.UUID       `json:"-" form:"-"`
		MobilePhone    string          `json:"-" form:"-"`
		MessageType    TextMessageType `json:"-" form:"-"`
		Body           string          `json:"-" form:"-"`
		DeliveryStatus *string         `json:"-" form:"-"`
		MessageID      *int64          `json:"-" form:"-"`
		CreatedBy      *uuid.UUID      `json:"-" form:"-"`
		CreatedAt      time.Time       `json:"-" form:"-"`
		UpdatedAt      time.Time       `json:"-" form:"-"`
		SentAt         *time.Time      `json:"-" form:"-"`
	}

	TextMessageFilter struct {
		TextMessageIDs,
		BuyerIDs []uuid.UUID
		DeliveryStatusBlank bool
		DeliveryStatuses    []string
		MessageIDs          []int64
		PaginationURL       string
		Sortings            []Sorting
		MapSortings         Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

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

func (q *TextMessageFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *TextMessageFilter {
	q.BuyerIDs = buyerIDs
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

func (q TextMessageType) Int() int {
	return int(q)
}

func (q TextMessageType) String() string {
	if q == TextMessageTypeOtp {
		return "otp"
	}
	return "sms"
}
