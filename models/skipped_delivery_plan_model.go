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
	ErrSkippedDeliveryPlanNotFound                          = errors.New("skipped delivery plan not found")
	ErrUniqueSkippedDeliveryPlanDateDeliveryPlanIdViolation = errors.New("date and delivery plan already exists")

	MapSkippedDeliveryPlanSortings = map[string]struct{}{
		SortingFieldID:                      {},
		SkippedDeliveryPlanSortingFieldDate: {},
	}
)

const (
	SkippedDeliveryPlanSortingFieldDate = "date"
)

type (
	SkippedDeliveryPlan struct {
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

	SkippedDeliveryPlanFilter struct {
		SkippedDeliveryPlanIDs,
		ExcludedSkippedDeliveryPlanIDs,
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

func NewSkippedDeliveryPlanFilter() *SkippedDeliveryPlanFilter {
	return &SkippedDeliveryPlanFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *SkippedDeliveryPlanFilter) WithSkippedDeliveryPlanIDs(skippedDeliveryPlanIDs ...uuid.UUID) *SkippedDeliveryPlanFilter {
	q.SkippedDeliveryPlanIDs = skippedDeliveryPlanIDs
	return q
}

func (q *SkippedDeliveryPlanFilter) WithExcludedSkippedDeliveryPlanIDs(excludedSkippedDeliveryPlanIDs ...uuid.UUID) *SkippedDeliveryPlanFilter {
	q.ExcludedSkippedDeliveryPlanIDs = excludedSkippedDeliveryPlanIDs
	return q
}

func (q *SkippedDeliveryPlanFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *SkippedDeliveryPlanFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *SkippedDeliveryPlanFilter) WithDeliveryPlanIDs(deliveryPlanIDs ...uuid.UUID) *SkippedDeliveryPlanFilter {
	q.DeliveryPlanIDs = deliveryPlanIDs
	return q
}

func (q *SkippedDeliveryPlanFilter) WithDates(dates ...time.Time) *SkippedDeliveryPlanFilter {
	q.Dates = dates
	return q
}

func (q *SkippedDeliveryPlanFilter) WithKeyword(keyword string) *SkippedDeliveryPlanFilter {
	q.Keyword = keyword
	return q
}

func (q *SkippedDeliveryPlanFilter) WithPaginationURL(paginationURL string) *SkippedDeliveryPlanFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *SkippedDeliveryPlanFilter) WithSortings(sortings ...Sorting) *SkippedDeliveryPlanFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *SkippedDeliveryPlanFilter) WithLimit(limit int64) *SkippedDeliveryPlanFilter {
	q.Limit = limit
	return q
}

func (q *SkippedDeliveryPlanFilter) WithPage(page int64) *SkippedDeliveryPlanFilter {
	q.Page = page
	return q
}

func (q *SkippedDeliveryPlanFilter) WithUrlValues(urlValues url.Values) *SkippedDeliveryPlanFilter {
	q.UrlValues = urlValues
	return q
}

func (q *SkippedDeliveryPlan) Validate() (err error) {
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

func (q *SkippedDeliveryPlan) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *SkippedDeliveryPlan) Populate() *SkippedDeliveryPlan {
	q.Date = q.Date.In(time.Local)
	if !q.Date.IsZero() {
		q.RawDate = q.Date.Format(time.DateOnly)
	}
	return q
}
