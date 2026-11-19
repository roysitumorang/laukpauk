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

var ErrCartNotFound = errors.New("cart not found")

type Cart struct {
	RowNo         int64       `json:"-" form:"-"`
	ID            uuid.UUID   `json:"id" form:"-"`
	UserID        *uuid.UUID  `json:"user_id" form:"-"`
	DeletedUserID *uuid.UUID  `json:"-" form:"-"`
	Quantity      int64       `json:"quantity" form:"-"`
	CreatedAt     time.Time   `json:"-" form:"-"`
	UpdatedAt     time.Time   `json:"-" form:"-"`
	Items         []*CartItem `json:"items" form:"-"`
}

func (q *Cart) Validate() error {
	for i, cartItem := range q.Items {
		if cartItem.ProductVariantID.Compare(uuid.Nil()) == 0 {
			return fmt.Errorf("items[%d].product_variant_id: is required", i)
		}
		cartItem.Quantity = max(cartItem.Quantity, 0)
		cartItem.Note = strings.TrimSpace(cartItem.Note)
		q.Items[i] = cartItem
	}
	return nil
}

func (q *Cart) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type CartItem struct {
	RowNo                   int64        `json:"-" form:"-"`
	ID                      uuid.UUID    `json:"id" form:"-"`
	CartID                  uuid.UUID    `json:"-" form:"-"`
	ProductVariantID        uuid.UUID    `json:"product_variant_id" form:"-"`
	ProductCategoryID       uuid.UUID    `json:"product_category_id" form:"-"`
	ProductCategoryName     string       `json:"product_category_name" form:"-"`
	ProductID               uuid.UUID    `json:"product_id" form:"-"`
	ProductName             string       `json:"product_name" form:"-"`
	ProductPhoto            string       `json:"product_photo" form:"-"`
	ProductReadyStock       bool         `json:"product_ready_stock" form:"-"`
	StockUnit               string       `json:"stock_unit" form:"stock_unit"`
	BasePrice               int64        `json:"base_price,omitempty" form:"base_price"`
	SellingPrice            int64        `json:"selling_price" form:"selling_price"`
	DiscountType            DiscountType `json:"discount_type" form:"discount_type"`
	DiscountTypeName        string       `json:"discount_type_name" form:"-"`
	DiscountRate            int64        `json:"discount_rate" form:"discount_rate"`
	DiscountMinimumQuantity int64        `json:"discount_minimum_quantity" form:"discount_minimum_quantity"`
	DiscountAmount          int64        `json:"discount_amount" form:"-"`
	SalePrice               int64        `json:"sale_price" form:"-"`
	Description             string       `json:"description" form:"description"`
	Quantity                int64        `json:"quantity" form:"-"`
	Note                    string       `json:"note" form:"-"`
	CreatedAt               time.Time    `json:"-" form:"-"`
	UpdatedAt               time.Time    `json:"-" form:"-"`
}

func (q *CartItem) ProductPhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.ProductPhoto, baseURLs...)
}

type CartFilter struct {
	UserIDs       []uuid.UUID
	PaginationURL string
	Sortings      []Sorting
	MapSortings   Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewCartFilter() *CartFilter {
	return &CartFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *CartFilter) WithUserIDs(userIDs ...uuid.UUID) *CartFilter {
	q.UserIDs = userIDs
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
