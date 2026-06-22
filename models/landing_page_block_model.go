package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/nyaruka/phonenumbers"
)

const (
	LandingPageBlockPhotoName = "%s-%d%s"
	LandingPageBlockPhotoPath = "images/landing-page/%s/%s"

	LandingPageBlockNameHero         LandingPageBlockName = "hero"
	LandingPageBlockNameFeature      LandingPageBlockName = "feature"
	LandingPageBlockNameWhy          LandingPageBlockName = "why"
	LandingPageBlockNameFreeShopping LandingPageBlockName = "free-shopping"
	LandingPageBlockNameCommitment   LandingPageBlockName = "commitment"
	LandingPageBlockNameKitchenHero  LandingPageBlockName = "kitchen-hero"
	LandingPageBlockNameFAQ          LandingPageBlockName = "faq"
)

type (
	LandingPageBlockName string
	LandingPageBlock     struct {
		RowNo            int64                   `json:"-" form:"-"`
		ID               uuid.UUID               `json:"-" form:"-"`
		RecordStatus     RecordStatus            `json:"-" form:"-"`
		Name             LandingPageBlockName    `json:"-" form:"-"`
		CustomAttributes map[string]string       `json:"-" form:"custom_attributes"`
		CreatedBy        uuid.UUID               `json:"-" form:"-"`
		CreatedAt        time.Time               `json:"-" form:"-"`
		UpdatedBy        uuid.UUID               `json:"-" form:"-"`
		UpdatedAt        time.Time               `json:"-" form:"-"`
		DeletedBy        *uuid.UUID              `json:"-" form:"-"`
		DeletedAt        *time.Time              `json:"-" form:"-"`
		DeletionReason   *string                 `json:"-" form:"-"`
		Items            []*LandingPageBlockItem `json:"-" form:"-"`
	}

	LandingPageBlockItem struct {
		RowNo                int64                `json:"-" form:"-"`
		ID                   uuid.UUID            `json:"-" form:"-"`
		RecordStatus         RecordStatus         `json:"-" form:"-"`
		LandingPageBlockName LandingPageBlockName `json:"-" form:"-"`
		Subject              string               `json:"-" form:"subject"`
		Body                 string               `json:"-" form:"body"`
		Photo                *string              `json:"-" form:"photo"`
		LinkText             *string              `json:"-" form:"link_text"`
		LinkURL              *string              `json:"-" form:"link_url"`
		CreatedBy            uuid.UUID            `json:"-" form:"-"`
		CreatedAt            time.Time            `json:"-" form:"-"`
		UpdatedBy            uuid.UUID            `json:"-" form:"-"`
		UpdatedAt            time.Time            `json:"-" form:"-"`
		DeletedBy            *uuid.UUID           `json:"-" form:"-"`
		DeletedAt            *time.Time           `json:"-" form:"-"`
		DeletionReason       *string              `json:"-" form:"-"`
	}

	LandingPageBlockFilter struct {
		LandingPageBlockIDs []uuid.UUID
		RecordStatus        []RecordStatus
		Names               []LandingPageBlockName
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}

	LandingPageBlockItemFilter struct {
		LandingPageBlockItemIDs []uuid.UUID
		LandingPageBlockNames   []LandingPageBlockName
		RecordStatus            []RecordStatus
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}

	ContactUs struct {
		RowNo       int64     `json:"-" form:"-"`
		ID          uuid.UUID `json:"-" form:"-"`
		Name        string    `json:"name" form:"name"`
		MobilePhone string    `json:"mobile_phone" form:"mobile_phone"`
		Message     string    `json:"message" form:"message"`
	}
)

func NewLandingPageBlockFilter() *LandingPageBlockFilter {
	return new(LandingPageBlockFilter)
}

func (q *LandingPageBlockFilter) WithLandingPageBlockIDs(LandingPageBlockItemIDs ...uuid.UUID) *LandingPageBlockFilter {
	q.LandingPageBlockIDs = LandingPageBlockItemIDs
	return q
}

