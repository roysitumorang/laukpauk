package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"github.com/dustin/go-humanize"
	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/consumers"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	TransactionService interface {
		FindTransactions(
			ctx context.Context,
			request *models.TransactionFilter,
		) ([]*models.Transaction, *models.Pagination, error)
		FindTransaction(
			ctx context.Context,
			request *models.TransactionFilter,
		) (*models.Transaction, error)
		CreateTransaction(
			ctx context.Context,
			buyer *models.Buyer,
			request *models.Transaction,
		) (*models.Transaction, error)
		UpdateTransaction(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Transaction,
		) (*models.Transaction, error)
		CancelOverdueTransactions(ctx context.Context) error
		WithdrawTransaction(
			ctx context.Context,
			request *models.Transaction,
		) (*models.Transaction, error)
	}

	transactionService struct {
		transactionRepository        repositories.TransactionRepository
		buyerRepository              repositories.BuyerRepository
		transactionCounterRepository repositories.TransactionCounterRepository
		cartRepository               repositories.CartRepository
		deliveryPlanRepository       repositories.DeliveryPlanRepository
		deliveryRouteRepository      repositories.DeliveryRouteRepository
		settingRepository            repositories.SettingRepository
		voucherRepository            repositories.VoucherRepository
		pointSpendingCapRepository   repositories.PointSpendingCapRepository
		pointRepository              repositories.PointRepository
		productVariantRepository     repositories.ProductVariantRepository
		paymentChannelRepository     repositories.PaymentChannelRepository
		xenditConsumer               *consumers.XenditConsumer
	}
)

func NewTransactionService(
	transactionRepository repositories.TransactionRepository,
	buyerRepository repositories.BuyerRepository,
	transactionCounterRepository repositories.TransactionCounterRepository,
	cartRepository repositories.CartRepository,
	deliveryPlanRepository repositories.DeliveryPlanRepository,
	deliveryRouteRepository repositories.DeliveryRouteRepository,
	settingRepository repositories.SettingRepository,
	voucherRepository repositories.VoucherRepository,
	pointSpendingCapRepository repositories.PointSpendingCapRepository,
	pointRepository repositories.PointRepository,
	productVariantRepository repositories.ProductVariantRepository,
	paymentChannelRepository repositories.PaymentChannelRepository,
	xenditConsumer *consumers.XenditConsumer,
) TransactionService {
	return &transactionService{
		transactionRepository:        transactionRepository,
		buyerRepository:              buyerRepository,
		transactionCounterRepository: transactionCounterRepository,
		cartRepository:               cartRepository,
		deliveryPlanRepository:       deliveryPlanRepository,
		deliveryRouteRepository:      deliveryRouteRepository,
		settingRepository:            settingRepository,
		voucherRepository:            voucherRepository,
		pointSpendingCapRepository:   pointSpendingCapRepository,
		pointRepository:              pointRepository,
		productVariantRepository:     productVariantRepository,
		paymentChannelRepository:     paymentChannelRepository,
		xenditConsumer:               xenditConsumer,
	}
}

func (q *transactionService) FindTransactions(
	ctx context.Context,
	request *models.TransactionFilter,
) ([]*models.Transaction, *models.Pagination, error) {
	ctxt := "TransactionService-FindTransactions"
	transactions, total, pages, err := q.transactionRepository.FindTransactions(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return nil, nil, err
	}
	n := len(transactions)
	rows := make([]*models.Transaction, n)
	if n > 0 {
		copy(rows, transactions)
	}
	pagination, err := models.NewPagination(
		total,
		pages,
		request.Limit,
		request.Page,
		request.PaginationURL,
		request.UrlValues,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewPagination")
		return nil, nil, err
	}
	return rows, pagination, nil
}

func (q *transactionService) FindTransaction(
	ctx context.Context,
	request *models.TransactionFilter,
) (*models.Transaction, error) {
	ctxt := "TransactionService-FindTransaction"
	response, err := q.transactionRepository.FindTransaction(ctx, request)
	if err != nil && !errors.Is(err, models.ErrTransactionNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
	}
	return response, err
}

