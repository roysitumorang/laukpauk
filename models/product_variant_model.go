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

	PriceTypes = []PriceType{PriceTypeAbsolute, PriceTypeRelative}

	MapPriceTypes = map[PriceType]struct{}{
		PriceTypeAbsolute: {},
		PriceTypeRelative: {},
	}
)

const (
	PriceTypeAbsolute PriceType = iota
	PriceTypeRelative
)

type (
	PriceType int

	ProductVariant struct {
		RowNo                   int64        `json:"-" form:"-"`
		ID                      uuid.UUID    `json:"-" form:"-"`
		RecordStatus            RecordStatus `json:"-" form:"-"`
		ProductID               uuid.UUID    `json:"-" form:"-"`
		StockUnit               string       `json:"-" form:"stock_unit"`
		Price                   int64        `json:"-" form:"price"`
		PriceType               PriceType    `json:"-" form:"price_type"`
		PriceVariantPercentage  float64      `json:"-" form:"price_variant_percentage"`
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
		PriceTypes     []PriceType
		PaginationURL  string
		Sortings       []Sorting
		MapSortings    *Map
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

func (q *ProductVariantFilter) WithPriceTypes(priceTypes ...PriceType) *ProductVariantFilter {
	q.PriceTypes = priceTypes
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

func (q *ProductVariant) Validate(productBasePrice int64) error {
	if q.StockUnit = strings.TrimSpace(q.StockUnit); q.StockUnit == "" {
		return errors.New("stock_unit: is required")
	}
	switch q.PriceType {
	case PriceTypeAbsolute:
		q.PriceVariantPercentage = 0
		if q.Price < 1 {
			return errors.New("price: requires a positive integer minimum 1")
		}
	case PriceTypeRelative:
		if q.PriceVariantPercentage < 1 || q.PriceVariantPercentage > 99 {
			return errors.New("price_variant_percentage: requires a positive decimal minimum 0 maximum 99")
		}
		basePrice, err := decimal.New(productBasePrice, 0)
		if err != nil {
			return err
		}
		variantPercentage, err := decimal.NewFromFloat64(q.PriceVariantPercentage)
		if err != nil {
			return err
		}
		if basePrice, err = basePrice.Mul(variantPercentage); err != nil {
			return err
		}
		if basePrice, err = basePrice.Quo(decimal.Hundred); err != nil {
			return err
		}
		price, remainder, err := basePrice.QuoRem(decimal.Hundred)
		if err != nil {
			return err
		}
		if remainder.IsPos() {
			if price, err = price.Add(decimal.One); err != nil {
				return err
			}
			if price, err = price.Mul(decimal.Hundred); err != nil {
				return err
			}
		}
		q.Price, _, _ = price.Ceil(0).Int64(0)
	default:
		return errors.New("price_type: not found")
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

func (q PriceType) Int() int {
	return int(q)
}

func (q PriceType) String() string {
	if q == PriceTypeRelative {
		return "tidak pas"
	}
	return "pas"
}

func (q PriceType) Equal(v PriceType) bool {
	return q.Int() == v.Int()
}
