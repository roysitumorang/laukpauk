package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/nyaruka/phonenumbers"
	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrTransactionNotFound                        = errors.New("transaction not found")
	ErrUniqueTransactionCartIdViolation           = errors.New("cart_id: already exists")
	ErrUniqueTransactionReferenceNumberViolation  = errors.New("reference_number: already exists")
	ErrUniqueTransactionProductVariantIdViolation = errors.New("product_variant_id: already exists")

	MapTransactionSortings = map[string]struct{}{
		SortingFieldID: {},
	}
)

const (
	TransactionReferenceNumberFormat = "LP%s%03d%s"
)

const (
	TransactionStatusPending TransactionStatus = iota
	TransactionStatusCancelled
	TransactionStatusPaid
	TransactionStatusRevoked
	TransactionStatusProcessed
	TransactionStatusShipped
	TransactionStatusCompleted
)

type (
	TransactionStatus int

	Transaction struct {
		RowNo                         int64              `json:"-" form:"-"`
		ID                            uuid.UUID          `json:"id" form:"id"`
		ParentID                      *uuid.UUID         `json:"parent_id" form:"parent_id"`
		Reordered                     bool               `json:"reordered" form:"reordered"`
		BuyerID                       uuid.UUID          `json:"buyer_id" form:"buyer_id"`
		CartID                        *uuid.UUID         `json:"cart_id" form:"cart_id"`
		ReferenceNumber               string             `json:"-" form:"-"`
		TransactionStatus             TransactionStatus  `json:"-" form:"-"`
		RecipientName                 string             `json:"recipient_name" form:"recipient_name"`
		RecipientMobilePhone          string             `json:"-" form:"-"`
		RawRecipientMobilePhone       string             `json:"recipient_mobile_phone" form:"recipient_mobile_phone"`
		BuyerNote                     string             `json:"buyer_note" form:"buyer_note"`
		UniqueCode                    int64              `json:"-" form:"-"`
		ShoppingCost                  int64              `json:"-" form:"-"`
		ReferrerID                    *uuid.UUID         `json:"-" form:"-"`
		ReferralCode                  *string            `json:"referral_code" form:"referral_code"`
		ReferralPoints                int64              `json:"-" form:"-"`
		VoucherID                     *uuid.UUID         `json:"-" form:"-"`
		VoucherType                   *VoucherType       `json:"-" form:"-"`
		VoucherCode                   *string            `json:"voucher_code" form:"voucher_code"`
		VoucherDiscount               int64              `json:"-" form:"-"`
		RewardPointsDiscounted        bool               `json:"reward_points_discounted" form:"reward_points_discounted"`
		RewardPointsDiscount          int64              `json:"-" form:"-"`
		WalletDiscounted              bool               `json:"wallet_discounted" form:"wallet_discounted"`
		WalletAmount                  int64              `json:"-" form:"-"`
		WalletDiscount                int64              `json:"-" form:"-"`
		DiscountedShoppingCost        int64              `json:"-" form:"-"`
		PaymentAmount                 int64              `json:"-" form:"-"`
		PaymentDueDate                time.Time          `json:"-" form:"-"`
		DeliveryCost                  int64              `json:"-" form:"-"`
		DiscountedDeliveryCost        int64              `json:"-" form:"-"`
		DeliveryRouteID               uuid.UUID          `json:"delivery_route_id" form:"-"`
		DeliveryLatitude              decimal.Decimal    `json:"delivery_latitude" form:"delivery_latitude"`
		DeliveryLongitude             decimal.Decimal    `json:"delivery_longitude" form:"delivery_longitude"`
		DeliveryDistanceKM            int64              `json:"delivery_distance_km" form:"-"`
		DeliveryAddress               string             `json:"delivery_address" form:"delivery_address"`
		DeliveryInstruction           string             `json:"delivery_instruction" form:"delivery_instruction"`
		DeliveryDatetime              time.Time          `json:"delivery_datetime" form:"delivery_datetime"`
		PaymentChannelID              *uuid.UUID         `json:"payment_channel_id" form:"-"`
		PaymentRequestID              *string            `json:"-" form:"-"`
		PaymentTokenID                *string            `json:"-" form:"-"`
		CreatedAt                     time.Time          `json:"created_at" form:"-"`
		UpdatedAt                     time.Time          `json:"updated_at" form:"-"`
		CancelledAt                   *time.Time         `json:"-" form:"-"`
		PaymentID                     *string            `json:"-" form:"-"`
		PaidAt                        *time.Time         `json:"-" form:"-"`
		RevokedAt                     *time.Time         `json:"-" form:"-"`
		RevocationReason              *string            `json:"-" form:"-"`
		ProcessedBy                   *uuid.UUID         `json:"-" form:"-"`
		ProcessedAt                   *time.Time         `json:"-" form:"-"`
		RevisedShoppingCost           int64              `json:"-" form:"-"`
		RevisedDiscountedShoppingCost int64              `json:"-" form:"-"`
		RevisedPaymentAmount          int64              `json:"-" form:"-"`
		RefundAmount                  *int64             `json:"-" form:"-"`
		RefundedAt                    *time.Time         `json:"-" form:"-"`
		RefundID                      *string            `json:"-" form:"-"`
		ShippedBy                     *uuid.UUID         `json:"-" form:"-"`
		ShippedAt                     *time.Time         `json:"-" form:"-"`
		CompletedBy                   *uuid.UUID         `json:"-" form:"-"`
		CompletedAt                   *time.Time         `json:"-" form:"-"`
		Items                         []*TransactionItem `json:"items" form:"items"`
	}

	TransactionItem struct {
		RowNo                   int64      `json:"-" form:"-"`
		ID                      uuid.UUID  `json:"id" form:"-"`
		TransactionID           uuid.UUID  `json:"-" form:"-"`
		CartItemID              *uuid.UUID `json:"cart_item_id,omitempty" form:"cart_item_id"`
		ProductVariantID        uuid.UUID  `json:"-" form:"-"`
		ProductID               uuid.UUID  `json:"-" form:"-"`
		ProductName             string     `json:"-" form:"-"`
		StockUnit               string     `json:"-" form:"-"`
		Price                   int64      `json:"-" form:"-"`
		DiscountMinimumQuantity int64      `json:"-" form:"-"`
		DiscountAmount          int64      `json:"-" form:"-"`
		Quantity                int64      `json:"quantity" form:"quantity"`
		Subtotal                int64      `json:"-" form:"-"`
		BuyerNote               string     `json:"buyer_note" form:"buyer_note"`
		RevisedQuantity         int64      `json:"revised_quantity" form:"revised_quantity"`
		RevisedSubtotal         int64      `json:"revised_subtotal" form:"-"`
		RevisionNote            string     `json:"revision_note" form:"revision_note"`
		OutOfStock              bool       `json:"out_of_stock" form:"out_of_stock"`
	}

	TransactionFilter struct {
		TransactionIDs,
		ExcludedTransactionIDs,
		BuyerIDs,
		CartIDs,
		ReferrerIDs,
		VoucherIDs,
		PaymentRequestIDs,
		PaymentTokenIDs,
		PaymentIDs []uuid.UUID
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

	PaymentWebhookNotification struct {
		PaymentCapture       PaymentEvent `json:"paymentCapture"`
		PaymentAuthorization PaymentEvent `json:"paymentAuthorization"`
		PaymentFailure       PaymentEvent `json:"paymentFailure"`
	}

	PaymentEvent struct {
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
)

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

func (q *TransactionFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *TransactionFilter {
	q.BuyerIDs = buyerIDs
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

func (q *TransactionFilter) WithPaymentRequestIDs(paymentRequestIDs ...uuid.UUID) *TransactionFilter {
	q.PaymentRequestIDs = paymentRequestIDs
	return q
}

func (q *TransactionFilter) WithPaymentTokenIDs(paymentTokenIDs ...uuid.UUID) *TransactionFilter {
	q.PaymentTokenIDs = paymentTokenIDs
	return q
}

func (q *TransactionFilter) WithPaymentIDs(paymentIDs ...uuid.UUID) *TransactionFilter {
	q.PaymentIDs = paymentIDs
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

func (q *Transaction) Validate() error {
	if q.New() &&
		(q.CartID == nil ||
			q.CartID.Compare(uuid.Nil()) == 0) {
		return errors.New("cart_id: is required")
	}
	if q.RecipientName = strings.TrimSpace(q.RecipientName); q.RecipientName == "" {
		return errors.New("recipient_name: is required")
	}
	if q.RawRecipientMobilePhone = strings.TrimSpace(q.RawRecipientMobilePhone); q.RawRecipientMobilePhone == "" {
		return errors.New("recipient_mobile_phone: is required")
	}
	mobilePhone, err := phonenumbers.Parse(q.RawRecipientMobilePhone, "ID")
	if err != nil {
		return errors.New("recipient_mobile_phone: is invalid")
	}
	q.RecipientMobilePhone = phonenumbers.Format(mobilePhone, phonenumbers.E164)
	q.BuyerNote = strings.TrimSpace(q.BuyerNote)
	if q.ReferralCode != nil {
		if *q.ReferralCode = strings.TrimSpace(*q.ReferralCode); *q.ReferralCode == "" {
			q.ReferralCode = nil
		}
	}
	if q.RewardPointsDiscounted {
	} else if q.VoucherCode != nil {
		if *q.VoucherCode = strings.TrimSpace(*q.VoucherCode); *q.VoucherCode == "" {
			q.VoucherCode = nil
		}
	}
	if q.DeliveryRouteID.Compare(uuid.Nil()) == 0 {
		return errors.New("delivery_route_id: is required")
	}
	if q.DeliveryAddress = strings.TrimSpace(q.DeliveryAddress); q.DeliveryAddress == "" {
		return errors.New("delivery_address: is required")
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
			item.Subtotal = max(0, item.Price-item.DiscountAmount) * item.Quantity
		} else {
			item.Subtotal = item.Price * item.Quantity
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
	q.DiscountedDeliveryCost = q.DeliveryCost
	if q.RewardPointsDiscounted && q.RewardPointsDiscount > 0 {
		rewardPointsAmount := q.RewardPointsDiscount
		if q.DiscountedShoppingCost > 0 {
			rewardPointsDiscount := min(q.DiscountedShoppingCost, rewardPointsAmount)
			q.DiscountedShoppingCost -= rewardPointsDiscount
			rewardPointsAmount -= rewardPointsDiscount
		}
		if q.DiscountedDeliveryCost > 0 && rewardPointsAmount > 0 {
			rewardPointsDiscount := min(q.DiscountedDeliveryCost, rewardPointsAmount)
			q.DiscountedDeliveryCost -= rewardPointsDiscount
		}
	} else if q.VoucherCode != nil && q.VoucherType != nil {
		if *q.VoucherType == VoucherTypeShoppingCost {
			q.DiscountedShoppingCost = max(0, q.DiscountedShoppingCost-q.VoucherDiscount)
		} else {
			q.DiscountedDeliveryCost = max(0, q.DeliveryCost-q.VoucherDiscount)
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
		if q.DiscountedDeliveryCost > 0 && walletAmount > 0 {
			walletDiscount := min(q.DiscountedDeliveryCost, q.WalletAmount)
			q.DiscountedDeliveryCost -= walletDiscount
			q.WalletDiscount += walletDiscount
		}
	}
	q.PaymentAmount = q.DiscountedShoppingCost + q.DiscountedDeliveryCost
	q.RevisedDiscountedShoppingCost = q.DiscountedShoppingCost
	q.RevisedPaymentAmount = q.PaymentAmount
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