func (q *transactionService) CreateTransaction(
	ctx context.Context,
	buyer *models.Buyer,
	request *models.Transaction,
) (*models.Transaction, error) {
	ctxt := "TransactionService-CreateTransaction"
	cart, err := q.cartRepository.FindCart(
		ctx,
		models.NewCartFilter().
			WithBuyerIDs(request.BuyerID),
	)
	if err != nil && !errors.Is(err, models.ErrCartNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCart")
		return nil, err
	}
	if cart == nil || request.CartID == nil || request.CartID.Compare(cart.ID) != 0 {
		return nil, errors.New("cart_id: not found")
	}
	mapCartItems := map[uuid.UUID]*models.CartItem{}
	for _, cartItem := range cart.Items {
		mapCartItems[cartItem.ID] = cartItem
	}
	deliveryRoute, err := q.deliveryRouteRepository.FindDeliveryRoute(
		ctx,
		models.NewDeliveryRouteFilter().
			WithDeliveryRouteIDs(request.DeliveryRouteID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryRouteNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryRoute")
		return nil, err
	}
	if deliveryRoute == nil {
		return nil, errors.New("delivery_route_id: not found")
	}
	request.DeliveryLongitude = deliveryRoute.DestinationLongitude
	request.DeliveryLatitude = deliveryRoute.DestinationLatitude
	request.DeliveryDistanceKM = deliveryRoute.DistanceKM
	setting, err := q.settingRepository.FindSetting(ctx)
	if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
		return nil, err
	}
	if setting == nil {
		return nil, models.ErrSettingNotFound
	}
	now := time.Now().In(time.Local)
	deliveryPlan, err := q.deliveryPlanRepository.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithTransactionDatetime(now),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return nil, err
	}
	if deliveryPlan == nil {
		return nil, models.ErrDeliveryPlanNotFound
	}
	if _, ok := deliveryPlan.MapDeliveryTimes[request.DeliveryDatetime.UnixNano()]; !ok {
		return nil, errors.New("delivery_time: not found")
	}
	request.DeliveryEstimatedCost = request.DeliveryDistanceKM * setting.DeliveryCostPerKM
	if n := len(request.Items); n > 0 {
		mapCartItemIDs := map[uuid.UUID]struct{}{}
		productVariantIDs := make([]uuid.UUID, len(request.Items))
		for i, item := range request.Items {
			if item.CartItemID == nil {
				return nil, fmt.Errorf("items[%d].cart_item_id: is required", i)
			}
			cartItem, ok := mapCartItems[*item.CartItemID]
			if !ok {
				return nil, fmt.Errorf("items[%d].cart_item_id: not found in your cart", i)
			}
			if _, ok = mapCartItemIDs[*item.CartItemID]; ok {
				return nil, fmt.Errorf("items[%d].cart_item_id: duplicated", i)
			}
			mapCartItemIDs[*item.CartItemID] = struct{}{}
			item.ProductVariantID = cartItem.ProductVariantID
			item.Quantity = cartItem.Quantity
			request.Items[i] = item
			productVariantIDs[i] = item.ProductVariantID
		}
		productVariants, _, _, err := q.productVariantRepository.FindProductVariants(
			ctx,
			models.NewProductVariantFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductVariantIDs(productVariantIDs...),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
			return nil, err
		}
		mapProductVariants := make(map[uuid.UUID]*models.ProductVariant, len(productVariants))
		for _, productVariant := range productVariants {
			mapProductVariants[productVariant.ID] = productVariant
		}
		for i, item := range request.Items {
			productVariant, ok := mapProductVariants[item.ProductVariantID]
			if !ok || !productVariant.ProductReadyStock {
				return nil, fmt.Errorf("items[%d].product_variant_id: not found or out of stock", i)
			}
			item.ProductID = productVariant.ProductID
			item.ProductName = productVariant.ProductName
			item.StockUnit = productVariant.StockUnit
			item.DiscountMinimumQuantity = productVariant.DiscountMinimumQuantity
			item.BasePrice = productVariant.BasePrice
			item.SellingPrice = productVariant.SellingPrice
			request.Items[i] = item
		}
	}
	request.CalculateShoppingCost()
	var sb strings.Builder
	if setting.ReferralPointsEnabled {
		if request.ReferralCode != nil {
			if request.ShoppingCost < setting.ReferralPointsMinimumShoppingCost {
				return nil, fmt.Errorf("referral_code: requires shopping cost minimum %s", humanize.Comma(setting.ReferralPointsMinimumShoppingCost))
			}
			referrer, err := q.buyerRepository.FindBuyer(
				ctx,
				models.NewBuyerFilter().
					WithReferralCodes(*request.ReferralCode),
			)
			if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
				return nil, err
			}
			if referrer == nil {
				return nil, errors.New("referral_code: is invalid")
			}
			_, total, _, err := q.transactionRepository.FindTransactions(
				ctx,
				models.NewTransactionFilter().
					WithBuyerIDs(request.BuyerID).
					WithExcludedTransactionStatuses(
						models.TransactionStatusCancelled,
						models.TransactionStatusWithdrawalPending,
						models.TransactionStatusWithdrawn,
					).
					WithReferrerIDs(referrer.ID).
					WithReturnCountOnly(true),
			)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
				return nil, err
			}
			if (buyer.ReferrerID != nil &&
				buyer.ReferrerID.Compare(referrer.ID) != 0) ||
				int(total) >= len(setting.ReferralPointsSlots) {
				return nil, errors.New("referral_code: is invalid")
			}
			request.ReferrerID = &referrer.ID
			request.ReferralPoints = setting.ReferralPointsSlots[total]
		} else {
			request.ReferrerID = nil
			request.ReferralPoints = 0
		}
	}
	request.RewardPointsDiscount = 0
	if request.RewardPointsDiscounted {
		pointSpendingCap, err := q.pointSpendingCapRepository.FindPointSpendingCap(
			ctx,
			models.NewPointSpendingCapFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithShoppingCosts(request.ShoppingCost),
		)
		if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
			return nil, err
		}
		if pointSpendingCap != nil {
			request.RewardPointsDiscount = pointSpendingCap.DisposablePoints
		}
	} else if setting.VoucherEnabled {
		if request.VoucherCode != nil {
			voucher, err := q.voucherRepository.FindVoucher(
				ctx,
				models.NewVoucherFilter().
					WithRecordStatuses(models.RecordStatusActive).
					WithCodes(*request.VoucherCode).
					WithTransactionDatetime(now).
					WithBuyerIDs(request.BuyerID),
			)
			if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
				return nil, err
			}
			if voucher == nil {
				return nil, models.ErrVoucherNotFound
			}
			if voucher.QuotaPerBuyer > 0 {
				_, total, _, err := q.transactionRepository.FindTransactions(
					ctx,
					models.NewTransactionFilter().
						WithBuyerIDs(request.BuyerID).
						WithExcludedTransactionStatuses(models.TransactionStatusCancelled).
						WithVoucherIDs(voucher.ID).
						WithReturnCountOnly(true),
				)
				if err != nil {
					helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
					return nil, err
				}
				if total >= int64(voucher.QuotaPerBuyer) {
					return nil, errors.New("voucher_code: quota exceeded")
				}
			}
			request.VoucherID = &voucher.ID
			request.VoucherDiscount = voucher.Discount
			request.VoucherType = &voucher.VoucherType
		} else {
			request.VoucherID = nil
			request.VoucherDiscount = 0
			request.VoucherType = nil
		}
	}
	request.CalculateDiscount()
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "%s%03d", now.Format("06"), now.YearDay())
	counter, err := q.transactionCounterRepository.SaveTransactionCounter(
		ctx,
		&models.TransactionCounter{
			TransactionWindow: sb.String(),
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveTransactionCounter")
		return nil, err
	}
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		models.TransactionReferenceNumberFormat,
		counter.TransactionWindow,
		counter.Hits,
		strings.ToUpper(helper.RandomString(4)),
	)
	if request.PaymentAmount > 0 {
		request.Status = models.TransactionStatusPending
	} else {
		request.Status = models.TransactionStatusPaid
		request.PaidAt = &now
	}
	request.ReferenceNumber = sb.String()
	request.CreatedAt = now
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return nil, err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.transactionRepository.CreateTransaction(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTransaction")
		return nil, err
	}
	if request.PaymentAmount > 0 {
		if request.PaymentChannelID == nil || request.PaymentChannelID.Compare(uuid.Nil()) == 0 {
			return nil, errors.New("payment_channel_id: is required")
		}
		paymentChannel, err := q.paymentChannelRepository.FindPaymentChannel(
			ctx,
			models.NewPaymentChannelFilter().
				WithPaymentChannelIDs(*request.PaymentChannelID).
				WithAmount(request.PaymentAmount),
		)
		if err != nil && !errors.Is(err, models.ErrPaymentChannelNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPaymentChannel")
			return nil, err
		}
		if paymentChannel == nil {
			return nil, errors.New("payment_channel_id: not found")
		}
		paymentRequestID, paymentTokenID, err := q.xenditConsumer.CreatePaymentRequest(
			ctx,
			response.ID.String(),
			paymentChannel.ChannelType,
			paymentChannel.Country,
			paymentChannel.Currency,
			request.PaymentAmount,
			"AUTOMATIC",
			paymentChannel.ChannelCode,
			helper.Sprintf("Pembayaran Transaksi %s", response.ReferenceNumber),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePaymentRequest")
			return nil, err
		}
		paymentRequestExpiry, err := decimal.NewFromFloat64(paymentChannel.PaymentRequestExpiry)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewFromFloat64")
			return nil, err
		}
		oneHour, err := decimal.New(3_600_000_000_000, 0)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNew")
			return nil, err
		}
		nanoseconds, err := paymentRequestExpiry.Mul(oneHour)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrMul")
			return nil, err
		}
		roundedNanoseconds, _, _ := nanoseconds.Floor(0).Int64(0)
		duration := min(time.Duration(roundedNanoseconds), setting.PaymentTerms)
		response.PaymentDueDate = new(now.Add(duration))
		response.PaymentChannelID = request.PaymentChannelID
		response.PaymentRequestID = &paymentRequestID
		response.PaymentTokenID = &paymentTokenID
		if response, err = q.transactionRepository.UpdateTransaction(ctx, tx, response); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
			return nil, err
		}
	}
	cart.BuyerID = nil
	cart.DeletedBuyerID = &request.BuyerID
	cart.UpdatedAt = now
	cart.Items = nil
	if _, err = q.cartRepository.SaveCart(ctx, tx, cart); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveCart")
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return nil, err
	}
	return response, nil
}
func (q *transactionService) UpdateTransaction(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Transaction,
) (*models.Transaction, error) {
	ctxt := "TransactionService-UpdateTransaction"
	response, err := q.transactionRepository.UpdateTransaction(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return nil, err
	}
	return response, nil
}

