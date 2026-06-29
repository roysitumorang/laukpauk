package models

import (
	"errors"
	"net/url"
	"time"
	"uuid"
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
		Product                 Product      `json:"-" form:"-"`
		MeasureUnit             string       `json:"-" form:"-"`
		Price                   int64        `json:"-" form:"-"`
		PriceType               PriceType    `json:"-" form:"-"`
		PriceVariantPercentage  float64      `json:"-" form:"-"`
		DiscountType            DiscountType `json:"-" form:"-"`
		DiscountRate            int64        `json:"-" form:"-"`
		DiscountMinimumQuantity int64        `json:"-" form:"-"`
		Description             string       `json:"-" form:"-"`
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
		ProductVariantIDs []uuid.UUID
		RecordStatuses []RecordStatus
		PaginationURL  string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewProductVariantFilter() *ProductVariantFilter {
	return new(ProductVariantFilter{UrlValues: url.Values{}})
}

func (q *ProductVariantFilter) WithProductIDs(productIDs ...uuid.UUID) *ProductVariantFilter {
	q.ProductIDs = productIDs
	return q
}

func (q *ProductVariantFilter) WithProductVariantIDs(productVariantIDs ...uuid.UUID) *ProductVariantFilter {
	q.ProductVariantIDs = productVariantIDs
	return q
}

func (q *ProductVariantFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *ProductVariantFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *ProductVariantFilter) WithPaginationURL(paginationURL string) *ProductVariantFilter {
	q.PaginationURL = paginationURL
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
	if q.Price == 0 {
		return errors.New("price: is required")
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
