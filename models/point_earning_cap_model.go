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
	ErrPointEarningCapNotFound                       = errors.New("point earning cap not found")
	ErrUniquePointEarningCapMinimumPurchaseViolation = errors.New("minimum_purchase: already exists")

	MapPointEarningCapSortings = map[string]struct{}{
		SortingFieldID: {},
		PointEarningCapSortingFieldMinimumPurchase: {},
	}
)

const (
	PointEarningCapSortingFieldMinimumPurchase = "minimum_purchase"
)

type (
	PointEarningCap struct {
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

	PointEarningCapFilter struct {
		PointEarningCapIDs,
		ExcludedPointEarningCapIDs []uuid.UUID
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

func NewPointEarningCapFilter() *PointEarningCapFilter {
	return new(PointEarningCapFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *PointEarningCapFilter) WithPointEarningCapIDs(pointEarningIDs ...uuid.UUID) *PointEarningCapFilter {
	q.PointEarningCapIDs = pointEarningIDs
	return q
}

func (q *PointEarningCapFilter) WithExcludedPointEarningCapIDs(excludedPointEarningIDs ...uuid.UUID) *PointEarningCapFilter {
	q.ExcludedPointEarningCapIDs = excludedPointEarningIDs
	return q
}

func (q *PointEarningCapFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *PointEarningCapFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *PointEarningCapFilter) WithMinimumPurchases(minimumPurchases ...int64) *PointEarningCapFilter {
	q.MinimumPurchases = minimumPurchases
	return q
}

func (q *PointEarningCapFilter) WithKeyword(keyword string) *PointEarningCapFilter {
	q.Keyword = keyword
	return q
}

func (q *PointEarningCapFilter) WithPaginationURL(paginationURL string) *PointEarningCapFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *PointEarningCapFilter) WithSortings(sortings ...Sorting) *PointEarningCapFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *PointEarningCapFilter) WithLimit(limit int64) *PointEarningCapFilter {
	q.Limit = limit
	return q
}

func (q *PointEarningCapFilter) WithPage(page int64) *PointEarningCapFilter {
	q.Page = page
	return q
}

func (q *PointEarningCapFilter) WithUrlValues(urlValues url.Values) *PointEarningCapFilter {
	q.UrlValues = urlValues
	return q
}

func (q *PointEarningCap) Validate() error {
	if q.RawMinimumPurchase < 1 {
		return errors.New("minimum_purchase: requires a positive integer")
	}
	q.MinimumPurchase = &q.RawMinimumPurchase
	if q.EarnedPoints < 1 {
		return errors.New("earned_points: requires a positive integer")
	}
	return nil
}

func (q *PointEarningCap) Populate() *PointEarningCap {
	if q.MinimumPurchase != nil {
		q.RawMinimumPurchase = *q.MinimumPurchase
	}
	return q
}

func (q *PointEarningCap) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}
