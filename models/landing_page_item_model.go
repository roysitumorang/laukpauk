package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrLandingPageItemNotFound = errors.New("landing page item not found")

type (
	LandingPageItem struct {
		RowNo           int64           `json:"-" form:"-"`
		ID              uuid.UUID       `json:"-" form:"id"`
		RecordStatus    RecordStatus    `json:"-" form:"-"`
		LandingPageName LandingPageName `json:"-" form:"-"`
		Subject         string          `json:"-" form:"subject"`
		Body            string          `json:"-" form:"body"`
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

	LandingPageItemFilter struct {
		LandingPageItemIDs []uuid.UUID
		LandingPageNames   []LandingPageName
		RecordStatus       []RecordStatus
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewLandingPageItemFilter() *LandingPageItemFilter {
	return &LandingPageItemFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *LandingPageItemFilter) WithLandingPageItemIDs(landingPageItemIDs ...uuid.UUID) *LandingPageItemFilter {
	q.LandingPageItemIDs = landingPageItemIDs
	return q
}

func (q *LandingPageItemFilter) WithLandingPageNames(landingPageNames ...LandingPageName) *LandingPageItemFilter {
	q.LandingPageNames = landingPageNames
	return q
}

func (q *LandingPageItemFilter) WithRecordStatus(recordStatus ...RecordStatus) *LandingPageItemFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *LandingPageItemFilter) WithKeyword(keyword string) *LandingPageItemFilter {
	q.Keyword = keyword
	return q
}

func (q *LandingPageItemFilter) WithPaginationURL(paginationURL string) *LandingPageItemFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *LandingPageItemFilter) WithSortings(sortings ...Sorting) *LandingPageItemFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *LandingPageItemFilter) WithLimit(limit int64) *LandingPageItemFilter {
	q.Limit = limit
	return q
}

func (q *LandingPageItemFilter) WithPage(page int64) *LandingPageItemFilter {
	q.Page = page
	return q
}

func (q *LandingPageItemFilter) WithUrlValues(urlValues url.Values) *LandingPageItemFilter {
	q.UrlValues = urlValues
	return q
}

func (q *LandingPageItem) Validate() error {
	q.Subject = strings.TrimSpace(q.Subject)
	q.Body = strings.TrimSpace(q.Body)
	if q.LinkText != nil {
		*q.LinkText = strings.TrimSpace(*q.LinkText)
	}
	if q.LinkURL != nil {
		*q.LinkURL = strings.TrimSpace(*q.LinkURL)
	}
	switch q.LandingPageName {
	case LandingPageNameHero:
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
	case LandingPageNameFeature, LandingPageNameWhy:
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
	case LandingPageNameFAQ:
		if q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == "" {
			return errors.New("body: is required")
		}
	}
	return nil
}

func (q *LandingPageItem) SetPhoto(photo *string) *LandingPageItem {
	q.Photo = photo
	return q
}

func (q *LandingPageItem) PhotoPath(baseURLs ...string) string {
	if q.Photo == nil {
		return ""
	}
	return helper.PhotoPath(*q.Photo, baseURLs...)
}
