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
	ErrVoucherNotFound            = errors.New("voucher not found")
	ErrUniqueVoucherCodeViolation = errors.New("code: already exists")

	VoucherTypes    = []VoucherType{VoucherTypeDeliveryCost, VoucherTypeShoppingCost}
	MapVoucherTypes = map[VoucherType]struct{}{
		VoucherTypeDeliveryCost: {},
		VoucherTypeShoppingCost: {},
	}
	MapVoucherSortings = map[string]struct{}{
		SortingFieldID:          {},
		VoucherSortingFieldName: {},
	}
)

const (
	VoucherTypeDeliveryCost VoucherType = iota
	VoucherTypeShoppingCost
)

const (
	VoucherSortingFieldName = "name"

	VoucherPhotoName         = "voucher-%d%s"
	VoucherPhotoPath         = "./images/voucher/%s"
	VoucherPhotoCompressPath = "./images/voucher/compress/%s"
)

type (
	VoucherType int

	Voucher struct {
		RowNo                 int64        `json:"-" form:"-"`
		ID                    uuid.UUID    `json:"-" form:"-"`
		RecordStatus          RecordStatus `json:"-" form:"-"`
		VoucherType           VoucherType  `json:"-" form:"voucher_type"`
		Code                  string       `json:"-" form:"code"`
		Photo                 string       `json:"-" form:"photo"`
		Name                  string       `json:"-" form:"name"`
		Budget                int64        `json:"-" form:"budget"`
		Discount              int64        `json:"-" form:"discount"`
		MinimumPurchase       int64        `json:"-" form:"minimum_purchase"`
		MaximumUsage          int          `json:"-" form:"maximum_usage"`
		EffectiveStartTime    time.Time    `json:"-" form:"-"`
		RawEffectiveStartTime string       `json:"-" form:"effective_start_time"`
		EffectiveEndTime      time.Time    `json:"-" form:"-"`
		RawEffectiveEndTime   string       `json:"-" form:"effective_end_time"`
		EffectiveStartDate    time.Time    `json:"-" form:"-"`
		RawEffectiveStartDate string       `json:"-" form:"effective_start_date"`
		EffectiveEndDate      time.Time    `json:"-" form:"-"`
		RawEffectiveEndDate   string       `json:"-" form:"effective_end_date"`
		Description           string       `json:"-" form:"description"`
		CreatedBy             uuid.UUID    `json:"-" form:"-"`
		CreatedAt             time.Time    `json:"-" form:"-"`
		UpdatedBy             uuid.UUID    `json:"-" form:"-"`
		UpdatedAt             time.Time    `json:"-" form:"-"`
		DeletedBy             *uuid.UUID   `json:"-" form:"-"`
		DeletedAt             *time.Time   `json:"-" form:"-"`
		DeletionReason        *string      `json:"-" form:"-"`
	}

	VoucherFilter struct {
		VoucherIDs,
		ExcludedVoucherIDs []uuid.UUID
		RecordStatuses []RecordStatus
		VoucherTypes   []VoucherType
		Codes          []string
		EffectiveStartTime,
		EffectiveEndTime,
		EffectiveStartDate,
		EffectiveEndDate time.Time
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings *Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewVoucherFilter() *VoucherFilter {
	return new(VoucherFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *VoucherFilter) WithVoucherIDs(voucherIDs ...uuid.UUID) *VoucherFilter {
	q.VoucherIDs = voucherIDs
	return q
}

func (q *VoucherFilter) WithExcludedVoucherIDs(excludedVoucherIDs ...uuid.UUID) *VoucherFilter {
	q.ExcludedVoucherIDs = excludedVoucherIDs
	return q
}

func (q *VoucherFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *VoucherFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *VoucherFilter) WithVoucherTypes(voucherTypes ...VoucherType) *VoucherFilter {
	q.VoucherTypes = voucherTypes
	return q
}

func (q *VoucherFilter) WithCodes(codes ...string) *VoucherFilter {
	q.Codes = codes
	return q
}

func (q *VoucherFilter) WithEffectiveTime(effectiveStartTime, effectiveEndTime time.Time) *VoucherFilter {
	q.EffectiveStartTime = effectiveStartTime
	q.EffectiveEndTime = effectiveEndTime
	return q
}

func (q *VoucherFilter) WithEffectiveDate(effectiveStartDate, effectiveEndDate time.Time) *VoucherFilter {
	q.EffectiveStartDate = effectiveStartDate
	q.EffectiveEndDate = effectiveEndDate
	return q
}

func (q *VoucherFilter) WithKeyword(keyword string) *VoucherFilter {
	q.Keyword = keyword
	return q
}

func (q *VoucherFilter) WithPaginationURL(paginationURL string) *VoucherFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *VoucherFilter) WithSortings(sortings ...Sorting) *VoucherFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *VoucherFilter) WithLimit(limit int64) *VoucherFilter {
	q.Limit = limit
	return q
}

func (q *VoucherFilter) WithPage(page int64) *VoucherFilter {
	q.Page = page
	return q
}

func (q *VoucherFilter) WithUrlValues(urlValues url.Values) *VoucherFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Voucher) Validate() (err error) {
	location := helper.LoadTimeZone()
	if _, ok := MapVoucherTypes[q.VoucherType]; !ok {
		return errors.New("voucher_type: not found")
	}
	if q.Code = strings.TrimSpace(q.Code); q.Code == "" {
		return errors.New("code: is required")
	}
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.Budget < 1 {
		return errors.New("budget: is required")
	}
	if q.Discount < 1 {
		return errors.New("discount: is required")
	}
	if q.MinimumPurchase < 1 {
		return errors.New("minimum_purchase: is required")
	}
	if q.MaximumUsage < 1 {
		return errors.New("maximum_usage: is required")
	}
	if q.RawEffectiveStartTime = strings.TrimSpace(q.RawEffectiveStartTime); q.RawEffectiveStartTime == "" {
		return errors.New("effective_start_time: is required")
	}
	if q.EffectiveStartTime, err = time.ParseInLocation(time.TimeOnly, q.RawEffectiveStartTime, location); err != nil {
		return fmt.Errorf("effective_start_time: should use format '%s'", time.TimeOnly)
	}
	if q.RawEffectiveEndTime = strings.TrimSpace(q.RawEffectiveEndTime); q.RawEffectiveEndTime == "" {
		return errors.New("effective_end_time: is required")
	}
	if q.EffectiveEndTime, err = time.ParseInLocation(time.TimeOnly, q.RawEffectiveEndTime, location); err != nil {
		return fmt.Errorf("effective_end_time: should use format '%s'", time.TimeOnly)
	}
	if !q.EffectiveEndTime.After(q.EffectiveStartTime) {
		return errors.New("effective_end_time: should exceed effective_start_time")
	}
	if q.RawEffectiveStartDate = strings.TrimSpace(q.RawEffectiveStartDate); q.RawEffectiveStartDate == "" {
		return errors.New("effective_start_date: is required")
	}
	effectiveStartDate, err := time.ParseInLocation(time.DateOnly, q.RawEffectiveStartDate, location)
	if err != nil {
		return fmt.Errorf("effective_start_date: should use format '%s'", time.DateOnly)
	}
	q.EffectiveStartDate = time.Date(
		effectiveStartDate.Year(),
		effectiveStartDate.Month(),
		effectiveStartDate.Day(),
		0,
		0,
		0,
		0,
		effectiveStartDate.Location(),
	)
	if q.RawEffectiveEndDate = strings.TrimSpace(q.RawEffectiveEndDate); q.RawEffectiveEndDate == "" {
		return errors.New("effective_end_date: is required")
	}
	effectiveEndDate, err := time.ParseInLocation(time.DateOnly, q.RawEffectiveEndDate, location)
	if err != nil {
		return fmt.Errorf("effective_end_date: should use format '%s'", time.DateOnly)
	}
	q.EffectiveEndDate = time.Date(
		effectiveEndDate.Year(),
		effectiveEndDate.Month(),
		effectiveEndDate.Day(),
		23,
		59,
		59,
		999_999,
		effectiveEndDate.Location(),
	)
	if !q.EffectiveEndDate.After(q.EffectiveStartDate) {
		return errors.New("effective_end_date: should exceed effective_start_date")
	}
	q.Description = strings.TrimSpace(q.Description)
	q.Populate()
	return nil
}

