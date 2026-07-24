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
	ErrUniqueDeliveryPlanTransactionWindowViolation = errors.New("transaction time frame already exists")

	DeliveryWeekdays = []DeliveryWeekday{
		DeliveryWeekdaySunday,
		DeliveryWeekdayWeekdays,
		DeliveryWeekdayFriday,
		DeliveryWeekdaySaturday,
	}
	MapDeliveryWeekdays = map[DeliveryWeekday]struct{}{
		DeliveryWeekdaySunday:   {},
		DeliveryWeekdayWeekdays: {},
		DeliveryWeekdayFriday:   {},
		DeliveryWeekdaySaturday: {},
	}

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

const (
	DeliveryWeekdaySunday   DeliveryWeekday = DeliveryWeekday(time.Sunday)
	DeliveryWeekdayWeekdays DeliveryWeekday = DeliveryWeekday(time.Sunday + 1)
	DeliveryWeekdayFriday   DeliveryWeekday = DeliveryWeekday(time.Friday)
	DeliveryWeekdaySaturday DeliveryWeekday = DeliveryWeekday(time.Saturday)
)

type (
	DeliverySpeed   int
	DeliveryWeekday int

	DeliveryPlan struct {
		RowNo                   int64               `json:"-" form:"-"`
		ID                      uuid.UUID           `json:"id" form:"-"`
		RecordStatus            RecordStatus        `json:"-" form:"-"`
		Name                    string              `json:"name" form:"name"`
		Weekday                 DeliveryWeekday     `json:"weekday" form:"weekday"`
		TransactionWindowMin    time.Time           `json:"transaction_window_min" form:"-"`
		RawTransactionWindowMin string              `json:"-" form:"transaction_window_min"`
		TransactionWindowMax    time.Time           `json:"transaction_window_max" form:"-"`
		RawTransactionWindowMax string              `json:"-" form:"transaction_window_max"`
		SameDay                 bool                `json:"same_day" form:"same_day"`
		SameDayWindowMin        *time.Time          `json:"-" form:"-"`
		RawSameDayWindowMin     string              `json:"-" form:"same_day_window_min"`
		SameDayWindowMax        *time.Time          `json:"-" form:"-"`
		RawSameDayWindowMax     string              `json:"-" form:"same_day_window_max"`
		NextDay                 bool                `json:"next_day" form:"next_day"`
		NextDayWindowMin        *time.Time          `json:"-" form:"-"`
		RawNextDayWindowMin     string              `json:"-" form:"next_day_window_min"`
		NextDayWindowMax        *time.Time          `json:"-" form:"-"`
		RawNextDayWindowMax     string              `json:"-" form:"next_day_window_max"`
		NextTwoDay              bool                `json:"next_two_day" form:"next_two_day"`
		NextTwoDayWindowMin     *time.Time          `json:"-" form:"-"`
		RawNextTwoDayWindowMin  string              `json:"-" form:"next_two_day_window_min"`
		NextTwoDayWindowMax     *time.Time          `json:"-" form:"-"`
		RawNextTwoDayWindowMax  string              `json:"-" form:"next_two_day_window_max"`
		DeliveryTimes           []*DeliveryPlanTime `json:"delivery_times" form:"-"`
		CreatedBy               uuid.UUID           `json:"-" form:"-"`
		CreatedAt               time.Time           `json:"-" form:"-"`
		UpdatedBy               uuid.UUID           `json:"-" form:"-"`
		UpdatedAt               time.Time           `json:"-" form:"-"`
		DeletedBy               *uuid.UUID          `json:"-" form:"-"`
		DeletedAt               *time.Time          `json:"-" form:"-"`
		DeletionReason          *string             `json:"-" form:"-"`
	}

	DeliveryPlanTime struct {
		DeliverySpeed DeliverySpeed `json:"-" form:"-"`
		Datetime      time.Time     `json:"datetime" form:"-"`
	}

	DeliveryPlanFilter struct {
		DeliveryPlanIDs,
		ExcludedDeliveryPlanIDs []uuid.UUID
		RecordStatuses       []RecordStatus
		Name                 string
		DeliveryWeekdays     []DeliveryWeekday
		TransactionDatetime  time.Time
		DeliveryTimeInterval time.Duration
		TransactionWindow    *DeliveryPlanTransactionWindow
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}

	DeliveryPlanTransactionWindow struct {
		Min,
		Max time.Time
	}
)

