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
	return new(transactionRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
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
	if total == 0 {
		return nil, 0, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`id
		, parent_id
		, buyer_id
		, idempotency_key
		, reference_number
		, transaction_status
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
		, payment_amount
		, payment_deadline
		, delivery_cost
		, discounted_delivery_cost
		, delivery_coordinates[0]
		, delivery_coordinates[1]
		, delivery_address
		, delivery_instruction
		, delivery_plan_id
		, delivery_speed
		, delivery_datetime
		, created_at
		, paid_at
		, payment_request_id
		, payment_id
		, payment_status
		, processed_by
		, processed_at
		, revised_payment_amount
		, refund_amount
		, refunded_at
		, refund_id
		, cancelled_by
		, cancelled_at
		, cancellation_reason
		, shipped_by
		, shipped_at
		, delivered_by
		, delivered_at`,
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
			, shopping_cost
			, buyer_note
			, out_of_stock
		FROM transaction_items
		WHERE transaction_id IN (`,
	)
	for rows.Next() {
		var transaction models.Transaction
		if err = rows.Scan(
			&transaction.ID,
			&transaction.ParentID,
			&transaction.BuyerID,
			&transaction.IdempotencyKey,
			&transaction.ReferenceNumber,
			&transaction.TransactionStatus,
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
			&transaction.PaymentAmount,
			&transaction.PaymentDeadline,
			&transaction.DeliveryCost,
			&transaction.DiscountedDeliveryCost,
			&transaction.DeliveryLongitude,
			&transaction.DeliveryLatitude,
			&transaction.DeliveryAddress,
			&transaction.DeliveryInstruction,
			&transaction.DeliveryPlanID,
			&transaction.DeliverySpeed,
			&transaction.DeliveryDatetime,
			&transaction.CreatedAt,
			&transaction.PaidAt,
			&transaction.PaymentRequestID,
			&transaction.PaymentID,
			&transaction.PaymentStatus,
			&transaction.ProcessedBy,
			&transaction.ProcessedAt,
			&transaction.RevisedPaymentAmount,
			&transaction.RefundAmount,
			&transaction.RefundedAt,
			&transaction.RefundID,
			&transaction.CancelledBy,
			&transaction.CancelledAt,
			&transaction.CancellationReason,
			&transaction.ShippedBy,
			&transaction.ShippedAt,
			&transaction.DeliveredBy,
			&transaction.DeliveredAt,
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
			, buyer_id
			, idempotency_key
			, reference_number
			, transaction_status
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
			, payment_amount
			, payment_deadline
			, delivery_cost
			, discounted_delivery_cost
			, delivery_coordinates[0]
			, delivery_coordinates[1]
			, delivery_address
			, delivery_instruction
			, delivery_plan_id
			, delivery_speed
			, delivery_datetime
			, created_at
			, paid_at
			, payment_request_id
			, payment_id
			, payment_status
			, processed_by
			, processed_at
			, revised_payment_amount
			, refund_amount
			, refunded_at
			, refund_id
			, cancelled_by
			, cancelled_at
			, cancellation_reason
			, shipped_by
			, shipped_at
			, delivered_by
			, delivered_at
		FROM transactions %s`,
		condition,
	)
	query := sb.String()
	var response models.Transaction
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.ParentID,
		&response.BuyerID,
		&response.IdempotencyKey,
		&response.ReferenceNumber,
		&response.TransactionStatus,
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
		&response.PaymentAmount,
		&response.PaymentDeadline,
		&response.DeliveryCost,
		&response.DiscountedDeliveryCost,
		&response.DeliveryLongitude,
		&response.DeliveryLatitude,
		&response.DeliveryAddress,
		&response.DeliveryInstruction,
		&response.DeliveryPlanID,
		&response.DeliverySpeed,
		&response.DeliveryDatetime,
		&response.CreatedAt,
		&response.PaidAt,
		&response.PaymentRequestID,
		&response.PaymentID,
		&response.PaymentStatus,
		&response.ProcessedBy,
		&response.ProcessedAt,
		&response.RevisedPaymentAmount,
		&response.RefundAmount,
		&response.RefundedAt,
		&response.RefundID,
		&response.CancelledBy,
		&response.CancelledAt,
		&response.CancellationReason,
		&response.ShippedBy,
		&response.ShippedAt,
		&response.DeliveredBy,
		&response.DeliveredAt,
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
			, shopping_cost
			, buyer_note
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
) (*models.Transaction, error) {
	ctxt := "TransactionRepository-CreateTransaction"
	var (
		sb    strings.Builder
		count int
	)
	for {
		sb.Reset()
		_, _ = fmt.Fprintf(
			&sb,
			models.TransactionReferenceNumberFormat,
			helper.RandomString(12),
		)
		request.ReferenceNumber = sb.String()
		if err := tx.QueryRow(
			ctx,
			"SELECT COUNT(1) FROM transactions WHERE LOWER(reference_number) = $1",
			strings.ToLower(request.ReferenceNumber),
		).Scan(&count); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		if count == 0 {
			break
		}
	}
	var response models.Transaction
	err := tx.QueryRow(
		ctx,
		`INSERT INTO transactions (
			parent_id
			, buyer_id
			, idempotency_key
			, reference_number
			, transaction_status
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
			, payment_amount
			, payment_deadline
			, delivery_cost
			, discounted_delivery_cost
			, delivery_coordinates
			, delivery_address
			, delivery_instruction
			, delivery_plan_id
			, delivery_speed
			, delivery_datetime
			, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			, $21, $22, $23, $24, $25, $26, $27, POINT($28, $29), $30
			, $31, $32, $34, $35
		)
		RETURNING id
			, parent_id
			, buyer_id
			, idempotency_key
			, reference_number
			, transaction_status
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
			, payment_amount
			, payment_deadline
			, delivery_cost
			, discounted_delivery_cost
			, delivery_coordinates[0]
			, delivery_coordinates[1]
			, delivery_address
			, delivery_instruction
			, delivery_plan_id
			, delivery_speed
			, delivery_datetime
			, created_at
			, paid_at
			, payment_request_id
			, payment_id
			, payment_status
			, processed_by
			, processed_at
			, revised_payment_amount
			, refund_amount
			, refunded_at
			, refund_id
			, cancelled_by
			, cancelled_at
			, cancellation_reason
			, shipped_by
			, shipped_at
			, delivered_by
			, delivered_at`,
		request.ParentID,
		request.BuyerID,
		request.IdempotencyKey,
		request.ReferenceNumber,
		request.TransactionStatus,
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
		request.PaymentAmount,
		request.PaymentDeadline,
		request.DeliveryCost,
		request.DiscountedDeliveryCost,
		request.DeliveryLongitude,
		request.DeliveryLatitude,
		request.DeliveryAddress,
		request.DeliveryInstruction,
		request.DeliveryPlanID,
		request.DeliverySpeed,
		request.DeliveryDatetime,
		request.CreatedAt,
	).Scan(
		&response.ID,
		&response.ParentID,
		&response.BuyerID,
		&response.IdempotencyKey,
		&response.ReferenceNumber,
		&response.TransactionStatus,
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
		&response.PaymentAmount,
		&response.PaymentDeadline,
		&response.DeliveryCost,
		&response.DiscountedDeliveryCost,
		&response.DeliveryLongitude,
		&response.DeliveryLatitude,
		&response.DeliveryAddress,
		&response.DeliveryInstruction,
		&response.DeliveryPlanID,
		&response.DeliverySpeed,
		&response.DeliveryDatetime,
		&response.CreatedAt,
		&response.PaidAt,
		&response.PaymentRequestID,
		&response.PaymentID,
		&response.PaymentStatus,
		&response.ProcessedBy,
		&response.ProcessedAt,
		&response.RevisedPaymentAmount,
		&response.RefundAmount,
		&response.RefundedAt,
		&response.RefundID,
		&response.CancelledBy,
		&response.CancelledAt,
		&response.CancellationReason,
		&response.ShippedBy,
		&response.ShippedAt,
		&response.DeliveredBy,
		&response.DeliveredAt,
	)
	if err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation {
			switch pgxErr.ConstraintName {
			case "transactions_buyer_id_idempotency_key_key":
				err = models.ErrUniqueTransactionIdempotencyKeyViolation
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
			, shopping_cost
			, buyer_note
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
			&requestItem.OutOfStock,
		)
		if n = len(params); n > 12 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(
			&sb,
			" ($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			n-11, n-10, n-9, n-8, n-7, n-6, n-5, n-4, n-3, n-2, n-1, n,
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
			, shopping_cost
			, buyer_note
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
		`UPDATE transactions SET
			transaction_status = $1
			, paid_at = $2
			, payment_request_id = $3
			, payment_id = $4
			, payment_status = $5
			, processed_by = $6
			, processed_at = $7
			, revised_payment_amount = $8
			, refund_amount = $9
			, refunded_at = $10
			, refund_id = $11
			, cancelled_by = $12
			, cancelled_at = $13
			, cancellation_reason = $14
			, shipped_by = $15
			, shipped_at = $16
			, delivered_by = $17
			, delivered_at = $18
		WHERE id = $19
		RETURNING id
			, parent_id
			, buyer_id
			, idempotency_key
			, reference_number
			, transaction_status
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
			, payment_amount
			, payment_deadline
			, delivery_cost
			, discounted_delivery_cost
			, delivery_coordinates[0]
			, delivery_coordinates[1]
			, delivery_address
			, delivery_instruction
			, delivery_plan_id
			, delivery_speed
			, delivery_datetime
			, created_at
			, paid_at
			, payment_request_id
			, payment_id
			, payment_status
			, processed_by
			, processed_at
			, revised_payment_amount
			, refund_amount
			, refunded_at
			, refund_id
			, cancelled_by
			, cancelled_at
			, cancellation_reason
			, shipped_by
			, shipped_at
			, delivered_by
			, delivered_at`,
		request.TransactionStatus,
		request.PaidAt,
		request.PaymentRequestID,
		request.PaymentID,
		request.PaymentStatus,
		request.ProcessedBy,
		request.ProcessedAt,
		request.RevisedPaymentAmount,
		request.RefundAmount,
		request.RefundedAt,
		request.RefundID,
		request.CancelledBy,
		request.CancelledAt,
		request.CancellationReason,
		request.ShippedBy,
		request.ShippedAt,
		request.DeliveredBy,
		request.DeliveredAt,
		request.ID,
	).Scan(
		&response.ID,
		&response.ParentID,
		&response.BuyerID,
		&response.IdempotencyKey,
		&response.ReferenceNumber,
		&response.TransactionStatus,
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
		&response.PaymentAmount,
		&response.PaymentDeadline,
		&response.DeliveryCost,
		&response.DiscountedDeliveryCost,
		&response.DeliveryLongitude,
		&response.DeliveryLatitude,
		&response.DeliveryAddress,
		&response.DeliveryInstruction,
		&response.DeliveryPlanID,
		&response.DeliverySpeed,
		&response.DeliveryDatetime,
		&response.CreatedAt,
		&response.PaidAt,
		&response.PaymentRequestID,
		&response.PaymentID,
		&response.PaymentStatus,
		&response.ProcessedBy,
		&response.ProcessedAt,
		&response.RevisedPaymentAmount,
		&response.RefundAmount,
		&response.RefundedAt,
		&response.RefundID,
		&response.CancelledBy,
		&response.CancelledAt,
		&response.CancellationReason,
		&response.ShippedBy,
		&response.ShippedAt,
		&response.DeliveredBy,
		&response.DeliveredAt,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	for i, requestItem := range request.Items {
		var responseItem models.TransactionItem
		if err = tx.QueryRow(
			ctx,
			`UPDATE transaction_items SET
				out_of_stock = $1
			WHERE transaction_id = $2
				AND id = $3
			RETURNING id
				, transaction_id
				, product_variant_id
				, product_id
				, product_name
				, stock_unit
				, price
				, discount_minimum_quantity
				, discount_amount
				, quantity
				, shopping_cost
				, buyer_note
				, out_of_stock`,
			requestItem.OutOfStock,
			response.ID,
			requestItem.ID,
		).Scan(
			&responseItem.ID,
			&responseItem.TransactionID,
			&responseItem.ProductVariantID,
			&responseItem.ProductID,
			&responseItem.ProductName,
			&responseItem.StockUnit,
			&responseItem.Price,
			&responseItem.DiscountMinimumQuantity,
			&responseItem.DiscountAmount,
			&responseItem.Quantity,
			&responseItem.Subtotal,
			&responseItem.BuyerNote,
			&responseItem.OutOfStock,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.Items[i] = &responseItem
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
	if len(request.IdempotencyKeys) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("idempotency_key IN (")
		for i, idempotencyKey := range request.IdempotencyKeys {
			params = append(params, idempotencyKey)
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
	if len(request.TransactionStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("transaction_status IN (")
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
	if len(request.ReferralCodes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("referral_code IN (")
		for i, referralCode := range request.ReferralCodes {
			params = append(params, referralCode)
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
