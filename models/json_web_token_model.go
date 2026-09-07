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
	ErrJsonWebTokenNotFound        = errors.New("json web token not found")
	ErrUniqueJsonWebTokenViolation = errors.New("token: already exists")
)

type (
	JsonWebToken struct {
		RowNo     int64     `json:"-" form:"-"`
		ID        uuid.UUID `json:"-" form:"-"`
		BuyerID   uuid.UUID `json:"buyer_id" form:"-"`
		CreatedAt time.Time `json:"-" form:"-"`
		ExpiredAt time.Time `json:"-" form:"-"`
	}

	JsonWebTokenFilter struct {
		JsonWebTokenIDs,
		BuyerIDs []uuid.UUID
		Expired *bool
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewJsonWebTokenFilter() *JsonWebTokenFilter {
	return &JsonWebTokenFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *JsonWebTokenFilter) WithJsonWebTokenIDs(jsonWebTokenIDs ...uuid.UUID) *JsonWebTokenFilter {
	q.JsonWebTokenIDs = jsonWebTokenIDs
	return q
}

func (q *JsonWebTokenFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *JsonWebTokenFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *JsonWebTokenFilter) WithExpired(expired *bool) *JsonWebTokenFilter {
	q.Expired = expired
	return q
}

func (q *JsonWebTokenFilter) WithKeyword(keyword string) *JsonWebTokenFilter {
	q.Keyword = keyword
	return q
}

func (q *JsonWebTokenFilter) WithPaginationURL(paginationURL string) *JsonWebTokenFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *JsonWebTokenFilter) WithSortings(sortings ...Sorting) *JsonWebTokenFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *JsonWebTokenFilter) WithLimit(limit int64) *JsonWebTokenFilter {
	q.Limit = limit
	return q
}

func (q *JsonWebTokenFilter) WithPage(page int64) *JsonWebTokenFilter {
	q.Page = page
	return q
}

func (q *JsonWebTokenFilter) WithUrlValues(urlValues url.Values) *JsonWebTokenFilter {
	q.UrlValues = urlValues
	return q
}

func (q *JsonWebToken) Validate() error {
	if q.BuyerID.Compare(uuid.Nil()) == 0 {
		return errors.New("buyer_id: is required")
	}
	return nil
}

func (q *JsonWebToken) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}