func NewDeliveryPlanFilter() *DeliveryPlanFilter {
	return &DeliveryPlanFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
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

func (q *DeliveryPlanFilter) WithDeliveryWeekdays(deliveryWeekdays ...DeliveryWeekday) *DeliveryPlanFilter {
	q.DeliveryWeekdays = deliveryWeekdays
	return q
}

func (q *DeliveryPlanFilter) WithTransactionDatetime(transactionDatetime time.Time) *DeliveryPlanFilter {
	q.TransactionDatetime = transactionDatetime
	return q
}

func (q *DeliveryPlanFilter) WithDeliveryTimeInterval(deliveryTimeInterval time.Duration) *DeliveryPlanFilter {
	q.DeliveryTimeInterval = deliveryTimeInterval
	return q
}

func (q *DeliveryPlanFilter) WithTransactionWindow(min, max time.Time) *DeliveryPlanFilter {
	q.TransactionWindow = &DeliveryPlanTransactionWindow{
		Min: min,
		Max: max,
	}
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

func (q *DeliveryPlan) Validate(interval time.Duration) (err error) {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if _, ok := MapDeliveryWeekdays[q.Weekday]; !ok {
		return errors.New("weekday: not found")
	}
	if q.RawTransactionWindowMin = strings.TrimSpace(q.RawTransactionWindowMin); q.RawTransactionWindowMin == "" {
		return errors.New("transaction_window_min: is required")
	}
	t, err := time.ParseInLocation(time.TimeOnly, q.RawTransactionWindowMin, time.Local)
	if err != nil {
		return fmt.Errorf("transaction_window_min: should use format '%s'", time.TimeOnly)
	}
	q.TransactionWindowMin = time.Date(
		2000,
		1,
		1,
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	if q.RawTransactionWindowMax = strings.TrimSpace(q.RawTransactionWindowMax); q.RawTransactionWindowMax == "" {
		return errors.New("transaction_window_max: is required")
	}
	if t, err = time.ParseInLocation(time.TimeOnly, q.RawTransactionWindowMax, time.Local); err != nil {
		return fmt.Errorf("transaction_window_max: should use format '%s'", time.TimeOnly)
	}
	q.TransactionWindowMax = time.Date(
		2000,
		1,
		1,
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	if !q.TransactionWindowMax.After(q.TransactionWindowMin) {
		return errors.New("transaction_window_max: should exceed transaction_window_min")
	}
	if q.SameDay {
		if q.RawSameDayWindowMin = strings.TrimSpace(q.RawSameDayWindowMin); q.RawSameDayWindowMin == "" {
			return errors.New("same_day_window_min: is required")
		}
		if t, err = time.ParseInLocation(time.TimeOnly, q.RawSameDayWindowMin, time.Local); err != nil {
			return fmt.Errorf("same_day_window_min: should use format '%s'", time.TimeOnly)
		}
		sameDayWindowMin := time.Date(
			2000,
			1,
			1,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if q.RawSameDayWindowMax = strings.TrimSpace(q.RawSameDayWindowMax); q.RawSameDayWindowMax == "" {
			return errors.New("same_day_window_max: is required")
		}
		if t, err = time.ParseInLocation(time.TimeOnly, q.RawSameDayWindowMax, time.Local); err != nil {
			return fmt.Errorf("same_day_window_max: should use format '%s'", time.TimeOnly)
		}
		sameDayWindowMax := time.Date(
			2000,
			1,
			1,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if !sameDayWindowMax.After(sameDayWindowMin) {
			return errors.New("same_day_window_max: should exceed same_day_window_min")
		}
		q.SameDayWindowMin = &sameDayWindowMin
		q.SameDayWindowMax = &sameDayWindowMax
	} else {
		q.RawSameDayWindowMin = ""
		q.SameDayWindowMin = nil
		q.RawSameDayWindowMax = ""
		q.SameDayWindowMax = nil
	}
	if q.NextDay {
		if q.RawNextDayWindowMin = strings.TrimSpace(q.RawNextDayWindowMin); q.RawNextDayWindowMin == "" {
			return errors.New("next_day_window_min: is required")
		}
		if t, err = time.ParseInLocation(time.TimeOnly, q.RawNextDayWindowMin, time.Local); err != nil {
			return fmt.Errorf("next_day_window_min: should use format '%s'", time.TimeOnly)
		}
		nextDayWindowMin := time.Date(
			2000,
			1,
			1,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if q.RawNextDayWindowMax = strings.TrimSpace(q.RawNextDayWindowMax); q.RawNextDayWindowMax == "" {
			return errors.New("next_day_window_max: is required")
		}
		if t, err = time.ParseInLocation(time.TimeOnly, q.RawNextDayWindowMax, time.Local); err != nil {
			return fmt.Errorf("next_day_window_max: should use format '%s'", time.TimeOnly)
		}
		nextDayWindowMax := time.Date(
			2000,
			1,
			1,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if !nextDayWindowMax.After(nextDayWindowMin) {
			return errors.New("next_day_window_max: should exceed next_day_window_min")
		}
		q.NextDayWindowMin = new(nextDayWindowMin.AddDate(0, 0, 1))
		q.NextDayWindowMax = new(nextDayWindowMax.AddDate(0, 0, 1))
	} else {
		q.RawNextDayWindowMin = ""
		q.NextDayWindowMin = nil
		q.RawNextDayWindowMax = ""
		q.NextDayWindowMax = nil
	}
	if q.NextTwoDay {
		if q.RawNextTwoDayWindowMin = strings.TrimSpace(q.RawNextTwoDayWindowMin); q.RawNextTwoDayWindowMin == "" {
			return errors.New("next_two_day_window_min: is required")
		}
		if t, err = time.ParseInLocation(time.TimeOnly, q.RawNextTwoDayWindowMin, time.Local); err != nil {
			return fmt.Errorf("next_two_day_window_min: should use format '%s'", time.TimeOnly)
		}
		nextTwoDayWindowMin := time.Date(
			2000,
			1,
			1,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if q.RawNextTwoDayWindowMax = strings.TrimSpace(q.RawNextTwoDayWindowMax); q.RawNextTwoDayWindowMax == "" {
			return errors.New("next_two_day_window_max: is required")
		}
		if t, err = time.ParseInLocation(time.TimeOnly, q.RawNextTwoDayWindowMax, time.Local); err != nil {
			return fmt.Errorf("next_two_day_window_max: should use format '%s'", time.TimeOnly)
		}
		nextTwoDayWindowMax := time.Date(
			2000,
			1,
			1,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if !nextTwoDayWindowMax.After(nextTwoDayWindowMin) {
			return errors.New("next_two_day_window_max: should exceed next_two_day_window_min")
		}
		q.NextTwoDayWindowMin = new(nextTwoDayWindowMin.AddDate(0, 0, 2))
		q.NextTwoDayWindowMax = new(nextTwoDayWindowMax.AddDate(0, 0, 2))
	} else {
		q.RawNextTwoDayWindowMin = ""
		q.NextTwoDayWindowMin = nil
		q.RawNextTwoDayWindowMax = ""
		q.NextTwoDayWindowMax = nil
	}
	return nil
}

func (q *DeliveryPlan) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *DeliveryPlan) UpdateDate(interval time.Duration) *DeliveryPlan {
	now := time.Now().In(time.Local)
	t := q.TransactionWindowMin.In(now.Location())
	q.TransactionWindowMin = time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	t = q.TransactionWindowMax.In(now.Location())
	q.TransactionWindowMax = time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	if interval == 0 {
		return q
	}
	if q.SameDay && q.SameDayWindowMin != nil && q.SameDayWindowMax != nil {
		sameDayWindowMin := *q.SameDayWindowMin
		sameDayWindowMax := *q.SameDayWindowMax
		for {
			q.DeliveryTimes = append(q.DeliveryTimes, &DeliveryPlanTime{
				DeliverySpeed: DeliverySpeedSameDay,
				Datetime: time.Date(
					now.Year(),
					now.Month(),
					now.Day(),
					sameDayWindowMin.Hour(),
					sameDayWindowMin.Minute(),
					sameDayWindowMin.Second(),
					sameDayWindowMin.Nanosecond(),
					sameDayWindowMin.Location(),
				),
			})
			if sameDayWindowMin.Equal(sameDayWindowMax) {
				break
			}
			sameDayWindowMin = sameDayWindowMin.Add(interval)
		}
	}
	if q.NextDay && q.NextDayWindowMin != nil && q.NextDayWindowMax != nil {
		nextDayWindowMin := *q.NextDayWindowMin
		nextDayWindowMax := *q.NextDayWindowMax
		for {
			q.DeliveryTimes = append(q.DeliveryTimes, &DeliveryPlanTime{
				DeliverySpeed: DeliverySpeedNextDay,
				Datetime: time.Date(
					now.Year(),
					now.Month(),
					now.Day(),
					nextDayWindowMin.Hour(),
					nextDayWindowMin.Minute(),
					nextDayWindowMin.Second(),
					nextDayWindowMin.Nanosecond(),
					nextDayWindowMin.Location(),
				),
			})
			if nextDayWindowMin.Equal(nextDayWindowMax) {
				break
			}
			nextDayWindowMin = nextDayWindowMin.Add(interval)
		}
	}
	if q.NextTwoDay && q.NextTwoDayWindowMin != nil && q.NextTwoDayWindowMax != nil {
		nextTwoDayWindowMin := *q.NextTwoDayWindowMin
		nextTwoDayWindowMax := *q.NextTwoDayWindowMax
		for {
			q.DeliveryTimes = append(q.DeliveryTimes, &DeliveryPlanTime{
				DeliverySpeed: DeliverySpeedNextTwoDay,
				Datetime: time.Date(
					now.Year(),
					now.Month(),
					now.Day(),
					nextTwoDayWindowMin.Hour(),
					nextTwoDayWindowMin.Minute(),
					nextTwoDayWindowMin.Second(),
					nextTwoDayWindowMin.Nanosecond(),
					nextTwoDayWindowMin.Location(),
				),
			})
			if nextTwoDayWindowMin.Equal(nextTwoDayWindowMax) {
				break
			}
			nextTwoDayWindowMin = nextTwoDayWindowMin.Add(interval)
		}
	}
	return q
}

func (q DeliveryWeekday) Int() int {
	return int(q)
}

func (q DeliveryWeekday) Equal(v DeliveryWeekday) bool {
	return q.Int() == v.Int()
}

func (q DeliveryWeekday) String() string {
	if q == DeliveryWeekdayWeekdays {
		return "Weekday"
	}
	if q == DeliveryWeekdayFriday {
		return "Jumat"
	}
	if q == DeliveryWeekdaySaturday {
		return "Sabtu"
	}
	return "Minggu"
}
