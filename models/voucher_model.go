package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/coregx/coregex"
	"github.com/roysitumorang/laukpauk/helper"
)

var ErrVoucherNotFound = errors.New("voucher not found")
var ErrUniqueVoucherCodeViolation = errors.New("code: already exists")

var VoucherCodeRegex = coregex.MustCompile("^[a-zA-Z0-9-]+$")

type VoucherType uint8

const (
	VoucherTypeDeliveryCost VoucherType = iota
	VoucherTypeShoppingCost
)

func (q VoucherType) String() string {
	if q == VoucherTypeShoppingCost {
		return "shopping cost"
	}
	return "delivery cost"
}

func (q VoucherType) Int() uint8 {
	return uint8(q)
}

func (q VoucherType) Equal(v VoucherType) bool {
	return q.Int() == v.Int()
}

var VoucherTypes = []VoucherType{
	VoucherTypeDeliveryCost,
	VoucherTypeShoppingCost,
}
var MapVoucherTypes = map[VoucherType]struct{}{
	VoucherTypeDeliveryCost: {},
	VoucherTypeShoppingCost: {},
}

type VoucherTargetAudience uint8

const (
	VoucherTargetAudienceAllBuyers VoucherTargetAudience = iota
	VoucherTargetAudienceNewBuyers
	VoucherTargetAudienceLoyalty
	VoucherTargetAudienceEarlyShopping
	VoucherTargetAudienceInactiveBuyers
)

func (q VoucherTargetAudience) Equal(v VoucherTargetAudience) bool {
	return q.Int() == v.Int()
}

func (q VoucherTargetAudience) Int() uint8 {
	return uint8(q)
}

func (q VoucherTargetAudience) String() string {
	if q == VoucherTargetAudienceNewBuyers {
		return "new buyers"
	}
	if q == VoucherTargetAudienceLoyalty {
		return "loyalty"
	}
	if q == VoucherTargetAudienceEarlyShopping {
		return "belanja lebih awal"
	}
	if q == VoucherTargetAudienceInactiveBuyers {
		return "inactive buyers"
	}
	return "all buyers"
}

var VoucherTargetAudiences = []VoucherTargetAudience{
	VoucherTargetAudienceAllBuyers,
	VoucherTargetAudienceNewBuyers,
	VoucherTargetAudienceLoyalty,
	VoucherTargetAudienceEarlyShopping,
	VoucherTargetAudienceInactiveBuyers,
}
var MapVoucherTargetAudiences = map[VoucherTargetAudience]struct{}{
	VoucherTargetAudienceAllBuyers:      {},
	VoucherTargetAudienceNewBuyers:      {},
	VoucherTargetAudienceLoyalty:        {},
	VoucherTargetAudienceEarlyShopping:  {},
	VoucherTargetAudienceInactiveBuyers: {},
}

const VoucherSortingFieldName = "name"

var MapVoucherSortings = map[string]struct{}{
	SortingFieldID:          {},
	VoucherSortingFieldName: {},
}

type Voucher struct {
	RowNo                 int64                 `json:"-" form:"-"`
	ID                    uuid.UUID             `json:"id" form:"id"`
	RecordStatus          RecordStatus          `json:"-" form:"-"`
	VoucherType           VoucherType           `json:"voucher_type" form:"voucher_type"`
	VoucherTypeName       string                `json:"voucher_type_name" form:"-"`
	TargetAudience        VoucherTargetAudience `json:"target_audience" form:"target_audience"`
	TargetAudienceName    string                `json:"target_audience_name" form:"-"`
	Code                  *string               `json:"code" form:"code"`
	DeletedCode           *string               `json:"-" form:"-"`
	Photo                 string                `json:"photo" form:"photo"`
	Name                  string                `json:"name" form:"name"`
	Discount              int64                 `json:"discount" form:"discount"`
	MinimumPurchase       int64                 `json:"minimum_purchase" form:"minimum_purchase"`
	QuotaPerBuyer         int64                 `json:"-" form:"quota_per_buyer"`
	EffectiveStartTime    time.Time             `json:"effective_start_time" form:"-"`
	RawEffectiveStartTime string                `json:"-" form:"effective_start_time"`
	EffectiveEndTime      time.Time             `json:"effective_end_time" form:"-"`
	RawEffectiveEndTime   string                `json:"-" form:"effective_end_time"`
	EffectiveStartDate    time.Time             `json:"effective_start_date" form:"-"`
	RawEffectiveStartDate string                `json:"-" form:"effective_start_date"`
	EffectiveEndDate      time.Time             `json:"effective_end_date" form:"-"`
	RawEffectiveEndDate   string                `json:"-" form:"effective_end_date"`
	Description           string                `json:"description" form:"description"`
	CreatedBy             uuid.UUID             `json:"-" form:"-"`
	CreatedAt             time.Time             `json:"-" form:"-"`
	UpdatedBy             uuid.UUID             `json:"-" form:"-"`
	UpdatedAt             time.Time             `json:"-" form:"-"`
	DeletedBy             *uuid.UUID            `json:"-" form:"-"`
	DeletedAt             *time.Time            `json:"-" form:"-"`
	DeletionReason        *string               `json:"-" form:"-"`
	BuyerIDs              []uuid.UUID           `json:"-" form:"-"`
}

