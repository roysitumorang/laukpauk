package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrDeliveryPlanNotFound                          = errors.New("delivery plan not found")
	ErrUniqueDeliveryPlanNameViolation               = errors.New("name: already exists")
	ErrUniqueDeliveryPlanDayViolation                = errors.New("day: already exists")
	ErrUniqueDeliveryTransactionWindowFrameViolation = errors.New("transaction time frame already exists")

	DeliveryPlanTypes = []DeliveryPlanType{
		DeliveryPlanTypeStandard,
		DeliveryPlanTypeCustom,
	}
	MapDeliveryPlanTypes = map[DeliveryPlanType]struct{}{
		DeliveryPlanTypeStandard: {},
		DeliveryPlanTypeCustom:   {},
	}

	MapDeliveryPlanSortings = map[string]struct{}{
		SortingFieldID:   {},
		SortingFieldName: {},
	}
)

const (
	DeliveryPlanTypeStandard DeliveryPlanType = iota
	DeliveryPlanTypeCustom
)

const (
	DeliverySpeedSameDay DeliverySpeed = iota
	DeliverySpeedNextDay
	DeliverySpeedNextTwoDay
)

type (
	DeliveryPlanType int
	DeliverySpeed    int

	DeliveryPlan struct {
		RowNo                   int64            `json:"-" form:"-"`
		ID                      uuid.UUID        `json:"-" form:"-"`
		RecordStatus            RecordStatus     `json:"-" form:"-"`
		DeliveryPlanType        DeliveryPlanType `json:"-" form:"delivery_plan_type"`
		Name                    *string          `json:"-" form:"name"`
		RawDay                  *string          `json:"-" form:"day"`
		Day                     *time.Time       `json:"-" form:"-"`
		TransactionWindowMin    time.Time        `json:"-" form:"-"`
		RawTransactionWindowMin string           `json:"-" form:"transaction_window_min"`
		TransactionWindowMax    time.Time        `json:"-" form:"-"`
		RawTransactionWindowMax string           `json:"-" form:"transaction_window_max"`
		SameDay                 bool             `json:"-" form:"same_day"`
		SameDayWindowMin        *time.Time       `json:"-" form:"-"`
		RawSameDayWindowMin     string           `json:"-" form:"same_day_window_min"`
		SameDayWindowMax        *time.Time       `json:"-" form:"-"`
		RawSameDayWindowMax     string           `json:"-" form:"same_day_window_max"`
		SameDayInterval         *string          `json:"-" form:"same_day_interval"`
		SameDayHours            []time.Time      `json:"-" form:"-"`
		NextDay                 bool             `json:"-" form:"next_day"`
		NextDayWindowMin        *time.Time       `json:"-" form:"-"`
		RawNextDayWindowMin     string           `json:"-" form:"next_day_window_min"`
		NextDayWindowMax        *time.Time       `json:"-" form:"-"`
		RawNextDayWindowMax     string           `json:"-" form:"next_day_window_max"`
		NextDayInterval         *string          `json:"-" form:"next_day_interval"`
		NextDayHours            []time.Time      `json:"-" form:"-"`
		NextTwoDay              bool             `json:"-" form:"next_two_day"`
		NextTwoDayWindowMin     *time.Time       `json:"-" form:"-"`
		RawNextTwoDayWindowMin  string           `json:"-" form:"next_two_day_window_min"`
		NextTwoDayWindowMax     *time.Time       `json:"-" form:"-"`
		RawNextTwoDayWindowMax  string           `json:"-" form:"next_two_day_window_max"`
		NextTwoDayInterval      *string          `json:"-" form:"next_two_day_interval"`
		NextTwoDayHours         []time.Time      `json:"-" form:"-"`
		CreatedBy               uuid.UUID        `json:"-" form:"-"`
		CreatedAt               time.Time        `json:"-" form:"-"`
		UpdatedBy               uuid.UUID        `json:"-" form:"-"`
		UpdatedAt               time.Time        `json:"-" form:"-"`
		DeletedBy               *uuid.UUID       `json:"-" form:"-"`
		DeletedAt               *time.Time       `json:"-" form:"-"`
		DeletionReason          *string          `json:"-" form:"-"`
	}

	DeliveryPlanFilter struct {
		DeliveryPlanIDs,
		ExcludedDeliveryPlanIDs []uuid.UUID
		RecordStatuses    []RecordStatus
		DeliveryPlanTypes []DeliveryPlanType
		Name              string
		Day               time.Time
		TransactionWindow *TransactionWindow
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings *Map
		Limit,
		Page int64
		UrlValues url.Values
	}

	TransactionWindow struct {
		Min,
		Max time.Time
	}
)

