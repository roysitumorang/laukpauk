package models

import (
	"net/url"
	"time"

	"github.com/google/uuid"
)

const (
	WebsiteBlockHeroPhotoName = "hero-%d%s"
	WebsiteBlockHeroPhotoPath = "./images/landing-page-imgs/hero/%s"

	WebsiteBlockNameHero         WebsiteBlockName = "hero"
	WebsiteBlockNameFeature      WebsiteBlockName = "fitur"
	WebsiteBlockNameWhy          WebsiteBlockName = "why"
	WebsiteBlockNameFreeShopping WebsiteBlockName = "uang-belanja-gratis"
	WebsiteBlockNameCommitment   WebsiteBlockName = "komitmen"
	WebsiteBlockNameKitchenHero  WebsiteBlockName = "pahlawan-dapur"
	WebsiteBlockNameFAQ          WebsiteBlockName = "bantuan"
)

type (
	WebsiteBlockName string
	WebsiteBlock     struct {
		RowNo          int64            `json:"-" form:"-"`
		ID             uuid.UUID        `json:"-" form:"-"`
		RecordStatus   RecordStatus     `json:"-" form:"-"`
		Name           WebsiteBlockName `json:"-" form:"-"`
		Attributes     any              `json:"-" form:"-"`
		CreatedBy      uuid.UUID        `json:"-" form:"-"`
		CreatedAt      time.Time        `json:"-" form:"-"`
		UpdatedBy      uuid.UUID        `json:"-" form:"-"`
		UpdatedAt      time.Time        `json:"-" form:"-"`
		DeletedBy      *uuid.UUID       `json:"-" form:"-"`
		DeletedAt      *time.Time       `json:"-" form:"-"`
		DeletionReason *string          `json:"-" form:"-"`
	}

	WebsiteBlockFilter struct {
		WebsiteBlockIDs []uuid.UUID
		RecordStatus    []RecordStatus
		Names           []WebsiteBlockName
		PaginationURL   string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewWebsiteBlockFilter() *WebsiteBlockFilter {
	return new(WebsiteBlockFilter{UrlValues: url.Values{}})
}

func (q *WebsiteBlockFilter) WithWebsiteBlockIDs(websiteBlockItemIDs ...uuid.UUID) *WebsiteBlockFilter {
	q.WebsiteBlockIDs = websiteBlockItemIDs
	return q
}

func (q *WebsiteBlockFilter) WithRecordStatus(recordStatus ...RecordStatus) *WebsiteBlockFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *WebsiteBlockFilter) WithNames(names ...WebsiteBlockName) *WebsiteBlockFilter {
	q.Names = names
	return q
}

func (q *WebsiteBlockFilter) WithPaginationURL(paginationURL string) *WebsiteBlockFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *WebsiteBlockFilter) WithLimit(limit int64) *WebsiteBlockFilter {
	q.Limit = limit
	return q
}

func (q *WebsiteBlockFilter) WithPage(page int64) *WebsiteBlockFilter {
	q.Page = page
	return q
}

func (q *WebsiteBlockFilter) WithUrlValues(urlValues url.Values) *WebsiteBlockFilter {
	q.UrlValues = urlValues
	return q
}

func (q *WebsiteBlock) Validate() error {
	return nil
}