func (q *LandingPageBlockFilter) WithRecordStatus(recordStatus ...RecordStatus) *LandingPageBlockFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *LandingPageBlockFilter) WithNames(names ...LandingPageBlockName) *LandingPageBlockFilter {
	q.Names = names
	return q
}

func (q *LandingPageBlockFilter) WithKeyword(keyword string) *LandingPageBlockFilter {
	q.Keyword = keyword
	return q
}

func (q *LandingPageBlockFilter) WithPaginationURL(paginationURL string) *LandingPageBlockFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *LandingPageBlockFilter) WithLimit(limit int64) *LandingPageBlockFilter {
	q.Limit = limit
	return q
}

func (q *LandingPageBlockFilter) WithPage(page int64) *LandingPageBlockFilter {
	q.Page = page
	return q
}

func (q *LandingPageBlockFilter) WithUrlValues(urlValues url.Values) *LandingPageBlockFilter {
	q.UrlValues = urlValues
	return q
}

func NewLandingPageBlockItemFilter() *LandingPageBlockItemFilter {
	return new(LandingPageBlockItemFilter{UrlValues: url.Values{}})
}

func (q *LandingPageBlockItemFilter) WithLandingPageBlockItemIDs(landingPageBlockItemIDs ...uuid.UUID) *LandingPageBlockItemFilter {
	q.LandingPageBlockItemIDs = landingPageBlockItemIDs
	return q
}

func (q *LandingPageBlockItemFilter) WithLandingPageBlockNames(names ...LandingPageBlockName) *LandingPageBlockItemFilter {
	q.LandingPageBlockNames = names
	return q
}

func (q *LandingPageBlockItemFilter) WithRecordStatus(recordStatus ...RecordStatus) *LandingPageBlockItemFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *LandingPageBlockItemFilter) WithKeyword(keyword string) *LandingPageBlockItemFilter {
	q.Keyword = keyword
	return q
}

func (q *LandingPageBlockItemFilter) WithPaginationURL(paginationURL string) *LandingPageBlockItemFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *LandingPageBlockItemFilter) WithLimit(limit int64) *LandingPageBlockItemFilter {
	q.Limit = limit
	return q
}

func (q *LandingPageBlockItemFilter) WithPage(page int64) *LandingPageBlockItemFilter {
	q.Page = page
	return q
}

func (q *LandingPageBlockItemFilter) WithUrlValues(urlValues url.Values) *LandingPageBlockItemFilter {
	q.UrlValues = urlValues
	return q
}

func (q *LandingPageBlockItem) Validate() error {
	q.Subject = strings.TrimSpace(q.Subject)
	q.Body = strings.TrimSpace(q.Body)

	if q.LinkText != nil {
		*q.LinkText = strings.TrimSpace(*q.LinkText)
	}
	if q.LinkURL != nil {
		*q.LinkURL = strings.TrimSpace(*q.LinkURL)
	}
	switch q.LandingPageBlockName {
	case LandingPageBlockNameHero:
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
	case LandingPageBlockNameFeature, LandingPageBlockNameWhy:
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
	case LandingPageBlockNameFAQ:
		if q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == "" {
			return errors.New("body: is required")
		}
	}
	return nil
}

func (q *LandingPageBlockItem) SetPhoto(photo *string) *LandingPageBlockItem {
	q.Photo = photo
	return q
}

func (q *LandingPageBlockItem) PhotoPath() *string {
	if q.Photo == nil {
		return nil
	}
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, LandingPageBlockPhotoPath, q.LandingPageBlockName, *q.Photo)
	return new(sb.String())
}

func (q *ContactUs) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.MobilePhone = strings.TrimSpace(q.MobilePhone); q.MobilePhone == "" {
		return errors.New("mobile_phone: is required")
	}
	mobilePhone, err := phonenumbers.Parse(q.MobilePhone, "ID")
	if err != nil {
		return fmt.Errorf("mobile_phone: %w", err)
	}
	q.MobilePhone = phonenumbers.Format(mobilePhone, phonenumbers.E164)
	if q.Message = strings.TrimSpace(q.Message); q.Message == "" {
		return errors.New("message: is required")
	}
	return nil
}
