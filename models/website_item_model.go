package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	WebsiteItemHeroPhotoName = "hero-%d%s"
	WebsiteItemHeroPhotoPath = "./images/landing-page-imgs/hero/%s"

	WebsiteItemTypeHero         WebsiteItemType = "hero"
	WebsiteItemTypeFeature      WebsiteItemType = "fitur"
	WebsiteItemTypeWhy          WebsiteItemType = "why"
	WebsiteItemTypeFreeShopping WebsiteItemType = "uang-belanja-gratis"
	WebsiteItemTypeCommitment   WebsiteItemType = "komitmen"
	WebsiteItemTypeKitchenHero  WebsiteItemType = "pahlawan-dapur"
	WebsiteItemTypeKitchenFAQ   WebsiteItemType = "bantuan"
)

type (
	WebsiteItemType string
	WebsiteItem     struct {
		RowNo           int64           `json:"-" form:"-"`
		ID              uuid.UUID       `json:"-" form:"-"`
		RecordStatus    RecordStatus    `json:"-" form:"-"`
		WebsiteItemType WebsiteItemType `json:"-" form:"-"`
		Subject         string          `json:"-" form:"subject"`
		Body            *string         `json:"-" form:"body"`
		Photo           *string         `json:"-" form:"photo"`
		LinkText        *string         `json:"-" form:"link_text"`
		LinkURL         *string         `json:"-" form:"link_url"`
		CreatedBy       uuid.UUID       `json:"-" form:"-"`
		CreatedAt       time.Time       `json:"-" form:"-"`
		UpdatedBy       uuid.UUID       `json:"-" form:"-"`
		UpdatedAt       time.Time       `json:"-" form:"-"`
		DeletedBy       *uuid.UUID      `json:"-" form:"-"`
		DeletedAt       *time.Time      `json:"-" form:"-"`
		DeletionReason  *string         `json:"-" form:"-"`
	}

	WebsiteItemFilter struct {
		WebsiteItemIDs   []uuid.UUID
		RecordStatus     []RecordStatus
		WebsiteItemTypes []WebsiteItemType
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewWebsiteItemFilter() *WebsiteItemFilter {
	return new(WebsiteItemFilter{UrlValues: url.Values{}})
}

func (q *WebsiteItemFilter) WithWebsiteItemIDs(websiteItemIDs ...uuid.UUID) *WebsiteItemFilter {
	q.WebsiteItemIDs = websiteItemIDs
	return q
}

func (q *WebsiteItemFilter) WithRecordStatus(recordStatus ...RecordStatus) *WebsiteItemFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *WebsiteItemFilter) WithWebsiteItemTypes(websiteItemTypes ...WebsiteItemType) *WebsiteItemFilter {
	q.WebsiteItemTypes = websiteItemTypes
	return q
}

func (q *WebsiteItemFilter) WithKeyword(keyword string) *WebsiteItemFilter {
	q.Keyword = keyword
	return q
}

func (q *WebsiteItemFilter) WithPaginationURL(paginationURL string) *WebsiteItemFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *WebsiteItemFilter) WithLimit(limit int64) *WebsiteItemFilter {
	q.Limit = limit
	return q
}

func (q *WebsiteItemFilter) WithPage(page int64) *WebsiteItemFilter {
	q.Page = page
	return q
}

func (q *WebsiteItemFilter) WithUrlValues(urlValues url.Values) *WebsiteItemFilter {
	q.UrlValues = urlValues
	return q
}

func (q *WebsiteItem) Validate() error {
	switch q.WebsiteItemType {
	case WebsiteItemTypeHero:
		if q.Subject = strings.TrimSpace(q.Subject); q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == nil ||
			strings.TrimSpace(*q.Body) == "" {
			return errors.New("body: is required")
		}
		if q.LinkText == nil ||
			strings.TrimSpace(*q.LinkText) == "" {
			return errors.New("link_text: is required")
		}
		if q.LinkURL == nil ||
			strings.TrimSpace(*q.LinkURL) == "" {
			return errors.New("link_url: is required")
		}
		if _, err := url.ParseRequestURI(*q.LinkURL); err != nil {
			return fmt.Errorf("link_url: %w", err)
		}
		if q.ID == uuid.Nil &&
			(q.Photo == nil ||
				*q.Photo == "") {
			return errors.New("photo: is required")
		}
	}
	return nil
}
