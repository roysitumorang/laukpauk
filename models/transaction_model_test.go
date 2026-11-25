package models

import "testing"

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
