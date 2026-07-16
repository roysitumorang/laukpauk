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
	ErrCartItemNotFound = errors.New("cart item not found")
)

type (
	CartItem struct {
		RowNo            int64     `json:"-" form:"-"`
		ID               uuid.UUID `json:"id" form:"-"`
		CartID           uuid.UUID `json:"-" form:"-"`
		ProductVariantID uuid.UUID `json:"product_variant_id" form:"-"`
		Quantity         int64     `json:"quantity" form:"-"`
		CreatedAt        time.Time `json:"-" form:"-"`
		UpdatedAt        time.Time `json:"-" form:"-"`
	}

	CartItemFilter struct {
		CartItemIDs,
		CartIDs,
		ProductVariantIDs []uuid.UUID
		PaginationURL string
		Sortings      []Sorting
		MapSortings   Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewCartItemFilter() *CartItemFilter {
	return &CartItemFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *CartItemFilter) WithCartItemIDs(cartItemIDs ...uuid.UUID) *CartItemFilter {
	q.CartItemIDs = cartItemIDs
	return q
}

func (q *CartItemFilter) WithCartIDs(cartIDs ...uuid.UUID) *CartItemFilter {
	q.CartIDs = cartIDs
	return q
}

func (q *CartItemFilter) WithProductVariantIDs(productVariantIDs ...uuid.UUID) *CartItemFilter {
	q.ProductVariantIDs = productVariantIDs
	return q
}

func (q *CartItemFilter) WithPaginationURL(paginationURL string) *CartItemFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *CartItemFilter) WithSortings(sortings ...Sorting) *CartItemFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *CartItemFilter) WithLimit(limit int64) *CartItemFilter {
	q.Limit = limit
	return q
}

func (q *CartItemFilter) WithPage(page int64) *CartItemFilter {
	q.Page = page
	return q
}

func (q *CartItemFilter) WithUrlValues(urlValues url.Values) *CartItemFilter {
	q.UrlValues = urlValues
	return q
}

func (q *CartItem) Validate() error {
	if q.ProductVariantID.Compare(uuid.Nil()) == 0 {
		return errors.New("product_variant_id: is required")
	}
	q.Quantity = max(q.Quantity, 1)
	return nil
}

func (q *CartItem) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}
