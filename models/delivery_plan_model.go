package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var ErrDeliveryPlanNotFound = errors.New("delivery plan not found")
var ErrUniqueDeliveryPlanTransactionWindowViolation = errors.New("transaction time frame already exists")

var DeliveryWeekdays = []DeliveryWeekday{
	DeliveryWeekdaySunday,
	DeliveryWeekdayWeekdays,
	DeliveryWeekdayFriday,
	DeliveryWeekdaySaturday,
}
var MapDeliveryWeekdays = map[DeliveryWeekday]struct{}{
	DeliveryWeekdaySunday:   {},
	DeliveryWeekdayWeekdays: {},
	DeliveryWeekdayFriday:   {},
	DeliveryWeekdaySaturday: {},
}
var MapDeliveryPlanSortings = map[string]struct{}{
	SortingFieldID:   {},
	SortingFieldName: {},
}

const DeliveryWeekdaySunday DeliveryWeekday = DeliveryWeekday(time.Sunday)
const DeliveryWeekdayWeekdays DeliveryWeekday = DeliveryWeekday(time.Sunday + 1)
const DeliveryWeekdayFriday DeliveryWeekday = DeliveryWeekday(time.Friday)
const DeliveryWeekdaySaturday DeliveryWeekday = DeliveryWeekday(time.Saturday)

type DeliveryWeekday int

