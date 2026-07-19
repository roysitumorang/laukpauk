package models

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrVariableNotFound            = errors.New("variable not found")
	ErrUniqueVariableNameViolation = errors.New("name: already exists")
	VariableNameRegex              = regexp.MustCompile("^[a-z][a-z_]+[a-z]$")
)

const (
	VariableNameSetting = "setting"
)

type (
	Variable struct {
		RowNo     int64          `json:"-" form:"-"`
		ID        uuid.UUID      `json:"-" form:"id"`
		Name      string         `json:"-" form:"name"`
		Value     map[string]any `json:"-" form:"value"`
		CreatedBy uuid.UUID      `json:"-" form:"-"`
		CreatedAt time.Time      `json:"-" form:"-"`
		UpdatedBy uuid.UUID      `json:"-" form:"-"`
		UpdatedAt time.Time      `json:"-" form:"-"`
	}

	VariableFilter struct {
		VariableIDs,
		ExcludedVariableIDs []uuid.UUID
		Names []string
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}

	VariableSetting struct {
		MaintenanceMode                      bool            `json:"maintenance_mode"`
		VoucherEnabled                       bool            `json:"voucher_enabled"`
		ReferralPointsEnabled                bool            `json:"referral_points_enabled"`
		ReferralPointsBudget                 int64           `json:"referral_points_budget"`
		ReferralPointsPerReferee             int64           `json:"referral_points_per_referee"`
		ReferralPointsSlots                  []int64         `json:"referral_points_slots"`
		RewardPointsAccumulationWindowInDays int             `json:"reward_points_accumulation_window_in_days"`
		DeliveryOriginLongitude              decimal.Decimal `json:"delivery_origin_longitude"`
		DeliveryOriginLatitude               decimal.Decimal `json:"delivery_origin_latitude"`
		DeliveryCostPerKM                    int64           `json:"delivery_cost_per_km"`
		DeliveryMaxDistanceKM                int64           `json:"delivery_max_distance_km"`
		WhatsappPhoneNumber                  string          `json:"whatsapp_phone_number"`
	}
)

func NewVariableFilter() *VariableFilter {
	return &VariableFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *VariableFilter) WithVariableIDs(settingIDs ...uuid.UUID) *VariableFilter {
	q.VariableIDs = settingIDs
	return q
}

func (q *VariableFilter) WithExcludedVariableIDs(excludedVariableIDs ...uuid.UUID) *VariableFilter {
	q.ExcludedVariableIDs = excludedVariableIDs
	return q
}

func (q *VariableFilter) WithNames(names ...string) *VariableFilter {
	q.Names = names
	return q
}

func (q *VariableFilter) WithKeyword(keyword string) *VariableFilter {
	q.Keyword = keyword
	return q
}

func (q *VariableFilter) WithPaginationURL(paginationURL string) *VariableFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *VariableFilter) WithSortings(sortings ...Sorting) *VariableFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *VariableFilter) WithLimit(limit int64) *VariableFilter {
	q.Limit = limit
	return q
}

func (q *VariableFilter) WithPage(page int64) *VariableFilter {
	q.Page = page
	return q
}

func (q *VariableFilter) WithUrlValues(urlValues url.Values) *VariableFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Variable) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if !VariableNameRegex.MatchString(q.Name) {
		return errors.New("name: is invalid")
	}
	return nil
}
