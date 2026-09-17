package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrPointSpendingCapNotFound = errors.New("point spending cap not found")
var ErrPointSpendingCapShoppingCostOverlapping = errors.New("shopping_cost: already exists")
var MapPointSpendingCapSortings = map[string]struct{}{
	SortingFieldID: {},
}

type PointSpendingCap struct {
	RowNo               int64        `json:"-" form:"-"`
	ID                  uuid.UUID    `json:"-" form:"id"`
	CompanyID           uuid.UUID    `json:"-" form:"-"`
	RecordStatus        RecordStatus `json:"-" form:"-"`
	MinimumShoppingCost int64        `json:"minimum_shopping_cost" form:"minimum_shopping_cost"`
	MaximumShoppingCost int64        `json:"maximum_shopping_cost" form:"maximum_shopping_cost"`
	DisposablePoints    int64        `json:"disposable_points" form:"disposable_points"`
	CreatedBy           uuid.UUID    `json:"-" form:"-"`
	CreatedAt           time.Time    `json:"-" form:"-"`
	UpdatedBy           uuid.UUID    `json:"-" form:"-"`
	UpdatedAt           time.Time    `json:"-" form:"-"`
	DeletedBy           *uuid.UUID   `json:"-" form:"-"`
	DeletedAt           *time.Time   `json:"-" form:"-"`
	DeletionReason      *string      `json:"-" form:"-"`
}

func (q *PointSpendingCap) Validate() error {
	if q.MinimumShoppingCost < 1 {
		return errors.New("minimum_shopping_cost: requires a positive integer")
	}
	if q.MaximumShoppingCost < 1 {
		return errors.New("maximum_shopping_cost: requires a positive integer")
	}
	if q.MinimumShoppingCost >= q.MaximumShoppingCost {
		return errors.New("maximum_shopping_cost: should exceed minimum_shopping_cost")
	}
	if q.DisposablePoints < 1 {
		return errors.New("disposable_points: requires a positive integer")
	}
	return nil
}

func (q *PointSpendingCap) Populate() *PointSpendingCap {
	return q
}

func (q *PointSpendingCap) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type PointSpendingCapFilter struct {
	PointSpendingCapIDs,
	ExcludedPointSpendingCapIDs,
	CompanyIDs []uuid.UUID
	MinimumShoppingCost,
	MaximumShoppingCost int64
	RecordStatuses []RecordStatus
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewPointSpendingCapFilter() *PointSpendingCapFilter {
	return &PointSpendingCapFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *PointSpendingCapFilter) WithPointSpendingCapIDs(pointSpendingIDs ...uuid.UUID) *PointSpendingCapFilter {
	q.PointSpendingCapIDs = pointSpendingIDs
	return q
}

func (q *PointSpendingCapFilter) WithExcludedPointSpendingCapIDs(excludedPointSpendingIDs ...uuid.UUID) *PointSpendingCapFilter {
	q.ExcludedPointSpendingCapIDs = excludedPointSpendingIDs
	return q
}

func (q *PointSpendingCapFilter) WithCompanyIDs(companyIDs ...uuid.UUID) *PointSpendingCapFilter {
	q.CompanyIDs = companyIDs
	return q
}

func (q *PointSpendingCapFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *PointSpendingCapFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *PointSpendingCapFilter) WithShoppingCost(minimumShoppingCost, maximumShoppingCost int64) *PointSpendingCapFilter {
	q.MinimumShoppingCost = minimumShoppingCost
	q.MaximumShoppingCost = maximumShoppingCost
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
