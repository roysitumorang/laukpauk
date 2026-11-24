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
	ErrPaymentChannelNotFound = errors.New("payment channel not found")
	MapPaymentChannelSortings = map[string]struct{}{
		SortingFieldID:                 {},
		PaymentChannelSortingFieldName: {},
	}
)

const (
	PaymentChannelSortingFieldName = "name"
)

type (
	PaymentChannel struct {
		RowNo                int64     `json:"-"`
		ID                   uuid.UUID `json:"id"`
		BusinessID           string    `json:"-"`
		Name                 string    `json:"name"`
		ChannelCode          string    `json:"channel_code"`
		Country              string    `json:"-"`
		Currency             string    `json:"-"`
		MinimumAmount        int64     `json:"-"`
		MaximumAmount        int64     `json:"-"`
		PaymentRequestExpiry float64   `json:"-"`
		Refund               bool      `json:"-"`
		ChannelType          string    `json:"-"`
		ChannelCategory      string    `json:"-"`
		CreatedAt            time.Time `json:"-"`
	}

	PaymentChannelFilter struct {
		PaymentChannelIDs []uuid.UUID
		Amount            int64
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewPaymentChannelFilter() *PaymentChannelFilter {
	return &PaymentChannelFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *PaymentChannelFilter) WithPaymentChannelIDs(paymentChannelIDs ...uuid.UUID) *PaymentChannelFilter {
	q.PaymentChannelIDs = paymentChannelIDs
	return q
}

func (q *PaymentChannelFilter) WithAmount(amount int64) *PaymentChannelFilter {
	q.Amount = amount
	return q
}

func (q *PaymentChannelFilter) WithKeyword(keyword string) *PaymentChannelFilter {
	q.Keyword = keyword
	return q
}

func (q *PaymentChannelFilter) WithPaginationURL(paginationURL string) *PaymentChannelFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *PaymentChannelFilter) WithSortings(sortings ...Sorting) *PaymentChannelFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *PaymentChannelFilter) WithLimit(limit int64) *PaymentChannelFilter {
	q.Limit = limit
	return q
}

func (q *PaymentChannelFilter) WithPage(page int64) *PaymentChannelFilter {
	q.Page = page
	return q
}

func (q *PaymentChannelFilter) WithUrlValues(urlValues url.Values) *PaymentChannelFilter {
	q.UrlValues = urlValues
	return q
}
