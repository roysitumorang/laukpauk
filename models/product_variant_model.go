package models

import (
	"errors"
	"net/url"
	"time"
	"uuid"
)

type (
	ProductVariant struct {
		RowNo                   int64        `json:"-" form:"-"`
		ID                      uuid.UUID    `json:"-" form:"-"`
		RecordStatus            RecordStatus `json:"-" form:"-"`
		ProductID               uuid.UUID    `json:"-" form:"-"`
		Product                 Product      `json:"-" form:"-"`
		MetricUnitID            uuid.UUID    `json:"-" form:"-"`
		MetricUnit              MetricUnit   `json:"-" form:"-"`
		MeasureMin              float64      `json:"-" form:"-"`
		MeasureMax              float64      `json:"-" form:"-"`
		Price                   int64        `json:"-" form:"-"`
		DiscountType            DiscountType `json:"-" form:"-"`
		DiscountAmount          int64        `json:"-" form:"-"`
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
		RecordStatus  []RecordStatus
		PaginationURL string
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

func (q *ProductVariantFilter) WithRecordStatus(recordStatus ...RecordStatus) *ProductVariantFilter {
	q.RecordStatus = recordStatus
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
	q.MeasureMin = max(0, q.MeasureMin)
	q.MeasureMax = max(0, q.MeasureMax)
	q.Price = max(0, q.Price)
	q.DiscountAmount = max(0, q.DiscountAmount)
	q.DiscountMinimumQuantity = max(0, q.DiscountMinimumQuantity)
	if q.MeasureMax == 0 {
		return errors.New("measure_max: is required")
	}
	if q.MeasureMax < q.MeasureMin {
		return errors.New("measure_max: should exceed measure_min")
	}
	if q.Price == 0 {
		return errors.New("price: is required")
	}
	return nil
}

func (q *ProductVariant) HumanizeDiscountType() string {
	switch q.DiscountType {
	case DiscountTypePercentage:
		return "persentase"
	case DiscountTypeFlat:
		fallthrough
	default:
		return "cash"
	}
}
