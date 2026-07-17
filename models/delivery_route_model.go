package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrDeliveryRouteNotFound = errors.New("delivery route not found")
)

type (
	DeliveryRoute struct {
		RowNo                int64           `json:"-" form:"-"`
		ID                   uuid.UUID       `json:"id" form:"-"`
		OriginLongitude      decimal.Decimal `json:"origin_longitude" form:"-"`
		OriginLatitude       decimal.Decimal `json:"origin_latitude" form:"-"`
		DestinationLongitude decimal.Decimal `json:"destination_longitude" form:"-"`
		DestinationLatitude  decimal.Decimal `json:"destination_latitude" form:"-"`
		DistanceMeters       int64           `json:"-" form:"-"`
		DistanceKM           int64           `json:"distance_km" form:"-"`
		CreatedAt            time.Time       `json:"-" form:"-"`
		UpdatedAt            time.Time       `json:"-" form:"-"`
	}

	DeliveryRouteFilter struct {
		DeliveryRouteIDs []uuid.UUID
		OriginLongitude,
		OriginLatitude,
		DestinationLongitude,
		DestinationLatitude *decimal.Decimal
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewDeliveryRouteFilter() *DeliveryRouteFilter {
	return &DeliveryRouteFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *DeliveryRouteFilter) WithDeliveryRouteIDs(deliveryRouteIDs ...uuid.UUID) *DeliveryRouteFilter {
	q.DeliveryRouteIDs = deliveryRouteIDs
	return q
}

func (q *DeliveryRouteFilter) WithOriginLongitude(originLongitude *decimal.Decimal) *DeliveryRouteFilter {
	q.OriginLongitude = originLongitude
	return q
}

func (q *DeliveryRouteFilter) WithOriginLatitude(originLatitude *decimal.Decimal) *DeliveryRouteFilter {
	q.OriginLatitude = originLatitude
	return q
}

func (q *DeliveryRouteFilter) WithDestinationLongitude(destinationLongitude *decimal.Decimal) *DeliveryRouteFilter {
	q.DestinationLongitude = destinationLongitude
	return q
}

func (q *DeliveryRouteFilter) WithDestinationLatitude(destinationLatitude *decimal.Decimal) *DeliveryRouteFilter {
	q.DestinationLatitude = destinationLatitude
	return q
}

func (q *DeliveryRouteFilter) WithKeyword(keyword string) *DeliveryRouteFilter {
	q.Keyword = keyword
	return q
}

func (q *DeliveryRouteFilter) WithPaginationURL(paginationURL string) *DeliveryRouteFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *DeliveryRouteFilter) WithSortings(sortings ...Sorting) *DeliveryRouteFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *DeliveryRouteFilter) WithLimit(limit int64) *DeliveryRouteFilter {
	q.Limit = limit
	return q
}

func (q *DeliveryRouteFilter) WithPage(page int64) *DeliveryRouteFilter {
	q.Page = page
	return q
}

func (q *DeliveryRouteFilter) WithUrlValues(urlValues url.Values) *DeliveryRouteFilter {
	q.UrlValues = urlValues
	return q
}

func (q *DeliveryRoute) Validate() (err error) {
	if q.DestinationLongitude.IsZero() {
		return errors.New("destination_longitude: is required")
	}
	if q.DestinationLatitude.IsZero() {
		return errors.New("destination_latitude: is required")
	}
	return nil
}

func (q *DeliveryRoute) Populate() *DeliveryRoute {
	return q
}
