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
	ErrIdempotencyKeyNotFound = errors.New("idempotency key not found")
)

type (
	IdempotencyKey struct {
		RowNo     int64     `json:"-" form:"-"`
		ID        uuid.UUID `json:"id" form:"-"`
		BuyerID   uuid.UUID `json:"buyer_id" form:"-"`
		Occupied  bool      `json:"occupied" form:"-"`
		CreatedAt time.Time `json:"created_at" form:"-"`
		ExpiredAt time.Time `json:"expired_at" form:"-"`
	}

	IdempotencyKeyFilter struct {
		IdempotencyKeyIDs,
		BuyerIDs []uuid.UUID
		Expired       *bool
		PaginationURL string
		Sortings      []Sorting
		MapSortings   Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewIdempotencyKeyFilter() *IdempotencyKeyFilter {
	return &IdempotencyKeyFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *IdempotencyKeyFilter) WithIdempotencyKeyIDs(idempotencyKeyIDs ...uuid.UUID) *IdempotencyKeyFilter {
	q.IdempotencyKeyIDs = idempotencyKeyIDs
	return q
}

func (q *IdempotencyKeyFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *IdempotencyKeyFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *IdempotencyKeyFilter) WithExpired(expired *bool) *IdempotencyKeyFilter {
	q.Expired = expired
	return q
}

func (q *IdempotencyKeyFilter) WithPaginationURL(paginationURL string) *IdempotencyKeyFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *IdempotencyKeyFilter) WithSortings(sortings ...Sorting) *IdempotencyKeyFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *IdempotencyKeyFilter) WithLimit(limit int64) *IdempotencyKeyFilter {
	q.Limit = limit
	return q
}

func (q *IdempotencyKeyFilter) WithPage(page int64) *IdempotencyKeyFilter {
	q.Page = page
	return q
}

func (q *IdempotencyKeyFilter) WithUrlValues(urlValues url.Values) *IdempotencyKeyFilter {
	q.UrlValues = urlValues
	return q
}

func (q *IdempotencyKey) Validate() error {
	if q.BuyerID.Compare(uuid.Nil()) == 0 {
		return errors.New("buyer_id: is required")
	}
	return nil
}

func (q *IdempotencyKey) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *IdempotencyKey) Available() bool {
	return !q.Occupied && q.ExpiredAt.After(time.Now())
}