func (q *transactionService) CancelOverdueTransactions(ctx context.Context) error {
	ctxt := "TransactionService-CancelOverdueTransactions"
	now := time.Now().In(time.Local)
	transactions, _, _, err := q.transactionRepository.FindTransactions(
		ctx,
		models.NewTransactionFilter().
			WithPaymentDueDate(now),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return err
	}
	if len(transactions) == 0 {
		return nil
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	for _, transaction := range transactions {
		transaction.Status = models.TransactionStatusCancelled
		transaction.CancelledAt = &now
		if _, err = q.transactionRepository.UpdateTransaction(ctx, tx, transaction); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
			return err
		}
		if transaction.PaymentRequestID != nil {
			if _, err = q.xenditConsumer.CancelPaymentRequest(ctx, *transaction.PaymentRequestID); err != nil {
				return err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return err
	}
	return nil
}

func (q *transactionService) WithdrawTransaction(
	ctx context.Context,
	request *models.Transaction,
) (*models.Transaction, error) {
	ctxt := "TransactionService-WithdrawTransaction"
	if request.PaymentID != nil {
		if _, err := q.xenditConsumer.CancelPayment(ctx, *request.PaymentID); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCancelPayment")
			return nil, err
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return nil, err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.transactionRepository.UpdateTransaction(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return nil, err
	}
	return response, nil
}
