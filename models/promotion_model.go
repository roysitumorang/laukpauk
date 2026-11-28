package models

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/coregx/coregex"
	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/helper"
)

var ErrPromotionNotFound = errors.New("promotion not found")
var ErrUniquePromotionCodeViolation = errors.New("code: already exists")
var PromotionCodeRegex = coregex.MustCompile("^[a-zA-Z0-9-]+$")
var MapPromotionSortings = map[string]struct{}{
	SortingFieldID: {},
}

type Promotion struct {
	RowNo              int64        `json:"-"`
	ID                 uuid.UUID    `json:"id"`
	RecordStatus       RecordStatus `json:"record_status"`
	Name               string       `json:"name"`
	Code               *string      `json:"code"`
	DeletedCode        *string      `json:"deleted_code"`
	Description        string       `json:"description"`
	Discount           int64        `json:"discount"`
	QuotaPerBuyer      int64        `json:"quota_per_buyer"`
	EffectiveStartDate time.Time    `json:"effective_start_date"`
	EffectiveEndDate   time.Time    `json:"effective_end_date"`
	CreatedBy          uuid.UUID    `json:"created_by"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedBy          uuid.UUID    `json:"updated_by"`
	UpdatedAt          time.Time    `json:"updated_at"`
	DeletedBy          *uuid.UUID   `json:"deleted_by"`
	DeletedAt          *time.Time   `json:"deleted_at"`
	DeletionReason     *string      `json:"deletion_reason"`
	Usages             int64        `json:"usages"`
}

func (q *Promotion) Validate() (err error) {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.Code != nil {
		if *q.Code = strings.TrimSpace(*q.Code); *q.Code == "" {
			q.Code = nil
		}
	}
	if q.RecordStatus == RecordStatusActive && q.Code == nil {
		return errors.New("code: is required")
	}
	if !PromotionCodeRegex.MatchString(*q.Code) {
		return errors.New("code: is invalid")
	}
	if q.Discount < 1 {
		return errors.New("discount: is required")
	}
	q.EffectiveStartDate = time.Date(
		q.EffectiveStartDate.Year(),
		q.EffectiveStartDate.Month(),
		q.EffectiveStartDate.Day(),
		0,
		0,
		0,
		0,
		q.EffectiveStartDate.Location(),
	)
	q.EffectiveEndDate = time.Date(
		q.EffectiveEndDate.Year(),
		q.EffectiveEndDate.Month(),
		q.EffectiveEndDate.Day(),
		23,
		59,
		59,
		999_999,
		q.EffectiveEndDate.Location(),
	)
	if !q.EffectiveEndDate.After(q.EffectiveStartDate) {
		return errors.New("effective_end_date: should exceed effective_start_date")
	}
	q.Description = strings.TrimSpace(q.Description)
	return nil
}

func (q *Promotion) New() bool {
	return q.ID == uuid.Nil
}

type PromotionFilter struct {
	PromotionIDs,
	ExcludedPromotionIDs []uuid.UUID
	RecordStatuses      []RecordStatus
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

func NewPromotionFilter() *PromotionFilter {
	return &PromotionFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *PromotionFilter) WithPromotionIDs(voucherIDs ...uuid.UUID) *PromotionFilter {
	q.PromotionIDs = voucherIDs
	return q
}

func (q *PromotionFilter) WithExcludedPromotionIDs(excludedPromotionIDs ...uuid.UUID) *PromotionFilter {
	q.ExcludedPromotionIDs = excludedPromotionIDs
	return q
}

func (q *PromotionFilter) WithRecordStatuses(recordStatuses ...RecordStatus) *PromotionFilter {
	q.RecordStatuses = recordStatuses
	return q
}

func (q *PromotionFilter) WithCodes(codes ...string) *PromotionFilter {
	q.Codes = codes
	return q
}

func (q *PromotionFilter) WithTransactionDatetime(transactionDatetime time.Time) *PromotionFilter {
	q.TransactionDatetime = transactionDatetime
	return q
}

func (q *PromotionFilter) WithKeyword(keyword string) *PromotionFilter {
	q.Keyword = keyword
	return q
}

func (q *PromotionFilter) WithPaginationURL(paginationURL string) *PromotionFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *PromotionFilter) WithSortings(sortings ...Sorting) *PromotionFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *PromotionFilter) WithLimit(limit int64) *PromotionFilter {
	q.Limit = limit
	return q
}

func (q *PromotionFilter) WithPage(page int64) *PromotionFilter {
	q.Page = page
	return q
}

func (q *PromotionFilter) WithUrlValues(urlValues url.Values) *PromotionFilter {
	q.UrlValues = urlValues
	return q
}
