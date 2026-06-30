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
	ErrProductVariantNotFound                   = errors.New("product variant not found")
	ErrUniqueProductVariantMeasureUnitViolation = errors.New("measure_unit: already exists")

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
		Product                 *Product     `json:"-" form:"-"`
		MeasureUnit             string       `json:"-" form:"measure_unit"`
		Price                   int64        `json:"-" form:"price"`
		PriceType               PriceType    `json:"-" form:"price_type"`
		PriceVariantPercentage  float64      `json:"-" form:"price_variant_percentage"`
		DiscountType            DiscountType `json:"-" form:"discount_type"`
		DiscountRate            int64        `json:"-" form:"discount_rate"`
		DiscountMinimumQuantity int64        `json:"-" form:"discount_minimum_quantity"`
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
		MeasureUnits   []string
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

func (q *ProductVariantFilter) WithMeasureUnits(measureUnits ...string) *ProductVariantFilter {
	q.MeasureUnits = measureUnits
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
	q.Price = max(0, q.Price)
	if q.PriceType == PriceTypeRelative {
		q.PriceVariantPercentage = min(max(0, q.PriceVariantPercentage), 100)
	} else {
		q.PriceVariantPercentage = 100
	}
	q.DiscountRate = max(0, q.DiscountRate)
	q.DiscountMinimumQuantity = max(0, q.DiscountMinimumQuantity)
	if q.MeasureUnit = strings.TrimSpace(q.MeasureUnit); q.MeasureUnit == "" {
		return errors.New("measure_unit: is required")
	}
	if q.Price == 0 {
		return errors.New("price: is required")
	}
	if _, ok := MapPriceTypes[q.PriceType]; !ok {
		return errors.New("price_type: not found")
	}
	if _, ok := MapDiscountTypes[q.DiscountType]; !ok {
		return errors.New("discount_type: not found")
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
