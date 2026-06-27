package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
)

const (
	ProductTypeStandalone ProductType = iota
	ProductTypeBundled
)
const (
	StockStatusOutOfStock StockStatus = iota
	StockStatusInStock
)

type (
	ProductType int
	StockStatus int

	Product struct {
		RowNo             int64           `json:"-" form:"-"`
		ID                uuid.UUID       `json:"-" form:"-"`
		RecordStatus      RecordStatus    `json:"-" form:"-"`
		ProductCategoryID uuid.UUID       `json:"-" form:"-"`
		ProductCategory   ProductCategory `json:"-" form:"-"`
		ProductType       ProductType     `json:"-" form:"-"`
		Name              string          `json:"-" form:"-"`
		Photo             string          `json:"-" form:"-"`
		Description       string          `json:"-" form:"-"`
		StockStatus       StockStatus     `json:"-" form:"-"`
		Pinned            bool            `json:"-" form:"-"`
		CreatedBy         uuid.UUID       `json:"-" form:"-"`
		CreatedAt         time.Time       `json:"-" form:"-"`
		UpdatedBy         uuid.UUID       `json:"-" form:"-"`
		UpdatedAt         time.Time       `json:"-" form:"-"`
		DeletedBy         *uuid.UUID      `json:"-" form:"-"`
		DeletedAt         *time.Time      `json:"-" form:"-"`
		DeletionReason    *string         `json:"-" form:"-"`
	}

	ProductFilter struct {
		ProductCategoryIDs,
		ProductIDs []uuid.UUID
		RecordStatus []RecordStatus
		ProductTypes []ProductType
		StockStatus  []StockStatus
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewProductFilter() *ProductFilter {
	return new(ProductFilter{UrlValues: url.Values{}})
}

func (q *ProductFilter) WithProductCategoryIDs(productCategoryIDs ...uuid.UUID) *ProductFilter {
	q.ProductCategoryIDs = productCategoryIDs
	return q
}

func (q *ProductFilter) WithProductIDs(productIDs ...uuid.UUID) *ProductFilter {
	q.ProductIDs = productIDs
	return q
}

func (q *ProductFilter) WithRecordStatus(recordStatus ...RecordStatus) *ProductFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *ProductFilter) WithProductTypes(productTypes ...ProductType) *ProductFilter {
	q.ProductTypes = productTypes
	return q
}

func (q *ProductFilter) WithStockStatus(stockStatus ...StockStatus) *ProductFilter {
	q.StockStatus = stockStatus
	return q
}

func (q *ProductFilter) WithKeyword(keyword string) *ProductFilter {
	q.Keyword = keyword
	return q
}

func (q *ProductFilter) WithPaginationURL(paginationURL string) *ProductFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *ProductFilter) WithLimit(limit int64) *ProductFilter {
	q.Limit = limit
	return q
}

func (q *ProductFilter) WithPage(page int64) *ProductFilter {
	q.Page = page
	return q
}

func (q *ProductFilter) WithUrlValues(urlValues url.Values) *ProductFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Product) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	q.Description = strings.TrimSpace(q.Description)
	return nil
}

func (q *Product) HumanizeProductType() string {
	switch q.ProductType {
	case ProductTypeBundled:
		return "paket"
	case ProductTypeStandalone:
		fallthrough
	default:
		return "non paket"
	}
}

func (q *Product) HumanizeStockStatus() string {
	switch q.StockStatus {
	case StockStatusInStock:
		return "tersedia"
	case StockStatusOutOfStock:
		fallthrough
	default:
		return "tidak tersedia"
	}
}
