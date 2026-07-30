package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrBlogNotFound            = errors.New("blog not found")
	ErrUniqueBlogSlugViolation = errors.New("slug: already exists")
	MapBlogSortings            = map[string]struct{}{
		SortingFieldID:          {},
		BlogSortingFieldSubject: {},
	}
)

const (
	BlogSortingFieldSubject = "subject"
)

type (
	Blog struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		Subject        string       `json:"-" form:"subject"`
		Body           string       `json:"-" form:"body"`
		Slug           *string      `json:"-" form:"slug"`
		DeletedSlug    *string      `json:"-" form:"-"`
		Photo          string       `json:"-" form:"photo"`
		CreatedBy      uuid.UUID    `json:"-" form:"-"`
		CreatedAt      time.Time    `json:"-" form:"-"`
		UpdatedBy      uuid.UUID    `json:"-" form:"-"`
		UpdatedAt      time.Time    `json:"-" form:"-"`
		DeletedBy      *uuid.UUID   `json:"-" form:"-"`
		DeletedAt      *time.Time   `json:"-" form:"-"`
		DeletionReason *string      `json:"-" form:"-"`
	}

	BlogFilter struct {
		BlogIDs,
		ExcludedBlogIDs []uuid.UUID
		RecordStatuses []RecordStatus
		Slugs          []*string
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewBlogFilter() *BlogFilter {
	return &BlogFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *BlogFilter) WithBlogIDs(blogIDs ...uuid.UUID) *BlogFilter {
	q.BlogIDs = blogIDs
	return q
}

func (q *BlogFilter) WithExcludedBlogIDs(excludedBlogIDs ...uuid.UUID) *BlogFilter {
	q.ExcludedBlogIDs = excludedBlogIDs
	return q
}

func (q *BlogFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *BlogFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *BlogFilter) WithSlugs(slugs ...*string) *BlogFilter {
	q.Slugs = slugs
	return q
}

func (q *BlogFilter) WithKeyword(keyword string) *BlogFilter {
	q.Keyword = keyword
	return q
}

func (q *BlogFilter) WithPaginationURL(paginationURL string) *BlogFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *BlogFilter) WithSortings(sortings ...Sorting) *BlogFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *BlogFilter) WithLimit(limit int64) *BlogFilter {
	q.Limit = limit
	return q
}

func (q *BlogFilter) WithPage(page int64) *BlogFilter {
	q.Page = page
	return q
}

func (q *BlogFilter) WithUrlValues(urlValues url.Values) *BlogFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Blog) Validate() error {
	if q.Subject = strings.TrimSpace(q.Subject); q.Subject == "" {
		return errors.New("subject: is required")
	}
	if q.Slug == nil || strings.TrimSpace(*q.Slug) == "" {
		return errors.New("slug: is required")
	}
	if !SlugRegex.MatchString(*q.Slug) {
		return errors.New("slug: invalid characters")
	}
	return nil
}

func (q *Blog) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Blog) SetPhoto(photo string) *Blog {
	q.Photo = photo
	return q
}

func (q *Blog) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}
