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
	ErrPointNotFound = errors.New("point not found")

	PointTypes    = []PointType{PointTypeReferral, PointTypeReward}
	MapPointTypes = map[PointType]struct{}{
		PointTypeReferral: {},
		PointTypeReward:   {},
	}
	MapRewardPointSortings = map[string]struct{}{
		SortingFieldID: {},
	}
)

const (
	PointTypeReferral PointType = iota
	PointTypeReward
)

type (
	PointType int

	Point struct {
		RowNo             int64      `json:"-" form:"-"`
		ID                uuid.UUID  `json:"-" form:"-"`
		PointType         PointType  `json:"-" form:"point_type"`
		BuyerID           uuid.UUID  `json:"-" form:"buyer_id"`
		BuyerName         string     `json:"-" form:"-"`
		BuyerUsername     string     `json:"-" form:"-"`
		BuyerMobilePhone  string     `json:"-" form:"-"`
		BuyerReferralCode string     `json:"-" form:"-"`
		Amount            int64      `json:"-" form:"amount"`
		TransactionID     *uuid.UUID `json:"-" form:"-"`
		Spendable         bool       `json:"-" form:"-"`
		Description       string     `json:"-" form:"description"`
		CreatedAt         time.Time  `json:"-" form:"-"`
		UpdatedAt         time.Time  `json:"-" form:"-"`
	}

	PointFilter struct {
		PointIDs,
		BuyerIDs,
		TransactionIDs []uuid.UUID
		PointTypes []PointType
		Spendable  *bool
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

func (q *PointFilter) WithPointTypes(pointTypes ...PointType) *PointFilter {
	q.PointTypes = pointTypes
	return q
}

func (q *PointFilter) WithSpendable(spendable *bool) *PointFilter {
	q.Spendable = spendable
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
	if q.BuyerID.Compare(uuid.Nil()) == 0 {
		return errors.New("buyer_id: is required")
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
	if q == PointTypeReward {
		return "reward"
	}
	return "referral"
}

func (q PointType) Equal(v PointType) bool {
	return q.Int() == v.Int()
}
