package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
)

const (
	BazaarTypePublic  BazaarType = "public"
	BazaarTypePrivate BazaarType = "private"

	BazaarPhotoName         = "bazaar-%d%s"
	BazaarPhotoPath         = "images/bazaar/%s"
	BazaarPhotoCompressPath = "images/bazaar/compress/%s"
)

type (
	BazaarType string

	Bazaar struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		Name           string       `json:"-" form:"name"`
		Address        string       `json:"-" form:"address"`
		Photo          string       `json:"-" form:"photo"`
		Longitude      float64      `json:"-" form:"longitude"`
		Latitude       float64      `json:"-" form:"latitude"`
		BazaarType     BazaarType   `json:"-" form:"bazaar_type"`
		CreatedBy      *uuid.UUID   `json:"-" form:"-"`
		CreatedAt      time.Time    `json:"-" form:"-"`
		UpdatedBy      *uuid.UUID   `json:"-" form:"-"`
		UpdatedAt      time.Time    `json:"-" form:"-"`
		DeletedBy      *uuid.UUID   `json:"-" form:"-"`
		DeletedAt      *time.Time   `json:"-" form:"-"`
		DeletionReason *string      `json:"-" form:"-"`
	}

	BazaarFilter struct {
		BazaarIDs    []uuid.UUID
		RecordStatus []RecordStatus
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewBazaarFilter() *BazaarFilter {
	return new(BazaarFilter)
}

func (q *BazaarFilter) WithBazaarIDs(bazaarIDs ...uuid.UUID) *BazaarFilter {
	q.BazaarIDs = bazaarIDs
	return q
}

func (q *BazaarFilter) WithRecordStatus(recordStatus ...RecordStatus) *BazaarFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *BazaarFilter) WithKeyword(keyword string) *BazaarFilter {
	q.Keyword = keyword
	return q
}

func (q *BazaarFilter) WithPaginationURL(paginationURL string) *BazaarFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *BazaarFilter) WithLimit(limit int64) *BazaarFilter {
	q.Limit = limit
	return q
}

func (q *BazaarFilter) WithPage(page int64) *BazaarFilter {
	q.Page = page
	return q
}

func (q *BazaarFilter) WithUrlValues(urlValues url.Values) *BazaarFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Bazaar) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.Address = strings.TrimSpace(q.Address); q.Address == "" {
		return errors.New("address: is required")
	}
	return nil
}

func (q *Bazaar) HumanizedBazaarType() string {
	switch q.BazaarType {
	case BazaarTypePrivate:
		return "pribadi"
	case BazaarTypePublic:
		fallthrough
	default:
		return "umum"
	}
}
