package models

import (
	"testing"

	"github.com/google/uuid"
)

// The wallet pays the shopping cost first, then the delivery cost with what is left of the balance.
func TestCalculateDiscountWalletCoversDeliveryWithRemainingBalance(t *testing.T) {
	q := Transaction{
		ShoppingCost:          6000,
		DeliveryEstimatedCost: 8000,
		WalletDiscounted:      true,
		WalletAmount:          10000,
	}
	q.CalculateDiscount()
	if q.WalletDiscount != 10000 {
		t.Fatalf("WalletDiscount = %d, want 10000", q.WalletDiscount)
	}
	if q.DiscountedShoppingCost != 0 || q.DiscountedDeliveryEstimatedCost != 4000 {
		t.Fatalf("discounted costs = %d + %d, want 0 + 4000", q.DiscountedShoppingCost, q.DiscountedDeliveryEstimatedCost)
	}
	if q.PaymentAmount != 4000 {
		t.Fatalf("PaymentAmount = %d, want 4000", q.PaymentAmount)
	}
}

func TestCalculateDiscountWalletLargerThanTotal(t *testing.T) {
	q := Transaction{
		ShoppingCost:          6000,
		DeliveryEstimatedCost: 8000,
		WalletDiscounted:      true,
		WalletAmount:          25000,
	}
	q.CalculateDiscount()
	if q.WalletDiscount != 14000 || q.PaymentAmount != 0 {
		t.Fatalf("WalletDiscount = %d, PaymentAmount = %d, want 14000 and 0", q.WalletDiscount, q.PaymentAmount)
	}
}

func TestReviseWalletCoversDeliveryWithRemainingBalance(t *testing.T) {
	q := Transaction{
		UniqueCode:            6000,
		DeliveryEstimatedCost: 8000,
		WalletDiscounted:      true,
		WalletAmount:          10000,
		PaymentAmount:         4000,
	}
	q.Revise()
	if q.WalletDiscount != 10000 {
		t.Fatalf("WalletDiscount = %d, want 10000", q.WalletDiscount)
	}
	if q.RevisedPaymentAmount != 4000 {
		t.Fatalf("RevisedPaymentAmount = %d, want 4000", q.RevisedPaymentAmount)
	}
}

func TestUniqueCodeRangeKeepsTheChannelLimits(t *testing.T) {
	cases := []struct {
		amount, minimum, maximum, from, to int64
		ok                                 bool
	}{
		{amount: 50000, minimum: 10000, maximum: 1000000, from: 0, to: 999, ok: true},
		{amount: 999500, minimum: 10000, maximum: 1000000, from: 0, to: 500, ok: true},
		{amount: 9500, minimum: 10000, maximum: 1000000, from: 500, to: 999, ok: true},
		{amount: 50000, minimum: 0, maximum: 0, from: 0, to: 999, ok: true},
		{amount: 1000001, minimum: 10000, maximum: 1000000, ok: false},
		{amount: 8000, minimum: 10000, maximum: 1000000, ok: false},
	}
	for _, c := range cases {
		from, to, ok := UniqueCodeRange(c.amount, c.minimum, c.maximum)
		if ok != c.ok || (ok && (from != c.from || to != c.to)) {
			t.Fatalf("UniqueCodeRange(%d, %d, %d) = %d, %d, %v; want %d, %d, %v",
				c.amount, c.minimum, c.maximum, from, to, ok, c.from, c.to, c.ok)
		}
	}
}

// The unique code is added after the discounts, so the buyer always pays it.
func TestApplyUniqueCodeIsPaidByTheBuyer(t *testing.T) {
	q := Transaction{
		ShoppingCost:          6000,
		DeliveryEstimatedCost: 8000,
		WalletDiscounted:      true,
		WalletAmount:          10000,
	}
	q.CalculateDiscount()
	q.ApplyUniqueCode(123)
	if q.UniqueCode != 123 || q.ShoppingCost != 6123 || q.PaymentAmount != 4123 || q.RevisedPaymentAmount != 4123 {
		t.Fatalf("UniqueCode = %d, ShoppingCost = %d, PaymentAmount = %d, RevisedPaymentAmount = %d; want 123, 6123, 4123, 4123",
			q.UniqueCode, q.ShoppingCost, q.PaymentAmount, q.RevisedPaymentAmount)
	}
	if q.WalletDiscount != 10000 {
		t.Fatalf("WalletDiscount = %d, want 10000", q.WalletDiscount)
	}
}

