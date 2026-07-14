package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrNonDeliveryDayNotFound                          = errors.New("non delivery day not found")
	ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation = errors.New("date and delivery plan already exists")

	MapNonDeliveryDaySortings = map[string]struct{}{
		SortingFieldID:                 {},
		NonDeliveryDaySortingFieldDate: {},
	}
)

const (
	NonDeliveryDaySortingFieldDate = "date"
)

type (
	NonDeliveryDay struct {
		RowNo            int64        `json:"-" form:"-"`
		ID               uuid.UUID    `json:"-" form:"-"`
		RecordStatus     RecordStatus `json:"-" form:"-"`
		Date             time.Time    `json:"-" form:"-"`
		RawDate          string       `json:"-" form:"date"`
		DeliveryPlanID   uuid.UUID    `json:"-" form:"delivery_plan_id"`
		DeliveryPlanName string       `json:"-" form:"-"`
		CreatedBy        uuid.UUID    `json:"-" form:"-"`
		CreatedAt        time.Time    `json:"-" form:"-"`
		DeletedBy        *uuid.UUID   `json:"-" form:"-"`
		DeletedAt        *time.Time   `json:"-" form:"-"`
		DeletionReason   *string      `json:"-" form:"-"`
	}

	NonDeliveryDayFilter struct {
		NonDeliveryDayIDs,
		ExcludedNonDeliveryDayIDs,
		DeliveryPlanIDs []uuid.UUID
		RecordStatuses []RecordStatus
		Dates          []time.Time
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewNonDeliveryDayFilter() *NonDeliveryDayFilter {
	return new(NonDeliveryDayFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *NonDeliveryDayFilter) WithNonDeliveryDayIDs(nonDeliveryDayIDs ...uuid.UUID) *NonDeliveryDayFilter {
	q.NonDeliveryDayIDs = nonDeliveryDayIDs
	return q
}

func (q *NonDeliveryDayFilter) WithExcludedNonDeliveryDayIDs(excludedNonDeliveryDayIDs ...uuid.UUID) *NonDeliveryDayFilter {
	q.ExcludedNonDeliveryDayIDs = excludedNonDeliveryDayIDs
	return q
}

func (q *NonDeliveryDayFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *NonDeliveryDayFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *NonDeliveryDayFilter) WithDeliveryPlanIDs(deliveryPlanIDs ...uuid.UUID) *NonDeliveryDayFilter {
	q.DeliveryPlanIDs = deliveryPlanIDs
	return q
}

func (q *NonDeliveryDayFilter) WithDates(dates ...time.Time) *NonDeliveryDayFilter {
	q.Dates = dates
	return q
}

func (q *NonDeliveryDayFilter) WithKeyword(keyword string) *NonDeliveryDayFilter {
	q.Keyword = keyword
	return q
}

func (q *NonDeliveryDayFilter) WithPaginationURL(paginationURL string) *NonDeliveryDayFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *NonDeliveryDayFilter) WithSortings(sortings ...Sorting) *NonDeliveryDayFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *NonDeliveryDayFilter) WithLimit(limit int64) *NonDeliveryDayFilter {
	q.Limit = limit
	return q
}

func (q *NonDeliveryDayFilter) WithPage(page int64) *NonDeliveryDayFilter {
	q.Page = page
	return q
}

func (q *NonDeliveryDayFilter) WithUrlValues(urlValues url.Values) *NonDeliveryDayFilter {
	q.UrlValues = urlValues
	return q
}

func (q *NonDeliveryDay) Validate() (err error) {
	location := helper.LoadTimeZone()
	if q.RawDate = strings.TrimSpace(q.RawDate); q.RawDate == "" {
		return errors.New("date: is required")
	}
	if q.Date, err = time.ParseInLocation(time.DateOnly, q.RawDate, location); err != nil {
		return fmt.Errorf("date: should use format '%s'", time.DateOnly)
	}
	if q.Date.Before(time.Now()) {
		return errors.New("date: should exceed current date")
	}
	return nil
}

func (q *NonDeliveryDay) Populate() *NonDeliveryDay {
	location := helper.LoadTimeZone()
	q.Date = q.Date.In(location)
	if !q.Date.IsZero() {
		q.RawDate = q.Date.Format(time.DateOnly)
	}
	return q
}