func NewDeliveryPlanFilter() *DeliveryPlanFilter {
	return new(DeliveryPlanFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *DeliveryPlanFilter) WithDeliveryPlanIDs(deliveryPlanIDs ...uuid.UUID) *DeliveryPlanFilter {
	q.DeliveryPlanIDs = deliveryPlanIDs
	return q
}

func (q *DeliveryPlanFilter) WithExcludedDeliveryPlanIDs(excludedDeliveryPlanIDs ...uuid.UUID) *DeliveryPlanFilter {
	q.ExcludedDeliveryPlanIDs = excludedDeliveryPlanIDs
	return q
}

func (q *DeliveryPlanFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *DeliveryPlanFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *DeliveryPlanFilter) WithDeliveryPlanTypes(deliveryPlanTypes ...DeliveryPlanType) *DeliveryPlanFilter {
	q.DeliveryPlanTypes = deliveryPlanTypes
	return q
}

func (q *DeliveryPlanFilter) WithName(name string) *DeliveryPlanFilter {
	q.Name = name
	return q
}

func (q *DeliveryPlanFilter) WithDay(day time.Time) *DeliveryPlanFilter {
	q.Day = day
	return q
}

func (q *DeliveryPlanFilter) WithTransactionWindow(min, max time.Time) *DeliveryPlanFilter {
	q.TransactionWindow = new(TransactionWindow{
		Min: min,
		Max: max,
	})
	return q
}

func (q *DeliveryPlanFilter) WithKeyword(keyword string) *DeliveryPlanFilter {
	q.Keyword = keyword
	return q
}

func (q *DeliveryPlanFilter) WithPaginationURL(paginationURL string) *DeliveryPlanFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *DeliveryPlanFilter) WithSortings(sortings ...Sorting) *DeliveryPlanFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *DeliveryPlanFilter) WithLimit(limit int64) *DeliveryPlanFilter {
	q.Limit = limit
	return q
}

func (q *DeliveryPlanFilter) WithPage(page int64) *DeliveryPlanFilter {
	q.Page = page
	return q
}

func (q *DeliveryPlanFilter) WithUrlValues(urlValues url.Values) *DeliveryPlanFilter {
	q.UrlValues = urlValues
	return q
}

