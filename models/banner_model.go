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
	RowNo          int64        `json:"-"`
	ID             uuid.UUID    `json:"id"`
	RecordStatus   RecordStatus `json:"record_status"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Photo          string       `json:"photo"`
	Visible        bool         `json:"visible"`
	CreatedBy      uuid.UUID    `json:"created_by"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedBy      uuid.UUID    `json:"updated_by"`
	UpdatedAt      time.Time    `json:"updated_at"`
	DeletedBy      *uuid.UUID   `json:"deleted_by"`
	DeletedAt      *time.Time   `json:"deleted_at"`
	DeletionReason *string      `json:"deletion_reason"`
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
	BannerIDs      []uuid.UUID
	RecordStatuses []RecordStatus
	Visible        *bool
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

func (q *BannerFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *BannerFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *BannerFilter) WithVisible(visible *bool) *BannerFilter {
	q.Visible = visible
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
