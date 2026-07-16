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
	ErrCartNotFound = errors.New("cart not found")
)

type (
	Cart struct {
		RowNo          int64       `json:"-" form:"-"`
		ID             uuid.UUID   `json:"id" form:"-"`
		BuyerID        *uuid.UUID  `json:"buyer_id" form:"-"`
		DeletedBuyerID *uuid.UUID  `json:"-" form:"-"`
		Quantity       int64       `json:"quantity" form:"-"`
		CreatedAt      time.Time   `json:"-" form:"-"`
		UpdatedAt      time.Time   `json:"-" form:"-"`
		Items          []*CartItem `json:"items" form:"-"`
	}

	CartFilter struct {
		CartIDs       []uuid.UUID
		PaginationURL string
		Sortings      []Sorting
		MapSortings   Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewCartFilter() *CartFilter {
	return &CartFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *CartFilter) WithCartIDs(cartIDs ...uuid.UUID) *CartFilter {
	q.CartIDs = cartIDs
	return q
}

func (q *CartFilter) WithPaginationURL(paginationURL string) *CartFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *CartFilter) WithSortings(sortings ...Sorting) *CartFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *CartFilter) WithLimit(limit int64) *CartFilter {
	q.Limit = limit
	return q
}

func (q *CartFilter) WithPage(page int64) *CartFilter {
	q.Page = page
	return q
}

func (q *CartFilter) WithUrlValues(urlValues url.Values) *CartFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Cart) Validate() error {
	if q.BuyerID.Compare(uuid.Nil()) == 0 {
		return errors.New("buyer_id: is required")
	}
	return nil
}

func (q *Cart) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}
