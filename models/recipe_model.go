package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrRecipeNotFound = errors.New("recipe message not found")
var ErrUniqueRecipeNameViolation = errors.New("name: already exists")
var MapRecipeSortings = map[string]struct{}{
	SortingFieldID:   {},
	SortingFieldName: {},
}

type Recipe struct {
	RowNo          int64        `json:"-" form:"-"`
	ID             uuid.UUID    `json:"-" form:"id"`
	RecordStatus   RecordStatus `json:"-" form:"-"`
	Photo          string       `json:"-" form:"photo"`
	Name           *string      `json:"-" form:"name"`
	DeletedName    *string      `json:"-" form:"-"`
	Description    string       `json:"-" form:"description"`
	InstagramLink  string       `json:"-" form:"instagram_link"`
	FacebookLink   string       `json:"-" form:"facebook_link"`
	CreatedBy      uuid.UUID    `json:"-" form:"-"`
	CreatedAt      time.Time    `json:"-" form:"-"`
	UpdatedBy      uuid.UUID    `json:"-" form:"-"`
	UpdatedAt      time.Time    `json:"-" form:"-"`
	DeletedBy      *uuid.UUID   `json:"-" form:"-"`
	DeletedAt      *time.Time   `json:"-" form:"-"`
	DeletionReason *string      `json:"-" form:"-"`
}

func (q *Recipe) Validate() error {
	if q.Name != nil {
		if *q.Name = strings.TrimSpace(*q.Name); *q.Name == "" {
			q.Name = nil
		}
	}
	if q.Name == nil {
		return errors.New("name: is required")
	}
	if q.InstagramLink = strings.TrimSpace(q.InstagramLink); q.InstagramLink == "" {
		return errors.New("instagram_link: is required")
	}
	_, err := url.ParseRequestURI(q.InstagramLink)
	if err != nil {
		return errors.New("instagram_link: is invalid")
	}
	if q.FacebookLink = strings.TrimSpace(q.FacebookLink); q.FacebookLink == "" {
		return errors.New("facebook_link: is required")
	}
	if _, err = url.ParseRequestURI(q.FacebookLink); err != nil {
		return errors.New("facebook_link: is invalid")
	}
	q.Description = strings.TrimSpace(q.Description)
	return nil
}

func (q *Recipe) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Recipe) SetPhoto(photo string) *Recipe {
	q.Photo = photo
	return q
}

func (q *Recipe) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

type RecipeFilter struct {
	RecipeIDs,
	ExcludedRecipeIDs []uuid.UUID
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

func NewRecipeFilter() *RecipeFilter {
	return &RecipeFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *RecipeFilter) WithRecipeIDs(recipeIDs ...uuid.UUID) *RecipeFilter {
	q.RecipeIDs = recipeIDs
	return q
}

func (q *RecipeFilter) WithExcludedRecipeIDs(excludedRecipeIDs ...uuid.UUID) *RecipeFilter {
	q.ExcludedRecipeIDs = excludedRecipeIDs
	return q
}

func (q *RecipeFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *RecipeFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *RecipeFilter) WithNames(names ...string) *RecipeFilter {
	q.Names = names
	return q
}

func (q *RecipeFilter) WithKeyword(keyword string) *RecipeFilter {
	q.Keyword = keyword
	return q
}

func (q *RecipeFilter) WithPaginationURL(paginationURL string) *RecipeFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *RecipeFilter) WithSortings(sortings ...Sorting) *RecipeFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *RecipeFilter) WithLimit(limit int64) *RecipeFilter {
	q.Limit = limit
	return q
}

func (q *RecipeFilter) WithPage(page int64) *RecipeFilter {
	q.Page = page
	return q
}

func (q *RecipeFilter) WithUrlValues(urlValues url.Values) *RecipeFilter {
	q.UrlValues = urlValues
	return q
}