func (q *DeliveryPlan) Validate() (err error) {
	location := helper.LoadTimeZone()
	switch q.DeliveryPlanType {
	case DeliveryPlanTypeCustom:
		if q.RawDay == nil {
			return errors.New("day: is required")
		}
		day, err := time.ParseInLocation(time.DateOnly, *q.RawDay, location)
		if err != nil {
			return fmt.Errorf("day: should use format '%s'", time.DateOnly)
		}
		q.Day = new(day)
		if !q.Day.After(time.Now()) {
			return errors.New("day: should exceed current day")
		}
		q.Name = nil
	case DeliveryPlanTypeStandard:
		if q.Name == nil {
			return errors.New("name: is required")
		}
		if *q.Name = strings.TrimSpace(*q.Name); *q.Name == "" {
			return errors.New("name: is required")
		}
		q.Day = nil
	default:
		return errors.New("delivery_plan_type: not found")
	}
	if q.RawTransactionWindowMin = strings.TrimSpace(q.RawTransactionWindowMin); q.RawTransactionWindowMin == "" {
		return errors.New("transaction_window_min: is required")
	}
	if q.TransactionWindowMin, err = time.ParseInLocation(time.TimeOnly, q.RawTransactionWindowMin, location); err != nil {
		return fmt.Errorf("transaction_window_min: should use format '%s'", time.TimeOnly)
	}
	if q.RawTransactionWindowMax = strings.TrimSpace(q.RawTransactionWindowMax); q.RawTransactionWindowMax == "" {
		return errors.New("transaction_window_max: is required")
	}
	if q.TransactionWindowMax, err = time.ParseInLocation(time.TimeOnly, q.RawTransactionWindowMax, location); err != nil {
		return fmt.Errorf("transaction_window_max: should use format '%s'", time.TimeOnly)
	}
	if !q.TransactionWindowMax.After(q.TransactionWindowMin) {
		return errors.New("transaction_window_max: should exceed transaction_window_min")
	}
	if q.SameDay {
		if q.RawSameDayWindowMin = strings.TrimSpace(q.RawSameDayWindowMin); q.RawSameDayWindowMin == "" {
			return errors.New("same_day_window_min: is required")
		}
		sameDayWindowMin, err := time.ParseInLocation(time.TimeOnly, q.RawSameDayWindowMin, location)
		if err != nil {
			return fmt.Errorf("same_day_window_min: should use format '%s'", time.TimeOnly)
		}
		q.SameDayWindowMin = new(sameDayWindowMin)
		if q.RawSameDayWindowMax = strings.TrimSpace(q.RawSameDayWindowMax); q.RawSameDayWindowMax == "" {
			return errors.New("same_day_window_max: is required")
		}
		sameDayWindowMax, err := time.ParseInLocation(time.TimeOnly, q.RawSameDayWindowMax, location)
		if err != nil {
			return fmt.Errorf("same_day_window_max: should use format '%s'", time.TimeOnly)
		}
		q.SameDayWindowMax = new(sameDayWindowMax)
		if !q.SameDayWindowMax.After(*q.SameDayWindowMin) {
			return errors.New("same_day_window_max: should exceed same_day_window_min")
		}
		if q.SameDayInterval == nil || strings.TrimSpace(*q.SameDayInterval) == "" {
			return errors.New("same_day_interval: is required")
		}
		interval, err := time.ParseDuration(*q.SameDayInterval)
		if err != nil {
			return errors.New("same_day_interval: is not valid")
		}
		t := *q.SameDayWindowMin
		for {
			q.SameDayHours = append(q.SameDayHours, t)
			if t.Equal(*q.SameDayWindowMax) {
				break
			}
			t = t.Add(interval)
		}
	} else {
		q.RawSameDayWindowMin = ""
		q.SameDayWindowMin = nil
		q.RawSameDayWindowMax = ""
		q.SameDayWindowMax = nil
		q.SameDayInterval = nil
	}
	if q.NextDay {
		if q.RawNextDayWindowMin = strings.TrimSpace(q.RawNextDayWindowMin); q.RawNextDayWindowMin == "" {
			return errors.New("next_day_window_min: is required")
		}
		nextDayWindowMin, err := time.ParseInLocation(time.TimeOnly, q.RawNextDayWindowMin, location)
		if err != nil {
			return fmt.Errorf("next_day_window_min: should use format '%s'", time.TimeOnly)
		}
		q.NextDayWindowMin = new(nextDayWindowMin)
		if q.RawNextDayWindowMax = strings.TrimSpace(q.RawNextDayWindowMax); q.RawNextDayWindowMax == "" {
			return errors.New("next_day_window_max: is required")
		}
		nextDayWindowMax, err := time.ParseInLocation(time.TimeOnly, q.RawNextDayWindowMax, location)
		if err != nil {
			return fmt.Errorf("next_day_window_max: should use format '%s'", time.TimeOnly)
		}
		q.NextDayWindowMax = new(nextDayWindowMax)
		if !q.NextDayWindowMax.After(*q.NextDayWindowMin) {
			return errors.New("next_day_window_max: should exceed next_day_window_min")
		}
		if q.NextDayInterval == nil || strings.TrimSpace(*q.NextDayInterval) == "" {
			return errors.New("next_day_interval: is required")
		}
		interval, err := time.ParseDuration(*q.NextDayInterval)
		if err != nil {
			return errors.New("next_day_interval: is not valid")
		}
		t := *q.NextDayWindowMin
		for {
			q.NextDayHours = append(q.NextDayHours, t)
			if t.Equal(*q.NextDayWindowMax) {
				break
			}
			t = t.Add(interval)
		}
	} else {
		q.RawNextDayWindowMin = ""
		q.NextDayWindowMin = nil
		q.RawNextDayWindowMax = ""
		q.NextDayWindowMax = nil
		q.NextDayInterval = nil
	}
	if q.NextTwoDay {
		if q.RawNextTwoDayWindowMin = strings.TrimSpace(q.RawNextTwoDayWindowMin); q.RawNextTwoDayWindowMin == "" {
			return errors.New("next_two_day_window_min: is required")
		}
		nextTwoDayWindowMin, err := time.ParseInLocation(time.TimeOnly, q.RawNextTwoDayWindowMin, location)
		if err != nil {
			return fmt.Errorf("next_two_day_window_min: should use format '%s'", time.TimeOnly)
		}
		q.NextTwoDayWindowMin = new(nextTwoDayWindowMin)
		if q.RawNextTwoDayWindowMax = strings.TrimSpace(q.RawNextTwoDayWindowMax); q.RawNextTwoDayWindowMax == "" {
			return errors.New("next_two_day_window_max: is required")
		}
		nextTwoDayWindowMax, err := time.ParseInLocation(time.TimeOnly, q.RawNextTwoDayWindowMax, location)
		if err != nil {
			return fmt.Errorf("next_two_day_window_max: should use format '%s'", time.TimeOnly)
		}
		q.NextTwoDayWindowMax = new(nextTwoDayWindowMax)
		if !q.NextTwoDayWindowMax.After(*q.NextTwoDayWindowMin) {
			return errors.New("next_two_day_window_max: should exceed next_two_day_window_min")
		}
		if q.NextTwoDayInterval == nil || strings.TrimSpace(*q.NextTwoDayInterval) == "" {
			return errors.New("next_two_day_interval: is required")
		}
		interval, err := time.ParseDuration(*q.NextTwoDayInterval)
		if err != nil {
			return errors.New("next_two_day_interval: is not valid")
		}
		t := *q.NextTwoDayWindowMin
		for {
			q.NextTwoDayHours = append(q.NextTwoDayHours, t)
			if t.Equal(*q.NextTwoDayWindowMax) {
				break
			}
			t = t.Add(interval)
		}
	} else {
		q.RawNextTwoDayWindowMin = ""
		q.NextTwoDayWindowMin = nil
		q.RawNextTwoDayWindowMax = ""
		q.NextTwoDayWindowMax = nil
		q.NextTwoDayInterval = nil
	}
	return nil
}

