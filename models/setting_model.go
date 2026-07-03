package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrSettingNotFound           = errors.New("setting not found")
	ErrUniqueSettingKeyViolation = errors.New("key: already exists")

	MapSettingSortings = map[string]struct{}{
		SortingFieldID:         {},
		SettingSortingFieldKey: {},
	}
)

const (
	SettingSortingFieldKey = "key"
)

type (
	Setting struct {
		RowNo       int64     `json:"-" form:"-"`
		ID          uuid.UUID `json:"-" form:"-"`
		Active      bool      `json:"-" form:"-"`
		Key         string    `json:"-" form:"key"`
		Value       string    `json:"-" form:"value"`
		Description string    `json:"-" form:"description"`
		CreatedBy   uuid.UUID `json:"-" form:"-"`
		CreatedAt   time.Time `json:"-" form:"-"`
		UpdatedBy   uuid.UUID `json:"-" form:"-"`
		UpdatedAt   time.Time `json:"-" form:"-"`
	}

	SettingFilter struct {
		SettingIDs,
		ExcludedSettingIDs []uuid.UUID
		Active *bool
		Keys   []string
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings *Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewSettingFilter() *SettingFilter {
	return new(SettingFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *SettingFilter) WithSettingIDs(settingIDs ...uuid.UUID) *SettingFilter {
	q.SettingIDs = settingIDs
	return q
}

func (q *SettingFilter) WithExcludedSettingIDs(excludedSettingIDs ...uuid.UUID) *SettingFilter {
	q.ExcludedSettingIDs = excludedSettingIDs
	return q
}

func (q *SettingFilter) WithActive(active *bool) *SettingFilter {
	q.Active = active
	return q
}

func (q *SettingFilter) WithKeys(keys ...string) *SettingFilter {
	q.Keys = keys
	return q
}

func (q *SettingFilter) WithKeyword(keyword string) *SettingFilter {
	q.Keyword = keyword
	return q
}

func (q *SettingFilter) WithPaginationURL(paginationURL string) *SettingFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *SettingFilter) WithSortings(sortings ...Sorting) *SettingFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *SettingFilter) WithLimit(limit int64) *SettingFilter {
	q.Limit = limit
	return q
}

func (q *SettingFilter) WithPage(page int64) *SettingFilter {
	q.Page = page
	return q
}

func (q *SettingFilter) WithUrlValues(urlValues url.Values) *SettingFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Setting) Validate() error {
	q.Key = strings.TrimSpace(q.Key)
	q.Key = strings.ReplaceAll(q.Key, " ", "")
	if q.Key == "" {
		return errors.New("key: is required")
	}
	if q.Value = strings.TrimSpace(q.Value); q.Value == "" {
		return errors.New("value: is required")
	}
	if q.Description = strings.TrimSpace(q.Description); q.Description == "" {
		return errors.New("description: is required")
	}
	return nil
}
