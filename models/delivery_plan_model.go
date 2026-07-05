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
	ErrDeliveryPlanNotFound             = errors.New("shipping plan not found")
	ErrUniqueDeliveryPlanNameViolation  = errors.New("name: already exists")
	ErrUniqueDeliveryTimeRangeViolation = errors.New("delivery plan time range already exists")

	DeliveryDayTimes    = []DeliveryDayTime{DeliveryDayTimeMorning, DeliveryDayTimeMidday, DeliveryDayTimeAfternoon}
	MapDeliveryDayTimes = map[DeliveryDayTime]struct{}{
		DeliveryDayTimeMorning:   {},
		DeliveryDayTimeMidday:    {},
		DeliveryDayTimeAfternoon: {},
	}
	MapDeliveryPlanSortings = map[string]struct{}{
		SortingFieldID:   {},
		SortingFieldName: {},
	}
)

const (
	DeliveryDayTimeMorning DeliveryDayTime = iota
	DeliveryDayTimeMidday
	DeliveryDayTimeAfternoon
)

type (
	DeliveryDayTime int

	DeliveryPlan struct {
		RowNo                      int64           `json:"-" form:"-"`
		ID                         uuid.UUID       `json:"-" form:"-"`
		RecordStatus               RecordStatus    `json:"-" form:"-"`
		Name                       string          `json:"-" form:"name"`
		TransactionTimeMin         time.Time       `json:"-" form:"-"`
		RawTransactionTimeMin      string          `json:"-" form:"transaction_time_min"`
		TransactionTimeMax         time.Time       `json:"-" form:"-"`
		RawTransactionTimeMax      string          `json:"-" form:"transaction_time_max"`
		TransactionTimeDescription string          `json:"-" form:"-"`
		DeliveryTimeMin            time.Time       `json:"-" form:"-"`
		RawDeliveryTimeMin         string          `json:"-" form:"delivery_time_min"`
		DeliveryTimeMax            time.Time       `json:"-" form:"-"`
		RawDeliveryTimeMax         string          `json:"-" form:"delivery_time_max"`
		DeliveryTimeDescription    string          `json:"-" form:"-"`
		DeliveryDayTime            DeliveryDayTime `json:"-" form:"delivery_daytime"`
		SameDay                    bool            `json:"-" form:"same_day"`
		NextDay                    bool            `json:"-" form:"next_day"`
		NextTwoDay                 bool            `json:"-" form:"next_two_day"`
		CreatedBy                  uuid.UUID       `json:"-" form:"-"`
		CreatedAt                  time.Time       `json:"-" form:"-"`
		UpdatedBy                  uuid.UUID       `json:"-" form:"-"`
		UpdatedAt                  time.Time       `json:"-" form:"-"`
		DeletedBy                  *uuid.UUID      `json:"-" form:"-"`
		DeletedAt                  *time.Time      `json:"-" form:"-"`
		DeletionReason             *string         `json:"-" form:"-"`
	}

	DeliveryPlanFilter struct {
		DeliveryPlanIDs,
		ExcludedDeliveryPlanIDs []uuid.UUID
		RecordStatuses []RecordStatus
		Name           string
		TransactionTimeMin,
		TransactionTimeMax *time.Time
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings *Map
		Limit,
		Page int64
		UrlValues url.Values
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

func (q *DeliveryPlanFilter) WithTransactionTimeMin(transactionTimeMin *time.Time) *DeliveryPlanFilter {
	q.TransactionTimeMin = transactionTimeMin
	return q
}

func (q *DeliveryPlanFilter) WithTransactionTimeMax(transactionTimeMax *time.Time) *DeliveryPlanFilter {
	q.TransactionTimeMax = transactionTimeMax
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
	if q.RawTransactionTimeMin = strings.TrimSpace(q.RawTransactionTimeMin); q.RawTransactionTimeMin == "" {
		return errors.New("transaction_time_min: is required")
	}
	if q.TransactionTimeMin, err = time.ParseInLocation(time.TimeOnly, q.RawTransactionTimeMin, location); err != nil {
		return fmt.Errorf("transaction_time_min: should use format '%s'", time.TimeOnly)
	}
	if q.RawTransactionTimeMax = strings.TrimSpace(q.RawTransactionTimeMax); q.RawTransactionTimeMax == "" {
		return errors.New("transaction_time_max: is required")
	}
	if q.TransactionTimeMax, err = time.ParseInLocation(time.TimeOnly, q.RawTransactionTimeMax, location); err != nil {
		return fmt.Errorf("transaction_time_max: should use format '%s'", time.TimeOnly)
	}
	if q.RawDeliveryTimeMin = strings.TrimSpace(q.RawDeliveryTimeMin); q.RawDeliveryTimeMin == "" {
		return errors.New("delivery_time_min: is required")
	}
	if q.DeliveryTimeMin, err = time.ParseInLocation(time.TimeOnly, q.RawDeliveryTimeMin, location); err != nil {
		return fmt.Errorf("delivery_time_min: should use format '%s'", time.TimeOnly)
	}
	if q.RawDeliveryTimeMax = strings.TrimSpace(q.RawDeliveryTimeMax); q.RawDeliveryTimeMax == "" {
		return errors.New("delivery_time_max: is required")
	}
	if q.DeliveryTimeMax, err = time.ParseInLocation(time.TimeOnly, q.RawDeliveryTimeMax, location); err != nil {
		return fmt.Errorf("delivery_time_max: should use format '%s'", time.TimeOnly)
	}
	if _, ok := MapDeliveryDayTimes[q.DeliveryDayTime]; !ok {
		return errors.New("delivery_daytime: not found")
	}
	return nil
}

func (q *DeliveryPlan) Populate() *DeliveryPlan {
	location := helper.LoadTimeZone()
	q.TransactionTimeMin = q.TransactionTimeMin.In(location)
	q.TransactionTimeMax = q.TransactionTimeMax.In(location)
	q.DeliveryTimeMin = q.DeliveryTimeMin.In(location)
	q.DeliveryTimeMax = q.DeliveryTimeMax.In(location)
	q.RawTransactionTimeMin = q.TransactionTimeMin.Format(time.TimeOnly)
	q.RawTransactionTimeMax = q.TransactionTimeMax.Format(time.TimeOnly)
	q.RawDeliveryTimeMin = q.DeliveryTimeMin.Format(time.TimeOnly)
	q.RawDeliveryTimeMax = q.DeliveryTimeMax.Format(time.TimeOnly)
	return q
}

func (q DeliveryDayTime) Int() int {
	return int(q)
}

func (q DeliveryDayTime) String() string {
	if q == DeliveryDayTimeAfternoon {
		return "sore"
	}
	if q == DeliveryDayTimeMidday {
		return "siang"
	}
	return "pagi"
}

func (q DeliveryDayTime) Equal(v DeliveryDayTime) bool {
	return q.Int() == v.Int()
}
