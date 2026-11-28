package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/helper"
)

var ErrCartNotFound = errors.New("cart not found")

type Cart struct {
	RowNo         int64       `json:"-"`
	ID            uuid.UUID   `json:"id"`
	UserID        *uuid.UUID  `json:"user_id"`
	DeletedUserID *uuid.UUID  `json:"-"`
	Quantity      int64       `json:"quantity"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Items         []*CartItem `json:"items"`
}

func (q *Cart) Validate() error {
	for i, cartItem := range q.Items {
		if cartItem.ProductVariantID == uuid.Nil {
			return fmt.Errorf("items[%d].product_variant_id: is required", i)
		}
		cartItem.Quantity = max(cartItem.Quantity, 0)
		cartItem.Note = strings.TrimSpace(cartItem.Note)
		q.Items[i] = cartItem
	}
	return nil
}

func (q *Cart) New() bool {
	return q.ID == uuid.Nil
}

type CartItem struct {
	RowNo                   int64        `json:"-"`
	ID                      uuid.UUID    `json:"id"`
	CartID                  uuid.UUID    `json:"-"`
	ProductVariantID        uuid.UUID    `json:"product_variant_id"`
	ProductCategoryID       uuid.UUID    `json:"product_category_id"`
	ProductCategoryName     string       `json:"product_category_name"`
	ProductID               uuid.UUID    `json:"product_id"`
	ProductName             string       `json:"product_name"`
	ProductPhoto            string       `json:"product_photo"`
	ProductReadyStock       bool         `json:"product_ready_stock"`
	StockUnit               string       `json:"stock_unit"`
	BasePrice               int64        `json:"base_price,omitempty"`
	SellingPrice            int64        `json:"selling_price"`
	DiscountType            DiscountType `json:"discount_type"`
	DiscountTypeName        string       `json:"discount_type_name"`
	DiscountRate            int64        `json:"discount_rate"`
	DiscountMinimumQuantity int64        `json:"discount_minimum_quantity"`
	DiscountAmount          int64        `json:"discount_amount"`
	SalePrice               int64        `json:"sale_price"`
	Description             string       `json:"description"`
	Quantity                int64        `json:"quantity"`
	Note                    string       `json:"note"`
	CreatedAt               time.Time    `json:"-"`
	UpdatedAt               time.Time    `json:"-"`
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
