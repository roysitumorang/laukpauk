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
	ErrTransactionDayNotFound            = errors.New("transaction day not found")
	ErrUniqueTransactionDayDateViolation = errors.New("date: already exists")
)

type (
	TransactionDay struct {
		RowNo              int64                `json:"-" form:"-"`
		ID                 uuid.UUID            `json:"id" form:"-"`
		Date               time.Time            `json:"date" form:"-"`
		Today              *bool                `json:"today" form:"-"`
		PublicHoliday      bool                 `json:"public_holiday" form:"-"`
		CreatedBy          *uuid.UUID           `json:"-" form:"-"`
		CreatedAt          time.Time            `json:"-" form:"-"`
		UpdatedBy          *uuid.UUID           `json:"-" form:"-"`
		UpdatedAt          time.Time            `json:"-" form:"-"`
		TransactionWindows []*TransactionWindow `json:"transaction_windows" form:"-"`
	}

	TransactionWindow struct {
		ID                   uuid.UUID       `json:"id" form:"-"`
		TransactionDayID     uuid.UUID       `json:"-" form:"-"`
		DeliveryPlanID       uuid.UUID       `json:"-" form:"-"`
		TransactionWindowMin time.Time       `json:"min" form:"-"`
		TransactionWindowMax time.Time       `json:"max" form:"-"`
		CreatedAt            time.Time       `json:"-" form:"-"`
		DeliveryTimes        []*DeliveryTime `json:"delivery_times" form:"-"`
	}

	DeliveryTime struct {
		ID                  uuid.UUID     `json:"id" form:"-"`
		TransactionWindowID uuid.UUID     `json:"-" form:"-"`
		DeliverySpeed       DeliverySpeed `json:"-" form:"-"`
		Datetime            time.Time     `json:"datetime" form:"-"`
		CreatedAt           time.Time     `json:"-" form:"-"`
	}

	TransactionDayFilter struct {
		TransactionDayIDs []uuid.UUID
		Dates             []time.Time
		Today,
		PublicHoliday *bool
		PaginationURL string
		Sortings      []Sorting
		MapSortings   Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewTransactionDayFilter() *TransactionDayFilter {
	return &TransactionDayFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *TransactionDayFilter) WithTransactionDayIDs(transactionDayIDs ...uuid.UUID) *TransactionDayFilter {
	q.TransactionDayIDs = transactionDayIDs
	return q
}

func (q *TransactionDayFilter) WithDates(dates ...time.Time) *TransactionDayFilter {
	q.Dates = dates
	return q
}

func (q *TransactionDayFilter) WithToday(today *bool) *TransactionDayFilter {
	q.Today = today
	return q
}

func (q *TransactionDayFilter) WithPublicHoliday(publicHoliday *bool) *TransactionDayFilter {
	q.PublicHoliday = publicHoliday
	return q
}

func (q *TransactionDayFilter) WithPaginationURL(paginationURL string) *TransactionDayFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *TransactionDayFilter) WithSortings(sortings ...Sorting) *TransactionDayFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *TransactionDayFilter) WithLimit(limit int64) *TransactionDayFilter {
	q.Limit = limit
	return q
}

func (q *TransactionDayFilter) WithPage(page int64) *TransactionDayFilter {
	q.Page = page
	return q
}

func (q *TransactionDayFilter) WithUrlValues(urlValues url.Values) *TransactionDayFilter {
	q.UrlValues = urlValues
	return q
}
