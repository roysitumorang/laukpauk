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

var ErrProductVariantNotFound = errors.New("product variant not found")
var ErrUniqueProductVariantStockUnitViolation = errors.New("stock_unit: already exists")

type ProductVariant struct {
	RowNo                   int64        `json:"-" form:"-"`
	ID                      uuid.UUID    `json:"id" form:"id"`
	CompanyID               uuid.UUID    `json:"-" form:"-"`
	RecordStatus            RecordStatus `json:"-" form:"-"`
	ProductCategoryID       uuid.UUID    `json:"-" form:"-"`
	ProductCategoryName     string       `json:"-" form:"-"`
	ProductID               uuid.UUID    `json:"-" form:"-"`
	ProductName             string       `json:"-" form:"-"`
	ProductReadyStock       bool         `json:"-" form:"-"`
	StockUnit               string       `json:"stock_unit" form:"stock_unit"`
	BasePrice               int64        `json:"-" form:"base_price"`
	SellingPrice            int64        `json:"selling_price" form:"selling_price"`
	DiscountType            DiscountType `json:"discount_type" form:"discount_type"`
	DiscountTypeName        string       `json:"discount_type_name" form:"-"`
	DiscountRate            int64        `json:"discount_rate" form:"discount_rate"`
	DiscountMinimumQuantity int64        `json:"discount_minimum_quantity" form:"discount_minimum_quantity"`
	DiscountAmount          int64        `json:"discount_amount" form:"-"`
	SalePrice               int64        `json:"sale_price" form:"-"`
	Description             string       `json:"description" form:"description"`
	CreatedBy               uuid.UUID    `json:"-" form:"-"`
	CreatedAt               time.Time    `json:"-" form:"-"`
	UpdatedBy               uuid.UUID    `json:"-" form:"-"`
	UpdatedAt               time.Time    `json:"-" form:"-"`
	DeletedBy               *uuid.UUID   `json:"-" form:"-"`
	DeletedAt               *time.Time   `json:"-" form:"-"`
	DeletionReason          *string      `json:"-" form:"-"`
}

func (q *ProductVariant) Validate() error {
	if q.StockUnit = strings.TrimSpace(q.StockUnit); q.StockUnit == "" {
		return errors.New("stock_unit: is required")
	}
	if q.BasePrice < 1 {
		return errors.New("base_price: requires a positive integer minimum 1")
	}
	if q.SellingPrice < 1 {
		return errors.New("selling_price: requires a positive integer minimum 1")
	}
	if q.BasePrice > q.SellingPrice {
		return errors.New("selling_price: should exceed or equal base_price")
	}
	switch q.DiscountType {
	case DiscountTypeFixedAmount:
		if q.DiscountRate < 0 {
			return errors.New("discount_rate: requires a positive integer minimum 0")
		}
		if q.DiscountRate >= q.SellingPrice {
			return errors.New("discount_rate: cannot equal or exceed price")
		}
		q.DiscountAmount = q.DiscountRate
		q.SalePrice = q.SellingPrice - q.DiscountAmount
	case DiscountTypePercentage:
		if q.DiscountRate < 0 || q.DiscountRate > 99 {
			return errors.New("discount_rate: requires a positive integer minimum 0 maximum 99")
		}
		sellingPrice, err := decimal.New(q.SellingPrice, 0)
		if err != nil {
			return err
		}
		discountRate, err := decimal.New(q.DiscountRate, 0)
		if err != nil {
			return err
		}
		discount, err := sellingPrice.Mul(discountRate)
		if err != nil {
			return err
		}
		if discount, err = discount.Quo(decimal.Hundred); err != nil {
			return err
		}
		q.DiscountAmount, _, _ = discount.Ceil(0).Int64(0)
		q.SalePrice = q.SellingPrice - q.DiscountAmount
	default:
		return errors.New("discount_type: not found")
	}
	if q.DiscountMinimumQuantity < 0 {
		return errors.New("discount_minimum_quantity: requires a positive integer minimum 0")
	}
	return nil
}

func (q *ProductVariant) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *ProductVariant) Populate() *ProductVariant {
	q.DiscountTypeName = q.DiscountType.String()
	return q
}

type ProductVariantFilter struct {
	CompanyIDs,
	ProductCategoryIDs,
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

func NewProductVariantFilter() *ProductVariantFilter {
	return &ProductVariantFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *ProductVariantFilter) WithProductCategoryIDs(productCategoryIDs ...uuid.UUID) *ProductVariantFilter {
	q.ProductCategoryIDs = productCategoryIDs
	return q
}

func (q *ProductVariantFilter) WithCompanyIDs(companyIDs ...uuid.UUID) *ProductVariantFilter {
	q.CompanyIDs = companyIDs
	return q
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
