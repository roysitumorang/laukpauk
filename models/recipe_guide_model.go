package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrRecipeGuideNotFound = errors.New("recipe guide not found")
var MapRecipeGuideSortings = map[string]struct{}{
	SortingFieldID: {},
}

type RecipeGuide struct {
	RowNo          int64        `json:"-" form:"-"`
	ID             uuid.UUID    `json:"id" form:"id"`
	RecordStatus   RecordStatus `json:"record_status" form:"-"`
	RecipeID       uuid.UUID    `json:"recipe_id" form:"-"`
	Instruction    string       `json:"instruction" form:"instruction"`
	CreatedBy      uuid.UUID    `json:"created_by" form:"-"`
	CreatedAt      time.Time    `json:"created_at" form:"-"`
	UpdatedBy      uuid.UUID    `json:"updated_by" form:"-"`
	UpdatedAt      time.Time    `json:"updated_at" form:"-"`
	DeletedBy      *uuid.UUID   `json:"deleted_by" form:"-"`
	DeletedAt      *time.Time   `json:"deleted_at" form:"-"`
	DeletionReason *string      `json:"deletion_reason" form:"-"`
}

func (q *RecipeGuide) Validate() error {
	if q.Instruction = strings.TrimSpace(q.Instruction); q.Instruction == "" {
		return errors.New("instruction: is required")
	}
	return nil
}

func (q *RecipeGuide) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type RecipeGuideFilter struct {
	RecipeIDs,
	RecipeGuideIDs,
	ExcludedRecipeGuideIDs []uuid.UUID
	RecordStatuses []RecordStatus
	PaginationURL  string
	Sortings       []Sorting
	MapSortings    Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewRecipeGuideFilter() *RecipeGuideFilter {
	return &RecipeGuideFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *RecipeGuideFilter) WithRecipeIDs(recipeIDs ...uuid.UUID) *RecipeGuideFilter {
	q.RecipeIDs = recipeIDs
	return q
}

func (q *RecipeGuideFilter) WithRecipeGuideIDs(recipeGuideIDs ...uuid.UUID) *RecipeGuideFilter {
	q.RecipeGuideIDs = recipeGuideIDs
	return q
}

func (q *RecipeGuideFilter) WithExcludedRecipeGuideIDs(excludedRecipeGuideIDs ...uuid.UUID) *RecipeGuideFilter {
	q.ExcludedRecipeGuideIDs = excludedRecipeGuideIDs
	return q
}

func (q *RecipeGuideFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *RecipeGuideFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *RecipeGuideFilter) WithPaginationURL(paginationURL string) *RecipeGuideFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *RecipeGuideFilter) WithSortings(sortings ...Sorting) *RecipeGuideFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *RecipeGuideFilter) WithLimit(limit int64) *RecipeGuideFilter {
	q.Limit = limit
	return q
}

func (q *RecipeGuideFilter) WithPage(page int64) *RecipeGuideFilter {
	q.Page = page
	return q
}

func (q *RecipeGuideFilter) WithUrlValues(urlValues url.Values) *RecipeGuideFilter {
	q.UrlValues = urlValues
	return q
}
