package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/nyaruka/phonenumbers"
	"github.com/roysitumorang/laukpauk/helper"
)

const DeliveryAddressSortingFieldMainFirst = "main"

var MapDeliveryAddressSortings = map[string]struct{}{
	SortingFieldID:                       {},
	DeliveryAddressSortingFieldMainFirst: {},
}
var ErrDeliveryAddressNotFound = errors.New("delivery address not found")

type DeliveryAddress struct {
	RowNo                int64           `json:"-"`
	ID                   uuid.UUID       `json:"id"`
	RecordStatus         RecordStatus    `json:"record_status"`
	UserID               uuid.UUID       `json:"user_id"`
	OriginLongitude      decimal.Decimal `json:"origin_longitude"`
	OriginLatitude       decimal.Decimal `json:"origin_latitude"`
	DestinationLongitude decimal.Decimal `json:"destination_longitude"`
	DestinationLatitude  decimal.Decimal `json:"destination_latitude"`
	DistanceMeters       int64           `json:"-"`
	DistanceKM           int64           `json:"distance_km"`
	StreetAddress        string          `json:"street_address"`
	StreetAddressNote    string          `json:"street_address_note"`
	Label                string          `json:"label"`
	RecipientName        string          `json:"recipient_name"`
	RecipientMobilePhone string          `json:"recipient_mobile_phone"`
	Main                 bool            `json:"main"`
	CreatedAt            time.Time       `json:"-"`
	UpdatedAt            time.Time       `json:"-"`
	DeletedAt            *time.Time      `json:"-"`
	DeletionReason       *string         `json:"-"`
}

func (q *DeliveryAddress) Validate() (err error) {
	if q.DestinationLongitude.IsZero() {
		return errors.New("destination_longitude: is required")
	}
	if q.DestinationLatitude.IsZero() {
		return errors.New("destination_latitude: is required")
	}
	if q.StreetAddress = strings.TrimSpace(q.StreetAddress); q.StreetAddress == "" {
		return errors.New("street_address: is required")
	}
	q.StreetAddressNote = strings.TrimSpace(q.StreetAddressNote)
	if q.Label = strings.TrimSpace(q.Label); q.Label == "" {
		return errors.New("label: is required")
	}
	if q.RecipientName = strings.TrimSpace(q.RecipientName); q.RecipientName == "" {
		return errors.New("recipient_name: is required")
	}
	if q.RecipientMobilePhone = strings.TrimSpace(q.RecipientMobilePhone); q.RecipientMobilePhone == "" {
		return errors.New("recipient_mobile_phone: is required")
	}
	mobilePhone, err := phonenumbers.Parse(q.RecipientMobilePhone, "ID")
	if err != nil {
		return fmt.Errorf("recipient_mobile_phone: %w", err)
	}
	q.RecipientMobilePhone = phonenumbers.Format(mobilePhone, phonenumbers.E164)
	return nil
}

func (q *DeliveryAddress) Populate() *DeliveryAddress {
	return q
}

type DeliveryAddressFilter struct {
	DeliveryAddressIDs,
	ExcludedDeliveryAddressIDs,
	UserIDs []uuid.UUID
	RecordStatuses []RecordStatus
	OriginLongitudes,
	OriginLatitudes,
	DestinationLongitudes,
	DestinationLatitudes []decimal.Decimal
	Label string
	Main  *bool
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewDeliveryAddressFilter() *DeliveryAddressFilter {
	return &DeliveryAddressFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *DeliveryAddressFilter) WithDeliveryAddressIDs(deliveryAddressIDs ...uuid.UUID) *DeliveryAddressFilter {
	q.DeliveryAddressIDs = deliveryAddressIDs
	return q
}

func (q *DeliveryAddressFilter) WithExcludedDeliveryAddressIDs(excludedDeliveryAddressIDs ...uuid.UUID) *DeliveryAddressFilter {
	q.ExcludedDeliveryAddressIDs = excludedDeliveryAddressIDs
	return q
}

func (q *DeliveryAddressFilter) WithUserIDs(userIDs ...uuid.UUID) *DeliveryAddressFilter {
	q.UserIDs = userIDs
	return q
}

func (q *DeliveryAddressFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *DeliveryAddressFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *DeliveryAddressFilter) WithOriginLongitudes(originLongitudes ...decimal.Decimal) *DeliveryAddressFilter {
	q.OriginLongitudes = originLongitudes
	return q
}

func (q *DeliveryAddressFilter) WithOriginLatitudes(originLatitudes ...decimal.Decimal) *DeliveryAddressFilter {
	q.OriginLatitudes = originLatitudes
	return q
}

func (q *DeliveryAddressFilter) WithDestinationLongitudes(destinationLongitudes ...decimal.Decimal) *DeliveryAddressFilter {
	q.DestinationLongitudes = destinationLongitudes
	return q
}

func (q *DeliveryAddressFilter) WithDestinationLatitudes(destinationLatitudes ...decimal.Decimal) *DeliveryAddressFilter {
	q.DestinationLatitudes = destinationLatitudes
	return q
}

func (q *DeliveryAddressFilter) WithLabel(label string) *DeliveryAddressFilter {
	q.Label = label
	return q
}

func (q *DeliveryAddressFilter) WithMain(main *bool) *DeliveryAddressFilter {
	q.Main = main
	return q
}

func (q *DeliveryAddressFilter) WithKeyword(keyword string) *DeliveryAddressFilter {
	q.Keyword = keyword
	return q
}

func (q *DeliveryAddressFilter) WithPaginationURL(paginationURL string) *DeliveryAddressFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *DeliveryAddressFilter) WithSortings(sortings ...Sorting) *DeliveryAddressFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *DeliveryAddressFilter) WithLimit(limit int64) *DeliveryAddressFilter {
	q.Limit = limit
	return q
}

func (q *DeliveryAddressFilter) WithPage(page int64) *DeliveryAddressFilter {
	q.Page = page
	return q
}

func (q *DeliveryAddressFilter) WithUrlValues(urlValues url.Values) *DeliveryAddressFilter {
	q.UrlValues = urlValues
	return q
}
