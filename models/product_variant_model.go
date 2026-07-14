package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrProductVariantNotFound                 = errors.New("product variant not found")
	ErrUniqueProductVariantStockUnitViolation = errors.New("stock_unit: already exists")
)

type (
	ProductVariant struct {
		RowNo                   int64        `json:"-" form:"-"`
		ID                      uuid.UUID    `json:"-" form:"-"`
		RecordStatus            RecordStatus `json:"-" form:"-"`
		ProductID               uuid.UUID    `json:"-" form:"-"`
		StockUnit               string       `json:"-" form:"stock_unit"`
		Price                   int64        `json:"-" form:"price"`
		DiscountType            DiscountType `json:"-" form:"discount_type"`
		DiscountRate            int64        `json:"-" form:"discount_rate"`
		DiscountMinimumQuantity int64        `json:"-" form:"discount_minimum_quantity"`
		DiscountAmount          int64        `json:"-" form:"-"`
		SalePrice               int64        `json:"-" form:"-"`
		Description             string       `json:"-" form:"description"`
		CreatedBy               uuid.UUID    `json:"-" form:"-"`
		CreatedAt               time.Time    `json:"-" form:"-"`
		UpdatedBy               uuid.UUID    `json:"-" form:"-"`
		UpdatedAt               time.Time    `json:"-" form:"-"`
		DeletedBy               *uuid.UUID   `json:"-" form:"-"`
		DeletedAt               *time.Time   `json:"-" form:"-"`
		DeletionReason          *string      `json:"-" form:"-"`
	}

	ProductVariantFilter struct {
		ProductIDs,
		ProductVariantIDs,
		ExcludedProductVariantIDs []uuid.UUID
		RecordStatuses []RecordStatus
		StockUnits     []string
		PaginationURL  string
		Sortings       []Sorting
		MapSortings    Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewProductVariantFilter() *ProductVariantFilter {
	return new(ProductVariantFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *ProductVariantFilter) WithProductIDs(productIDs ...uuid.UUID) *ProductVariantFilter {
	q.ProductIDs = productIDs
	return q
}

func (q *ProductVariantFilter) WithProductVariantIDs(productVariantIDs ...uuid.UUID) *ProductVariantFilter {
	q.ProductVariantIDs = productVariantIDs
	return q
}

func (q *ProductVariantFilter) WithExcludedProductVariantIDs(excludedProductVariantIDs ...uuid.UUID) *ProductVariantFilter {
	q.ExcludedProductVariantIDs = excludedProductVariantIDs
	return q
}

func (q *ProductVariantFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *ProductVariantFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *ProductVariantFilter) WithStockUnits(stockUnits ...string) *ProductVariantFilter {
	q.StockUnits = stockUnits
	return q
}

func (q *ProductVariantFilter) WithPaginationURL(paginationURL string) *ProductVariantFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *ProductVariantFilter) WithSortings(sortings ...Sorting) *ProductVariantFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *ProductVariantFilter) WithLimit(limit int64) *ProductVariantFilter {
	q.Limit = limit
	return q
}

func (q *ProductVariantFilter) WithPage(page int64) *ProductVariantFilter {
	q.Page = page
	return q
}

func (q *ProductVariantFilter) WithUrlValues(urlValues url.Values) *ProductVariantFilter {
	q.UrlValues = urlValues
	return q
}

func (q *ProductVariant) Validate() error {
	if q.StockUnit = strings.TrimSpace(q.StockUnit); q.StockUnit == "" {
		return errors.New("stock_unit: is required")
	}
	if q.Price < 1 {
		return errors.New("price: requires a positive integer minimum 1")
	}
	switch q.DiscountType {
	case DiscountTypeFixedAmount:
		if q.DiscountRate < 0 {
			return errors.New("discount_rate: requires a positive integer minimum 0")
		}
		if q.DiscountRate >= q.Price {
			return errors.New("discount_rate: cannot equal or exceed price")
		}
		q.DiscountAmount = q.DiscountRate
		q.SalePrice = q.Price - q.DiscountAmount
	case DiscountTypePercentage:
		if q.DiscountRate < 0 || q.DiscountRate > 99 {
			return errors.New("discount_rate: requires a positive integer minimum 0 maximum 99")
		}
		price, err := decimal.New(q.Price, 0)
		if err != nil {
			return err
		}
		discountRate, err := decimal.New(q.DiscountRate, 0)
		if err != nil {
			return err
		}
		discount, err := price.Mul(discountRate)
		if err != nil {
			return err
		}
		if discount, err = discount.Quo(decimal.Hundred); err != nil {
			return err
		}
		q.DiscountAmount, _, _ = discount.Ceil(0).Int64(0)
		q.SalePrice = q.Price - q.DiscountAmount
	default:
		return errors.New("discount_type: not found")
	}
	if q.DiscountMinimumQuantity < 0 {
		return errors.New("discount_minimum_quantity: requires a positive integer minimum 0")
	}
	return nil
}
