package models

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ArticlePhotoName         = "artikel-%d%s"
	ArticlePhotoPath         = "./images/artikel/%s"
	ArticlePhotoCompressPath = "./images/artikel/compress/%s"
)

type (
	Article struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		Subject        string       `json:"-" form:"subject"`
		Body           string       `json:"-" form:"body"`
		Photo          string       `json:"-" form:"photo"`
		Link           string       `json:"-" form:"link"`
		CreatedBy      uuid.UUID    `json:"-" form:"-"`
		CreatedAt      time.Time    `json:"-" form:"-"`
		UpdatedBy      uuid.UUID    `json:"-" form:"-"`
		UpdatedAt      time.Time    `json:"-" form:"-"`
		DeletedBy      *uuid.UUID   `json:"-" form:"-"`
		DeletedAt      *time.Time   `json:"-" form:"-"`
		DeletionReason *string      `json:"-" form:"-"`
	}

	ArticleFilter struct {
		ArticleIDs   []uuid.UUID
		RecordStatus []RecordStatus
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewArticleFilter() *ArticleFilter {
	return new(ArticleFilter{UrlValues: url.Values{}})
}

func (q *ArticleFilter) WithArticleIDs(articleIDs ...uuid.UUID) *ArticleFilter {
	q.ArticleIDs = articleIDs
	return q
}

func (q *ArticleFilter) WithRecordStatus(recordStatus ...RecordStatus) *ArticleFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *ArticleFilter) WithKeyword(keyword string) *ArticleFilter {
	q.Keyword = keyword
	return q
}

func (q *ArticleFilter) WithPaginationURL(paginationURL string) *ArticleFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *ArticleFilter) WithLimit(limit int64) *ArticleFilter {
	q.Limit = limit
	return q
}

func (q *ArticleFilter) WithPage(page int64) *ArticleFilter {
	q.Page = page
	return q
}

func (q *ArticleFilter) WithUrlValues(urlValues url.Values) *ArticleFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Article) Validate() error {
	if q.Subject = strings.TrimSpace(q.Subject); q.Subject == "" {
		return errors.New("subject: is required")
	}
	if q.Link = strings.TrimSpace(q.Link); q.Link == "" {
		return errors.New("link: is required")
	}
	if _, err := url.ParseRequestURI(q.Link); err != nil {
		return err
	}
	return nil
}
