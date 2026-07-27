package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	TransactionRepository interface {
		FindTransactions(
			ctx context.Context,
			request *models.TransactionFilter,
		) ([]*models.Transaction, int64, int64, error)
		FindTransaction(
			ctx context.Context,
			request *models.TransactionFilter,
		) (*models.Transaction, error)
		CreateTransaction(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Transaction,
			sb *strings.Builder,
		) (*models.Transaction, error)
		UpdateTransaction(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Transaction,
		) (*models.Transaction, error)
	}

	transactionRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewTransactionRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) TransactionRepository {
	return &transactionRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *transactionRepository) FindTransactions(
	ctx context.Context,
	request *models.TransactionFilter,
) ([]*models.Transaction, int64, int64, error) {
	ctxt := "TransactionRepository-FindTransactions"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM transactions %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(&total)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return nil, 0, 0, err
	}
	if total == 0 || request.ReturnCountOnly {
		return nil, total, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`id
		, parent_id
		, reordered
		, buyer_id
		, cart_id
		, reference_number
		, status
		, recipient_name
		, recipient_mobile_phone
		, buyer_note
		, unique_code
		, shopping_cost
		, referrer_id
		, referral_code
		, referral_points
		, voucher_id
		, voucher_type
		, voucher_code
		, voucher_discount
		, reward_points_discounted
		, reward_points_discount
		, wallet_discounted
		, wallet_amount
		, wallet_discount
		, discounted_shopping_cost
		, delivery_cost
		, discounted_delivery_cost
		, delivery_route_id
		, delivery_longitude
		, delivery_latitude
		, delivery_distance_km
		, delivery_address
		, delivery_instruction
		, delivery_datetime
		, payment_amount
		, payment_due_date
		, payment_channel_id
		, payment_request_id
		, payment_token_id
		, created_at
		, updated_at
		, cancelled_at
		, payment_status
		, payment_failure_code
		, payment_captured_at
		, payment_id
		, paid_at
		, withdrawal_requested_at
		, withdrawal_reason
		, withdrawn_by
		, withdrawn_at
		, processed_by
		, processed_at
		, revised_shopping_cost
		, revised_discounted_shopping_cost
		, revised_payment_amount
		, refund_amount
		, refunded_at
		, refund_id
		, shipped_by
		, shipped_at
		, completed_by
		, completed_at`,
	)
	sb.Reset()
	_, _ = sb.WriteString(query)
	sortBy := request.Sortings
	if len(sortBy) == 0 {
		sortBy = append(sortBy, models.DefaultSortBy)
	}
	_, _ = sb.WriteString(" ORDER BY")
	for i, item := range sortBy {
		if i > 0 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " %s %s", item.Field, item.Direction)
	}
	pages := int64(1)
	var offset int64
	if request.Limit > 0 {
		totalDecimal, err := decimal.New(total, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		perPageDecimal, err := decimal.New(request.Limit, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		pagesDecimal, err := totalDecimal.Quo(perPageDecimal)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuo")
			return nil, 0, 0, err
		}
		pages, _, _ = pagesDecimal.Ceil(0).Int64(0)
		offset = request.Page * request.Limit
		_, _ = fmt.Fprintf(&sb, " LIMIT %d OFFSET %d", request.Limit, offset)
	}
	rows, err := q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var (
		response []*models.Transaction
		n        int
	)
	params = []any{}
	mapTransactionIDs := map[uuid.UUID]int{}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT
			id
			, transaction_id
			, product_variant_id
			, product_id
			, product_name
			, stock_unit
			, price
			, discount_minimum_quantity
			, discount_amount
			, quantity
			, subtotal
			, buyer_note
			, revised_quantity
			, revised_subtotal
			, revision_note
			, out_of_stock
		FROM transaction_items
		WHERE transaction_id IN (`,
	)
	for rows.Next() {
		var transaction models.Transaction
		if err = rows.Scan(
			&transaction.ID,
			&transaction.ParentID,
			&transaction.Reordered,
			&transaction.BuyerID,
			&transaction.CartID,
			&transaction.ReferenceNumber,
			&transaction.Status,
			&transaction.RecipientName,
			&transaction.RecipientMobilePhone,
			&transaction.BuyerNote,
			&transaction.UniqueCode,
			&transaction.ShoppingCost,
			&transaction.ReferrerID,
			&transaction.ReferralCode,
			&transaction.ReferralPoints,
			&transaction.VoucherID,
			&transaction.VoucherType,
			&transaction.VoucherCode,
			&transaction.VoucherDiscount,
			&transaction.RewardPointsDiscounted,
			&transaction.RewardPointsDiscount,
			&transaction.WalletDiscounted,
			&transaction.WalletAmount,
			&transaction.WalletDiscount,
			&transaction.DiscountedShoppingCost,
			&transaction.DeliveryCost,
			&transaction.DiscountedDeliveryCost,
			&transaction.DeliveryRouteID,
			&transaction.DeliveryLongitude,
			&transaction.DeliveryLatitude,
			&transaction.DeliveryDistanceKM,
			&transaction.DeliveryAddress,
			&transaction.DeliveryInstruction,
			&transaction.DeliveryDatetime,
			&transaction.PaymentAmount,
			&transaction.PaymentDueDate,
			&transaction.PaymentChannelID,
			&transaction.PaymentRequestID,
			&transaction.PaymentTokenID,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.CancelledAt,
			&transaction.PaymentStatus,
			&transaction.PaymentFailureCode,
			&transaction.PaymentCapturedAt,
			&transaction.PaymentID,
			&transaction.PaidAt,
			&transaction.WithdrawalRequestedAt,
			&transaction.WithdrawalReason,
			&transaction.WithdrawnBy,
			&transaction.WithdrawnAt,
			&transaction.ProcessedBy,
			&transaction.ProcessedAt,
			&transaction.RevisedShoppingCost,
			&transaction.RevisedDiscountedShoppingCost,
			&transaction.RevisedPaymentAmount,
			&transaction.RefundAmount,
			&transaction.RefundedAt,
			&transaction.RefundID,
			&transaction.ShippedBy,
			&transaction.ShippedAt,
			&transaction.CompletedBy,
			&transaction.CompletedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		transaction.RowNo = offset
		response = append(response, &transaction)
		params = append(params, transaction.ID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, "$%d", n)
		mapTransactionIDs[transaction.ID] = n - 1
	}
	_, _ = sb.WriteString(") ORDER BY id")
	if len(params) == 0 {
		return response, total, pages, nil
	}
	rows, err = q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var i int
	for rows.Next() {
		var transactionItem models.TransactionItem
		if err = rows.Scan(
			&transactionItem.ID,
			&transactionItem.TransactionID,
			&transactionItem.ProductVariantID,
			&transactionItem.ProductID,
			&transactionItem.ProductName,
			&transactionItem.StockUnit,
			&transactionItem.Price,
			&transactionItem.DiscountMinimumQuantity,
			&transactionItem.DiscountAmount,
			&transactionItem.Quantity,
			&transactionItem.Subtotal,
			&transactionItem.BuyerNote,
			&transactionItem.RevisedQuantity,
			&transactionItem.RevisedSubtotal,
			&transactionItem.RevisionNote,
			&transactionItem.OutOfStock,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		i = mapTransactionIDs[transactionItem.TransactionID]
		response[i].Items = append(response[i].Items, &transactionItem)
	}
	return response, total, pages, nil
}

func (q *transactionRepository) FindTransaction(
	ctx context.Context,
	request *models.TransactionFilter,
) (*models.Transaction, error) {
	ctxt := "TransactionRepository-FindTransaction"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, parent_id
			, reordered
			, buyer_id
			, cart_id
			, reference_number
			, status
			, recipient_name
			, recipient_mobile_phone
			, buyer_note
			, unique_code
			, shopping_cost
			, referrer_id
			, referral_code
			, referral_points
			, voucher_id
			, voucher_type
			, voucher_code
			, voucher_discount
			, reward_points_discounted
			, reward_points_discount
			, wallet_discounted
			, wallet_amount
			, wallet_discount
			, discounted_shopping_cost
			, delivery_cost
			, discounted_delivery_cost
			, delivery_route_id
			, delivery_longitude
			, delivery_latitude
			, delivery_distance_km
			, delivery_address
			, delivery_instruction
			, delivery_datetime
			, payment_amount
			, payment_due_date
			, payment_channel_id
			, payment_request_id
			, payment_token_id
			, created_at
			, updated_at
			, cancelled_at
			, payment_status
			, payment_failure_code
			, payment_captured_at
			, payment_id
			, paid_at
			, withdrawal_requested_at
			, withdrawal_reason
			, withdrawn_by
			, withdrawn_at
			, processed_by
			, processed_at
			, revised_shopping_cost
			, revised_discounted_shopping_cost
			, revised_payment_amount
			, refund_amount
			, refunded_at
			, refund_id
			, shipped_by
			, shipped_at
			, completed_by
			, completed_at
		FROM transactions %s`,
		condition,
	)
	query := sb.String()
	var response models.Transaction
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.ParentID,
		&response.Reordered,
		&response.BuyerID,
		&response.CartID,
		&response.ReferenceNumber,
		&response.Status,
		&response.RecipientName,
		&response.RecipientMobilePhone,
		&response.BuyerNote,
		&response.UniqueCode,
		&response.ShoppingCost,
		&response.ReferrerID,
		&response.ReferralCode,
		&response.ReferralPoints,
		&response.VoucherID,
		&response.VoucherType,
		&response.VoucherCode,
		&response.VoucherDiscount,
		&response.RewardPointsDiscounted,
		&response.RewardPointsDiscount,
		&response.WalletDiscounted,
		&response.WalletAmount,
		&response.WalletDiscount,
		&response.DiscountedShoppingCost,
		&response.DeliveryCost,
		&response.DiscountedDeliveryCost,
		&response.DeliveryRouteID,
		&response.DeliveryLongitude,
		&response.DeliveryLatitude,
		&response.DeliveryDistanceKM,
		&response.DeliveryAddress,
		&response.DeliveryInstruction,
		&response.DeliveryDatetime,
		&response.PaymentAmount,
		&response.PaymentDueDate,
		&response.PaymentChannelID,
		&response.PaymentRequestID,
		&response.PaymentTokenID,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.CancelledAt,
		&response.PaymentStatus,
		&response.PaymentFailureCode,
		&response.PaymentCapturedAt,
		&response.PaymentID,
		&response.PaidAt,
		&response.WithdrawalRequestedAt,
		&response.WithdrawalReason,
		&response.WithdrawnBy,
		&response.WithdrawnAt,
		&response.ProcessedBy,
		&response.ProcessedAt,
		&response.RevisedShoppingCost,
		&response.RevisedDiscountedShoppingCost,
		&response.RevisedPaymentAmount,
		&response.RefundAmount,
		&response.RefundedAt,
		&response.RefundID,
		&response.ShippedBy,
		&response.ShippedAt,
		&response.CompletedBy,
		&response.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrTransactionNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			id
			, transaction_id
			, product_variant_id
			, product_id
			, product_name
			, stock_unit
			, price
			, discount_minimum_quantity
			, discount_amount
			, quantity
			, subtotal
			, buyer_note
			, revised_quantity
			, revised_subtotal
			, revision_note
			, out_of_stock
		FROM transaction_items
		WHERE transaction_id = $1
		ORDER BY id`,
		response.ID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var transactionItem models.TransactionItem
		if err = rows.Scan(
			&transactionItem.ID,
			&transactionItem.TransactionID,
			&transactionItem.ProductVariantID,
			&transactionItem.ProductID,
			&transactionItem.ProductName,
			&transactionItem.StockUnit,
			&transactionItem.Price,
			&transactionItem.DiscountMinimumQuantity,
			&transactionItem.DiscountAmount,
			&transactionItem.Quantity,
			&transactionItem.Subtotal,
			&transactionItem.BuyerNote,
			&transactionItem.RevisedQuantity,
			&transactionItem.RevisedSubtotal,
			&transactionItem.RevisionNote,
			&transactionItem.OutOfStock,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.Items = append(response.Items, &transactionItem)
	}
	return &response, nil
}

func (q *transactionRepository) CreateTransaction(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Transaction,
	sb *strings.Builder,
) (*models.Transaction, error) {
	ctxt := "TransactionRepository-CreateTransaction"
	var response models.Transaction
	err := tx.QueryRow(
		ctx,
		`INSERT INTO transactions (
			parent_id
			, reordered
			, buyer_id
			, cart_id
			, reference_number
			, status
			, recipient_name
			, recipient_mobile_phone
			, buyer_note
			, unique_code
			, shopping_cost
			, referrer_id
			, referral_code
			, referral_points
			, voucher_id
			, voucher_type
			, voucher_code
			, voucher_discount
			, reward_points_discounted
			, reward_points_discount
			, wallet_discounted
			, wallet_amount
			, wallet_discount
			, discounted_shopping_cost
			, delivery_cost
			, discounted_delivery_cost
			, delivery_route_id
			, delivery_longitude
			, delivery_latitude
			, delivery_distance_km
			, delivery_address
			, delivery_instruction
			, delivery_datetime
			, payment_amount
			, created_at
			, updated_at
			, paid_at
			, revised_shopping_cost
			, revised_discounted_shopping_cost
			, revised_payment_amount
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30
			, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40
		)
		RETURNING id
			, parent_id
			, reordered
			, buyer_id
			, cart_id
			, reference_number
			, status
			, recipient_name
			, recipient_mobile_phone
			, buyer_note
			, unique_code
			, shopping_cost
			, referrer_id
			, referral_code
			, referral_points
			, voucher_id
			, voucher_type
			, voucher_code
			, voucher_discount
			, reward_points_discounted
			, reward_points_discount
			, wallet_discounted
			, wallet_amount
			, wallet_discount
			, discounted_shopping_cost
			, delivery_cost
			, discounted_delivery_cost
			, delivery_route_id
			, delivery_longitude
			, delivery_latitude
			, delivery_distance_km
			, delivery_address
			, delivery_instruction
			, delivery_datetime
			, payment_amount
			, payment_due_date
			, payment_channel_id
			, payment_request_id
			, payment_token_id
			, created_at
			, updated_at
			, cancelled_at
			, payment_status
			, payment_failure_code
			, payment_captured_at
			, payment_id
			, paid_at
			, withdrawal_requested_at
			, withdrawal_reason
			, withdrawn_by
			, withdrawn_at
			, processed_by
			, processed_at
			, revised_shopping_cost
			, revised_discounted_shopping_cost
			, revised_payment_amount
			, refund_amount
			, refunded_at
			, refund_id
			, shipped_by
			, shipped_at
			, completed_by
			, completed_at`,
		request.ParentID,
		request.Reordered,
		request.BuyerID,
		request.CartID,
		request.ReferenceNumber,
		request.Status,
		request.RecipientName,
		request.RecipientMobilePhone,
		request.BuyerNote,
		request.UniqueCode,
		request.ShoppingCost,
		request.ReferrerID,
		request.ReferralCode,
		request.ReferralPoints,
		request.VoucherID,
		request.VoucherType,
		request.VoucherCode,
		request.VoucherDiscount,
		request.RewardPointsDiscounted,
		request.RewardPointsDiscount,
		request.WalletDiscounted,
		request.WalletAmount,
		request.WalletDiscount,
		request.DiscountedShoppingCost,
		request.DeliveryCost,
		request.DiscountedDeliveryCost,
		request.DeliveryRouteID,
		request.DeliveryLongitude,
		request.DeliveryLatitude,
		request.DeliveryDistanceKM,
		request.DeliveryAddress,
		request.DeliveryInstruction,
		request.DeliveryDatetime,
		request.PaymentAmount,
		request.CreatedAt,
		request.UpdatedAt,
		request.PaidAt,
		request.RevisedShoppingCost,
		request.RevisedDiscountedShoppingCost,
		request.RevisedPaymentAmount,
	).Scan(
		&response.ID,
		&response.ParentID,
		&response.Reordered,
		&response.BuyerID,
		&response.CartID,
		&response.ReferenceNumber,
		&response.Status,
		&response.RecipientName,
		&response.RecipientMobilePhone,
		&response.BuyerNote,
		&response.UniqueCode,
		&response.ShoppingCost,
		&response.ReferrerID,
		&response.ReferralCode,
		&response.ReferralPoints,
		&response.VoucherID,
		&response.VoucherType,
		&response.VoucherCode,
		&response.VoucherDiscount,
		&response.RewardPointsDiscounted,
		&response.RewardPointsDiscount,
		&response.WalletDiscounted,
		&response.WalletAmount,
		&response.WalletDiscount,
		&response.DiscountedShoppingCost,
		&response.DeliveryCost,
		&response.DiscountedDeliveryCost,
		&response.DeliveryRouteID,
		&response.DeliveryLongitude,
		&response.DeliveryLatitude,
		&response.DeliveryDistanceKM,
		&response.DeliveryAddress,
		&response.DeliveryInstruction,
		&response.DeliveryDatetime,
		&response.PaymentAmount,
		&response.PaymentDueDate,
		&response.PaymentChannelID,
		&response.PaymentRequestID,
		&response.PaymentTokenID,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.CancelledAt,
		&response.PaymentStatus,
		&response.PaymentFailureCode,
		&response.PaymentCapturedAt,
		&response.PaymentID,
		&response.PaidAt,
		&response.WithdrawalRequestedAt,
		&response.WithdrawalReason,
		&response.WithdrawnBy,
		&response.WithdrawnAt,
		&response.ProcessedBy,
		&response.ProcessedAt,
		&response.RevisedShoppingCost,
		&response.RevisedDiscountedShoppingCost,
		&response.RevisedPaymentAmount,
		&response.RefundAmount,
		&response.RefundedAt,
		&response.RefundID,
		&response.ShippedBy,
		&response.ShippedAt,
		&response.CompletedBy,
		&response.CompletedAt,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "transactions_buyer_id_cart_id_key":
				err = models.ErrUniqueTransactionCartIdViolation
			case "transactions_unique_reference_number_idx":
				err = models.ErrUniqueTransactionReferenceNumberViolation
			}
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	if len(request.Items) == 0 {
		return &response, nil
	}
	var params []any
	sb.Reset()
	_, _ = sb.WriteString(
		`INSERT INTO transaction_items (
			transaction_id
			, product_variant_id
			, product_id
			, product_name
			, stock_unit
			, price
			, discount_minimum_quantity
			, discount_amount
			, quantity
			, subtotal
			, buyer_note
			, revised_quantity
			, revised_subtotal
			, revision_note
			, out_of_stock
		) VALUES`,
	)
	var n int
	for _, requestItem := range request.Items {
		params = append(
			params,
			response.ID,
			&requestItem.ProductVariantID,
			&requestItem.ProductID,
			&requestItem.ProductName,
			&requestItem.StockUnit,
			&requestItem.Price,
			&requestItem.DiscountMinimumQuantity,
			&requestItem.DiscountAmount,
			&requestItem.Quantity,
			&requestItem.Subtotal,
			&requestItem.BuyerNote,
			&requestItem.RevisedQuantity,
			&requestItem.RevisedSubtotal,
			&requestItem.RevisionNote,
			&requestItem.OutOfStock,
		)
		if n = len(params); n > 15 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(
			sb,
			" ($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			n-14, n-13, n-12, n-11, n-10, n-9, n-8, n-7, n-6, n-5, n-4, n-3, n-2, n-1, n,
		)
	}
	_, _ = sb.WriteString(
		` RETURNING id
			, transaction_id
			, product_variant_id
			, product_id
			, product_name
			, stock_unit
			, price
			, discount_minimum_quantity
			, discount_amount
			, quantity
			, subtotal
			, buyer_note
			, revised_quantity
			, revised_subtotal
			, revision_note
			, out_of_stock`,
	)
	rows, err := tx.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var transactionItem models.TransactionItem
		if err = rows.Scan(
			&transactionItem.ID,
			&transactionItem.TransactionID,
			&transactionItem.ProductVariantID,
			&transactionItem.ProductID,
			&transactionItem.ProductName,
			&transactionItem.StockUnit,
			&transactionItem.Price,
			&transactionItem.DiscountMinimumQuantity,
			&transactionItem.DiscountAmount,
			&transactionItem.Quantity,
			&transactionItem.Subtotal,
			&transactionItem.BuyerNote,
			&transactionItem.RevisedQuantity,
			&transactionItem.RevisedSubtotal,
			&transactionItem.RevisionNote,
			&transactionItem.OutOfStock,
		); err != nil {
			if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
				pgxErr.Code == pgerrcode.UniqueViolation &&
				pgxErr.ConstraintName == "transaction_items_transaction_id_product_variant_id_key" {
				err = models.ErrUniqueTransactionProductVariantIdViolation
			} else {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			}
			return nil, err
		}
		response.Items = append(response.Items, &transactionItem)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	return &response, nil
}

func (q *transactionRepository) UpdateTransaction(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Transaction,
) (*models.Transaction, error) {
	ctxt := "TransactionRepository-UpdateTransaction"
	response := models.Transaction{
		Items: make([]*models.TransactionItem, len(request.Items)),
	}
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, parent_id
			, reordered
			, buyer_id
			, cart_id
			, reference_number
			, status
			, recipient_name
			, recipient_mobile_phone
			, buyer_note
			, unique_code
			, shopping_cost
			, referrer_id
			, referral_code
			, referral_points
			, voucher_id
			, voucher_type
			, voucher_code
			, voucher_discount
			, reward_points_discounted
			, reward_points_discount
			, wallet_discounted
			, wallet_amount
			, wallet_discount
			, discounted_shopping_cost
			, delivery_cost
			, discounted_delivery_cost
			, delivery_route_id
			, delivery_longitude
			, delivery_latitude
			, delivery_distance_km
			, delivery_address
			, delivery_instruction
			, delivery_datetime
			, payment_amount
			, payment_due_date
			, payment_channel_id
			, payment_request_id
			, payment_token_id
			, created_at
			, updated_at
			, cancelled_at
			, payment_status
			, payment_failure_code
			, payment_captured_at
			, payment_id
			, paid_at
			, withdrawal_requested_at
			, withdrawal_reason
			, withdrawn_by
			, withdrawn_at
			, processed_by
			, processed_at
			, revised_shopping_cost
			, revised_discounted_shopping_cost
			, revised_payment_amount
			, refund_amount
			, refunded_at
			, refund_id
			, shipped_by
			, shipped_at
			, completed_by
			, completed_at
		FROM transactions
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.ParentID,
		&response.Reordered,
		&response.BuyerID,
		&response.CartID,
		&response.ReferenceNumber,
		&response.Status,
		&response.RecipientName,
		&response.RecipientMobilePhone,
		&response.BuyerNote,
		&response.UniqueCode,
		&response.ShoppingCost,
		&response.ReferrerID,
		&response.ReferralCode,
		&response.ReferralPoints,
		&response.VoucherID,
		&response.VoucherType,
		&response.VoucherCode,
		&response.VoucherDiscount,
		&response.RewardPointsDiscounted,
		&response.RewardPointsDiscount,
		&response.WalletDiscounted,
		&response.WalletAmount,
		&response.WalletDiscount,
		&response.DiscountedShoppingCost,
		&response.DeliveryCost,
		&response.DiscountedDeliveryCost,
		&response.DeliveryRouteID,
		&response.DeliveryLongitude,
		&response.DeliveryLatitude,
		&response.DeliveryDistanceKM,
		&response.DeliveryAddress,
		&response.DeliveryInstruction,
		&response.DeliveryDatetime,
		&response.PaymentAmount,
		&response.PaymentDueDate,
		&response.PaymentChannelID,
		&response.PaymentRequestID,
		&response.PaymentTokenID,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.CancelledAt,
		&response.PaymentStatus,
		&response.PaymentFailureCode,
		&response.PaymentCapturedAt,
		&response.PaymentID,
		&response.PaidAt,
		&response.WithdrawalRequestedAt,
		&response.WithdrawalReason,
		&response.WithdrawnBy,
		&response.WithdrawnAt,
		&response.ProcessedBy,
		&response.ProcessedAt,
		&response.RevisedShoppingCost,
		&response.RevisedDiscountedShoppingCost,
		&response.RevisedPaymentAmount,
		&response.RefundAmount,
		&response.RefundedAt,
		&response.RefundID,
		&response.ShippedBy,
		&response.ShippedAt,
		&response.CompletedBy,
		&response.CompletedAt,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	_, _ = sb.WriteString("UPDATE transactions SET")
	if response.Status < request.Status {
		params = append(params, request.Status)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " status = $%d", n)
	}
	if response.PaymentDueDate == nil && request.PaymentDueDate != nil {
		params = append(params, request.PaymentDueDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_due_date = $%d", n)
	}
	if response.PaymentStatus == nil && request.PaymentStatus != nil {
		params = append(params, request.PaymentStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_status = $%d", n)
	}
	if response.PaymentFailureCode == nil && request.PaymentFailureCode != nil {
		params = append(params, request.PaymentFailureCode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_failure_code = $%d", n)
	}
	if response.PaymentCapturedAt == nil && request.PaymentCapturedAt != nil {
		params = append(params, request.PaymentCapturedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_captured_at = $%d", n)
	}
	if response.PaymentChannelID == nil && request.PaymentChannelID != nil {
		params = append(params, request.PaymentChannelID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_channel_id = $%d", n)
	}
	if response.PaymentRequestID == nil && request.PaymentRequestID != nil {
		params = append(params, request.PaymentRequestID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_request_id = $%d", n)
	}
	if response.PaymentTokenID == nil && request.PaymentTokenID != nil {
		params = append(params, request.PaymentTokenID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_token_id = $%d", n)
	}
	if response.CancelledAt == nil && request.CancelledAt != nil {
		params = append(params, request.CancelledAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " cancelled_at = $%d", n)
	}
	if response.PaymentID == nil && request.PaymentID != nil {
		params = append(params, request.PaymentID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " payment_id = $%d", n)
	}
	if response.PaidAt == nil && request.PaidAt != nil {
		params = append(params, request.PaidAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " paid_at = $%d", n)
	}
	if response.WithdrawalRequestedAt == nil && request.WithdrawalRequestedAt != nil {
		params = append(params, request.WithdrawalRequestedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " withdrawal_requested_at = $%d", n)
	}
	if response.WithdrawalReason == nil && request.WithdrawalReason != nil {
		params = append(params, request.WithdrawalReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " withdrawal_reason = $%d", n)
	}
	if response.WithdrawnBy == nil && request.WithdrawnBy != nil {
		params = append(params, request.WithdrawnBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " withdrawn_by = $%d", n)
	}
	if response.WithdrawnAt == nil && request.WithdrawnAt != nil {
		params = append(params, request.WithdrawnAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " withdrawn_at = $%d", n)
	}
	if response.ProcessedBy == nil && request.ProcessedBy != nil {
		params = append(params, request.ProcessedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " processed_by = $%d", n)
	}
	if response.ProcessedAt == nil && request.ProcessedAt != nil {
		params = append(params, request.ProcessedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " processed_at = $%d", n)
	}
	if response.RevisedShoppingCost != request.RevisedShoppingCost {
		params = append(params, request.RevisedShoppingCost)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " revised_shopping_cost = $%d", n)
	}
	if response.RevisedDiscountedShoppingCost != request.RevisedDiscountedShoppingCost {
		params = append(params, request.RevisedDiscountedShoppingCost)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " revised_discounted_shopping_cost = $%d", n)
	}
	if response.RevisedPaymentAmount != request.RevisedPaymentAmount {
		params = append(params, request.RevisedPaymentAmount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " revised_payment_amount = $%d", n)
	}
	if response.RefundAmount == nil && request.RefundAmount != nil {
		params = append(params, request.RefundAmount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " refund_amount = $%d", n)
	}
	if response.RefundedAt == nil && request.RefundedAt != nil {
		params = append(params, request.RefundedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " refunded_at = $%d", n)
	}
	if response.RefundID == nil && request.RefundID != nil {
		params = append(params, request.RefundID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " refund_id = $%d", n)
	}
	if response.ShippedBy == nil && request.ShippedBy != nil {
		params = append(params, request.ShippedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " shipped_by = $%d", n)
	}
	if response.ShippedAt == nil && request.ShippedAt != nil {
		params = append(params, request.ShippedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " shipped_at = $%d", n)
	}
	if response.CompletedBy == nil && request.CompletedBy != nil {
		params = append(params, request.CompletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " completed_by = $%d", n)
	}
	if response.CompletedAt == nil && request.CompletedAt != nil {
		params = append(params, request.CompletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " completed_at = $%d", n)
	}
	if response.CompletedAt == nil && request.CompletedAt != nil {
		params = append(params, request.CompletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " completed_at = $%d", n)
	}
	if len(params) > 0 {
		params = append(params, request.UpdatedAt, request.ID)
		n = len(params)
		_, _ = fmt.Fprintf(
			&sb,
			`, updated_at = $%d
			WHERE id = $%d
			RETURNING id
				, parent_id
				, reordered
				, buyer_id
				, cart_id
				, reference_number
				, status
				, recipient_name
				, recipient_mobile_phone
				, buyer_note
				, unique_code
				, shopping_cost
				, referrer_id
				, referral_code
				, referral_points
				, voucher_id
				, voucher_type
				, voucher_code
				, voucher_discount
				, reward_points_discounted
				, reward_points_discount
				, wallet_discounted
				, wallet_amount
				, wallet_discount
				, discounted_shopping_cost
				, delivery_cost
				, discounted_delivery_cost
				, delivery_route_id
				, delivery_longitude
				, delivery_latitude
				, delivery_distance_km
				, delivery_address
				, delivery_instruction
				, delivery_datetime
				, payment_amount
				, payment_due_date
				, payment_channel_id
				, payment_request_id
				, payment_token_id
				, created_at
				, updated_at
				, cancelled_at
				, payment_status
				, payment_failure_code
				, payment_captured_at
				, payment_id
				, paid_at
				, withdrawal_requested_at
				, withdrawal_reason
				, withdrawn_by
				, withdrawn_at
				, processed_by
				, processed_at
				, revised_shopping_cost
				, revised_discounted_shopping_cost
				, revised_payment_amount
				, refund_amount
				, refunded_at
				, refund_id
				, shipped_by
				, shipped_at
				, completed_by
				, completed_at`,
			n-1, n,
		)
		if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
			&response.ID,
			&response.ParentID,
			&response.Reordered,
			&response.BuyerID,
			&response.CartID,
			&response.ReferenceNumber,
			&response.Status,
			&response.RecipientName,
			&response.RecipientMobilePhone,
			&response.BuyerNote,
			&response.UniqueCode,
			&response.ShoppingCost,
			&response.ReferrerID,
			&response.ReferralCode,
			&response.ReferralPoints,
			&response.VoucherID,
			&response.VoucherType,
			&response.VoucherCode,
			&response.VoucherDiscount,
			&response.RewardPointsDiscounted,
			&response.RewardPointsDiscount,
			&response.WalletDiscounted,
			&response.WalletAmount,
			&response.WalletDiscount,
			&response.DiscountedShoppingCost,
			&response.DeliveryCost,
			&response.DiscountedDeliveryCost,
			&response.DeliveryRouteID,
			&response.DeliveryLongitude,
			&response.DeliveryLatitude,
			&response.DeliveryDistanceKM,
			&response.DeliveryAddress,
			&response.DeliveryInstruction,
			&response.DeliveryDatetime,
			&response.PaymentAmount,
			&response.PaymentDueDate,
			&response.PaymentChannelID,
			&response.PaymentRequestID,
			&response.PaymentTokenID,
			&response.CreatedAt,
			&response.UpdatedAt,
			&response.CancelledAt,
			&response.PaymentStatus,
			&response.PaymentFailureCode,
			&response.PaymentCapturedAt,
			&response.PaymentID,
			&response.PaidAt,
			&response.WithdrawalRequestedAt,
			&response.WithdrawalReason,
			&response.WithdrawnBy,
			&response.WithdrawnAt,
			&response.ProcessedBy,
			&response.ProcessedAt,
			&response.RevisedShoppingCost,
			&response.RevisedDiscountedShoppingCost,
			&response.RevisedPaymentAmount,
			&response.RefundAmount,
			&response.RefundedAt,
			&response.RefundID,
			&response.ShippedBy,
			&response.ShippedAt,
			&response.CompletedBy,
			&response.CompletedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
	}
	for i, itemRequest := range request.Items {
		var itemResponse models.TransactionItem
		if err = tx.QueryRow(
			ctx,
			`SELECT
				id
				, transaction_id
				, product_variant_id
				, product_id
				, product_name
				, stock_unit
				, price
				, discount_minimum_quantity
				, discount_amount
				, quantity
				, subtotal
				, buyer_note
				, revised_quantity
				, revised_subtotal
				, revision_note
				, out_of_stock
			FROM transaction_items
			WHERE transaction_id = $1
				AND id = $2
			FOR NO KEY UPDATE`,
			request.ID,
			itemRequest.ID,
		).Scan(
			&itemResponse.ID,
			&itemResponse.TransactionID,
			&itemResponse.ProductVariantID,
			&itemResponse.ProductID,
			&itemResponse.ProductName,
			&itemResponse.StockUnit,
			&itemResponse.Price,
			&itemResponse.DiscountMinimumQuantity,
			&itemResponse.DiscountAmount,
			&itemResponse.Quantity,
			&itemResponse.Subtotal,
			&itemResponse.BuyerNote,
			&itemResponse.RevisedQuantity,
			&itemResponse.RevisedSubtotal,
			&itemResponse.RevisionNote,
			&itemResponse.OutOfStock,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		params = params[:0]
		sb.Reset()
		_, _ = sb.WriteString("UPDATE transaction_items SET")
		if itemResponse.RevisedQuantity != itemRequest.RevisedQuantity {
			params = append(params, itemRequest.RevisedQuantity)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " revised_quantity = $%d", n)
		}
		if itemResponse.RevisedSubtotal != itemRequest.RevisedSubtotal {
			params = append(params, itemRequest.RevisedSubtotal)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " revised_subtotal = $%d", n)
		}
		if itemResponse.RevisionNote != itemRequest.RevisionNote {
			params = append(params, itemRequest.RevisionNote)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " revision_note = $%d", n)
		}
		if itemResponse.OutOfStock != itemRequest.OutOfStock {
			params = append(params, itemRequest.OutOfStock)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sb, " out_of_stock = $%d", n)
		}
		if len(params) == 0 {
			response.Items[i] = &itemResponse
			continue
		}
		params = append(params, itemRequest.ID)
		_, _ = fmt.Fprintf(
			&sb,
			` WHERE id = $%d
			RETURNING id
				, transaction_id
				, cart_item_id
				, product_variant_id
				, product_id
				, product_name
				, stock_unit
				, price
				, discount_minimum_quantity
				, discount_amount
				, quantity
				, subtotal
				, buyer_note
				, revised_quantity
				, revised_subtotal
				, revision_note
				, out_of_stock`,
			len(params),
		)
		if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
			&itemResponse.ID,
			&itemResponse.TransactionID,
			&itemResponse.CartItemID,
			&itemResponse.ProductVariantID,
			&itemResponse.ProductID,
			&itemResponse.ProductName,
			&itemResponse.StockUnit,
			&itemResponse.Price,
			&itemResponse.DiscountMinimumQuantity,
			&itemResponse.DiscountAmount,
			&itemResponse.Quantity,
			&itemResponse.Subtotal,
			&itemResponse.BuyerNote,
			&itemResponse.RevisedQuantity,
			&itemResponse.RevisedSubtotal,
			&itemResponse.RevisionNote,
			&itemResponse.OutOfStock,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.Items[i] = &itemResponse
	}
	return &response, nil
}

func (q *transactionRepository) buildCondition(request *models.TransactionFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.TransactionIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, transactionID := range request.TransactionIDs {
			params = append(params, transactionID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedTransactionIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedTransactionID := range request.ExcludedTransactionIDs {
			params = append(params, excludedTransactionID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.BuyerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("buyer_id IN (")
		for i, buyerID := range request.BuyerIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.CartIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("cart_id IN (")
		for i, cartID := range request.CartIDs {
			params = append(params, cartID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.VoucherIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("voucher_id IN (")
		for i, voucherID := range request.VoucherIDs {
			params = append(params, voucherID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.PaymentRequestIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("payment_request_id IN (")
		for i, paymentRequestID := range request.PaymentRequestIDs {
			params = append(params, paymentRequestID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.PaymentTokenIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("payment_token_id IN (")
		for i, paymentTokenID := range request.PaymentTokenIDs {
			params = append(params, paymentTokenID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.PaymentIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("payment_id IN (")
		for i, paymentID := range request.PaymentIDs {
			params = append(params, paymentID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if !request.PaymentDueDate.IsZero() {
		params = append(params, request.PaymentDueDate)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "payment_due_date < $%d", len(params))
		conds = append(conds, sb.String())
	}
	if len(request.TransactionStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("status IN (")
		for i, transactionStatus := range request.TransactionStatuses {
			params = append(params, transactionStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedTransactionStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("status NOT IN (")
		for i, excludedTransactionStatus := range request.ExcludedTransactionStatuses {
			params = append(params, excludedTransactionStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ReferrerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("referrer_id IN (")
		for i, referrerID := range request.ReferrerIDs {
			params = append(params, referrerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		sb.Reset()
		_, _ = fmt.Fprintf(sb, `LOWER(reference_number) LIKE $%d`, len(params))
		conds = append(conds, sb.String())
	}
	if len(conds) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = fmt.Fprintf(sb, " %s", cond)
		}
		condition = sb.String()
	}
	return
}