func (q *DeliveryPlan) Populate() *DeliveryPlan {
	location := helper.LoadTimeZone()
	if q.Day != nil {
		*q.Day = q.Day.In(location)
		q.RawDay = new(q.Day.Format(time.DateOnly))
	}
	q.TransactionWindowMin = q.TransactionWindowMin.In(location)
	q.TransactionWindowMax = q.TransactionWindowMax.In(location)
	if q.SameDayWindowMin != nil {
		*q.SameDayWindowMin = q.SameDayWindowMin.In(location)
	}
	if q.SameDayWindowMax != nil {
		*q.SameDayWindowMax = q.SameDayWindowMax.In(location)
	}
	if q.NextDayWindowMin != nil {
		*q.NextDayWindowMin = q.NextDayWindowMin.In(location)
	}
	if q.NextDayWindowMax != nil {
		*q.NextDayWindowMax = q.NextDayWindowMax.In(location)
	}
	if q.NextTwoDayWindowMin != nil {
		*q.NextTwoDayWindowMin = q.NextTwoDayWindowMin.In(location)
	}
	if q.NextTwoDayWindowMax != nil {
		*q.NextTwoDayWindowMax = q.NextTwoDayWindowMax.In(location)
	}
	return q
}

func (q DeliveryPlanType) Int() int {
	return int(q)
}

func (q DeliveryPlanType) String() string {
	if q == DeliveryPlanTypeCustom {
		return "custom"
	}
	return "standard"
}

func (q DeliveryPlanType) Equal(v DeliveryPlanType) bool {
	return q.Int() == v.Int()
}
