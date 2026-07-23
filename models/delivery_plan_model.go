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
		RowNo                   int64           `json:"-" form:"-"`
		ID                      uuid.UUID       `json:"id" form:"-"`
		RecordStatus            RecordStatus    `json:"-" form:"-"`
		Name                    string          `json:"name" form:"name"`
		Weekday                 DeliveryWeekday `json:"weekday" form:"weekday"`
		TransactionWindowMin    time.Time       `json:"transaction_window_min" form:"-"`
		RawTransactionWindowMin string          `json:"-" form:"transaction_window_min"`
		TransactionWindowMax    time.Time       `json:"transaction_window_max" form:"-"`
		RawTransactionWindowMax string          `json:"-" form:"transaction_window_max"`
		SameDay                 bool            `json:"same_day" form:"same_day"`
		SameDayWindowMin        *time.Time      `json:"-" form:"-"`
		RawSameDayWindowMin     string          `json:"-" form:"same_day_window_min"`
		SameDayWindowMax        *time.Time      `json:"-" form:"-"`
		RawSameDayWindowMax     string          `json:"-" form:"same_day_window_max"`
		SameDayDeliveryTimes    []*DeliveryTime `json:"same_day_delivery_times" form:"-"`
		NextDay                 bool            `json:"next_day" form:"next_day"`
		NextDayWindowMin        *time.Time      `json:"-" form:"-"`
		RawNextDayWindowMin     string          `json:"-" form:"next_day_window_min"`
		NextDayWindowMax        *time.Time      `json:"-" form:"-"`
		RawNextDayWindowMax     string          `json:"-" form:"next_day_window_max"`
		NextDayDeliveryTimes    []*DeliveryTime `json:"next_day_delivery_times" form:"-"`
		NextTwoDay              bool            `json:"next_two_day" form:"next_two_day"`
		NextTwoDayWindowMin     *time.Time      `json:"-" form:"-"`
		RawNextTwoDayWindowMin  string          `json:"-" form:"next_two_day_window_min"`
		NextTwoDayWindowMax     *time.Time      `json:"-" form:"-"`
		RawNextTwoDayWindowMax  string          `json:"-" form:"next_two_day_window_max"`
		NextTwoDayDeliveryTimes []*DeliveryTime `json:"next_two_day_delivery_times" form:"-"`
		CreatedBy               uuid.UUID       `json:"-" form:"-"`
		CreatedAt               time.Time       `json:"-" form:"-"`
		UpdatedBy               uuid.UUID       `json:"-" form:"-"`
		UpdatedAt               time.Time       `json:"-" form:"-"`
		DeletedBy               *uuid.UUID      `json:"-" form:"-"`
		DeletedAt               *time.Time      `json:"-" form:"-"`
		DeletionReason          *string         `json:"-" form:"-"`
	}

	DeliveryTime struct {
		ID             uuid.UUID     `json:"id" form:"-"`
		RecordStatus   RecordStatus  `json:"-" form:"-"`
		DeliveryPlanID uuid.UUID     `json:"-" form:"-"`
		DeliverySpeed  DeliverySpeed `json:"-" form:"-"`
		Timestamp      time.Time     `json:"timestamp" form:"-"`
		CreatedBy      uuid.UUID     `json:"-" form:"-"`
		CreatedAt      time.Time     `json:"-" form:"-"`
		DeletedBy      *uuid.UUID    `json:"-" form:"-"`
		DeletedAt      *time.Time    `json:"-" form:"-"`
		DeletionReason *string       `json:"-" form:"-"`
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
		q.SameDayWindowMin = new(sameDayWindowMin)
		q.SameDayWindowMax = new(sameDayWindowMax)
		for {
			q.SameDayDeliveryTimes = append(q.SameDayDeliveryTimes, &DeliveryTime{
				DeliverySpeed: DeliverySpeedSameDay,
				Timestamp:     sameDayWindowMin,
			})
			if sameDayWindowMin.Equal(sameDayWindowMax) {
				break
			}
			sameDayWindowMin = sameDayWindowMin.Add(interval)
		}
	} else {
		q.RawSameDayWindowMin = ""
		q.SameDayWindowMin = nil
		q.RawSameDayWindowMax = ""
		q.SameDayWindowMax = nil
		q.SameDayDeliveryTimes = nil
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
		nextDayWindowMin = nextDayWindowMin.AddDate(0, 0, 1)
		nextDayWindowMax = nextDayWindowMax.AddDate(0, 0, 1)
		q.NextDayWindowMin = new(nextDayWindowMin)
		q.NextDayWindowMax = new(nextDayWindowMax)
		for {
			q.NextDayDeliveryTimes = append(q.NextDayDeliveryTimes, &DeliveryTime{
				DeliverySpeed: DeliverySpeedNextDay,
				Timestamp:     nextDayWindowMin,
			})
			if nextDayWindowMin.Equal(nextDayWindowMax) {
				break
			}
			nextDayWindowMin = nextDayWindowMin.Add(interval)
		}
	} else {
		q.RawNextDayWindowMin = ""
		q.NextDayWindowMin = nil
		q.RawNextDayWindowMax = ""
		q.NextDayWindowMax = nil
		q.NextDayDeliveryTimes = nil
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
		nextTwoDayWindowMin = nextTwoDayWindowMin.AddDate(0, 0, 2)
		nextTwoDayWindowMax = nextTwoDayWindowMax.AddDate(0, 0, 2)
		q.NextTwoDayWindowMin = new(nextTwoDayWindowMin)
		q.NextTwoDayWindowMax = new(nextTwoDayWindowMax)
		for {
			q.NextTwoDayDeliveryTimes = append(q.NextTwoDayDeliveryTimes, &DeliveryTime{
				DeliverySpeed: DeliverySpeedNextTwoDay,
				Timestamp:     nextTwoDayWindowMin,
			})
			if nextTwoDayWindowMin.Equal(nextTwoDayWindowMax) {
				break
			}
			nextTwoDayWindowMin = nextTwoDayWindowMin.Add(interval)
		}
	} else {
		q.RawNextTwoDayWindowMin = ""
		q.NextTwoDayWindowMin = nil
		q.RawNextTwoDayWindowMax = ""
		q.NextTwoDayWindowMax = nil
		q.NextTwoDayDeliveryTimes = nil
	}
	return nil
}

