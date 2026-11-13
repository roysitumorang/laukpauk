package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrProductCategoryNotFound = errors.New("product category not found")
var ErrUniqueProductCategoryNameViolation = errors.New("name: already exists")
var MapProductCategorySortings = map[string]struct{}{
	SortingFieldID:   {},
	SortingFieldName: {},
}

const SortProductCategoriesByID = "id"
const SortProductCategoriesByName = "name"

type ProductCategory struct {
	RowNo          int64        `json:"-" form:"-"`
	ID             uuid.UUID    `json:"id" form:"id"`
	RecordStatus   RecordStatus `json:"record_status" form:"-"`
	Name           string       `json:"name" form:"name"`
	Description    string       `json:"description" form:"description"`
	Photo          string       `json:"photo" form:"photo"`
	CreatedBy      uuid.UUID    `json:"created_by" form:"-"`
	CreatedAt      time.Time    `json:"created_at" form:"-"`
	UpdatedBy      uuid.UUID    `json:"updated_by" form:"-"`
	UpdatedAt      time.Time    `json:"updated_at" form:"-"`
	DeletedBy      *uuid.UUID   `json:"deleted_by" form:"-"`
	DeletedAt      *time.Time   `json:"deleted_at" form:"-"`
	DeletionReason *string      `json:"deletion_reason" form:"-"`
}

func (q *ProductCategory) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	return nil
}

func (q *ProductCategory) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *ProductCategory) SetPhoto(photo string) *ProductCategory {
	q.Photo = photo
	return q
}

func (q *ProductCategory) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

type ProductCategoryFilter struct {
	ProductCategoryIDs,
	ExcludedProductCategoryIDs []uuid.UUID
	RecordStatuses []RecordStatus
	Names          []string
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewProductCategoryFilter() *ProductCategoryFilter {
	return &ProductCategoryFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *ProductCategoryFilter) WithProductCategoryIDs(productCategoryIDs ...uuid.UUID) *ProductCategoryFilter {
	q.ProductCategoryIDs = productCategoryIDs
	return q
}

func (q *ProductCategoryFilter) WithExcludedProductCategoryIDs(excludedProductCategoryIDs ...uuid.UUID) *ProductCategoryFilter {
	q.ExcludedProductCategoryIDs = excludedProductCategoryIDs
	return q
}

func (q *ProductCategoryFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *ProductCategoryFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *ProductCategoryFilter) WithNames(names ...string) *ProductCategoryFilter {
	q.Names = names
	return q
}

func (q *ProductCategoryFilter) WithKeyword(keyword string) *ProductCategoryFilter {
	q.Keyword = keyword
	return q
}

func (q *ProductCategoryFilter) WithPaginationURL(paginationURL string) *ProductCategoryFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *ProductCategoryFilter) WithSortings(sortings ...Sorting) *ProductCategoryFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *ProductCategoryFilter) WithLimit(limit int64) *ProductCategoryFilter {
	q.Limit = limit
	return q
}

func (q *ProductCategoryFilter) WithPage(page int64) *ProductCategoryFilter {
	q.Page = page
	return q
}

func (q *ProductCategoryFilter) WithUrlValues(urlValues url.Values) *ProductCategoryFilter {
	q.UrlValues = urlValues
	return q
}
