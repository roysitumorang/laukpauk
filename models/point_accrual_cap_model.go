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
	ErrPointAccrualCapNotFound                       = errors.New("point accrual cap not found")
	ErrUniquePointAccrualCapMinimumPurchaseViolation = errors.New("minimum_purchase: already exists")

	MapPointAccrualCapSortings = map[string]struct{}{
		SortingFieldID: {},
		PointAccrualCapSortingFieldMinimumPurchase: {},
	}
)

const (
	PointAccrualCapSortingFieldMinimumPurchase = "minimum_purchase"
)

type (
	PointAccrualCap struct {
		RowNo                  int64        `json:"-" form:"-"`
		ID                     uuid.UUID    `json:"-" form:"id"`
		RecordStatus           RecordStatus `json:"-" form:"-"`
		RawMinimumPurchase     int64        `json:"-" form:"minimum_purchase"`
		MinimumPurchase        *int64       `json:"-" form:"-"`
		DeletedMinimumPurchase *int64       `json:"-" form:"-"`
		EarnedPoints           int64        `json:"-" form:"earned_points"`
		CreatedBy              uuid.UUID    `json:"-" form:"-"`
		CreatedAt              time.Time    `json:"-" form:"-"`
		UpdatedBy              uuid.UUID    `json:"-" form:"-"`
		UpdatedAt              time.Time    `json:"-" form:"-"`
		DeletedBy              *uuid.UUID   `json:"-" form:"-"`
		DeletedAt              *time.Time   `json:"-" form:"-"`
		DeletionReason         *string      `json:"-" form:"-"`
	}

	PointAccrualCapFilter struct {
		PointAccrualCapIDs,
		ExcludedPointAccrualCapIDs []uuid.UUID
		RecordStatuses   []RecordStatus
		MinimumPurchases []int64
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewPointAccrualCapFilter() *PointAccrualCapFilter {
	return new(PointAccrualCapFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *PointAccrualCapFilter) WithPointAccrualCapIDs(pointAccrualIDs ...uuid.UUID) *PointAccrualCapFilter {
	q.PointAccrualCapIDs = pointAccrualIDs
	return q
}

func (q *PointAccrualCapFilter) WithExcludedPointAccrualCapIDs(excludedPointAccrualIDs ...uuid.UUID) *PointAccrualCapFilter {
	q.ExcludedPointAccrualCapIDs = excludedPointAccrualIDs
	return q
}

func (q *PointAccrualCapFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *PointAccrualCapFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *PointAccrualCapFilter) WithMinimumPurchases(minimumPurchases ...int64) *PointAccrualCapFilter {
	q.MinimumPurchases = minimumPurchases
	return q
}

func (q *PointAccrualCapFilter) WithKeyword(keyword string) *PointAccrualCapFilter {
	q.Keyword = keyword
	return q
}

func (q *PointAccrualCapFilter) WithPaginationURL(paginationURL string) *PointAccrualCapFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *PointAccrualCapFilter) WithSortings(sortings ...Sorting) *PointAccrualCapFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *PointAccrualCapFilter) WithLimit(limit int64) *PointAccrualCapFilter {
	q.Limit = limit
	return q
}

func (q *PointAccrualCapFilter) WithPage(page int64) *PointAccrualCapFilter {
	q.Page = page
	return q
}

func (q *PointAccrualCapFilter) WithUrlValues(urlValues url.Values) *PointAccrualCapFilter {
	q.UrlValues = urlValues
	return q
}

func (q *PointAccrualCap) Validate() error {
	if q.RawMinimumPurchase < 1 {
		return errors.New("minimum_purchase: requires a positive integer")
	}
	q.MinimumPurchase = &q.RawMinimumPurchase
	if q.EarnedPoints < 1 {
		return errors.New("earned_points: requires a positive integer")
	}
	return nil
}

func (q *PointAccrualCap) Populate() *PointAccrualCap {
	if q.MinimumPurchase != nil {
		q.RawMinimumPurchase = *q.MinimumPurchase
	}
	return q
}
