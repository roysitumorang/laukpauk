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
	ErrPointNotFound = errors.New("setting not found")

	PointTypes    = []PointType{PointTypeByTransaction, PointTypeByAdmin}
	MapPointTypes = map[PointType]struct{}{
		PointTypeByTransaction: {},
		PointTypeByAdmin:       {},
	}
	MapPointSortings = map[string]struct{}{
		SortingFieldID: {},
	}
)

const (
	PointTypeByTransaction PointType = iota
	PointTypeByAdmin
)

type (
	PointType int

	Point struct {
		RowNo             int64        `json:"-" form:"-"`
		ID                uuid.UUID    `json:"-" form:"-"`
		RecordStatus      RecordStatus `json:"-" form:"-"`
		PointType         PointType    `json:"-" form:"-"`
		BuyerID           uuid.UUID    `json:"-" form:"buyer_id"`
		BuyerName         string       `json:"-" form:"-"`
		BuyerUsername     string       `json:"-" form:"-"`
		BuyerMobilePhone  string       `json:"-" form:"-"`
		BuyerReferralCode string       `json:"-" form:"-"`
		Amount            int64        `json:"-" form:"amount"`
		Description       string       `json:"-" form:"description"`
		CreatedBy         *uuid.UUID   `json:"-" form:"-"`
		CreatedAt         time.Time    `json:"-" form:"-"`
		UpdatedBy         *uuid.UUID   `json:"-" form:"-"`
		UpdatedAt         time.Time    `json:"-" form:"-"`
		DeletedBy         *uuid.UUID   `json:"-" form:"-"`
		DeletedAt         *time.Time   `json:"-" form:"-"`
		DeletionReason    *string      `json:"-" form:"-"`
	}

	PointFilter struct {
		PointIDs,
		BuyerIDs []uuid.UUID
		RecordStatuses []RecordStatus
		PointTypes     []PointType
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings *Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewPointFilter() *PointFilter {
	return new(PointFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *PointFilter) WithPointIDs(settingIDs ...uuid.UUID) *PointFilter {
	q.PointIDs = settingIDs
	return q
}

func (q *PointFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *PointFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *PointFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *PointFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *PointFilter) WithPointTypes(pointTypes ...PointType) *PointFilter {
	q.PointTypes = pointTypes
	return q
}

func (q *PointFilter) WithKeyword(keyword string) *PointFilter {
	q.Keyword = keyword
	return q
}

func (q *PointFilter) WithPaginationURL(paginationURL string) *PointFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *PointFilter) WithSortings(sortings ...Sorting) *PointFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *PointFilter) WithLimit(limit int64) *PointFilter {
	q.Limit = limit
	return q
}

func (q *PointFilter) WithPage(page int64) *PointFilter {
	q.Page = page
	return q
}

func (q *PointFilter) WithUrlValues(urlValues url.Values) *PointFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Point) Validate() error {
	if _, ok := MapPointTypes[q.PointType]; !ok {
		return errors.New("point_type: not found")
	}
	if q.Amount == 0 {
		return errors.New("amount: is required")
	}
	if q.Description = strings.TrimSpace(q.Description); q.Description == "" {
		return errors.New("description: is required")
	}
	return nil
}

func (q PointType) Int() int {
	return int(q)
}

func (q PointType) String() string {
	if q == PointTypeByAdmin {
		return "admin"
	}
	return "transaction"
}

func (q PointType) Equal(v PointType) bool {
	return q.Int() == v.Int()
}
