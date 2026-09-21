package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

const BannerSortingFieldName = "name"

var ErrBannerNotFound = errors.New("banner not found")
var MapBannerSortings = map[string]struct{}{
	SortingFieldID:         {},
	BannerSortingFieldName: {},
}

type Banner struct {
	RowNo          int64        `json:"-" form:"-"`
	ID             uuid.UUID    `json:"id" form:"id"`
	CompanyID      uuid.UUID    `json:"-" form:"-"`
	RecordStatus   RecordStatus `json:"-" form:"-"`
	Name           string       `json:"name" form:"name"`
	Description    string       `json:"description" form:"description"`
	Photo          string       `json:"photo" form:"photo"`
	Visible        bool         `json:"-" form:"-"`
	CreatedBy      uuid.UUID    `json:"-" form:"-"`
	CreatedAt      time.Time    `json:"-" form:"-"`
	UpdatedBy      uuid.UUID    `json:"-" form:"-"`
	UpdatedAt      time.Time    `json:"-" form:"-"`
	DeletedBy      *uuid.UUID   `json:"-" form:"-"`
	DeletedAt      *time.Time   `json:"-" form:"-"`
	DeletionReason *string      `json:"-" form:"-"`
}

func (q *Banner) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	q.Description = strings.TrimSpace(q.Description)
	if q.Photo == "" {
		return errors.New("photo: is required")
	}
	return nil
}

func (q *Banner) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Banner) SetPhoto(photo string) *Banner {
	q.Photo = photo
	return q
}

func (q *Banner) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
}

type BannerFilter struct {
	BannerIDs,
	CompanyIDs []uuid.UUID
	RecordStatuses []RecordStatus
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewBannerFilter() *BannerFilter {
	return &BannerFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *BannerFilter) WithBannerIDs(bannerIDs ...uuid.UUID) *BannerFilter {
	q.BannerIDs = bannerIDs
	return q
}

func (q *BannerFilter) WithCompanyIDs(companyIDs ...uuid.UUID) *BannerFilter {
	q.CompanyIDs = companyIDs
	return q
}

func (q *BannerFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *BannerFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *BannerFilter) WithKeyword(keyword string) *BannerFilter {
	q.Keyword = keyword
	return q
}

func (q *BannerFilter) WithPaginationURL(paginationURL string) *BannerFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *BannerFilter) WithSortings(sortings ...Sorting) *BannerFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *BannerFilter) WithLimit(limit int64) *BannerFilter {
	q.Limit = limit
	return q
}

func (q *BannerFilter) WithPage(page int64) *BannerFilter {
	q.Page = page
	return q
}

func (q *BannerFilter) WithUrlValues(urlValues url.Values) *BannerFilter {
	q.UrlValues = urlValues
	return q
}
