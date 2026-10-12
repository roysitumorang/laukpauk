package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrOneTimePasswordNotFound = errors.New("OTP not found")
var ErrUniqueOneTimePasswordTrackingIdViolation = errors.New("tracking_id: already exists")
var ErrOneTimePasswordMaxAttemptsExceeded = errors.New("max attempts exceeded")

type OneTimePassword struct {
	RowNo         int64     `json:"-" form:"-"`
	ID            uuid.UUID `json:"-" form:"id"`
	SellerID      uuid.UUID `json:"-" form:"-"`
	BuyerID       uuid.UUID `json:"-" form:"-"`
	Hash          string    `json:"-" form:"-"`
	TrackingID    uuid.UUID `json:"tracking_id" form:"-"`
	AttemptsCount uint8     `json:"-" form:"-"`
	Valid         bool      `json:"-" form:"-"`
	CreatedAt     time.Time `json:"-" form:"-"`
	ExpiredAt     time.Time `json:"-" form:"-"`
}

type OneTimePasswordFilter struct {
	OneTimePasswordIDs,
	SellerIDs,
	BuyerIDs,
	TrackingIDs []uuid.UUID
	Valid         *bool
	PaginationURL string
	Sortings      []Sorting
	MapSortings   Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewOneTimePasswordFilter() *OneTimePasswordFilter {
	return &OneTimePasswordFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *OneTimePasswordFilter) WithOneTimePasswordIDs(oneTimePasswordIDs ...uuid.UUID) *OneTimePasswordFilter {
	q.OneTimePasswordIDs = oneTimePasswordIDs
	return q
}

func (q *OneTimePasswordFilter) WithSellerIDs(sellerIDs ...uuid.UUID) *OneTimePasswordFilter {
	q.SellerIDs = sellerIDs
	return q
}

func (q *OneTimePasswordFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *OneTimePasswordFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *OneTimePasswordFilter) WithTrackingIDs(trackingIDs ...uuid.UUID) *OneTimePasswordFilter {
	q.TrackingIDs = trackingIDs
	return q
}

func (q *OneTimePasswordFilter) WithValid(valid *bool) *OneTimePasswordFilter {
	q.Valid = valid
	return q
}

func (q *OneTimePasswordFilter) WithPaginationURL(paginationURL string) *OneTimePasswordFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *OneTimePasswordFilter) WithSortings(sortings ...Sorting) *OneTimePasswordFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *OneTimePasswordFilter) WithLimit(limit int64) *OneTimePasswordFilter {
	q.Limit = limit
	return q
}

func (q *OneTimePasswordFilter) WithPage(page int64) *OneTimePasswordFilter {
	q.Page = page
	return q
}

func (q *OneTimePasswordFilter) WithUrlValues(urlValues url.Values) *OneTimePasswordFilter {
	q.UrlValues = urlValues
	return q
}
