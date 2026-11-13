package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrPointNotFound = errors.New("point not found")

type PointType uint8

const (
	PointTypeReferral PointType = iota
	PointTypeReward
)

func (q PointType) Int() uint8 {
	return uint8(q)
}

func (q PointType) String() string {
	if q == PointTypeReward {
		return "reward"
	}
	return "referral"
}

func (q PointType) Equal(v PointType) bool {
	return q.Int() == v.Int()
}

var PointTypes = []PointType{
	PointTypeReferral,
	PointTypeReward,
}
var MapPointTypes = map[PointType]struct{}{
	PointTypeReferral: {},
	PointTypeReward:   {},
}

var MapRewardPointSortings = map[string]struct{}{
	SortingFieldID: {},
}

type Point struct {
	RowNo                      int64        `json:"-" form:"-"`
	ID                         uuid.UUID    `json:"-" form:"id"`
	RecordStatus               RecordStatus `json:"-" form:"-"`
	PointType                  PointType    `json:"-" form:"point_type"`
	BuyerID                    uuid.UUID    `json:"-" form:"buyer_id"`
	BuyerName                  string       `json:"-" form:"-"`
	BuyerUsername              string       `json:"-" form:"-"`
	BuyerMobilePhone           string       `json:"-" form:"-"`
	BuyerReferralCode          string       `json:"-" form:"-"`
	Amount                     int64        `json:"-" form:"amount"`
	TransactionID              *uuid.UUID   `json:"-" form:"-"`
	TransactionReferenceNumber *string      `json:"-" form:"-"`
	Description                string       `json:"-" form:"description"`
	CreatedBy                  *uuid.UUID   `json:"-" form:"-"`
	CreatedAt                  time.Time    `json:"-" form:"-"`
	UpdatedBy                  *uuid.UUID   `json:"-" form:"-"`
	UpdatedAt                  time.Time    `json:"-" form:"-"`
	DeletedBy                  *uuid.UUID   `json:"-" form:"-"`
	DeletedAt                  *time.Time   `json:"-" form:"-"`
	DeletionReason             *string      `json:"-" form:"-"`
}

func (q *Point) Validate() error {
	if _, ok := MapPointTypes[q.PointType]; !ok {
		return errors.New("point_type: not found")
	}
	if q.BuyerID.Compare(uuid.Nil()) == 0 {
		return errors.New("buyer_id: is required")
	}
	if q.Amount < 1 {
		return errors.New("amount: requires a positive integer")
	}
	if q.Description = strings.TrimSpace(q.Description); q.Description == "" {
		return errors.New("description: is required")
	}
	return nil
}

func (q *Point) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Point) Spendable() bool {
	return q.PointType.Equal(PointTypeReward)
}

type PointFilter struct {
	PointIDs,
	BuyerIDs,
	TransactionIDs []uuid.UUID
	RecordStatuses []RecordStatus
	PointTypes     []PointType
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewPointFilter() *PointFilter {
	return &PointFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *PointFilter) WithPointIDs(pointIDs ...uuid.UUID) *PointFilter {
	q.PointIDs = pointIDs
	return q
}

func (q *PointFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *PointFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *PointFilter) WithTransactionIDs(transactionIDs ...uuid.UUID) *PointFilter {
	q.TransactionIDs = transactionIDs
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
