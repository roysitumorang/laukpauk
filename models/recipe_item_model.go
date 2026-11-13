package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrRecipeItemNotFound = errors.New("recipe item not found")
var ErrUniqueRecipeItemProductIdViolation = errors.New("product_id: already exists")
var MapRecipeItemSortings = map[string]struct{}{
	SortingFieldID: {},
}

type RecipeItem struct {
	RowNo                   int64        `json:"-" form:"-"`
	ID                      uuid.UUID    `json:"id" form:"id"`
	RecordStatus            RecordStatus `json:"record_status" form:"-"`
	RecipeID                uuid.UUID    `json:"recipe_id" form:"-"`
	ProductID               uuid.UUID    `json:"product_id" form:"-"`
	ProductName             string       `json:"product_name" form:"-"`
	ProductPhoto            string       `json:"product_photo" form:"-"`
	ProductReadyStock       bool         `json:"product_ready_stock" form:"-"`
	ProductVariantID        uuid.UUID    `json:"product_variant_id" form:"product_variant_id"`
	ProductVariantStockUnit string       `json:"product_variant_stock_unit" form:"-"`
	Quantity                int64        `json:"quantity" form:"quantity"`
	Note                    string       `json:"note" form:"note"`
	CreatedBy               uuid.UUID    `json:"created_by" form:"-"`
	CreatedAt               time.Time    `json:"created_at" form:"-"`
	UpdatedBy               uuid.UUID    `json:"updated_by" form:"-"`
	UpdatedAt               time.Time    `json:"updated_at" form:"-"`
	DeletedBy               *uuid.UUID   `json:"deleted_by" form:"-"`
	DeletedAt               *time.Time   `json:"deleted_at" form:"-"`
	DeletionReason          *string      `json:"deletion_reason" form:"-"`
}

func (q *RecipeItem) Validate() error {
	if q.ProductVariantID.Compare(uuid.Nil()) == 0 {
		return errors.New("product_variant_id: is required")
	}
	if q.Quantity < 1 {
		return errors.New("quantity: is required")
	}
	q.Note = strings.TrimSpace(q.Note)
	return nil
}

func (q *RecipeItem) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *RecipeItem) ProductPhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.ProductPhoto, baseURLs...)
}

type RecipeItemFilter struct {
	RecipeIDs,
	ProductIDs,
	ProductVariantIDs,
	RecipeItemIDs,
	ExcludedRecipeItemIDs []uuid.UUID
	ProductReadyStock *bool
	RecordStatuses    []RecordStatus
	PaginationURL     string
	Sortings          []Sorting
	MapSortings       Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewRecipeItemFilter() *RecipeItemFilter {
	return &RecipeItemFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *RecipeItemFilter) WithRecipeIDs(recipeIDs ...uuid.UUID) *RecipeItemFilter {
	q.RecipeIDs = recipeIDs
	return q
}

func (q *RecipeItemFilter) WithProductIDs(productIDs ...uuid.UUID) *RecipeItemFilter {
	q.ProductIDs = productIDs
	return q
}

func (q *RecipeItemFilter) WithProductVariantIDs(productVariantIDs ...uuid.UUID) *RecipeItemFilter {
	q.ProductVariantIDs = productVariantIDs
	return q
}

func (q *RecipeItemFilter) WithRecipeItemIDs(recipeGuideIDs ...uuid.UUID) *RecipeItemFilter {
	q.RecipeItemIDs = recipeGuideIDs
	return q
}

func (q *RecipeItemFilter) WithExcludedRecipeItemIDs(excludedRecipeItemIDs ...uuid.UUID) *RecipeItemFilter {
	q.ExcludedRecipeItemIDs = excludedRecipeItemIDs
	return q
}

func (q *RecipeItemFilter) WithProductReadyStock(productReadyStock *bool) *RecipeItemFilter {
	q.ProductReadyStock = productReadyStock
	return q
}

func (q *RecipeItemFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *RecipeItemFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *RecipeItemFilter) WithPaginationURL(paginationURL string) *RecipeItemFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *RecipeItemFilter) WithSortings(sortings ...Sorting) *RecipeItemFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *RecipeItemFilter) WithLimit(limit int64) *RecipeItemFilter {
	q.Limit = limit
	return q
}

func (q *RecipeItemFilter) WithPage(page int64) *RecipeItemFilter {
	q.Page = page
	return q
}

func (q *RecipeItemFilter) WithUrlValues(urlValues url.Values) *RecipeItemFilter {
	q.UrlValues = urlValues
	return q
}
