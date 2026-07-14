package models

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrTransactionNotFound                        = errors.New("transaction not found")
	ErrUniqueTransactionIdempotencyKeyViolation   = errors.New("idempotency_key: already exists")
	ErrUniqueTransactionReferenceNumberViolation  = errors.New("reference_number: already exists")
	ErrUniqueTransactionCodeViolation             = errors.New("unique_code: already exists")
	ErrUniqueTransactionProductVariantIdViolation = errors.New("product_variant_id: already exists")

	MapTransactionSortings = map[string]struct{}{
		SortingFieldID: {},
	}
)

const (
	TransactionReferenceNumberFormat = "LP%s"
)

const (
	TransactionStatusPending TransactionStatus = iota
	TransactionStatusCancelled
	TransactionStatusPaid
	TransactionStatusProcessed
	TransactionStatusShipped
	TransactionStatusDelivered
)

type (
	TransactionStatus int

	Transaction struct {
		RowNo                int64              `json:"-" form:"-"`
		ID                   uuid.UUID          `json:"-" form:"id"`
		ParentID             *uuid.UUID         `json:"-" form:"parent_id"`
		BuyerID              uuid.UUID          `json:"-" form:"buyer_id"`
		IdempotencyKey       uuid.UUID          `json:"-" form:"idempotency_key"`
		ReferenceNumber      string             `json:"-" form:"-"`
		TransactionStatus    TransactionStatus  `json:"-" form:"-"`
		RecipientName        string             `json:"-" form:"recipient_name"`
		RecipientMobilePhone string             `json:"-" form:"recipient_mobile_phone"`
		BuyerNote            string             `json:"-" form:"buyer_note"`
		UniqueCode           string             `json:"-" form:"-"`
		Subtotal             int64              `json:"-" form:"-"`
		ReferrerID           *uuid.UUID         `json:"-" form:"-"`
		ReferralCode         *string            `json:"-" form:"referral_code"`
		VoucherID            *uuid.UUID         `json:"-" form:"-"`
		VoucherType          *VoucherType       `json:"-" form:"-"`
		VoucherCode          *string            `json:"-" form:"voucher_code"`
		VoucherDiscount      int64              `json:"-" form:"-"`
		PointDiscount        int64              `json:"-" form:"point_discount"`
		WalletDiscount       int64              `json:"-" form:"wallet_discount"`
		ShoppingCost         int64              `json:"-" form:"-"`
		PaymentAmount        int64              `json:"-" form:"-"`
		PaymentDeadline      time.Time          `json:"-" form:"payment_deadline"`
		DeliveryCost         int64              `json:"-" form:"delivery_cost"`
		DeliveryLatitude     float64            `json:"-" form:"delivery_latitude"`
		DeliveryLongitude    float64            `json:"-" form:"delivery_longitude"`
		DeliveryAddress      string             `json:"-" form:"delivery_address"`
		DeliveryInstruction  string             `json:"-" form:"delivery_instruction"`
		DeliveryPlanID       uuid.UUID          `json:"-" form:"delivery_plan_id"`
		DeliverySpeed        DeliverySpeed      `json:"-" form:"-"`
		DeliveryDatetime     time.Time          `json:"-" form:"delivery_datetime"`
		CreatedAt            time.Time          `json:"-" form:"-"`
		PaidAt               *time.Time         `json:"-" form:"-"`
		PaymentRequestID     *uuid.UUID         `json:"-" form:"-"`
		PaymentID            *uuid.UUID         `json:"-" form:"-"`
		PaymentStatus        *string            `json:"-" form:"-"`
		ProcessedBy          *uuid.UUID         `json:"-" form:"-"`
		ProcessedAt          *time.Time         `json:"-" form:"-"`
		RevisedPaymentAmount *int64             `json:"-" form:"-"`
		RefundAmount         *int64             `json:"-" form:"-"`
		RefundedAt           *time.Time         `json:"-" form:"-"`
		RefundID             *uuid.UUID         `json:"-" form:"-"`
		CancelledBy          *uuid.UUID         `json:"-" form:"-"`
		CancelledAt          *time.Time         `json:"-" form:"-"`
		CancellationReason   *string            `json:"-" form:"-"`
		ShippedBy            *uuid.UUID         `json:"-" form:"-"`
		ShippedAt            *time.Time         `json:"-" form:"-"`
		DeliveredBy          *uuid.UUID         `json:"-" form:"-"`
		DeliveredAt          *time.Time         `json:"-" form:"-"`
		Items                []*TransactionItem `json:"-" form:"items"`
	}

	TransactionItem struct {
		RowNo                   int64     `json:"-" form:"-"`
		ID                      uuid.UUID `json:"-" form:"-"`
		TransactionID           uuid.UUID `json:"-" form:"-"`
		ProductVariantID        uuid.UUID `json:"-" form:"product_variant_id"`
		ProductID               uuid.UUID `json:"-" form:"-"`
		ProductName             string    `json:"-" form:"-"`
		StockUnit               string    `json:"-" form:"-"`
		Price                   int64     `json:"-" form:"-"`
		DiscountMinimumQuantity int64     `json:"-" form:"-"`
		DiscountAmount          int64     `json:"-" form:"-"`
		Quantity                int64     `json:"-" form:"quantity"`
		Subtotal                int64     `json:"-" form:"-"`
		BuyerNote               string    `json:"-" form:"buyer_note"`
		OutOfStock              bool      `json:"-" form:"out_of_stock"`
	}

	TransactionFilter struct {
		TransactionIDs,
		ExcludedTransactionIDs,
		BuyerIDs,
		IdempotencyKeys,
		PaymentRequestIDs []uuid.UUID
		TransactionStatuses []TransactionStatus
		ReferralCodes       []string
		Keyword,
		PaginationURL string
		Sortings    []Sorting
		MapSortings Map
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewTransactionFilter() *TransactionFilter {
	return new(TransactionFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *TransactionFilter) WithTransactionIDs(transactionIDs ...uuid.UUID) *TransactionFilter {
	q.TransactionIDs = transactionIDs
	return q
}

func (q *TransactionFilter) WithExcludedTransactionIDs(excludedTransactionIDs ...uuid.UUID) *TransactionFilter {
	q.ExcludedTransactionIDs = excludedTransactionIDs
	return q
}

func (q *TransactionFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *TransactionFilter {
	q.BuyerIDs = buyerIDs
	return q
}

func (q *TransactionFilter) WithIdempotencyKeys(idempotencyKeys ...uuid.UUID) *TransactionFilter {
	q.IdempotencyKeys = idempotencyKeys
	return q
}

func (q *TransactionFilter) WithPaymentRequestIDs(paymentRequestIDs ...uuid.UUID) *TransactionFilter {
	q.PaymentRequestIDs = paymentRequestIDs
	return q
}

func (q *TransactionFilter) WithTransactionStatuses(transactionStatuses ...TransactionStatus) *TransactionFilter {
	q.TransactionStatuses = transactionStatuses
	return q
}

func (q *TransactionFilter) WithReferralCodes(referralCodes ...string) *TransactionFilter {
	q.ReferralCodes = referralCodes
	return q
}

func (q *TransactionFilter) WithKeyword(keyword string) *TransactionFilter {
	q.Keyword = keyword
	return q
}

func (q *TransactionFilter) WithPaginationURL(paginationURL string) *TransactionFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *TransactionFilter) WithSortings(sortings ...Sorting) *TransactionFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *TransactionFilter) WithLimit(limit int64) *TransactionFilter {
	q.Limit = limit
	return q
}

func (q *TransactionFilter) WithPage(page int64) *TransactionFilter {
	q.Page = page
	return q
}

func (q *TransactionFilter) WithUrlValues(urlValues url.Values) *TransactionFilter {
	q.UrlValues = urlValues
	return q
}

func (q *Transaction) Validate() error {
	return nil
}

func (q TransactionStatus) Int() int {
	return int(q)
}

func (q TransactionStatus) String() string {
	switch q {
	case TransactionStatusCancelled:
		return "cancelled"
	case TransactionStatusPaid:
		return "paid"
	case TransactionStatusProcessed:
		return "processed"
	case TransactionStatusShipped:
		return "otw"
	case TransactionStatusDelivered:
		return "delivered"
	case TransactionStatusPending:
		fallthrough
	default:
		return "pending"
	}
}

func (q TransactionStatus) Equal(v TransactionStatus) bool {
	return q.Int() == v.Int()
}
