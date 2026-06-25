package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"
)

const (
	BlockPhotoName = "%s-%d%s"
	BlockPhotoPath = "images/landing-page/%s/%s"

	BlockNameHero         BlockName = "hero"
	BlockNameFeature      BlockName = "feature"
	BlockNameWhy          BlockName = "why"
	BlockNameFreeShopping BlockName = "free-shopping"
	BlockNameCommitment   BlockName = "commitment"
	BlockNameKitchenHero  BlockName = "kitchen-hero"
	BlockNameFAQ          BlockName = "faq"
)

type (
	BlockName string
	Block     struct {
		RowNo            int64             `json:"-" form:"-"`
		ID               uuid.UUID         `json:"-" form:"-"`
		RecordStatus     RecordStatus      `json:"-" form:"-"`
		Name             BlockName         `json:"-" form:"-"`
		CustomAttributes map[string]string `json:"-" form:"custom_attributes"`
		CreatedBy        uuid.UUID         `json:"-" form:"-"`
		CreatedAt        time.Time         `json:"-" form:"-"`
		UpdatedBy        uuid.UUID         `json:"-" form:"-"`
		UpdatedAt        time.Time         `json:"-" form:"-"`
		DeletedBy        *uuid.UUID        `json:"-" form:"-"`
		DeletedAt        *time.Time        `json:"-" form:"-"`
		DeletionReason   *string           `json:"-" form:"-"`
		Items            []*BlockItem      `json:"-" form:"-"`
	}

	BlockItem struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		BlockName      BlockName    `json:"-" form:"-"`
		Subject        string       `json:"-" form:"subject"`
		Body           string       `json:"-" form:"body"`
		Photo          *string      `json:"-" form:"photo"`
		LinkText       *string      `json:"-" form:"link_text"`
		LinkURL        *string      `json:"-" form:"link_url"`
		CreatedBy      uuid.UUID    `json:"-" form:"-"`
		CreatedAt      time.Time    `json:"-" form:"-"`
		UpdatedBy      uuid.UUID    `json:"-" form:"-"`
		UpdatedAt      time.Time    `json:"-" form:"-"`
		DeletedBy      *uuid.UUID   `json:"-" form:"-"`
		DeletedAt      *time.Time   `json:"-" form:"-"`
		DeletionReason *string      `json:"-" form:"-"`
	}

	BlockFilter struct {
		BlockIDs     []uuid.UUID
		RecordStatus []RecordStatus
		Names        []BlockName
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}

	BlockItemFilter struct {
		BlockItemIDs []uuid.UUID
		BlockNames   []BlockName
		RecordStatus []RecordStatus
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewBlockFilter() *BlockFilter {
	return new(BlockFilter)
}

func (q *BlockFilter) WithBlockIDs(blockIDs ...uuid.UUID) *BlockFilter {
	q.BlockIDs = blockIDs
	return q
}

func (q *BlockFilter) WithRecordStatus(recordStatus ...RecordStatus) *BlockFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *BlockFilter) WithNames(names ...BlockName) *BlockFilter {
	q.Names = names
	return q
}

func (q *BlockFilter) WithKeyword(keyword string) *BlockFilter {
	q.Keyword = keyword
	return q
}

func (q *BlockFilter) WithPaginationURL(paginationURL string) *BlockFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *BlockFilter) WithLimit(limit int64) *BlockFilter {
	q.Limit = limit
	return q
}

func (q *BlockFilter) WithPage(page int64) *BlockFilter {
	q.Page = page
	return q
}

func (q *BlockFilter) WithUrlValues(urlValues url.Values) *BlockFilter {
	q.UrlValues = urlValues
	return q
}

func NewBlockItemFilter() *BlockItemFilter {
	return new(BlockItemFilter{UrlValues: url.Values{}})
}

func (q *BlockItemFilter) WithBlockItemIDs(blockItemIDs ...uuid.UUID) *BlockItemFilter {
	q.BlockItemIDs = blockItemIDs
	return q
}

func (q *BlockItemFilter) WithBlockNames(blockNames ...BlockName) *BlockItemFilter {
	q.BlockNames = blockNames
	return q
}

func (q *BlockItemFilter) WithRecordStatus(recordStatus ...RecordStatus) *BlockItemFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *BlockItemFilter) WithKeyword(keyword string) *BlockItemFilter {
	q.Keyword = keyword
	return q
}

func (q *BlockItemFilter) WithPaginationURL(paginationURL string) *BlockItemFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *BlockItemFilter) WithLimit(limit int64) *BlockItemFilter {
	q.Limit = limit
	return q
}

func (q *BlockItemFilter) WithPage(page int64) *BlockItemFilter {
	q.Page = page
	return q
}

func (q *BlockItemFilter) WithUrlValues(urlValues url.Values) *BlockItemFilter {
	q.UrlValues = urlValues
	return q
}

func (q *BlockItem) Validate() error {
	q.Subject = strings.TrimSpace(q.Subject)
	q.Body = strings.TrimSpace(q.Body)

	if q.LinkText != nil {
		*q.LinkText = strings.TrimSpace(*q.LinkText)
	}
	if q.LinkURL != nil {
		*q.LinkURL = strings.TrimSpace(*q.LinkURL)
	}
	switch q.BlockName {
	case BlockNameHero:
		if q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == "" {
			return errors.New("body: is required")
		}
		if q.LinkText == nil ||
			*q.LinkText == "" {
			return errors.New("link_text: is required")
		}
		if q.LinkURL == nil ||
			*q.LinkURL == "" {
			return errors.New("link_url: is required")
		}
		if _, err := url.ParseRequestURI(*q.LinkURL); err != nil {
			return errors.New("link_url: invalid URL")
		}
		if q.ID == uuid.Nil() &&
			(q.Photo == nil ||
				*q.Photo == "") {
			return errors.New("photo: is required")
		}
	case BlockNameFeature, BlockNameWhy:
		if q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == "" {
			return errors.New("body: is required")
		}
		q.LinkText = nil
		q.LinkURL = nil
		if q.ID == uuid.Nil() &&
			(q.Photo == nil ||
				*q.Photo == "") {
			return errors.New("photo: is required")
		}
	case BlockNameFAQ:
		if q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == "" {
			return errors.New("body: is required")
		}
	}
	return nil
}

func (q *BlockItem) SetPhoto(photo *string) *BlockItem {
	q.Photo = photo
	return q
}

func (q *BlockItem) PhotoPath() *string {
	if q.Photo == nil {
		return nil
	}
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, BlockPhotoPath, q.BlockName, *q.Photo)
	return new(sb.String())
}
