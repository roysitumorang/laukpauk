package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrLandingPageNotFound = errors.New("landing page not found")
var MapLandingPageSortings = map[string]struct{}{
	SortingFieldID: {},
}

const (
	LandingPageNameHeroSlider   LandingPageName = "hero-slider"
	LandingPageNameFeature      LandingPageName = "feature"
	LandingPageNameWhy          LandingPageName = "why"
	LandingPageNameFreeShopping LandingPageName = "free-shopping"
	LandingPageNameCommitment   LandingPageName = "commitment"
	LandingPageNameKitchenHero  LandingPageName = "kitchen-hero"
	LandingPageNameFAQ          LandingPageName = "faq"
)

type (
	LandingPageName string
	LandingPage     struct {
		RowNo            int64              `json:"-" form:"-"`
		ID               uuid.UUID          `json:"-" form:"id"`
		RecordStatus     RecordStatus       `json:"-" form:"-"`
		Name             LandingPageName    `json:"-" form:"-"`
		CustomAttributes map[string]string  `json:"-" form:"custom_attributes"`
		CreatedBy        uuid.UUID          `json:"-" form:"-"`
		CreatedAt        time.Time          `json:"-" form:"-"`
		UpdatedBy        uuid.UUID          `json:"-" form:"-"`
		UpdatedAt        time.Time          `json:"-" form:"-"`
		DeletedBy        *uuid.UUID         `json:"-" form:"-"`
		DeletedAt        *time.Time         `json:"-" form:"-"`
		DeletionReason   *string            `json:"-" form:"-"`
		Items            []*LandingPageItem `json:"-" form:"-"`
	}

	LandingPageFilter struct {
		LandingPageIDs []uuid.UUID
		RecordStatus   []RecordStatus
		Names          []LandingPageName
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewLandingPageFilter() *LandingPageFilter {
	return &LandingPageFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *LandingPageFilter) WithLandingPageIDs(landingPageIDs ...uuid.UUID) *LandingPageFilter {
	q.LandingPageIDs = landingPageIDs
	return q
}

func (q *LandingPageFilter) WithRecordStatus(recordStatus ...RecordStatus) *LandingPageFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *LandingPageFilter) WithNames(names ...LandingPageName) *LandingPageFilter {
	q.Names = names
	return q
}

func (q *LandingPageFilter) WithKeyword(keyword string) *LandingPageFilter {
	q.Keyword = keyword
	return q
}

func (q *LandingPageFilter) WithPaginationURL(paginationURL string) *LandingPageFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *LandingPageFilter) WithSortings(sortings ...Sorting) *LandingPageFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *LandingPageFilter) WithLimit(limit int64) *LandingPageFilter {
	q.Limit = limit
	return q
}

func (q *LandingPageFilter) WithPage(page int64) *LandingPageFilter {
	q.Page = page
	return q
}

func (q *LandingPageFilter) WithUrlValues(urlValues url.Values) *LandingPageFilter {
	q.UrlValues = urlValues
	return q
}
