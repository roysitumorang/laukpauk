package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
)

var (
	ErrUniqueProductCategoryNameViolation = errors.New("name: already exists")
)

const (
	ProductCategoryPhotoName         = "product-category-%d%s"
	ProductCategoryPhotoPath         = "./images/product_category/%s"
	ProductCategoryPhotoCompressPath = "./images/product_category/compress/%s"
)

type (
	ProductCategory struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		Name           string       `json:"-" form:"name"`
		Description    string       `json:"-" form:"description"`
		Photo          string       `json:"-" form:"photo"`
		CreatedBy      uuid.UUID    `json:"-" form:"-"`
		CreatedAt      time.Time    `json:"-" form:"-"`
		UpdatedBy      uuid.UUID    `json:"-" form:"-"`
		UpdatedAt      time.Time    `json:"-" form:"-"`
		DeletedBy      *uuid.UUID   `json:"-" form:"-"`
		DeletedAt      *time.Time   `json:"-" form:"-"`
		DeletionReason *string      `json:"-" form:"-"`
	}

	ProductCategoryFilter struct {
		ProductCategoryIDs []uuid.UUID
		RecordStatus       []RecordStatus
		Names              []string
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewProductCategoryFilter() *ProductCategoryFilter {
	return new(ProductCategoryFilter{UrlValues: url.Values{}})
}

func (q *ProductCategoryFilter) WithProductCategoryIDs(productCategoryIDs ...uuid.UUID) *ProductCategoryFilter {
	q.ProductCategoryIDs = productCategoryIDs
	return q
}

func (q *ProductCategoryFilter) WithRecordStatus(recordStatus ...RecordStatus) *ProductCategoryFilter {
	q.RecordStatus = recordStatus
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

func (q *ProductCategory) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	return nil
}

func (q *ProductCategory) SetPhoto(photo string) *ProductCategory {
	q.Photo = photo
	return q
}
