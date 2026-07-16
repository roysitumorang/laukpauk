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
	ErrDeliveryPlanNotFound                         = errors.New("delivery plan not found")
	ErrUniqueDeliveryPlanNameViolation              = errors.New("name: already exists")
	ErrUniqueDeliveryPlanTransactionWindowViolation = errors.New("transaction time frame already exists")

	MapDeliveryPlanSortings = map[string]struct{}{
		SortingFieldID:   {},
		SortingFieldName: {},
	}
)

const (
	DeliverySpeedSameDay DeliverySpeed = iota
	DeliverySpeedNextDay
	DeliverySpeedNextTwoDay
)

type (
	DeliverySpeed int

	DeliveryPlan struct {
		RowNo                   int64          `json:"-" form:"-"`
		ID                      uuid.UUID      `json:"id" form:"-"`
		RecordStatus            RecordStatus   `json:"-" form:"-"`
		Name                    string         `json:"name" form:"name"`
		TransactionWindowMin    time.Time      `json:"transaction_window_min" form:"-"`
		RawTransactionWindowMin string         `json:"-" form:"transaction_window_min"`
		TransactionWindowMax    time.Time      `json:"transaction_window_max" form:"-"`
		RawTransactionWindowMax string         `json:"-" form:"transaction_window_max"`
		SameDay                 bool           `json:"same_day" form:"same_day"`
		SameDayWindowMin        *time.Time     `json:"-" form:"-"`
		RawSameDayWindowMin     string         `json:"-" form:"same_day_window_min"`
		SameDayWindowMax        *time.Time     `json:"-" form:"-"`
		RawSameDayWindowMax     string         `json:"-" form:"same_day_window_max"`
		SameDayInterval         *time.Duration `json:"-" form:"-"`
		RawSameDayInterval      *string        `json:"-" form:"same_day_interval"`
		SameDayHours            []time.Time    `json:"same_day_hours" form:"-"`
		NextDay                 bool           `json:"next_day" form:"next_day"`
		NextDayWindowMin        *time.Time     `json:"-" form:"-"`
		RawNextDayWindowMin     string         `json:"-" form:"next_day_window_min"`
		NextDayWindowMax        *time.Time     `json:"-" form:"-"`
		RawNextDayWindowMax     string         `json:"-" form:"next_day_window_max"`
		NextDayInterval         *time.Duration `json:"-" form:"-"`
		RawNextDayInterval      *string        `json:"-" form:"next_day_interval"`
		NextDayHours            []time.Time    `json:"next_day_hours" form:"-"`
		NextTwoDay              bool           `json:"next_two_day" form:"next_two_day"`
		NextTwoDayWindowMin     *time.Time     `json:"-" form:"-"`
		RawNextTwoDayWindowMin  string         `json:"-" form:"next_two_day_window_min"`
		NextTwoDayWindowMax     *time.Time     `json:"-" form:"-"`
		RawNextTwoDayWindowMax  string         `json:"-" form:"next_two_day_window_max"`
		NextTwoDayInterval      *time.Duration `json:"-" form:"-"`
		RawNextTwoDayInterval   *string        `json:"-" form:"next_two_day_interval"`
		NextTwoDayHours         []time.Time    `json:"next_two_day_hours" form:"-"`
		CreatedBy               uuid.UUID      `json:"-" form:"-"`
		CreatedAt               time.Time      `json:"-" form:"-"`
		UpdatedBy               uuid.UUID      `json:"-" form:"-"`
		UpdatedAt               time.Time      `json:"-" form:"-"`
		DeletedBy               *uuid.UUID     `json:"-" form:"-"`
		DeletedAt               *time.Time     `json:"-" form:"-"`
		DeletionReason          *string        `json:"-" form:"-"`
	}

	DeliveryPlanFilter struct {
		DeliveryPlanIDs,
		ExcludedDeliveryPlanIDs []uuid.UUID
		RecordStatuses    []RecordStatus
		Name              string
		TransactionWindow *TransactionWindow
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
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

func (q *DeliveryPlanFilter) WithName(name string) *DeliveryPlanFilter {
	q.Name = name
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
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
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
		if q.RawSameDayWindowMax = strings.TrimSpace(q.RawSameDayWindowMax); q.RawSameDayWindowMax == "" {
			return errors.New("same_day_window_max: is required")
		}
		sameDayWindowMax, err := time.ParseInLocation(time.TimeOnly, q.RawSameDayWindowMax, location)
		if err != nil {
			return fmt.Errorf("same_day_window_max: should use format '%s'", time.TimeOnly)
		}
		if !sameDayWindowMax.After(sameDayWindowMin) {
			return errors.New("same_day_window_max: should exceed same_day_window_min")
		}
		if q.RawSameDayInterval == nil || strings.TrimSpace(*q.RawSameDayInterval) == "" {
			return errors.New("same_day_interval: is required")
		}
		if _, err = time.ParseDuration(*q.RawSameDayInterval); err != nil {
			return errors.New("same_day_interval: is not valid")
		}
		q.SameDayWindowMin = &sameDayWindowMin
		q.SameDayWindowMax = &sameDayWindowMax
	} else {
		q.RawSameDayWindowMin = ""
		q.SameDayWindowMin = nil
		q.RawSameDayWindowMax = ""
		q.SameDayWindowMax = nil
		q.RawSameDayInterval = nil
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
		if q.RawNextDayWindowMax = strings.TrimSpace(q.RawNextDayWindowMax); q.RawNextDayWindowMax == "" {
			return errors.New("next_day_window_max: is required")
		}
		nextDayWindowMax, err := time.ParseInLocation(time.TimeOnly, q.RawNextDayWindowMax, location)
		if err != nil {
			return fmt.Errorf("next_day_window_max: should use format '%s'", time.TimeOnly)
		}
		if !nextDayWindowMax.After(nextDayWindowMin) {
			return errors.New("next_day_window_max: should exceed next_day_window_min")
		}
		if q.RawNextDayInterval == nil || strings.TrimSpace(*q.RawNextDayInterval) == "" {
			return errors.New("next_day_interval: is required")
		}
		if _, err = time.ParseDuration(*q.RawNextDayInterval); err != nil {
			return errors.New("next_day_interval: is not valid")
		}
		nextDayWindowMin = nextDayWindowMin.AddDate(0, 0, 1)
		nextDayWindowMax = nextDayWindowMax.AddDate(0, 0, 1)
		q.NextDayWindowMin = &nextDayWindowMin
		q.NextDayWindowMax = &nextDayWindowMax
	} else {
		q.RawNextDayWindowMin = ""
		q.NextDayWindowMin = nil
		q.RawNextDayWindowMax = ""
		q.NextDayWindowMax = nil
		q.RawNextDayInterval = nil
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
		if q.RawNextTwoDayWindowMax = strings.TrimSpace(q.RawNextTwoDayWindowMax); q.RawNextTwoDayWindowMax == "" {
			return errors.New("next_two_day_window_max: is required")
		}
		nextTwoDayWindowMax, err := time.ParseInLocation(time.TimeOnly, q.RawNextTwoDayWindowMax, location)
		if err != nil {
			return fmt.Errorf("next_two_day_window_max: should use format '%s'", time.TimeOnly)
		}
		if !nextTwoDayWindowMax.After(nextTwoDayWindowMin) {
			return errors.New("next_two_day_window_max: should exceed next_two_day_window_min")
		}
		if q.RawNextTwoDayInterval == nil || strings.TrimSpace(*q.RawNextTwoDayInterval) == "" {
			return errors.New("next_two_day_interval: is required")
		}
		if _, err = time.ParseDuration(*q.RawNextTwoDayInterval); err != nil {
			return errors.New("next_two_day_interval: is not valid")
		}
		nextTwoDayWindowMin = nextTwoDayWindowMin.AddDate(0, 0, 2)
		nextTwoDayWindowMax = nextTwoDayWindowMax.AddDate(0, 0, 2)
		q.NextTwoDayWindowMin = &nextTwoDayWindowMin
		q.NextTwoDayWindowMax = &nextTwoDayWindowMax
	} else {
		q.RawNextTwoDayWindowMin = ""
		q.NextTwoDayWindowMin = nil
		q.RawNextTwoDayWindowMax = ""
		q.NextTwoDayWindowMax = nil
		q.RawNextTwoDayInterval = nil
		q.NextTwoDayInterval = nil
	}
	return nil
}

func (q *DeliveryPlan) Populate() *DeliveryPlan {
	location := helper.LoadTimeZone()
	now := time.Now().In(location)
	q.TransactionWindowMin = q.TransactionWindowMin.In(location)
	q.TransactionWindowMin = time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		q.TransactionWindowMin.Hour(),
		q.TransactionWindowMin.Minute(),
		q.TransactionWindowMin.Second(),
		q.TransactionWindowMin.Nanosecond(),
		q.TransactionWindowMin.Location(),
	)
	q.TransactionWindowMax = q.TransactionWindowMax.In(location)
	q.TransactionWindowMax = time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		q.TransactionWindowMax.Hour(),
		q.TransactionWindowMax.Minute(),
		q.TransactionWindowMax.Second(),
		q.TransactionWindowMax.Nanosecond(),
		q.TransactionWindowMax.Location(),
	)
	if q.SameDay && q.SameDayWindowMin != nil && q.SameDayWindowMax != nil && q.RawSameDayInterval != nil {
		interval, _ := time.ParseDuration(*q.RawSameDayInterval)
		q.SameDayInterval = &interval
		*q.SameDayWindowMin = q.SameDayWindowMin.In(location)
		*q.SameDayWindowMin = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.SameDayWindowMin.Hour(),
			q.SameDayWindowMin.Minute(),
			q.SameDayWindowMin.Second(),
			q.SameDayWindowMin.Nanosecond(),
			q.SameDayWindowMin.Location(),
		)
		*q.SameDayWindowMax = q.SameDayWindowMax.In(location)
		*q.SameDayWindowMax = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.SameDayWindowMax.Hour(),
			q.SameDayWindowMax.Minute(),
			q.SameDayWindowMax.Second(),
			q.SameDayWindowMax.Nanosecond(),
			q.SameDayWindowMax.Location(),
		)
		sameDayWindowMin := *q.SameDayWindowMin
		sameDayWindowMax := *q.SameDayWindowMax
		for {
			q.SameDayHours = append(q.SameDayHours, sameDayWindowMin)
			if sameDayWindowMin.Equal(sameDayWindowMax) {
				break
			}
			sameDayWindowMin = sameDayWindowMin.Add(*q.SameDayInterval)
		}
	}
	now = now.AddDate(0, 0, 1)
	if q.NextDay && q.NextDayWindowMin != nil && q.NextDayWindowMax != nil && q.RawNextDayInterval != nil {
		interval, _ := time.ParseDuration(*q.RawNextDayInterval)
		q.NextDayInterval = &interval
		*q.NextDayWindowMin = q.NextDayWindowMin.In(location)
		*q.NextDayWindowMin = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.NextDayWindowMin.Hour(),
			q.NextDayWindowMin.Minute(),
			q.NextDayWindowMin.Second(),
			q.NextDayWindowMin.Nanosecond(),
			q.NextDayWindowMin.Location(),
		)
		*q.NextDayWindowMax = q.NextDayWindowMax.In(location)
		*q.NextDayWindowMax = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.NextDayWindowMax.Hour(),
			q.NextDayWindowMax.Minute(),
			q.NextDayWindowMax.Second(),
			q.NextDayWindowMax.Nanosecond(),
			q.NextDayWindowMax.Location(),
		)
		nextDayWindowMin := *q.NextDayWindowMin
		nextDayWindowMax := *q.NextDayWindowMax
		for {
			q.NextDayHours = append(q.NextDayHours, nextDayWindowMin)
			if nextDayWindowMin.Equal(nextDayWindowMax) {
				break
			}
			nextDayWindowMin = nextDayWindowMin.Add(*q.NextDayInterval)
		}
	}
	now = now.AddDate(0, 0, 1)
	if q.NextTwoDay && q.NextTwoDayWindowMin != nil && q.NextTwoDayWindowMax != nil && q.RawNextTwoDayInterval != nil {
		interval, _ := time.ParseDuration(*q.RawNextTwoDayInterval)
		q.NextTwoDayInterval = &interval
		*q.NextTwoDayWindowMin = q.NextTwoDayWindowMin.In(location)
		*q.NextTwoDayWindowMin = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.NextTwoDayWindowMin.Hour(),
			q.NextTwoDayWindowMin.Minute(),
			q.NextTwoDayWindowMin.Second(),
			q.NextTwoDayWindowMin.Nanosecond(),
			q.NextTwoDayWindowMin.Location(),
		)
		*q.NextTwoDayWindowMax = q.NextTwoDayWindowMax.In(location)
		*q.NextTwoDayWindowMax = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.NextTwoDayWindowMax.Hour(),
			q.NextTwoDayWindowMax.Minute(),
			q.NextTwoDayWindowMax.Second(),
			q.NextTwoDayWindowMax.Nanosecond(),
			q.NextTwoDayWindowMax.Location(),
		)
		nextTwoDayWindowMin := *q.NextTwoDayWindowMin
		nextTwoDayWindowMax := *q.NextTwoDayWindowMax
		for {
			q.NextTwoDayHours = append(q.NextTwoDayHours, nextTwoDayWindowMin)
			if nextTwoDayWindowMin.Equal(nextTwoDayWindowMax) {
				break
			}
			nextTwoDayWindowMin = nextTwoDayWindowMin.Add(*q.NextTwoDayInterval)
		}
	}
	return q
}