func (q *Voucher) Validate() (err error) {
	if _, ok := MapVoucherTypes[q.VoucherType]; !ok {
		return errors.New("voucher_type: not found")
	}
	if _, ok := MapVoucherTargetAudiences[q.TargetAudience]; !ok {
		return errors.New("target_audience: not found")
	}
	if q.Code != nil {
		if *q.Code = strings.TrimSpace(*q.Code); *q.Code == "" {
			q.Code = nil
		}
	}
	if q.Code == nil {
		return errors.New("code: is required")
	}
	if !VoucherCodeRegex.MatchString(*q.Code) {
		return errors.New("code: is invalid")
	}
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.Discount < 1 {
		return errors.New("discount: is required")
	}
	if q.MinimumPurchase < 1 {
		return errors.New("minimum_purchase: is required")
	}
	if q.QuotaPerBuyer < 1 {
		return errors.New("quota_per_buyer: is required")
	}
	if q.RawEffectiveStartTime = strings.TrimSpace(q.RawEffectiveStartTime); q.RawEffectiveStartTime == "" {
		return errors.New("effective_start_time: is required")
	}
	if q.EffectiveStartTime, err = time.ParseInLocation(time.TimeOnly, q.RawEffectiveStartTime, time.Local); err != nil {
		return fmt.Errorf("effective_start_time: should use format '%s'", time.TimeOnly)
	}
	q.EffectiveStartTime = time.Date(
		1,
		1,
		1,
		q.EffectiveStartTime.Hour(),
		q.EffectiveStartTime.Minute(),
		q.EffectiveStartTime.Second(),
		0,
		q.EffectiveStartTime.Location(),
	)
	if q.RawEffectiveEndTime = strings.TrimSpace(q.RawEffectiveEndTime); q.RawEffectiveEndTime == "" {
		return errors.New("effective_end_time: is required")
	}
	if q.EffectiveEndTime, err = time.ParseInLocation(time.TimeOnly, q.RawEffectiveEndTime, time.Local); err != nil {
		return fmt.Errorf("effective_end_time: should use format '%s'", time.TimeOnly)
	}
	q.EffectiveEndTime = time.Date(
		1,
		1,
		1,
		q.EffectiveEndTime.Hour(),
		q.EffectiveEndTime.Minute(),
		q.EffectiveEndTime.Second(),
		999_999,
		q.EffectiveEndTime.Location(),
	)
	if !q.EffectiveEndTime.After(q.EffectiveStartTime) {
		return errors.New("effective_end_time: should exceed effective_start_time")
	}
	if q.RawEffectiveStartDate = strings.TrimSpace(q.RawEffectiveStartDate); q.RawEffectiveStartDate == "" {
		return errors.New("effective_start_date: is required")
	}
	effectiveStartDate, err := time.ParseInLocation(time.DateOnly, q.RawEffectiveStartDate, time.Local)
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
	effectiveEndDate, err := time.ParseInLocation(time.DateOnly, q.RawEffectiveEndDate, time.Local)
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

func (q *Voucher) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Voucher) Populate() *Voucher {
	q.VoucherTypeName = q.VoucherType.String()
	q.TargetAudienceName = q.TargetAudience.String()
	q.EffectiveStartTime = q.EffectiveStartTime.In(time.Local)
	q.RawEffectiveStartTime = q.EffectiveStartTime.Format(time.TimeOnly)
	q.EffectiveEndTime = q.EffectiveEndTime.In(time.Local)
	q.RawEffectiveEndTime = q.EffectiveEndTime.Format(time.TimeOnly)
	q.EffectiveStartDate = q.EffectiveStartDate.In(time.Local)
	q.RawEffectiveStartDate = q.EffectiveStartDate.Format(time.DateOnly)
	q.EffectiveEndDate = q.EffectiveEndDate.In(time.Local)
	q.RawEffectiveEndDate = q.EffectiveEndDate.Format(time.DateOnly)
	return q
}

func (q *Voucher) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.Photo, baseURLs...)
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

type VoucherFilter struct {
	VoucherIDs,
	ExcludedVoucherIDs,
	BuyerIDs []uuid.UUID
	RecordStatuses      []RecordStatus
	VoucherTypes        []VoucherType
	TargetAudiences     []VoucherTargetAudience
	Codes               []string
	TransactionDatetime time.Time
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewVoucherFilter() *VoucherFilter {
	return &VoucherFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *VoucherFilter) WithVoucherIDs(voucherIDs ...uuid.UUID) *VoucherFilter {
	q.VoucherIDs = voucherIDs
	return q
}

func (q *VoucherFilter) WithExcludedVoucherIDs(excludedVoucherIDs ...uuid.UUID) *VoucherFilter {
	q.ExcludedVoucherIDs = excludedVoucherIDs
	return q
}

func (q *VoucherFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *VoucherFilter {
	q.BuyerIDs = buyerIDs
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

func (q *VoucherFilter) WithTargetAudiences(targetAudiences ...VoucherTargetAudience) *VoucherFilter {
	q.TargetAudiences = targetAudiences
	return q
}

func (q *VoucherFilter) WithCodes(codes ...string) *VoucherFilter {
	q.Codes = codes
	return q
}

func (q *VoucherFilter) WithTransactionDatetime(transactionDatetime time.Time) *VoucherFilter {
	q.TransactionDatetime = transactionDatetime
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
