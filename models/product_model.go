package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrProductNotFound = errors.New("product not found")
var ErrUniqueProductNameViolation = errors.New("name: already exists")
var ProductTypes = []ProductType{
	ProductTypeStandalone,
	ProductTypeBundled,
}
var MapProductTypes = map[ProductType]struct{}{
	ProductTypeStandalone: {},
	ProductTypeBundled:    {},
}
var MapProductSortings = map[string]struct{}{
	SortingFieldID:   {},
	SortingFieldName: {},
}

const (
	ProductTypeStandalone ProductType = iota
	ProductTypeBundled
)

type ProductType int

func (q ProductType) Int() int {
	return int(q)
}

func (q ProductType) String() string {
	if q == ProductTypeBundled {
		return "paket"
	}
	return "non paket"
}

type Product struct {
	RowNo               int64             `json:"-" form:"-"`
	ID                  uuid.UUID         `json:"id" form:"id"`
	RecordStatus        RecordStatus      `json:"-" form:"-"`
	ProductCategoryID   uuid.UUID         `json:"category_id" form:"product_category_id"`
	ProductCategoryName string            `json:"category_name" form:"-"`
	ProductType         ProductType       `json:"product_type" form:"product_type"`
	ProductTypeName     string            `json:"product_type_name" form:"-"`
	Name                *string           `json:"name" form:"name"`
	DeletedName         *string           `json:"-" form:"-"`
	Photo               string            `json:"photo" form:"photo"`
	Description         string            `json:"description" form:"description"`
	ReadyStock          bool              `json:"ready_stock" form:"ready_stock"`
	Pinned              bool              `json:"-" form:"pinned"`
	Hits                int64             `json:"-" form:"-"`
	CreatedBy           uuid.UUID         `json:"-" form:"-"`
	CreatedAt           time.Time         `json:"-" form:"-"`
	UpdatedBy           uuid.UUID         `json:"-" form:"-"`
	UpdatedAt           time.Time         `json:"-" form:"-"`
	DeletedBy           *uuid.UUID        `json:"-" form:"-"`
	DeletedAt           *time.Time        `json:"-" form:"-"`
	DeletionReason      *string           `json:"-" form:"-"`
	Variants            []*ProductVariant `json:"variants" form:"-"`
}

func (q *Product) Validate() error {
	if q.ProductCategoryID.Compare(uuid.Nil()) == 0 {
		return errors.New("product_category_id: is required")
	}
	if q.Name != nil {
		if *q.Name = strings.TrimSpace(*q.Name); *q.Name == "" {
			q.Name = nil
		}
	}
	if q.Name == nil {
		return errors.New("name: is required")
	}
	if _, ok := MapProductTypes[q.ProductType]; !ok {
		return errors.New("product_type: invalid")
	}
	q.Description = strings.TrimSpace(q.Description)
	return nil
}

func (q *Product) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Product) SetPhoto(photo string) *Product {
	q.Photo = photo
	return q
}

func (q *Product) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

func (q *Product) StockStatus() string {
	if !q.ReadyStock {
		return "tidak tersedia"
	}
	return "tersedia"
}

func (q *Product) Populate() *Product {
	q.ProductTypeName = q.ProductType.String()
	return q
}

type ProductFilter struct {
	ProductCategoryIDs,
	ProductIDs,
	ExcludedProductIDs []uuid.UUID
	RecordStatuses []RecordStatus
	ProductTypes   []ProductType
	Names          []string
	ReadyStock     *bool
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewProductFilter() *ProductFilter {
	return &ProductFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
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

func (q *ProductFilter) WithNames(names ...*string) *ProductFilter {
	q.Names = []string{}
	for _, name := range names {
		if name != nil {
			q.Names = append(q.Names, *name)
		}
	}
	return q
}

func (q *ProductFilter) WithReadyStock(readyStock *bool) *ProductFilter {
	q.ReadyStock = readyStock
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

func (q *ProductFilter) WithSortings(sortings ...Sorting) *ProductFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
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
