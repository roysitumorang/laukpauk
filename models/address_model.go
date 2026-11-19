package models

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/nyaruka/phonenumbers"
	"github.com/roysitumorang/laukpauk/helper"
)

const AddressSortingFieldMainFirst = "main"

var MapAddressSortings = map[string]struct{}{
	SortingFieldID:               {},
	AddressSortingFieldMainFirst: {},
}
var ErrAddressNotFound = errors.New("address not found")

type Address struct {
	RowNo                int64        `json:"-"`
	ID                   uuid.UUID    `json:"id"`
	RecordStatus         RecordStatus `json:"record_status"`
	UserID               uuid.UUID    `json:"user_id"`
	Longitude            string       `json:"longitude"`
	Latitude             string       `json:"latitude"`
	StreetAddress        string       `json:"street_address"`
	DeliveryNote         string       `json:"delivery_note"`
	Label                string       `json:"label"`
	RecipientName        string       `json:"recipient_name"`
	RecipientMobilePhone string       `json:"recipient_mobile_phone"`
	Main                 bool         `json:"main"`
	CreatedAt            time.Time    `json:"-"`
	UpdatedAt            time.Time    `json:"-"`
	DeletedAt            *time.Time   `json:"-"`
	DeletionReason       *string      `json:"-"`
}

func (q *Address) Validate() error {
	if q.Longitude = strings.TrimSpace(q.Longitude); q.Longitude == "" {
		return errors.New("longitude: is required")
	}
	_, err := strconv.ParseFloat(q.Longitude, 64)
	if err != nil {
		return errors.New("longitude: is not a valid numeric")
	}
	if q.Latitude = strings.TrimSpace(q.Latitude); q.Latitude == "" {
		return errors.New("latitude: is required")
	}
	if _, err = strconv.ParseFloat(q.Latitude, 64); err != nil {
		return errors.New("latitude: is not a valid numeric")
	}
	if q.StreetAddress = strings.TrimSpace(q.StreetAddress); q.StreetAddress == "" {
		return errors.New("street_address: is required")
	}
	q.DeliveryNote = strings.TrimSpace(q.DeliveryNote)
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

func (q *Address) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

type AddressFilter struct {
	AddressIDs,
	UserIDs []uuid.UUID
	RecordStatuses []RecordStatus
	Longitude,
	Latitude,
	Label string
	Main *bool
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewAddressFilter() *AddressFilter {
	return &AddressFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *AddressFilter) WithAddressIDs(addressIDs ...uuid.UUID) *AddressFilter {
	q.AddressIDs = addressIDs
	return q
}

func (q *AddressFilter) WithUserIDs(userIDs ...uuid.UUID) *AddressFilter {
	q.UserIDs = userIDs
	return q
}

func (q *AddressFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *AddressFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *AddressFilter) WithLongitude(longitude string) *AddressFilter {
	q.Longitude = longitude
	return q
}

func (q *AddressFilter) WithLatitude(latitude string) *AddressFilter {
	q.Latitude = latitude
	return q
}

func (q *AddressFilter) WithLabel(label string) *AddressFilter {
	q.Label = label
	return q
}

func (q *AddressFilter) WithMain(main *bool) *AddressFilter {
	q.Main = main
	return q
}

func (q *AddressFilter) WithKeyword(keyword string) *AddressFilter {
	q.Keyword = keyword
	return q
}

func (q *AddressFilter) WithPaginationURL(paginationURL string) *AddressFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *AddressFilter) WithSortings(sortings ...Sorting) *AddressFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *AddressFilter) WithLimit(limit int64) *AddressFilter {
	q.Limit = limit
	return q
}

func (q *AddressFilter) WithPage(page int64) *AddressFilter {
	q.Page = page
	return q
}

func (q *AddressFilter) WithUrlValues(urlValues url.Values) *AddressFilter {
	q.UrlValues = urlValues
	return q
}
