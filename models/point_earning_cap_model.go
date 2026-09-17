package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrPointEarningCapNotFound = errors.New("point earning cap not found")
var ErrPointEarningCapShoppingCostOverlapping = errors.New("shopping_cost: already exists")
var MapPointEarningCapSortings = map[string]struct{}{
	SortingFieldID: {},
}

type PointEarningCap struct {
	RowNo               int64        `json:"-" form:"-"`
	ID                  uuid.UUID    `json:"-" form:"id"`
	CompanyID           uuid.UUID    `json:"-" form:"-"`
	RecordStatus        RecordStatus `json:"-" form:"-"`
	MinimumShoppingCost int64        `json:"-" form:"minimum_shopping_cost"`
	MaximumShoppingCost int64        `json:"-" form:"maximum_shopping_cost"`
	EarnedPoints        int64        `json:"-" form:"earned_points"`
	CreatedBy           uuid.UUID    `json:"-" form:"-"`
	CreatedAt           time.Time    `json:"-" form:"-"`
	UpdatedBy           uuid.UUID    `json:"-" form:"-"`
	UpdatedAt           time.Time    `json:"-" form:"-"`
	DeletedBy           *uuid.UUID   `json:"-" form:"-"`
	DeletedAt           *time.Time   `json:"-" form:"-"`
	DeletionReason      *string      `json:"-" form:"-"`
}

func (q *PointEarningCap) Validate() error {
	if q.MinimumShoppingCost < 1 {
		return errors.New("minimum_shopping_cost: requires a positive integer")
	}
	if q.MaximumShoppingCost < 1 {
		return errors.New("maximum_shopping_cost: requires a positive integer")
	}
	if q.MinimumShoppingCost >= q.MaximumShoppingCost {
		return errors.New("maximum_shopping_cost: should exceed minimum_shopping_cost")
	}
	if q.EarnedPoints < 1 {
		return errors.New("earned_points: requires a positive integer")
	}
	return nil
}

func (q *PointEarningCap) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type PointEarningCapFilter struct {
	PointEarningCapIDs,
	ExcludedPointEarningCapIDs,
	CompanyIDs []uuid.UUID
	RecordStatuses []RecordStatus
	MinimumShoppingCost,
	MaximumShoppingCost int64
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewPointEarningCapFilter() *PointEarningCapFilter {
	return &PointEarningCapFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *PointEarningCapFilter) WithPointEarningCapIDs(pointEarningIDs ...uuid.UUID) *PointEarningCapFilter {
	q.PointEarningCapIDs = pointEarningIDs
	return q
}

func (q *PointEarningCapFilter) WithExcludedPointEarningCapIDs(excludedPointEarningIDs ...uuid.UUID) *PointEarningCapFilter {
	q.ExcludedPointEarningCapIDs = excludedPointEarningIDs
	return q
}

func (q *PointEarningCapFilter) WithCompanyIDs(companyIDs ...uuid.UUID) *PointEarningCapFilter {
	q.CompanyIDs = companyIDs
	return q
}

func (q *PointEarningCapFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *PointEarningCapFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *PointEarningCapFilter) WithShoppingCost(minimumShoppingCost, maximumShoppingCost int64) *PointEarningCapFilter {
	q.MinimumShoppingCost = minimumShoppingCost
	q.MaximumShoppingCost = maximumShoppingCost
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
