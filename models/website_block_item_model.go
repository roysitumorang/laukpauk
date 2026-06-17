package models

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	WebsiteBlockItem struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		WebsiteBlockID uuid.UUID    `json:"-" form:"-"`
		Subject        string       `json:"-" form:"subject"`
		Body           string       `json:"-" form:"body"`
		Photo          string       `json:"-" form:"photo"`
		LinkText       string       `json:"-" form:"link_text"`
		LinkURL        string       `json:"-" form:"link_url"`
		CreatedBy      uuid.UUID    `json:"-" form:"-"`
		CreatedAt      time.Time    `json:"-" form:"-"`
		UpdatedBy      uuid.UUID    `json:"-" form:"-"`
		UpdatedAt      time.Time    `json:"-" form:"-"`
		DeletedBy      *uuid.UUID   `json:"-" form:"-"`
		DeletedAt      *time.Time   `json:"-" form:"-"`
		DeletionReason *string      `json:"-" form:"-"`
	}

	WebsiteBlockItemFilter struct {
		WebsiteBlockItemIDs,
		WebsiteBlockIDs []uuid.UUID
		WebsiteBlockNames []WebsiteBlockName
		RecordStatus      []RecordStatus
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewWebsiteBlockItemFilter() *WebsiteBlockItemFilter {
	return new(WebsiteBlockItemFilter{UrlValues: url.Values{}})
}

func (q *WebsiteBlockItemFilter) WithWebsiteBlockItemIDs(websiteBlockItemIDs ...uuid.UUID) *WebsiteBlockItemFilter {
	q.WebsiteBlockItemIDs = websiteBlockItemIDs
	return q
}

func (q *WebsiteBlockItemFilter) WithWebsiteBlockIDs(websiteBlockIDs ...uuid.UUID) *WebsiteBlockItemFilter {
	q.WebsiteBlockIDs = websiteBlockIDs
	return q
}

func (q *WebsiteBlockItemFilter) WithWebsiteBlockNames(websiteBlockNames ...WebsiteBlockName) *WebsiteBlockItemFilter {
	q.WebsiteBlockNames = websiteBlockNames
	return q
}

func (q *WebsiteBlockItemFilter) WithRecordStatus(recordStatus ...RecordStatus) *WebsiteBlockItemFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *WebsiteBlockItemFilter) WithKeyword(keyword string) *WebsiteBlockItemFilter {
	q.Keyword = keyword
	return q
}

func (q *WebsiteBlockItemFilter) WithPaginationURL(paginationURL string) *WebsiteBlockItemFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *WebsiteBlockItemFilter) WithLimit(limit int64) *WebsiteBlockItemFilter {
	q.Limit = limit
	return q
}

func (q *WebsiteBlockItemFilter) WithPage(page int64) *WebsiteBlockItemFilter {
	q.Page = page
	return q
}

func (q *WebsiteBlockItemFilter) WithUrlValues(urlValues url.Values) *WebsiteBlockItemFilter {
	q.UrlValues = urlValues
	return q
}

func (q *WebsiteBlockItem) Validate(name WebsiteBlockName) error {
	switch name {
	case WebsiteBlockNameHero:
		if q.Subject = strings.TrimSpace(q.Subject); q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body = strings.TrimSpace(q.Body); q.Body == "" {
			return errors.New("body: is required")
		}
		if q.LinkText = strings.TrimSpace(q.LinkText); q.LinkText == "" {
			return errors.New("link_text: is required")
		}
		if q.LinkURL = strings.TrimSpace(q.LinkURL); q.LinkURL == "" {
			return errors.New("link_url: is required")
		}
		if _, err := url.ParseRequestURI(q.LinkURL); err != nil {
			return errors.New("link_url: invalid URL")
		}
		if q.ID == uuid.Nil &&
			q.Photo == "" {
			return errors.New("photo: is required")
		}
	}
	return nil
}
