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
	ErrCustomDeliveryDayNotFound                          = errors.New("custom delivery day not found")
	ErrUniqueCustomDeliveryDayDateDeliveryPlanIdViolation = errors.New("date and delivery plan already exists")

	MapCustomDeliveryDaySortings = map[string]struct{}{
		SortingFieldID:                    {},
		CustomDeliveryDaySortingFieldDate: {},
	}
)

const (
	CustomDeliveryDaySortingFieldDate = "date"
)

type (
	CustomDeliveryDay struct {
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

	CustomDeliveryDayFilter struct {
		CustomDeliveryDayIDs,
		ExcludedCustomDeliveryDayIDs,
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

func NewCustomDeliveryDayFilter() *CustomDeliveryDayFilter {
	return &CustomDeliveryDayFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *CustomDeliveryDayFilter) WithCustomDeliveryDayIDs(customDeliveryDayIDs ...uuid.UUID) *CustomDeliveryDayFilter {
	q.CustomDeliveryDayIDs = customDeliveryDayIDs
	return q
}

func (q *CustomDeliveryDayFilter) WithExcludedCustomDeliveryDayIDs(excludedCustomDeliveryDayIDs ...uuid.UUID) *CustomDeliveryDayFilter {
	q.ExcludedCustomDeliveryDayIDs = excludedCustomDeliveryDayIDs
	return q
}

func (q *CustomDeliveryDayFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *CustomDeliveryDayFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *CustomDeliveryDayFilter) WithDeliveryPlanIDs(deliveryPlanIDs ...uuid.UUID) *CustomDeliveryDayFilter {
	q.DeliveryPlanIDs = deliveryPlanIDs
	return q
}

func (q *CustomDeliveryDayFilter) WithDates(dates ...time.Time) *CustomDeliveryDayFilter {
	q.Dates = dates
	return q
}

func (q *CustomDeliveryDayFilter) WithKeyword(keyword string) *CustomDeliveryDayFilter {
	q.Keyword = keyword
	return q
}

func (q *CustomDeliveryDayFilter) WithPaginationURL(paginationURL string) *CustomDeliveryDayFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *CustomDeliveryDayFilter) WithSortings(sortings ...Sorting) *CustomDeliveryDayFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *CustomDeliveryDayFilter) WithLimit(limit int64) *CustomDeliveryDayFilter {
	q.Limit = limit
	return q
}

func (q *CustomDeliveryDayFilter) WithPage(page int64) *CustomDeliveryDayFilter {
	q.Page = page
	return q
}

func (q *CustomDeliveryDayFilter) WithUrlValues(urlValues url.Values) *CustomDeliveryDayFilter {
	q.UrlValues = urlValues
	return q
}

func (q *CustomDeliveryDay) Validate() (err error) {
	if q.RawDate = strings.TrimSpace(q.RawDate); q.RawDate == "" {
		return errors.New("date: is required")
	}
	if q.Date, err = time.ParseInLocation(time.DateOnly, q.RawDate, time.Local); err != nil {
		return fmt.Errorf("date: should use format '%s'", time.DateOnly)
	}
	now := time.Now().In(time.Local)
	t := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)
	if q.Date.Before(t) {
		return errors.New("date: should equal or exceed current date")
	}
	return nil
}

func (q *CustomDeliveryDay) Populate() *CustomDeliveryDay {
	q.Date = q.Date.In(time.Local)
	if !q.Date.IsZero() {
		q.RawDate = q.Date.Format(time.DateOnly)
	}
	return q
}
