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

	CartItem struct {
		RowNo            int64     `json:"-" form:"-"`
		ID               uuid.UUID `json:"id" form:"-"`
		CartID           uuid.UUID `json:"-" form:"-"`
		ProductVariantID uuid.UUID `json:"product_variant_id" form:"-"`
		Quantity         int64     `json:"quantity" form:"-"`
		CreatedAt        time.Time `json:"-" form:"-"`
		UpdatedAt        time.Time `json:"-" form:"-"`
	}

	CartFilter struct {
		BuyerIDs      []uuid.UUID
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

func (q *CartFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *CartFilter {
	q.BuyerIDs = buyerIDs
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
	for i, cartItem := range q.Items {
		if cartItem.ProductVariantID.Compare(uuid.Nil()) == 0 {
			return fmt.Errorf("items[%d].product_variant_id: is required", i)
		}
		cartItem.Quantity = max(cartItem.Quantity, 0)
		q.Items[i] = cartItem
	}
	return nil
}

func (q *Cart) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}
