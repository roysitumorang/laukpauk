package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
)

const (
	BlogPhotoName         = "blog-%d%s"
	BlogPhotoPath         = "./images/blog/%s"
	BlogPhotoCompressPath = "./images/blog/compress/%s"
)

type (
	Blog struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		Subject        string       `json:"-" form:"subject"`
		Body           string       `json:"-" form:"body"`
		Slug           string       `json:"-" form:"slug"`
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
		BlogIDs      []uuid.UUID
		RecordStatus []RecordStatus
		Slugs        []string
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewBlogFilter() *BlogFilter {
	return new(BlogFilter{UrlValues: url.Values{}})
}

func (q *BlogFilter) WithBlogIDs(blogIDs ...uuid.UUID) *BlogFilter {
	q.BlogIDs = blogIDs
	return q
}

func (q *BlogFilter) WithRecordStatus(recordStatus ...RecordStatus) *BlogFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *BlogFilter) WithSlugs(slugs ...string) *BlogFilter {
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
	if q.Slug = strings.TrimSpace(q.Slug); q.Slug == "" {
		return errors.New("slug: is required")
	}
	if !SlugRegex.MatchString(q.Slug) {
		return errors.New("slug: invalid characters")
	}
	return nil
}
