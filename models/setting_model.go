package models

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrSettingNotFound            = errors.New("setting not found")
	ErrUniqueSettingNameViolation = errors.New("name: already exists")
	SettingNameRegex              = regexp.MustCompile("^[a-z][a-z_]+[a-z]$")

	MapSettingSortings = map[string]struct{}{
		SortingFieldID:          {},
		SettingSortingFieldName: {},
	}

	SettingNames = []SettingName{
		SettingNameMaintenanceMode,
		SettingNameVoucherEnabled,
		SettingNameReferralPointsEnabled,
		SettingNameReferralPointsBudget,
		SettingNameReferralPointsPerReferee,
		SettingNameReferralPointsAllocations,
		SettingNameRewardPointsAccumulationWindowInDays,
		SettingNameDeliveryOriginLatitude,
		SettingNameDeliveryOriginLongitude,
		SettingNameDeliveryCostPerKM,
		SettingNameDeliveryMaxDistanceKM,
		SettingNameWhatsappPhoneNumber,
	}

	MapSettingNames = map[string]struct{}{
		SettingNameMaintenanceMode.String():                      {},
		SettingNameVoucherEnabled.String():                       {},
		SettingNameReferralPointsEnabled.String():                {},
		SettingNameReferralPointsBudget.String():                 {},
		SettingNameReferralPointsPerReferee.String():             {},
		SettingNameReferralPointsAllocations.String():            {},
		SettingNameRewardPointsAccumulationWindowInDays.String(): {},
		SettingNameDeliveryOriginLatitude.String():               {},
		SettingNameDeliveryOriginLongitude.String():              {},
		SettingNameDeliveryCostPerKM.String():                    {},
		SettingNameDeliveryMaxDistanceKM.String():                {},
		SettingNameWhatsappPhoneNumber.String():                  {},
	}
)

const (
	SettingSortingFieldName = "name"

	SettingNameMaintenanceMode                      SettingName = "maintenance_mode"
	SettingNameVoucherEnabled                       SettingName = "voucher_enabled"
	SettingNameReferralPointsEnabled                SettingName = "referral_points_enabled"
	SettingNameReferralPointsBudget                 SettingName = "referral_points_budget"
	SettingNameReferralPointsPerReferee             SettingName = "referral_points_per_referee"
	SettingNameReferralPointsAllocations            SettingName = "referral_points_allocations"
	SettingNameRewardPointsAccumulationWindowInDays SettingName = "reward_points_accumulation_window_in_days"
	SettingNameDeliveryOriginLongitude              SettingName = "delivery_origin_longitude"
	SettingNameDeliveryOriginLatitude               SettingName = "delivery_origin_latitude"
	SettingNameDeliveryCostPerKM                    SettingName = "delivery_cost_per_km"
	SettingNameDeliveryMaxDistanceKM                SettingName = "delivery_max_distance_km"
	SettingNameWhatsappPhoneNumber                  SettingName = "whatsapp_phone_number"
)

type (
	SettingName string

	Setting struct {
		RowNo       int64      `json:"-" form:"-"`
		ID          uuid.UUID  `json:"-" form:"id"`
		Name        string     `json:"-" form:"name"`
		Value       string     `json:"-" form:"value"`
		Description string     `json:"-" form:"description"`
		CreatedBy   *uuid.UUID `json:"-" form:"-"`
		CreatedAt   time.Time  `json:"-" form:"-"`
		UpdatedBy   *uuid.UUID `json:"-" form:"-"`
		UpdatedAt   time.Time  `json:"-" form:"-"`
	}

	SettingFilter struct {
		SettingIDs,
		ExcludedSettingIDs []uuid.UUID
		Names        []string
		SettingNames []SettingName
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewSettingFilter() *SettingFilter {
	return &SettingFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *SettingFilter) WithSettingIDs(settingIDs ...uuid.UUID) *SettingFilter {
	q.SettingIDs = settingIDs
	return q
}

func (q *SettingFilter) WithExcludedSettingIDs(excludedSettingIDs ...uuid.UUID) *SettingFilter {
	q.ExcludedSettingIDs = excludedSettingIDs
	return q
}

func (q *SettingFilter) WithNames(names ...string) *SettingFilter {
	q.Names = names
	return q
}

func (q *SettingFilter) WithSettingNames(settingNames ...SettingName) *SettingFilter {
	q.SettingNames = settingNames
	return q
}

func (q *SettingFilter) WithKeyword(keyword string) *SettingFilter {
	q.Keyword = keyword
	return q
}

func (q *SettingFilter) WithPaginationURL(paginationURL string) *SettingFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *SettingFilter) WithSortings(sortings ...Sorting) *SettingFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *SettingFilter) WithLimit(limit int64) *SettingFilter {
	q.Limit = limit
	return q
}

func (q *SettingFilter) WithPage(page int64) *SettingFilter {
	q.Page = page
	return q
}

func (q *SettingFilter) WithUrlValues(urlValues url.Values) *SettingFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Setting) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if !SettingNameRegex.MatchString(q.Name) {
		return errors.New("name: is invalid")
	}
	if q.Value = strings.TrimSpace(q.Value); q.Value == "" {
		return errors.New("value: is required")
	}
	if q.Description = strings.TrimSpace(q.Description); q.Description == "" {
		return errors.New("description: is required")
	}
	return nil
}

func (q *Setting) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Setting) NameEditable() bool {
	_, ok := MapSettingNames[q.Name]
	return !ok
}

func (q SettingName) String() string {
	return string(q)
}
