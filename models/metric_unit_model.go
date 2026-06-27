package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
)

type (
	MetricUnit struct {
		RowNo          int64        `json:"-" form:"-"`
		ID             uuid.UUID    `json:"-" form:"-"`
		RecordStatus   RecordStatus `json:"-" form:"-"`
		Name           string       `json:"-" form:"-"`
		CreatedBy      uuid.UUID    `json:"-" form:"-"`
		CreatedAt      time.Time    `json:"-" form:"-"`
		UpdatedBy      uuid.UUID    `json:"-" form:"-"`
		UpdatedAt      time.Time    `json:"-" form:"-"`
		DeletedBy      *uuid.UUID   `json:"-" form:"-"`
		DeletedAt      *time.Time   `json:"-" form:"-"`
		DeletionReason *string      `json:"-" form:"-"`
	}

	MetricUnitFilter struct {
		MetricUnitIDs []uuid.UUID
		RecordStatus  []RecordStatus
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewMetricUnitFilter() *MetricUnitFilter {
	return new(MetricUnitFilter{UrlValues: url.Values{}})
}

func (q *MetricUnitFilter) WithMetricUnitIDs(metricUnitIDs ...uuid.UUID) *MetricUnitFilter {
	q.MetricUnitIDs = metricUnitIDs
	return q
}

func (q *MetricUnitFilter) WithRecordStatus(recordStatus ...RecordStatus) *MetricUnitFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *MetricUnitFilter) WithKeyword(keyword string) *MetricUnitFilter {
	q.Keyword = keyword
	return q
}

func (q *MetricUnitFilter) WithPaginationURL(paginationURL string) *MetricUnitFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *MetricUnitFilter) WithLimit(limit int64) *MetricUnitFilter {
	q.Limit = limit
	return q
}

func (q *MetricUnitFilter) WithPage(page int64) *MetricUnitFilter {
	q.Page = page
	return q
}

func (q *MetricUnitFilter) WithUrlValues(urlValues url.Values) *MetricUnitFilter {
	q.UrlValues = urlValues
	return q
}

func (q *MetricUnit) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	return nil
}