func (q *DeliveryPlan) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *DeliveryPlan) UpdateDate(timestamp time.Time) {
	t := q.TransactionWindowMin.In(timestamp.Location())
	q.TransactionWindowMin = time.Date(
		timestamp.Year(),
		timestamp.Month(),
		timestamp.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	t = q.TransactionWindowMax.In(timestamp.Location())
	q.TransactionWindowMax = time.Date(
		timestamp.Year(),
		timestamp.Month(),
		timestamp.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	if q.SameDay {
		for i := range q.SameDayDeliveryTimes {
			t = q.SameDayDeliveryTimes[i].Timestamp.In(timestamp.Location())
			q.SameDayDeliveryTimes[i].Timestamp = time.Date(
				timestamp.Year(),
				timestamp.Month(),
				timestamp.Day(),
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				t.Location(),
			)
		}
	}
	if q.NextDay {
		for i := range q.NextDayDeliveryTimes {
			t = q.NextDayDeliveryTimes[i].Timestamp.In(timestamp.Location())
			q.NextDayDeliveryTimes[i].Timestamp = time.Date(
				timestamp.Year(),
				timestamp.Month(),
				timestamp.Day()+1,
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				t.Location(),
			)
		}
	}
	if q.NextTwoDay {
		for i := range q.NextTwoDayDeliveryTimes {
			t = q.NextTwoDayDeliveryTimes[i].Timestamp.In(timestamp.Location())
			q.NextTwoDayDeliveryTimes[i].Timestamp = time.Date(
				timestamp.Year(),
				timestamp.Month(),
				timestamp.Day()+2,
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				t.Location(),
			)
		}
	}
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
