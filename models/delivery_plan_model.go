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
	DeliveryWeekdaySunday   DeliveryWeekday = DeliveryWeekday(time.Sunday)
	DeliveryWeekdayWeekdays DeliveryWeekday = DeliveryWeekday(time.Sunday + 1)
	DeliveryWeekdayFriday   DeliveryWeekday = DeliveryWeekday(time.Friday)
	DeliveryWeekdaySaturday DeliveryWeekday = DeliveryWeekday(time.Saturday)
)

type (
	DeliveryWeekday int

	DeliveryPlan struct {
		RowNo                   int64              `json:"-" form:"-"`
		ID                      uuid.UUID          `json:"id" form:"-"`
		RecordStatus            RecordStatus       `json:"-" form:"-"`
		Name                    string             `json:"name" form:"name"`
		Weekday                 DeliveryWeekday    `json:"weekday" form:"weekday"`
		TransactionWindowMin    time.Time          `json:"transaction_window_min" form:"-"`
		RawTransactionWindowMin string             `json:"-" form:"transaction_window_min"`
		TransactionWindowMax    time.Time          `json:"transaction_window_max" form:"-"`
		RawTransactionWindowMax string             `json:"-" form:"transaction_window_max"`
		DeliveryTimeInterval    time.Duration      `json:"-" form:"-"`
		RawDeliveryTimeInterval string             `json:"delivery_time_interval" form:"delivery_time_interval"`
		SameDay                 bool               `json:"same_day" form:"same_day"`
		SameDayWindowMin        *time.Time         `json:"-" form:"-"`
		RawSameDayWindowMin     string             `json:"-" form:"same_day_window_min"`
		SameDayWindowMax        *time.Time         `json:"-" form:"-"`
		RawSameDayWindowMax     string             `json:"-" form:"same_day_window_max"`
		NextDay                 bool               `json:"next_day" form:"next_day"`
		NextDayWindowMin        *time.Time         `json:"-" form:"-"`
		RawNextDayWindowMin     string             `json:"-" form:"next_day_window_min"`
		NextDayWindowMax        *time.Time         `json:"-" form:"-"`
		RawNextDayWindowMax     string             `json:"-" form:"next_day_window_max"`
		NextTwoDay              bool               `json:"next_two_day" form:"next_two_day"`
		NextTwoDayWindowMin     *time.Time         `json:"-" form:"-"`
		RawNextTwoDayWindowMin  string             `json:"-" form:"next_two_day_window_min"`
		NextTwoDayWindowMax     *time.Time         `json:"-" form:"-"`
		RawNextTwoDayWindowMax  string             `json:"-" form:"next_two_day_window_max"`
		DeliveryTimes           []time.Time        `json:"delivery_times" form:"-"`
		MapDeliveryTimes        map[int64]struct{} `json:"-" form:"-"`
		CreatedBy               uuid.UUID          `json:"-" form:"-"`
		CreatedAt               time.Time          `json:"-" form:"-"`
		UpdatedBy               uuid.UUID          `json:"-" form:"-"`
		UpdatedAt               time.Time          `json:"-" form:"-"`
		DeletedBy               *uuid.UUID         `json:"-" form:"-"`
		DeletedAt               *time.Time         `json:"-" form:"-"`
		DeletionReason          *string            `json:"-" form:"-"`
	}

	DeliveryPlanFilter struct {
		DeliveryPlanIDs,
		ExcludedDeliveryPlanIDs []uuid.UUID
		RecordStatuses      []RecordStatus
		Name                string
		DeliveryWeekdays    []DeliveryWeekday
		TransactionDatetime time.Time
		TransactionWindow   *DeliveryPlanTransactionWindow
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

func (q *DeliveryPlan) Validate() (err error) {
	now := time.Now().In(time.Local)
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
		now.Year(),
		now.Month(),
		now.Day(),
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
		now.Year(),
		now.Month(),
		now.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	if !q.TransactionWindowMax.After(q.TransactionWindowMin) {
		return errors.New("transaction_window_max: should exceed transaction_window_min")
	}
	if q.RawDeliveryTimeInterval = strings.TrimSpace(q.RawDeliveryTimeInterval); q.RawDeliveryTimeInterval == "" {
		return errors.New("delivery_time_interval: is required")
	}
	if q.DeliveryTimeInterval, err = time.ParseDuration(q.RawDeliveryTimeInterval); err != nil {
		return errors.New("delivery_time_interval: is invalid")
	}
	if q.SameDay {
		if q.RawSameDayWindowMin = strings.TrimSpace(q.RawSameDayWindowMin); q.RawSameDayWindowMin == "" {
			return errors.New("same_day_window_min: is required")
		}
		if t, err = time.ParseInLocation(time.TimeOnly, q.RawSameDayWindowMin, time.Local); err != nil {
			return fmt.Errorf("same_day_window_min: should use format '%s'", time.TimeOnly)
		}
		sameDayWindowMin := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
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
			now.Year(),
			now.Month(),
			now.Day(),
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
			now.Year(),
			now.Month(),
			now.Day()+1,
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
			now.Year(),
			now.Month(),
			now.Day()+1,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if !nextDayWindowMax.After(nextDayWindowMin) {
			return errors.New("next_day_window_max: should exceed next_day_window_min")
		}
		q.NextDayWindowMin = &nextDayWindowMin
		q.NextDayWindowMax = &nextDayWindowMax
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
			now.Year(),
			now.Month(),
			now.Day()+2,
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
			now.Year(),
			now.Month(),
			now.Day()+2,
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			t.Location(),
		)
		if !nextTwoDayWindowMax.After(nextTwoDayWindowMin) {
			return errors.New("next_two_day_window_max: should exceed next_two_day_window_min")
		}
		q.NextTwoDayWindowMin = &nextTwoDayWindowMin
		q.NextTwoDayWindowMax = &nextTwoDayWindowMax
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

func (q *DeliveryPlan) Populate() *DeliveryPlan {
	q.RawTransactionWindowMin = q.TransactionWindowMin.In(time.Local).Format(time.TimeOnly)
	q.RawTransactionWindowMax = q.TransactionWindowMax.In(time.Local).Format(time.TimeOnly)
	q.RawDeliveryTimeInterval = q.DeliveryTimeInterval.String()
	if q.SameDay && q.SameDayWindowMin != nil && q.SameDayWindowMax != nil {
		q.RawSameDayWindowMin = q.SameDayWindowMin.In(time.Local).Format(time.TimeOnly)
		q.RawSameDayWindowMax = q.SameDayWindowMax.In(time.Local).Format(time.TimeOnly)
		sameDayWindowMin := *q.SameDayWindowMin
		sameDayWindowMax := *q.SameDayWindowMax
		for !sameDayWindowMin.After(sameDayWindowMax) {
			q.DeliveryTimes = append(q.DeliveryTimes, sameDayWindowMin)
			sameDayWindowMin = sameDayWindowMin.Add(q.DeliveryTimeInterval)
		}
	}
	if q.NextDay && q.NextDayWindowMin != nil && q.NextDayWindowMax != nil {
		q.RawNextDayWindowMin = q.NextDayWindowMin.In(time.Local).Format(time.TimeOnly)
		q.RawNextDayWindowMax = q.NextDayWindowMax.In(time.Local).Format(time.TimeOnly)
		nextDayWindowMin := *q.NextDayWindowMin
		nextDayWindowMax := *q.NextDayWindowMax
		for !nextDayWindowMin.After(nextDayWindowMax) {
			q.DeliveryTimes = append(q.DeliveryTimes, nextDayWindowMin)
			nextDayWindowMin = nextDayWindowMin.Add(q.DeliveryTimeInterval)
		}
	}
	if q.NextTwoDay && q.NextTwoDayWindowMin != nil && q.NextTwoDayWindowMax != nil {
		q.RawNextTwoDayWindowMin = q.NextTwoDayWindowMin.In(time.Local).Format(time.TimeOnly)
		q.RawNextTwoDayWindowMax = q.NextTwoDayWindowMax.In(time.Local).Format(time.TimeOnly)
		nextTwoDayWindowMin := *q.NextTwoDayWindowMin
		nextTwoDayWindowMax := *q.NextTwoDayWindowMax
		for !nextTwoDayWindowMin.After(nextTwoDayWindowMax) {
			q.DeliveryTimes = append(q.DeliveryTimes, nextTwoDayWindowMin)
			nextTwoDayWindowMin = nextTwoDayWindowMin.Add(q.DeliveryTimeInterval)
		}
	}
	q.MapDeliveryTimes = make(map[int64]struct{}, len(q.DeliveryTimes))
	for i := range q.DeliveryTimes {
		q.MapDeliveryTimes[q.DeliveryTimes[i].UnixNano()] = struct{}{}
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
