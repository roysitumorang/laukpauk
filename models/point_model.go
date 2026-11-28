package models

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
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
	RowNo                      int64        `json:"-"`
	ID                         uuid.UUID    `json:"id"`
	RecordStatus               RecordStatus `json:"record_status"`
	PointType                  PointType    `json:"point_type"`
	UserID                     uuid.UUID    `json:"user_id"`
	BuyerName                  string       `json:"buyer_name"`
	BuyerUsername              string       `json:"buyer_username"`
	BuyerMobilePhone           string       `json:"buyer_mobile_phone"`
	BuyerReferralCode          string       `json:"buyer_referral_code"`
	Amount                     int64        `json:"amount"`
	TransactionID              *uuid.UUID   `json:"transaction_id"`
	TransactionReferenceNumber *string      `json:"transaction_reference_number"`
	Description                string       `json:"description"`
	CreatedBy                  *uuid.UUID   `json:"created_by"`
	CreatedAt                  time.Time    `json:"created_at"`
	UpdatedBy                  *uuid.UUID   `json:"updated_by"`
	UpdatedAt                  time.Time    `json:"updated_at"`
	DeletedBy                  *uuid.UUID   `json:"deleted_by"`
	DeletedAt                  *time.Time   `json:"deleted_at"`
	DeletionReason             *string      `json:"deletion_reason"`
}

func (q *Point) Validate() error {
	if _, ok := MapPointTypes[q.PointType]; !ok {
		return errors.New("point_type: not found")
	}
	if q.UserID == uuid.Nil {
		return errors.New("user_id: is required")
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
	return q.ID == uuid.Nil
}

func (q *Point) Spendable() bool {
	return q.PointType.Equal(PointTypeReward)
}

type PointFilter struct {
	PointIDs,
	UserIDs,
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

func (q *PointFilter) WithUserIDs(userIDs ...uuid.UUID) *PointFilter {
	q.UserIDs = userIDs
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

// SetBuyer copies the buyer's display fields onto the point.
func (q *Point) SetBuyer(buyer *User) *Point {
	q.BuyerName = buyer.Name
	q.BuyerMobilePhone = buyer.MobilePhone
	q.BuyerUsername, q.BuyerReferralCode = "", ""
	if buyer.Username != nil {
		q.BuyerUsername = *buyer.Username
	}
	if buyer.ReferralCode != nil {
		q.BuyerReferralCode = *buyer.ReferralCode
	}
	return q
}
