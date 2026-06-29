package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
)

var (
	ErrProductNotFound            = errors.New("product not found")
	ErrUniqueProductNameViolation = errors.New("name: already exists")

	ProductTypes = []ProductType{ProductTypeStandalone, ProductTypeBundled}

	MapProductTypes = map[ProductType]struct{}{
		ProductTypeStandalone: {},
		ProductTypeBundled:    {},
	}

	StockStatuses    = []StockStatus{StockStatusInStock, StockStatusOutOfStock}
	MapStockStatuses = map[StockStatus]struct{}{
		StockStatusInStock:    {},
		StockStatusOutOfStock: {},
	}
)

const (
	ProductPhotoName         = "product-%d%s"
	ProductPhotoPath         = "./images/product/%s"
	ProductPhotoCompressPath = "./images/product/compress/%s"

	SortProductsByID   = "id"
	SortProductsByName = "name"
)

const (
	ProductTypeStandalone ProductType = iota
	ProductTypeBundled
)
const (
	StockStatusInStock StockStatus = iota
	StockStatusOutOfStock
)

type (
	ProductType int
	StockStatus int

	Product struct {
		RowNo             int64           `json:"-" form:"-"`
		ID                uuid.UUID       `json:"-" form:"-"`
		RecordStatus      RecordStatus    `json:"-" form:"-"`
		ProductCategoryID uuid.UUID       `json:"-" form:"product_category_id"`
		ProductCategory   ProductCategory `json:"-" form:"-"`
		ProductType       ProductType     `json:"-" form:"product_type"`
		Name              string          `json:"-" form:"name"`
		Photo             string          `json:"-" form:"photo"`
		Description       string          `json:"-" form:"description"`
		StockStatus       StockStatus     `json:"-" form:"stock_status"`
		Pinned            bool            `json:"-" form:"pinned"`
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
		ProductIDs,
		ExcludedProductIDs []uuid.UUID
		RecordStatuses []RecordStatus
		ProductTypes   []ProductType
		Names          []string
		StockStatuses  []StockStatus
		Keyword,
		PaginationURL string
		SortBy []SortBy
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

func (q *ProductFilter) WithExcludedProductIDs(excludedProductIDs ...uuid.UUID) *ProductFilter {
	q.ExcludedProductIDs = excludedProductIDs
	return q
}

func (q *ProductFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *ProductFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *ProductFilter) WithProductTypes(productTypes ...ProductType) *ProductFilter {
	q.ProductTypes = productTypes
	return q
}

func (q *ProductFilter) WithNames(names ...string) *ProductFilter {
	q.Names = names
	return q
}

func (q *ProductFilter) WithStockStatus(stockStatuses ...StockStatus) *ProductFilter {
	q.StockStatuses = stockStatuses
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

func (q *ProductFilter) WithSortBy(sortBy ...SortBy) *ProductFilter {
	q.SortBy = sortBy
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
	if q.ProductCategoryID.Compare(uuid.Nil()) == 0 {
		return errors.New("product_category_id: is required")
	}
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if _, ok := MapProductTypes[q.ProductType]; !ok {
		return errors.New("product_type: invalid")
	}
	if _, ok := MapStockStatuses[q.StockStatus]; !ok {
		return errors.New("stock_status: invalid")
	}
	q.Description = strings.TrimSpace(q.Description)
	return nil
}

func (q *Product) SetPhoto(photo string) *Product {
	q.Photo = photo
	return q
}

func (q ProductType) Int() int {
	return int(q)
}

func (q ProductType) String() string {
	if q == ProductTypeBundled {
		return "paket"
	}
	return "non paket"
}

func (q StockStatus) Int() int {
	return int(q)
}

func (q StockStatus) String() string {
	if q == StockStatusOutOfStock {
		return "tidak tersedia"
	}
	return "tersedia"
}
