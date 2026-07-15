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
	ErrPointSpendingCapNotFound                       = errors.New("point spending cap not found")
	ErrUniquePointSpendingCapMinimumPurchaseViolation = errors.New("minimum_purchase: already exists")

	MapPointSpendingCapSortings = map[string]struct{}{
		SortingFieldID: {},
		PointSpendingCapSortingFieldMinimumPurchase: {},
	}
)

const (
	PointSpendingCapSortingFieldMinimumPurchase = "minimum_purchase"
)

type (
	PointSpendingCap struct {
		RowNo                  int64        `json:"-" form:"-"`
		ID                     uuid.UUID    `json:"-" form:"id"`
		RecordStatus           RecordStatus `json:"-" form:"-"`
		RawMinimumPurchase     int64        `json:"-" form:"minimum_purchase"`
		MinimumPurchase        *int64       `json:"-" form:"-"`
		DeletedMinimumPurchase *int64       `json:"-" form:"-"`
		DisposablePoints       int64        `json:"-" form:"disposable_points"`
		CreatedBy              uuid.UUID    `json:"-" form:"-"`
		CreatedAt              time.Time    `json:"-" form:"-"`
		UpdatedBy              uuid.UUID    `json:"-" form:"-"`
		UpdatedAt              time.Time    `json:"-" form:"-"`
		DeletedBy              *uuid.UUID   `json:"-" form:"-"`
		DeletedAt              *time.Time   `json:"-" form:"-"`
		DeletionReason         *string      `json:"-" form:"-"`
	}

	PointSpendingCapFilter struct {
		PointSpendingCapIDs,
		ExcludedPointSpendingCapIDs []uuid.UUID
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

func NewPointSpendingCapFilter() *PointSpendingCapFilter {
	return new(PointSpendingCapFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *PointSpendingCapFilter) WithPointSpendingCapIDs(pointSpendingIDs ...uuid.UUID) *PointSpendingCapFilter {
	q.PointSpendingCapIDs = pointSpendingIDs
	return q
}

func (q *PointSpendingCapFilter) WithExcludedPointSpendingCapIDs(excludedPointSpendingIDs ...uuid.UUID) *PointSpendingCapFilter {
	q.ExcludedPointSpendingCapIDs = excludedPointSpendingIDs
	return q
}

func (q *PointSpendingCapFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *PointSpendingCapFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *PointSpendingCapFilter) WithMinimumPurchases(minimumPurchases ...int64) *PointSpendingCapFilter {
	q.MinimumPurchases = minimumPurchases
	return q
}

func (q *PointSpendingCapFilter) WithKeyword(keyword string) *PointSpendingCapFilter {
	q.Keyword = keyword
	return q
}

func (q *PointSpendingCapFilter) WithPaginationURL(paginationURL string) *PointSpendingCapFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *PointSpendingCapFilter) WithSortings(sortings ...Sorting) *PointSpendingCapFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *PointSpendingCapFilter) WithLimit(limit int64) *PointSpendingCapFilter {
	q.Limit = limit
	return q
}

func (q *PointSpendingCapFilter) WithPage(page int64) *PointSpendingCapFilter {
	q.Page = page
	return q
}

func (q *PointSpendingCapFilter) WithUrlValues(urlValues url.Values) *PointSpendingCapFilter {
	q.UrlValues = urlValues
	return q
}

func (q *PointSpendingCap) Validate() error {
	if q.RawMinimumPurchase < 1 {
		return errors.New("minimum_purchase: requires a positive integer")
	}
	q.MinimumPurchase = &q.RawMinimumPurchase
	if q.DisposablePoints < 1 {
		return errors.New("disposable_points: requires a positive integer")
	}
	return nil
}

func (q *PointSpendingCap) Populate() *PointSpendingCap {
	if q.MinimumPurchase != nil {
		q.RawMinimumPurchase = *q.MinimumPurchase
	}
	return q
}

func (q *PointSpendingCap) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}