// Only the points actually used are kept as the discount (they are what is debited and refunded).
func TestCalculateDiscountKeepsOnlyTheUsedPoints(t *testing.T) {
	q := Transaction{
		ShoppingCost:           6000,
		DeliveryEstimatedCost:  8000,
		RewardPointsDiscounted: true,
		RewardPointsDiscount:   20000,
	}
	q.CalculateDiscount()
	if q.RewardPointsDiscount != 14000 || q.PaymentAmount != 0 {
		t.Fatalf("RewardPointsDiscount = %d, PaymentAmount = %d; want 14000 and 0", q.RewardPointsDiscount, q.PaymentAmount)
	}
}

// The promotion comes off the shopping cost before the voucher and the wallet; only what is used is recorded.
func TestCalculateDiscountPromotionComesOffTheShoppingCost(t *testing.T) {
	promotionID := uuid.Nil
	voucherType := VoucherTypeShoppingCost
	voucherID := uuid.Nil
	q := Transaction{
		ShoppingCost:          50000,
		DeliveryEstimatedCost: 8000,
		PromotionID:           &promotionID,
		PromotionDiscount:     10000,
		VoucherID:             &voucherID,
		VoucherType:           &voucherType,
		VoucherDiscount:       5000,
		WalletDiscounted:      true,
		WalletAmount:          20000,
	}
	q.CalculateDiscount()
	if q.PromotionDiscount != 10000 {
		t.Fatalf("PromotionDiscount = %d, want 10000", q.PromotionDiscount)
	}
	if q.DiscountedShoppingCost != 15000 || q.WalletDiscount != 20000 {
		t.Fatalf("DiscountedShoppingCost = %d, WalletDiscount = %d, want 15000 and 20000",
			q.DiscountedShoppingCost, q.WalletDiscount)
	}
	if q.PaymentAmount != 23000 {
		t.Fatalf("PaymentAmount = %d, want 23000", q.PaymentAmount)
	}
}

func TestCalculateDiscountPromotionLargerThanTheShoppingCost(t *testing.T) {
	promotionID := uuid.Nil
	q := Transaction{
		ShoppingCost:          6000,
		DeliveryEstimatedCost: 8000,
		PromotionID:           &promotionID,
		PromotionDiscount:     10000,
	}
	q.CalculateDiscount()
	if q.PromotionDiscount != 6000 || q.DiscountedShoppingCost != 0 || q.PaymentAmount != 8000 {
		t.Fatalf("PromotionDiscount = %d, DiscountedShoppingCost = %d, PaymentAmount = %d, want 6000, 0 and 8000",
			q.PromotionDiscount, q.DiscountedShoppingCost, q.PaymentAmount)
	}
}

func TestCalculateDiscountWithoutPromotionIgnoresItsDiscount(t *testing.T) {
	q := Transaction{ShoppingCost: 6000, PromotionDiscount: 1000}
	q.CalculateDiscount()
	if q.PromotionDiscount != 0 || q.PaymentAmount != 6000 {
		t.Fatalf("PromotionDiscount = %d, PaymentAmount = %d, want 0 and 6000", q.PromotionDiscount, q.PaymentAmount)
	}
}

// A revision keeps the promotion off the items, never off the unique code.
func TestRevisePromotionLeavesTheUniqueCode(t *testing.T) {
	promotionID := uuid.Nil
	q := Transaction{
		UniqueCode:        123,
		PromotionID:       &promotionID,
		PromotionDiscount: 10000,
		PaymentAmount:     20123,
		Items: []*TransactionItem{
			{SellingPrice: 5000, Quantity: 6, RevisedQuantity: 2},
		},
	}
	q.Revise()
	if q.RevisedPaymentAmount != 123 || q.RefundAmount != 20000 {
		t.Fatalf("RevisedPaymentAmount = %d, RefundAmount = %d, want 123 and 20000", q.RevisedPaymentAmount, q.RefundAmount)
	}
}
