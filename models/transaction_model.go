package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
)

var ErrTransactionNotFound = errors.New("transaction not found")
var ErrUniqueTransactionCartIdViolation = errors.New("cart_id: already exists")
var ErrUniqueTransactionReferenceNumberViolation = errors.New("reference_number: already exists")
var ErrUniqueTransactionProductVariantIdViolation = errors.New("product_variant_id: already exists")
var MapTransactionSortings = map[string]struct{}{
	SortingFieldID: {},
}

const TransactionReferenceNumberFormat = "LP%s%03d%s"

type TransactionStatus uint8

const (
	TransactionStatusPending TransactionStatus = iota
	TransactionStatusCancelled
	TransactionStatusPaid
	TransactionStatusWithdrawalPending
	TransactionStatusWithdrawn
	TransactionStatusProcessed
	TransactionStatusShipped
	TransactionStatusComplained
	TransactionStatusCompleted
)

func (q TransactionStatus) Int() uint8 {
	return uint8(q)
}

func (q TransactionStatus) String() string {
	switch q {
	case TransactionStatusCancelled:
		return "cancelled"
	case TransactionStatusPaid:
		return "paid"
	case TransactionStatusWithdrawalPending:
		return "withdrawal_pending"
	case TransactionStatusWithdrawn:
		return "withdrawn"
	case TransactionStatusProcessed:
		return "processed"
	case TransactionStatusShipped:
		return "otw"
	case TransactionStatusCompleted:
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

type Transaction struct {
	RowNo                           int64              `json:"-"`
	ID                              uuid.UUID          `json:"id"`
	ParentID                        *uuid.UUID         `json:"parent_id"`
	Reordered                       bool               `json:"reordered"`
	UserID                          uuid.UUID          `json:"user_id"`
	CartID                          *uuid.UUID         `json:"cart_id"`
	ReferenceNumber                 string             `json:"reference_number"`
	Status                          TransactionStatus  `json:"status"`
	RecipientName                   string             `json:"recipient_name"`
	RecipientMobilePhone            string             `json:"recipient_mobile_phone"`
	BuyerNote                       string             `json:"buyer_note"`
	UniqueCode                      int64              `json:"unique_code"`
	ShoppingCost                    int64              `json:"shopping_cost"`
	ReferrerID                      *uuid.UUID         `json:"referrer_id"`
	ReferralCode                    *string            `json:"referral_code"`
	ReferralPoints                  int64              `json:"referral_points"`
	VoucherID                       *uuid.UUID         `json:"voucher_id"`
	VoucherType                     *VoucherType       `json:"voucher_type"`
	VoucherDiscount                 int64              `json:"voucher_discount"`
	RewardPointsDiscounted          bool               `json:"reward_points_discounted"`
	RewardPointsDiscount            int64              `json:"reward_points_discount"`
	WalletDiscounted                bool               `json:"wallet_discounted"`
	WalletAmount                    int64              `json:"wallet_amount"`
	WalletDiscount                  int64              `json:"wallet_discount"`
	DiscountedShoppingCost          int64              `json:"discounted_shopping_cost"`
	DeliveryEstimatedCost           int64              `json:"delivery_estimated_cost"`
	DiscountedDeliveryEstimatedCost int64              `json:"discounted_delivery_estimated_cost"`
	DeliveryAddressID               uuid.UUID          `json:"delivery_address_id"`
	DeliveryDestinationLatitude     decimal.Decimal    `json:"delivery_destination_latitude"`
	DeliveryDestinationLongitude    decimal.Decimal    `json:"delivery_destination_longitude"`
	DeliveryDistanceKM              int64              `json:"delivery_distance_km"`
	DeliveryStreetAddress           string             `json:"delivery_street_address"`
	DeliveryStreetAddressNote       string             `json:"delivery_street_address_note"`
	DeliveryInstruction             string             `json:"delivery_instruction"`
	DeliveryDatetime                time.Time          `json:"delivery_datetime"`
	PaymentAmount                   int64              `json:"payment_amount"`
	PaymentDueDate                  *time.Time         `json:"payment_due_date"`
	PaymentChannelID                *uuid.UUID         `json:"payment_channel_id"`
	PaymentRequestID                *string            `json:"payment_request_id"`
	PaymentTokenID                  *string            `json:"payment_token_id"`
	PaymentChannelName              *string            `json:"payment_channel_name"`
	PaymentVirtualAccountNumber     *string            `json:"payment_virtual_account_number"`
	PaymentVirtualAccountExpiredAt  *time.Time         `json:"payment_virtual_account_expired_at"`
	CreatedAt                       time.Time          `json:"created_at"`
	UpdatedAt                       time.Time          `json:"updated_at"`
	CancelledAt                     *time.Time         `json:"cancelled_at"`
	PaymentStatus                   *string            `json:"payment_status"`
	PaymentFailureCode              *string            `json:"payment_failure_code"`
	PaymentCapturedAt               *time.Time         `json:"payment_captured_at"`
	PaymentID                       *string            `json:"payment_id"`
	PaidAt                          *time.Time         `json:"paid_at"`
	WithdrawalRequestedAt           *time.Time         `json:"withdrawal_requested_at"`
	WithdrawalReason                *string            `json:"withdrawal_reason"`
	WithdrawnBy                     *uuid.UUID         `json:"withdrawn_by"`
	WithdrawnAt                     *time.Time         `json:"withdrawn_at"`
	ProcessedBy                     *uuid.UUID         `json:"processed_by"`
	ProcessedAt                     *time.Time         `json:"processed_at"`
	RevisedShoppingCost             int64              `json:"revised_shopping_cost"`
	RevisedDiscountedShoppingCost   int64              `json:"revised_discounted_shopping_cost"`
	RevisedPaymentAmount            int64              `json:"revised_payment_amount"`
	RefundAmount                    int64              `json:"refund_amount"`
	RefundedAt                      *time.Time         `json:"refunded_at"`
	RefundID                        *string            `json:"refund_id"`
	ShippedBy                       *uuid.UUID         `json:"shipped_by"`
	ShippedAt                       *time.Time         `json:"shipped_at"`
	DeliveryActualCost              int64              `json:"delivery_actual_cost"`
	CompletedBy                     *uuid.UUID         `json:"completed_by"`
	CompletedAt                     *time.Time         `json:"completed_at"`
	Items                           []*TransactionItem `json:"items"`
}

type TransactionItem struct {
	RowNo         int64     `json:"-"`
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	// CartItemID is input only: it picks the cart item a new transaction item
	// comes from. It is not stored, so it is empty when read back.
	CartItemID              *uuid.UUID `json:"cart_item_id,omitempty"`
	ProductVariantID        uuid.UUID  `json:"product_variant_id"`
	ProductID               uuid.UUID  `json:"product_id"`
	ProductName             string     `json:"product_name"`
	ProductPhoto            string     `json:"product_photo"`
	StockUnit               string     `json:"stock_unit"`
	BasePrice               int64      `json:"-"`
	SellingPrice            int64      `json:"selling_price"`
	DiscountMinimumQuantity int64      `json:"discount_minimum_quantity"`
	DiscountAmount          int64      `json:"discount_amount"`
	Quantity                int64      `json:"quantity"`
	Subtotal                int64      `json:"subtotal"`
	BuyerNote               string     `json:"buyer_note"`
	RevisedQuantity         int64      `json:"revised_quantity"`
	RevisedSubtotal         int64      `json:"revised_subtotal"`
	RevisionNote            string     `json:"revision_note"`
	OutOfStock              bool       `json:"out_of_stock"`
}

func (q *Transaction) Validate() error {
	if q.New() &&
		(q.CartID == nil ||
			q.CartID.Compare(uuid.Nil()) == 0) {
		return errors.New("cart_id: is required")
	}
	q.BuyerNote = strings.TrimSpace(q.BuyerNote)
	if q.ReferralCode != nil {
		if *q.ReferralCode = strings.TrimSpace(*q.ReferralCode); *q.ReferralCode == "" {
			q.ReferralCode = nil
		}
	}
	if !q.RewardPointsDiscounted &&
		q.VoucherID != nil &&
		q.VoucherID.Compare(uuid.Nil()) == 0 {
		q.VoucherID = nil
	}
	if q.DeliveryAddressID.Compare(uuid.Nil()) == 0 {
		return errors.New("delivery_address_id: is required")
	}
	if q.DeliveryStreetAddress = strings.TrimSpace(q.DeliveryStreetAddress); q.DeliveryStreetAddress == "" {
		return errors.New("delivery_street_address: is required")
	}
	q.DeliveryInstruction = strings.TrimSpace(q.DeliveryInstruction)
	if q.DeliveryDatetime.IsZero() {
		return errors.New("delivery_datetime: is required")
	}
	if !q.DeliveryDatetime.After(time.Now()) {
		return errors.New("delivery_datetime: should exceed current date")
	}
	if len(q.Items) == 0 {
		return errors.New("items: is required")
	}
	for i, item := range q.Items {
		if item.CartItemID == nil || item.CartItemID.Compare(uuid.Nil()) == 0 {
			return fmt.Errorf("items[%d].cart_item_id: is required", i)
		}
		item.BuyerNote = strings.TrimSpace(item.BuyerNote)
	}
	return nil
}

func (q *Transaction) New() bool {
	return q.ID.Compare(uuid.Nil()) == 0
}

func (q *Transaction) CalculateShoppingCost() {
	q.UniqueCode = helper.RandomInteger(1000)
	q.ShoppingCost = q.UniqueCode
	for i, item := range q.Items {
		if item.OutOfStock {
			item.Subtotal = 0
		} else if (item.DiscountMinimumQuantity > 0 &&
			item.Quantity >= item.DiscountMinimumQuantity) ||
			item.DiscountAmount > 0 {
			item.Subtotal = max(0, item.SellingPrice-item.DiscountAmount) * item.Quantity
		} else {
			item.Subtotal = item.SellingPrice * item.Quantity
		}
		item.RevisedQuantity = item.Quantity
		item.RevisedSubtotal = item.Subtotal
		q.ShoppingCost += item.Subtotal
		q.Items[i] = item
	}
	q.RevisedShoppingCost = q.ShoppingCost
}

func (q *Transaction) CalculateDiscount() {
	q.DiscountedShoppingCost = q.ShoppingCost
	q.DiscountedDeliveryEstimatedCost = q.DeliveryEstimatedCost
	if q.RewardPointsDiscounted && q.RewardPointsDiscount > 0 {
		rewardPointsAmount := q.RewardPointsDiscount
		if q.DiscountedShoppingCost > 0 {
			rewardPointsDiscount := min(q.DiscountedShoppingCost, rewardPointsAmount)
			q.DiscountedShoppingCost -= rewardPointsDiscount
			rewardPointsAmount -= rewardPointsDiscount
		}
		if q.DiscountedDeliveryEstimatedCost > 0 && rewardPointsAmount > 0 {
			rewardPointsDiscount := min(q.DiscountedDeliveryEstimatedCost, rewardPointsAmount)
			q.DiscountedDeliveryEstimatedCost -= rewardPointsDiscount
		}
	} else if q.VoucherID != nil && q.VoucherType != nil {
		if *q.VoucherType == VoucherTypeShoppingCost {
			q.DiscountedShoppingCost = max(0, q.DiscountedShoppingCost-q.VoucherDiscount)
		} else {
			q.DiscountedDeliveryEstimatedCost = max(0, q.DeliveryEstimatedCost-q.VoucherDiscount)
		}
	}
	q.WalletDiscount = 0
	if q.WalletDiscounted && q.WalletAmount > 0 {
		walletAmount := q.WalletAmount
		if q.DiscountedShoppingCost > 0 {
			walletDiscount := min(q.DiscountedShoppingCost, q.WalletAmount)
			q.DiscountedShoppingCost -= walletDiscount
			walletAmount -= walletDiscount
			q.WalletDiscount += walletDiscount
		}
		if q.DiscountedDeliveryEstimatedCost > 0 && walletAmount > 0 {
			walletDiscount := min(q.DiscountedDeliveryEstimatedCost, walletAmount)
			q.DiscountedDeliveryEstimatedCost -= walletDiscount
			q.WalletDiscount += walletDiscount
		}
	}
	q.DeliveryActualCost = q.DiscountedDeliveryEstimatedCost
	q.PaymentAmount = q.DiscountedShoppingCost + q.DiscountedDeliveryEstimatedCost
	q.RevisedDiscountedShoppingCost = q.DiscountedShoppingCost
	q.RevisedPaymentAmount = q.PaymentAmount
}

func (q *Transaction) Revise() {
	q.RevisedShoppingCost = q.UniqueCode
	for i, item := range q.Items {
		if item.OutOfStock {
			item.RevisedSubtotal = 0
		} else if (item.DiscountMinimumQuantity > 0 &&
			item.RevisedQuantity >= item.DiscountMinimumQuantity) ||
			item.DiscountAmount > 0 {
			item.RevisedSubtotal = max(0, item.SellingPrice-item.DiscountAmount) * item.RevisedQuantity
		} else {
			item.RevisedSubtotal = item.SellingPrice * item.RevisedQuantity
		}
		q.RevisedShoppingCost += item.RevisedSubtotal
		q.Items[i] = item
	}
	q.RevisedDiscountedShoppingCost = q.RevisedShoppingCost
	q.DiscountedDeliveryEstimatedCost = q.DeliveryEstimatedCost
	if q.RewardPointsDiscounted && q.RewardPointsDiscount > 0 {
		rewardPointsAmount := q.RewardPointsDiscount
		if q.RevisedDiscountedShoppingCost > 0 {
			rewardPointsDiscount := min(q.RevisedDiscountedShoppingCost, rewardPointsAmount)
			q.RevisedDiscountedShoppingCost -= rewardPointsDiscount
			rewardPointsAmount -= rewardPointsDiscount
		}
		if q.DiscountedDeliveryEstimatedCost > 0 && rewardPointsAmount > 0 {
			rewardPointsDiscount := min(q.DiscountedDeliveryEstimatedCost, rewardPointsAmount)
			q.DiscountedDeliveryEstimatedCost -= rewardPointsDiscount
		}
	} else if q.VoucherID != nil && q.VoucherType != nil {
		if *q.VoucherType == VoucherTypeShoppingCost {
			q.RevisedDiscountedShoppingCost = max(0, q.RevisedDiscountedShoppingCost-q.VoucherDiscount)
		} else {
			q.DiscountedDeliveryEstimatedCost = max(0, q.DeliveryEstimatedCost-q.VoucherDiscount)
		}
	}
	q.WalletDiscount = 0
	if q.WalletDiscounted && q.WalletAmount > 0 {
		walletAmount := q.WalletAmount
		if q.RevisedDiscountedShoppingCost > 0 {
			walletDiscount := min(q.RevisedDiscountedShoppingCost, q.WalletAmount)
			q.RevisedDiscountedShoppingCost -= walletDiscount
			walletAmount -= walletDiscount
			q.WalletDiscount += walletDiscount
		}
		if q.DiscountedDeliveryEstimatedCost > 0 && walletAmount > 0 {
			walletDiscount := min(q.DiscountedDeliveryEstimatedCost, walletAmount)
			q.DiscountedDeliveryEstimatedCost -= walletDiscount
			q.WalletDiscount += walletDiscount
		}
	}
	q.RevisedPaymentAmount = q.RevisedDiscountedShoppingCost + q.DiscountedDeliveryEstimatedCost
	q.RefundAmount = q.PaymentAmount - q.RevisedPaymentAmount
}

func (q *TransactionItem) PhotoPath(baseURLs ...string) string {
	return helper.PhotoPath(q.ProductPhoto, baseURLs...)
}

type TransactionFilter struct {
	TransactionIDs,
	ExcludedTransactionIDs,
	UserIDs,
	CartIDs,
	ReferrerIDs,
	VoucherIDs []uuid.UUID
	PaymentRequestIDs,
	PaymentTokenIDs,
	PaymentIDs []string
	PaymentDueDate,
	ShippedAt time.Time
	TransactionStatuses,
	ExcludedTransactionStatuses []TransactionStatus
	ReturnCountOnly bool
	Keyword,
	PaginationURL string
	Sortings    []Sorting
	MapSortings Map
	Limit,
	Page int64
	UrlValues url.Values
}

func NewTransactionFilter() *TransactionFilter {
	return &TransactionFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	}
}

func (q *TransactionFilter) WithTransactionIDs(transactionIDs ...uuid.UUID) *TransactionFilter {
	q.TransactionIDs = transactionIDs
	return q
}

func (q *TransactionFilter) WithExcludedTransactionIDs(excludedTransactionIDs ...uuid.UUID) *TransactionFilter {
	q.ExcludedTransactionIDs = excludedTransactionIDs
	return q
}

func (q *TransactionFilter) WithUserIDs(userIDs ...uuid.UUID) *TransactionFilter {
	q.UserIDs = userIDs
	return q
}

func (q *TransactionFilter) WithCartIDs(cartIDs ...uuid.UUID) *TransactionFilter {
	q.CartIDs = cartIDs
	return q
}

func (q *TransactionFilter) WithReferrerIDs(referrerIDs ...uuid.UUID) *TransactionFilter {
	q.ReferrerIDs = referrerIDs
	return q
}

func (q *TransactionFilter) WithVoucherIDs(voucherIDs ...uuid.UUID) *TransactionFilter {
	q.VoucherIDs = voucherIDs
	return q
}

func (q *TransactionFilter) WithPaymentRequestIDs(paymentRequestIDs ...string) *TransactionFilter {
	q.PaymentRequestIDs = paymentRequestIDs
	return q
}

func (q *TransactionFilter) WithPaymentTokenIDs(paymentTokenIDs ...string) *TransactionFilter {
	q.PaymentTokenIDs = paymentTokenIDs
	return q
}

func (q *TransactionFilter) WithPaymentIDs(paymentIDs ...string) *TransactionFilter {
	q.PaymentIDs = paymentIDs
	return q
}

func (q *TransactionFilter) WithPaymentDueDate(paymentDueDate time.Time) *TransactionFilter {
	q.PaymentDueDate = paymentDueDate
	return q
}

func (q *TransactionFilter) WithShippedAt(shippedAt time.Time) *TransactionFilter {
	q.ShippedAt = shippedAt
	return q
}

func (q *TransactionFilter) WithTransactionStatuses(transactionStatuses ...TransactionStatus) *TransactionFilter {
	q.TransactionStatuses = transactionStatuses
	return q
}

func (q *TransactionFilter) WithExcludedTransactionStatuses(excludedTransactionStatuses ...TransactionStatus) *TransactionFilter {
	q.ExcludedTransactionStatuses = excludedTransactionStatuses
	return q
}

func (q *TransactionFilter) WithReturnCountOnly(returnCountOnly bool) *TransactionFilter {
	q.ReturnCountOnly = returnCountOnly
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

type TransactionPaymentWebhookNotification struct {
	PaymentCapture       TransactionPaymentEvent `json:"paymentCapture"`
	PaymentAuthorization TransactionPaymentEvent `json:"paymentAuthorization"`
	PaymentFailure       TransactionPaymentEvent `json:"paymentFailure"`
}

type TransactionPaymentEvent struct {
	Value struct {
		Event      string    `json:"event"`
		BusinessID string    `json:"business_id"`
		Created    time.Time `json:"created"`
		Data       struct {
			PaymentID         string         `json:"payment_id"`
			BusinessID        string         `json:"business_id"`
			Status            string         `json:"status"`
			PaymentRequestID  string         `json:"payment_request_id"`
			RequestAmount     int64          `json:"request_amount"`
			CustomerID        string         `json:"customer_id"`
			ChannelCode       string         `json:"channel_code"`
			Country           string         `json:"country"`
			Currency          string         `json:"currency"`
			ReferenceID       string         `json:"reference_id"`
			Description       string         `json:"description"`
			FailureCode       string         `json:"failure_code"`
			ChannelProperties map[string]any `json:"channel_properties"`
			ChannelType       string         `json:"type"`
			Created           time.Time      `json:"created"`
			Updated           time.Time      `json:"updated"`
		} `json:"data"`
	} `json:"value"`
}

type TransactionWithdrawal struct {
	Reason string `json:"reason"`
}

func (q *TransactionWithdrawal) Validate() error {
	if q.Reason = strings.TrimSpace(q.Reason); q.Reason == "" {
		return errors.New("reason: is required")
	}
	return nil
}

type TransactionShipment struct {
	DeliveryActualCost int64 `json:"delivery_actual_cost"`
}

func (q *TransactionShipment) Validate() error {
	if q.DeliveryActualCost < 1 {
		return errors.New("delivery_actual_cost: is required")
	}
	return nil
}
