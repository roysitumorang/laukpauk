package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrPointEarningNotFound = errors.New("point earning not found")
var ErrPointEarningUniqueBuyerIdStartDateEndDate = errors.New("buyer_id, start_date & end_date already exists")
var MapPointEarningSortings = map[string]struct{}{
	SortingFieldID: {},
}

type PointEarning struct {
	RowNo     int64     `json:"-" form:"-"`
	ID        uuid.UUID `json:"-" form:"-"`
	BuyerID   uuid.UUID `json:"-" form:"-"`
	StartDate time.Time `json:"-" form:"-"`
	EndDate   time.Time `json:"-" form:"-"`
	Amount    int       `json:"-" form:"-"`
	CreatedAt time.Time `json:"-" form:"-"`
	UpdatedAt time.Time `json:"-" form:"-"`
}

func (q *PointEarning) Validate() error {
	return nil
}

func (q *PointEarning) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type PointEarningFilter struct {
	PointEarningIDs,
	BuyerIDs []uuid.UUID
	StartDate,
	EndDate time.Time
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewPointEarningFilter() *PointEarningFilter {
	return &PointEarningFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *PointEarningFilter) WithPointEarningIDs(pointEarningIDs ...uuid.UUID) *PointEarningFilter {
	q.PointEarningIDs = pointEarningIDs
	return q
}

func (q *PointEarningFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *PointEarningFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *PointEarningFilter) WithKeyword(keyword string) *PointEarningFilter {
	q.Keyword = keyword
	return q
}

func (q *PointEarningFilter) WithPaginationURL(paginationURL string) *PointEarningFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *PointEarningFilter) WithSortings(sortings ...Sorting) *PointEarningFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *PointEarningFilter) WithLimit(limit int64) *PointEarningFilter {
	q.Limit = limit
	return q
}

func (q *PointEarningFilter) WithPage(page int64) *PointEarningFilter {
	q.Page = page
	return q
}

func (q *PointEarningFilter) WithUrlValues(urlValues url.Values) *PointEarningFilter {
	q.UrlValues = urlValues
	return q
}
