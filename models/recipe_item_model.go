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
	RowNo             int64        `json:"-" form:"-"`
	ID                uuid.UUID    `json:"-" form:"id"`
	CompanyID         uuid.UUID    `json:"-" form:"-"`
	RecordStatus      RecordStatus `json:"-" form:"-"`
	RecipeID          uuid.UUID    `json:"-" form:"-"`
	ProductID         uuid.UUID    `json:"-" form:"product_id"`
	ProductName       string       `json:"-" form:"-"`
	ProductPhoto      string       `json:"-" form:"-"`
	ProductReadyStock bool         `json:"-" form:"-"`
	Dose              string       `json:"-" form:"dose"`
	Note              string       `json:"-" form:"note"`
	CreatedBy         uuid.UUID    `json:"-" form:"-"`
	CreatedAt         time.Time    `json:"-" form:"-"`
	UpdatedBy         uuid.UUID    `json:"-" form:"-"`
	UpdatedAt         time.Time    `json:"-" form:"-"`
	DeletedBy         *uuid.UUID   `json:"-" form:"-"`
	DeletedAt         *time.Time   `json:"-" form:"-"`
	DeletionReason    *string      `json:"-" form:"-"`
}

func (q *RecipeItem) Validate() error {
	if q.ProductID.Compare(uuid.Nil()) == 0 {
		return errors.New("product_id: is required")
	}
	if q.Dose = strings.TrimSpace(q.Dose); q.Dose == "" {
		return errors.New("dose: is required")
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
	CompanyIDs,
	ProductIDs,
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

func (q *RecipeItemFilter) WithCompanyIDs(companyIDs ...uuid.UUID) *RecipeItemFilter {
	q.CompanyIDs = companyIDs
	return q
}

func (q *RecipeItemFilter) WithProductIDs(productIDs ...uuid.UUID) *RecipeItemFilter {
	q.ProductIDs = productIDs
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
