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
	ErrRewardPointNotFound = errors.New("setting not found")

	RewardPointTypes    = []RewardPointType{RewardPointTypeByTransaction, RewardPointTypeByAdmin}
	MapRewardPointTypes = map[RewardPointType]struct{}{
		RewardPointTypeByTransaction: {},
		RewardPointTypeByAdmin:       {},
	}
	MapRewardPointSortings = map[string]struct{}{
		SortingFieldID: {},
	}
)

const (
	RewardPointTypeByTransaction RewardPointType = iota
	RewardPointTypeByAdmin
)

type (
	RewardPointType int

	RewardPoint struct {
		RowNo             int64           `json:"-" form:"-"`
		ID                uuid.UUID       `json:"-" form:"-"`
		RecordStatus      RecordStatus    `json:"-" form:"-"`
		RewardPointType   RewardPointType `json:"-" form:"-"`
		BuyerID           uuid.UUID       `json:"-" form:"buyer_id"`
		BuyerName         string          `json:"-" form:"-"`
		BuyerUsername     string          `json:"-" form:"-"`
		BuyerMobilePhone  string          `json:"-" form:"-"`
		BuyerReferralCode string          `json:"-" form:"-"`
		Amount            int64           `json:"-" form:"amount"`
		Description       string          `json:"-" form:"description"`
		CreatedBy         *uuid.UUID      `json:"-" form:"-"`
		CreatedAt         time.Time       `json:"-" form:"-"`
		UpdatedBy         *uuid.UUID      `json:"-" form:"-"`
		UpdatedAt         time.Time       `json:"-" form:"-"`
		DeletedBy         *uuid.UUID      `json:"-" form:"-"`
		DeletedAt         *time.Time      `json:"-" form:"-"`
		DeletionReason    *string         `json:"-" form:"-"`
	}

	RewardPointFilter struct {
		RewardPointIDs,
		BuyerIDs []uuid.UUID
		RecordStatuses   []RecordStatus
		RewardPointTypes []RewardPointType
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings *Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewRewardPointFilter() *RewardPointFilter {
	return new(RewardPointFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *RewardPointFilter) WithRewardPointIDs(rewardPointIDs ...uuid.UUID) *RewardPointFilter {
	q.RewardPointIDs = rewardPointIDs
	return q
}

func (q *RewardPointFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *RewardPointFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *RewardPointFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *RewardPointFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *RewardPointFilter) WithRewardPointTypes(rewardPointTypes ...RewardPointType) *RewardPointFilter {
	q.RewardPointTypes = rewardPointTypes
	return q
}

func (q *RewardPointFilter) WithKeyword(keyword string) *RewardPointFilter {
	q.Keyword = keyword
	return q
}

func (q *RewardPointFilter) WithPaginationURL(paginationURL string) *RewardPointFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *RewardPointFilter) WithSortings(sortings ...Sorting) *RewardPointFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *RewardPointFilter) WithLimit(limit int64) *RewardPointFilter {
	q.Limit = limit
	return q
}

func (q *RewardPointFilter) WithPage(page int64) *RewardPointFilter {
	q.Page = page
	return q
}

func (q *RewardPointFilter) WithUrlValues(urlValues url.Values) *RewardPointFilter {
	q.UrlValues = urlValues
	return q
}

func (q *RewardPoint) Validate() error {
	if q.BuyerID.Compare(uuid.Nil()) == 0 {
		return errors.New("buyer_id: is required")
	}
	if _, ok := MapRewardPointTypes[q.RewardPointType]; !ok {
		return errors.New("reward_point_type: not found")
	}
	if q.Amount == 0 {
		return errors.New("amount: is required")
	}
	if q.Description = strings.TrimSpace(q.Description); q.Description == "" {
		return errors.New("description: is required")
	}
	return nil
}

func (q RewardPointType) Int() int {
	return int(q)
}

func (q RewardPointType) String() string {
	if q == RewardPointTypeByAdmin {
		return "admin"
	}
	return "transaction"
}

func (q RewardPointType) Equal(v RewardPointType) bool {
	return q.Int() == v.Int()
}