func (q *Voucher) Populate() *Voucher {
	location := helper.LoadTimeZone()
	q.EffectiveStartTime = q.EffectiveStartTime.In(location)
	q.RawEffectiveStartTime = q.EffectiveStartTime.Format(time.TimeOnly)
	q.EffectiveEndTime = q.EffectiveEndTime.In(location)
	q.RawEffectiveEndTime = q.EffectiveEndTime.Format(time.TimeOnly)
	q.EffectiveStartDate = q.EffectiveStartDate.In(location)
	q.RawEffectiveStartDate = q.EffectiveStartDate.Format(time.DateOnly)
	q.EffectiveEndDate = q.EffectiveEndDate.In(location)
	q.RawEffectiveEndDate = q.EffectiveEndDate.Format(time.DateOnly)
	return q
}

func (q *Voucher) Status() string {
	now := time.Now()
	if q.EffectiveEndDate.Before(now) {
		return "expired"
	}
	if q.EffectiveStartDate.After(now) {
		return "pending"
	}
	return "active"
}

func (q VoucherType) Int() int {
	return int(q)
}

func (q VoucherType) String() string {
	if q == VoucherTypeShoppingCost {
		return "shopping cost"
	}
	return "delivery cost"
}

func (q VoucherType) Equal(v VoucherType) bool {
	return q.Int() == v.Int()
}