type DeliveryPlan struct {
	RowNo                int64              `json:"-"`
	ID                   uuid.UUID          `json:"id"`
	RecordStatus         RecordStatus       `json:"record_status"`
	Name                 string             `json:"name"`
	Weekday              DeliveryWeekday    `json:"weekday"`
	TransactionWindowMin time.Time          `json:"transaction_window_min"`
	TransactionWindowMax time.Time          `json:"transaction_window_max"`
	DeliveryTimeInterval Duration           `json:"delivery_time_interval"`
	SameDay              bool               `json:"same_day"`
	SameDayWindowMin     *time.Time         `json:"same_day_window_min"`
	SameDayWindowMax     *time.Time         `json:"same_day_window_max"`
	NextDay              bool               `json:"next_day"`
	NextDayWindowMin     *time.Time         `json:"next_day_window_min"`
	NextDayWindowMax     *time.Time         `json:"next_day_window_max"`
	NextTwoDay           bool               `json:"next_two_day"`
	NextTwoDayWindowMin  *time.Time         `json:"next_two_day_window_min"`
	NextTwoDayWindowMax  *time.Time         `json:"next_two_day_window_max"`
	DeliveryTimes        []time.Time        `json:"delivery_times"`
	MapDeliveryTimes     map[int64]struct{} `json:"-"`
	CreatedBy            uuid.UUID          `json:"created_by"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedBy            uuid.UUID          `json:"updated_by"`
	UpdatedAt            time.Time          `json:"updated_at"`
	DeletedBy            *uuid.UUID         `json:"deleted_by"`
	DeletedAt            *time.Time         `json:"deleted_at"`
	DeletionReason       *string            `json:"deletion_reason"`
}

type DeliveryPlanFilter struct {
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

type DeliveryPlanTransactionWindow struct {
	Min,
	Max time.Time
}

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
	if !q.TransactionWindowMax.After(q.TransactionWindowMin) {
		return errors.New("transaction_window_max: should exceed transaction_window_min")
	}
	if q.SameDay {
		if q.SameDayWindowMin == nil {
			return errors.New("same_day_window_min: is required")
		}
		sameDayWindowMin := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.SameDayWindowMin.Hour(),
			q.SameDayWindowMin.Minute(),
			q.SameDayWindowMin.Second(),
			q.SameDayWindowMin.Nanosecond(),
			q.SameDayWindowMin.Location(),
		)
		if q.SameDayWindowMax == nil {
			return errors.New("same_day_window_max: is required")
		}
		sameDayWindowMax := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			q.SameDayWindowMax.Hour(),
			q.SameDayWindowMax.Minute(),
			q.SameDayWindowMax.Second(),
			q.SameDayWindowMax.Nanosecond(),
			q.SameDayWindowMax.Location(),
		)
		if !sameDayWindowMax.After(sameDayWindowMin) {
			return errors.New("same_day_window_max: should exceed same_day_window_min")
		}
		q.SameDayWindowMin = &sameDayWindowMin
		q.SameDayWindowMax = &sameDayWindowMax
	} else {
		q.SameDayWindowMin = nil
		q.SameDayWindowMax = nil
	}
	if q.NextDay {
		if q.NextDayWindowMin == nil {
			return errors.New("next_day_window_min: is required")
		}
		nextDayWindowMin := time.Date(
			now.Year(),
			now.Month(),
			now.Day()+1,
			q.NextDayWindowMin.Hour(),
			q.NextDayWindowMin.Minute(),
			q.NextDayWindowMin.Second(),
			q.NextDayWindowMin.Nanosecond(),
			q.NextDayWindowMin.Location(),
		)
		if q.NextDayWindowMax == nil {
			return errors.New("next_day_window_max: is required")
		}
		nextDayWindowMax := time.Date(
			now.Year(),
			now.Month(),
			now.Day()+1,
			q.NextDayWindowMax.Hour(),
			q.NextDayWindowMax.Minute(),
			q.NextDayWindowMax.Second(),
			q.NextDayWindowMax.Nanosecond(),
			q.NextDayWindowMax.Location(),
		)
		if !nextDayWindowMax.After(nextDayWindowMin) {
			return errors.New("next_day_window_max: should exceed next_day_window_min")
		}
		q.NextDayWindowMin = &nextDayWindowMin
		q.NextDayWindowMax = &nextDayWindowMax
	} else {
		q.NextDayWindowMin = nil
		q.NextDayWindowMax = nil
	}
	if q.NextTwoDay {
		if q.NextTwoDayWindowMin == nil {
			return errors.New("next_two_day_window_min: is required")
		}
		nextTwoDayWindowMin := time.Date(
			now.Year(),
			now.Month(),
			now.Day()+2,
			q.NextTwoDayWindowMin.Hour(),
			q.NextTwoDayWindowMin.Minute(),
			q.NextTwoDayWindowMin.Second(),
			q.NextTwoDayWindowMin.Nanosecond(),
			q.NextTwoDayWindowMin.Location(),
		)
		if q.NextTwoDayWindowMax == nil {
			return errors.New("next_two_day_window_max: is required")
		}
		nextTwoDayWindowMax := time.Date(
			now.Year(),
			now.Month(),
			now.Day()+2,
			q.NextTwoDayWindowMax.Hour(),
			q.NextTwoDayWindowMax.Minute(),
			q.NextTwoDayWindowMax.Second(),
			q.NextTwoDayWindowMax.Nanosecond(),
			q.NextTwoDayWindowMax.Location(),
		)
		if !nextTwoDayWindowMax.After(nextTwoDayWindowMin) {
			return errors.New("next_two_day_window_max: should exceed next_two_day_window_min")
		}
		q.NextTwoDayWindowMin = &nextTwoDayWindowMin
		q.NextTwoDayWindowMax = &nextTwoDayWindowMax
	} else {
		q.NextTwoDayWindowMin = nil
		q.NextTwoDayWindowMax = nil
	}
	return nil
}

func (q *DeliveryPlan) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *DeliveryPlan) Populate() *DeliveryPlan {
	deliveryTimeInterval := time.Duration(q.DeliveryTimeInterval)
	if q.SameDay && q.SameDayWindowMin != nil && q.SameDayWindowMax != nil {
		sameDayWindowMin := *q.SameDayWindowMin
		sameDayWindowMax := *q.SameDayWindowMax
		for !sameDayWindowMin.After(sameDayWindowMax) {
			q.DeliveryTimes = append(q.DeliveryTimes, sameDayWindowMin)
			sameDayWindowMin = sameDayWindowMin.Add(deliveryTimeInterval)
		}
	}
	if q.NextDay && q.NextDayWindowMin != nil && q.NextDayWindowMax != nil {
		nextDayWindowMin := *q.NextDayWindowMin
		nextDayWindowMax := *q.NextDayWindowMax
		for !nextDayWindowMin.After(nextDayWindowMax) {
			q.DeliveryTimes = append(q.DeliveryTimes, nextDayWindowMin)
			nextDayWindowMin = nextDayWindowMin.Add(deliveryTimeInterval)
		}
	}
	if q.NextTwoDay && q.NextTwoDayWindowMin != nil && q.NextTwoDayWindowMax != nil {
		nextTwoDayWindowMin := *q.NextTwoDayWindowMin
		nextTwoDayWindowMax := *q.NextTwoDayWindowMax
		for !nextTwoDayWindowMin.After(nextTwoDayWindowMax) {
			q.DeliveryTimes = append(q.DeliveryTimes, nextTwoDayWindowMin)
			nextTwoDayWindowMin = nextTwoDayWindowMin.Add(deliveryTimeInterval)
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
