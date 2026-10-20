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
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/roysitumorang/laukpauk/repositories"
)

type TransactionService interface {
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
		company *models.Company,
		buyer *models.Buyer,
		request *models.Transaction,
	) (*models.Transaction, []uuid.UUID, error)
	UpdateTransaction(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Transaction,
	) (*models.Transaction, error)
	AutoCancelOverdueTransactions(ctx context.Context) error
	RequestWithdrawal(
		ctx context.Context,
		request *models.Transaction,
	) (*models.Transaction, error)
	WithdrawTransaction(
		ctx context.Context,
		request *models.Transaction,
	) (*models.Transaction, error)
	CompleteTransaction(
		ctx context.Context,
		request *models.Transaction,
	) (*models.Transaction, error)
	AutoCompleteTransactions(ctx context.Context) error
}

type transactionService struct {
	transactionRepository        repositories.TransactionRepository
	buyerRepository              repositories.BuyerRepository
	transactionCounterRepository repositories.TransactionCounterRepository
	cartRepository               repositories.CartRepository
	deliveryPlanRepository       repositories.DeliveryPlanRepository
	deliveryRouteRepository      repositories.DeliveryRouteRepository
	companyRepository            repositories.CompanyRepository
	voucherRepository            repositories.VoucherRepository
	pointSpendingCapRepository   repositories.PointSpendingCapRepository
	pointRepository              repositories.PointRepository
	productVariantRepository     repositories.ProductVariantRepository
	paymentChannelRepository     repositories.PaymentChannelRepository
	sellerRepository             repositories.SellerRepository
	notificationRepository       repositories.NotificationRepository
	xenditClient                 *externals.XenditClient
}

func NewTransactionService(
	transactionRepository repositories.TransactionRepository,
	buyerRepository repositories.BuyerRepository,
	transactionCounterRepository repositories.TransactionCounterRepository,
	cartRepository repositories.CartRepository,
	deliveryPlanRepository repositories.DeliveryPlanRepository,
	deliveryRouteRepository repositories.DeliveryRouteRepository,
	companyRepository repositories.CompanyRepository,
	voucherRepository repositories.VoucherRepository,
	pointSpendingCapRepository repositories.PointSpendingCapRepository,
	pointRepository repositories.PointRepository,
	productVariantRepository repositories.ProductVariantRepository,
	paymentChannelRepository repositories.PaymentChannelRepository,
	sellerRepository repositories.SellerRepository,
	notificationRepository repositories.NotificationRepository,
	xenditClient *externals.XenditClient,
) TransactionService {
	return &transactionService{
		transactionRepository:        transactionRepository,
		buyerRepository:              buyerRepository,
		transactionCounterRepository: transactionCounterRepository,
		cartRepository:               cartRepository,
		deliveryPlanRepository:       deliveryPlanRepository,
		deliveryRouteRepository:      deliveryRouteRepository,
		companyRepository:            companyRepository,
		voucherRepository:            voucherRepository,
		pointSpendingCapRepository:   pointSpendingCapRepository,
		pointRepository:              pointRepository,
		productVariantRepository:     productVariantRepository,
		paymentChannelRepository:     paymentChannelRepository,
		sellerRepository:             sellerRepository,
		notificationRepository:       notificationRepository,
		xenditClient:                 xenditClient,
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
	company *models.Company,
	buyer *models.Buyer,
	request *models.Transaction,
) (*models.Transaction, []uuid.UUID, error) {
	ctxt := "TransactionService-CreateTransaction"
	cart, err := q.cartRepository.FindCart(
		ctx,
		models.NewCartFilter().
			WithBuyerIDs(buyer.ID),
	)
	if err != nil && !errors.Is(err, models.ErrCartNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCart")
		return nil, nil, err
	}
	if cart == nil || request.CartID == nil || request.CartID.Compare(cart.ID) != 0 {
		return nil, nil, errors.New("cart_id: not found")
	}
	mapCartItems := map[uuid.UUID]*models.CartItem{}
	for _, cartItem := range cart.Items {
		mapCartItems[cartItem.ID] = cartItem
	}
	deliveryRoute, err := q.deliveryRouteRepository.FindDeliveryRoute(
		ctx,
		models.NewDeliveryRouteFilter().
			WithCompanyIDs(company.ID).
			WithDeliveryRouteIDs(request.DeliveryRouteID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryRouteNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryRoute")
		return nil, nil, err
	}
	if deliveryRoute == nil {
		return nil, nil, errors.New("delivery_route_id: not found")
	}
	request.DeliveryLongitude = deliveryRoute.DestinationLongitude
	request.DeliveryLatitude = deliveryRoute.DestinationLatitude
	request.DeliveryDistanceKM = deliveryRoute.DistanceKM
	now := time.Now().In(time.Local)
	deliveryPlan, err := q.deliveryPlanRepository.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithCompanyIDs(company.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithTransactionDatetime(now),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return nil, nil, err
	}
	if deliveryPlan == nil {
		return nil, nil, models.ErrDeliveryPlanNotFound
	}
	if _, ok := deliveryPlan.MapDeliveryTimes[request.DeliveryDatetime.UnixNano()]; !ok {
		return nil, nil, errors.New("delivery_time: not found")
	}
	request.DeliveryEstimatedCost = request.DeliveryDistanceKM * company.DeliveryCostPerKM
	if n := len(request.Items); n > 0 {
		mapCartItemIDs := map[uuid.UUID]struct{}{}
		productVariantIDs := make([]uuid.UUID, len(request.Items))
		for i, item := range request.Items {
			if item.CartItemID == nil {
				return nil, nil, fmt.Errorf("items[%d].cart_item_id: is required", i)
			}
			cartItem, ok := mapCartItems[*item.CartItemID]
			if !ok {
				return nil, nil, fmt.Errorf("items[%d].cart_item_id: not found in your cart", i)
			}
			if _, ok = mapCartItemIDs[*item.CartItemID]; ok {
				return nil, nil, fmt.Errorf("items[%d].cart_item_id: duplicated", i)
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
				WithCompanyIDs(company.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithProductVariantIDs(productVariantIDs...),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
			return nil, nil, err
		}
		mapProductVariants := make(map[uuid.UUID]*models.ProductVariant, len(productVariants))
		for _, productVariant := range productVariants {
			mapProductVariants[productVariant.ID] = productVariant
		}
		for i, item := range request.Items {
			productVariant, ok := mapProductVariants[item.ProductVariantID]
			if !ok || !productVariant.ProductReadyStock {
				return nil, nil, fmt.Errorf("items[%d].product_variant_id: not found or out of stock", i)
			}
			item.ProductID = productVariant.ProductID
			item.ProductName = productVariant.ProductName
			item.ProductPhoto = productVariant.ProductPhoto
			item.StockUnit = productVariant.StockUnit
			item.DiscountMinimumQuantity = productVariant.DiscountMinimumQuantity
			item.BasePrice = productVariant.BasePrice
			item.SellingPrice = productVariant.SellingPrice
			request.Items[i] = item
		}
	}
	request.CalculateShoppingCost()
	if company.ReferralPointsEnabled {
		if request.ReferralCode != nil {
			if request.ShoppingCost < company.ReferralPointsMinimumShoppingCost {
				return nil, nil, fmt.Errorf("referral_code: requires shopping cost minimum %s", humanize.Comma(company.ReferralPointsMinimumShoppingCost))
			}
			referrer, err := q.buyerRepository.FindBuyer(
				ctx,
				models.NewBuyerFilter().
					WithCompanyIDs(company.ID).
					WithReferralCodes(*request.ReferralCode),
			)
			if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
				return nil, nil, err
			}
			if referrer == nil {
				return nil, nil, errors.New("referral_code: is invalid")
			}
			_, total, _, err := q.transactionRepository.FindTransactions(
				ctx,
				models.NewTransactionFilter().
					WithCompanyIDs(company.ID).
					WithBuyerIDs(buyer.ID).
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
				return nil, nil, err
			}
			if (buyer.ReferrerID != nil &&
				buyer.ReferrerID.Compare(referrer.ID) != 0) ||
				int(total) >= len(company.ReferralPointsSlots) {
				return nil, nil, errors.New("referral_code: is invalid")
			}
			request.ReferrerID = &referrer.ID
			request.ReferralPoints = company.ReferralPointsSlots[total]
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
				WithCompanyIDs(company.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithShoppingCost(request.ShoppingCost, request.ShoppingCost),
		)
		if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
			return nil, nil, err
		}
		if pointSpendingCap != nil {
			request.RewardPointsDiscount = pointSpendingCap.DisposablePoints
		}
	} else if company.VoucherEnabled {
		if request.VoucherCode != nil {
			voucher, err := q.voucherRepository.FindVoucher(
				ctx,
				models.NewVoucherFilter().
					WithCompanyIDs(company.ID).
					WithRecordStatuses(models.RecordStatusActive).
					WithCodes(*request.VoucherCode).
					WithTransactionDatetime(now).
					WithBuyerIDs(buyer.ID),
			)
			if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
				return nil, nil, err
			}
			if voucher == nil {
				return nil, nil, models.ErrVoucherNotFound
			}
			if voucher.QuotaPerBuyer > 0 {
				_, total, _, err := q.transactionRepository.FindTransactions(
					ctx,
					models.NewTransactionFilter().
						WithCompanyIDs(company.ID).
						WithBuyerIDs(buyer.ID).
						WithExcludedTransactionStatuses(models.TransactionStatusCancelled).
						WithVoucherIDs(voucher.ID).
						WithReturnCountOnly(true),
				)
				if err != nil {
					helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
					return nil, nil, err
				}
				if total >= voucher.QuotaPerBuyer {
					return nil, nil, errors.New("voucher_code: quota exceeded")
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
	sellers, _, _, err := q.sellerRepository.FindSellers(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(company.ID).
			WithStatuses(models.SellerStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSellers")
		return nil, nil, err
	}
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s%03d", now.Format("06"), now.YearDay())
	counter, err := q.transactionCounterRepository.SaveTransactionCounter(
		ctx,
		&models.TransactionCounter{
			CompanyID:         company.ID,
			TransactionWindow: sb.String(),
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveTransactionCounter")
		return nil, nil, err
	}
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
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
		return nil, nil, err
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
		return nil, nil, err
	}
	if request.PaymentAmount > 0 {
		if request.PaymentChannelID == nil || request.PaymentChannelID.Compare(uuid.Nil()) == 0 {
			return nil, nil, errors.New("payment_channel_id: is required")
		}
		paymentChannel, err := q.paymentChannelRepository.FindPaymentChannel(
			ctx,
			models.NewPaymentChannelFilter().
				WithPaymentChannelIDs(*request.PaymentChannelID).
				WithAmount(request.PaymentAmount),
		)
		if err != nil && !errors.Is(err, models.ErrPaymentChannelNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPaymentChannel")
			return nil, nil, err
		}
		if paymentChannel == nil {
			return nil, nil, errors.New("payment_channel_id: not found")
		}
		paymentRequestID, paymentTokenID, virtualAccountNumber, virtualAccountExpiredAt, err := q.xenditClient.CreatePaymentRequest(
			ctx,
			response.BuyerID.String(),
			response.ID.String(),
			helper.Sprintf("Transaksi %s", response.ReferenceNumber),
			paymentChannel.Country,
			paymentChannel.Currency,
			request.PaymentAmount,
			"AUTOMATIC",
			paymentChannel.ChannelCode,
			helper.Sprintf("Pembayaran Transaksi %s", response.ReferenceNumber),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePaymentRequest")
			return nil, nil, err
		}
		duration := company.PaymentTerms
		if virtualAccountExpiredAt != nil {
			duration = models.Duration{Duration: virtualAccountExpiredAt.Sub(response.CreatedAt) - time.Second}
		}
		response.PaymentDueDate = new(now.Add(duration.Duration))
		response.PaymentChannelID = request.PaymentChannelID
		response.PaymentChannelName = &paymentChannel.Name
		response.PaymentRequestID = &paymentRequestID
		response.PaymentTokenID = paymentTokenID
		response.PaymentVirtualAccountNumber = virtualAccountNumber
		response.PaymentVirtualAccountExpiredAt = virtualAccountExpiredAt
		if response, err = q.transactionRepository.UpdateTransaction(ctx, tx, response); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
			return nil, nil, err
		}
	}
	cart.BuyerID = nil
	cart.DeletedBuyerID = &buyer.ID
	cart.UpdatedAt = now
	cart.Items = nil
	if _, err = q.cartRepository.SaveCart(ctx, tx, cart); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveCart")
		return nil, nil, err
	}
	notificationIDs := make([]uuid.UUID, len(sellers))
	for i, seller := range sellers {
		notification, err := q.notificationRepository.CreateNotification(
			ctx,
			tx,
			&models.Notification{
				SellerID:         seller.ID,
				NotificationType: models.NotificationTypeTransaction,
				SellerLink:       new(helper.Sprintf("/seller/transactions/%s", response.ID.String())),
				LinkID:           &response.ID,
				Subject:          "New Order",
				Body:             helper.Sprintf(`New Order Reference No "%s"`, response.ReferenceNumber),
				Params: map[string]string{
					"id": response.ID.String(),
				},
				CreatedByBuyerID: buyer.ID,
				CreatedAt:        response.CreatedAt,
			},
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateNotification")
			return nil, nil, err
		}
		seller.UnreadNotifications = 1
		if _, err = q.sellerRepository.UpdateSeller(ctx, tx, seller); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSeller")
			return nil, nil, err
		}
		notificationIDs[i] = notification.ID
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return nil, nil, err
	}
	return response, notificationIDs, nil
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

func (q *transactionService) AutoCancelOverdueTransactions(ctx context.Context) error {
	ctxt := "TransactionService-AutoCancelOverdueTransactions"
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
		if transaction.PaymentRequestID != nil &&
			transaction.PaymentDueDate != nil &&
			transaction.PaymentDueDate.After(now) {
			if _, err = q.xenditClient.CancelPaymentRequest(ctx, *transaction.PaymentRequestID); err != nil {
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

func (q *transactionService) RequestWithdrawal(
	ctx context.Context,
	request *models.Transaction,
) (*models.Transaction, error) {
	ctxt := "TransactionService-RequestWithdrawal"
	setting, err := q.companyRepository.FindCompany(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(request.CompanyID),
	)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
		return nil, err
	}
	if setting == nil {
		return nil, models.ErrCompanyNotFound
	}
	now := time.Now().In(time.Local)
	if setting.TransactionAutoWithdrawalEnabled &&
		setting.TransactionAutoWithdrawalWindowMin != nil &&
		setting.TransactionAutoWithdrawalWindowMax != nil &&
		!request.CreatedAt.Before(*setting.TransactionAutoWithdrawalWindowMin) &&
		!request.CreatedAt.After(*setting.TransactionAutoWithdrawalWindowMax) {
		request.Status = models.TransactionStatusWithdrawn
		request.WithdrawnAt = &now
	} else {
		request.Status = models.TransactionStatusWithdrawalPending
		request.WithdrawalRequestedAt = &now
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

func (q *transactionService) WithdrawTransaction(
	ctx context.Context,
	request *models.Transaction,
) (*models.Transaction, error) {
	ctxt := "TransactionService-WithdrawTransaction"
	if request.PaymentID != nil {
		if _, err := q.xenditClient.CancelPayment(ctx, *request.PaymentID); err != nil {
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

func (q *transactionService) CompleteTransaction(
	ctx context.Context,
	request *models.Transaction,
) (*models.Transaction, error) {
	ctxt := "TransactionService-CompleteTransaction"
	buyerIDs := []uuid.UUID{request.BuyerID}
	if request.ReferrerID != nil {
		buyerIDs = append(buyerIDs, *request.ReferrerID)
	}
	buyers, _, _, err := q.buyerRepository.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(buyerIDs...),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return nil, err
	}
	mapBuyers := make(map[uuid.UUID]*models.Buyer, len(buyers))
	for _, buyer := range buyers {
		mapBuyers[buyer.ID] = buyer
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
	request.Status = models.TransactionStatusCompleted
	request.CompletedAt = new(time.Now())
	response, err := q.transactionRepository.UpdateTransaction(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return nil, err
	}
	if buyer, ok := mapBuyers[response.BuyerID]; ok && buyer.ReferrerID == nil && response.ReferrerID != nil {
		buyer.ReferrerID = response.ReferrerID
		buyer.ReferralPoints = 0
		buyer.RewardPoints = 0
		buyer.Wallet = max(0, response.RefundAmount)
		if _, err = q.buyerRepository.UpdateBuyer(ctx, tx, buyer); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
			return nil, err
		}
	}
	if response.ReferrerID != nil {
		if buyer, ok := mapBuyers[*response.ReferrerID]; ok {
			buyer.ReferralPoints = -response.ReferralPoints
			buyer.RewardPoints = response.RewardPointsDiscount
			buyer.Wallet = 0
			if _, err = q.buyerRepository.UpdateBuyer(ctx, tx, buyer); err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
				return nil, err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return nil, err
	}
	return response, nil
}

func (q *transactionService) AutoCompleteTransactions(ctx context.Context) error {
	ctxt := "TransactionService-AutoCompleteTransactions"
	transactions, _, _, err := q.transactionRepository.FindTransactions(
		ctx,
		models.NewTransactionFilter().
			WithTransactionStatuses(models.TransactionStatusShipped),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return err
	}
	if len(transactions) == 0 {
		return nil
	}
	now := time.Now().In(time.Local)
	var companyIDs, buyerIDs []uuid.UUID
	mapCompanyIDs := map[uuid.UUID]struct{}{}
	mapBuyerIDs := map[uuid.UUID]struct{}{}
	for _, transaction := range transactions {
		if _, ok := mapCompanyIDs[transaction.CompanyID]; !ok {
			companyIDs = append(buyerIDs, transaction.CompanyID)
			mapCompanyIDs[transaction.CompanyID] = struct{}{}
		}
		if _, ok := mapBuyerIDs[transaction.BuyerID]; !ok {
			buyerIDs = append(buyerIDs, transaction.BuyerID)
			mapBuyerIDs[transaction.BuyerID] = struct{}{}
		}
		if transaction.ReferrerID != nil {
			if _, ok := mapBuyerIDs[*transaction.ReferrerID]; !ok {
				buyerIDs = append(buyerIDs, *transaction.ReferrerID)
				mapBuyerIDs[*transaction.ReferrerID] = struct{}{}
			}
		}
	}
	companies, _, _, err := q.companyRepository.FindCompanies(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(companyIDs...),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompanies")
		return err
	}
	mapCompanies := make(map[uuid.UUID]*models.Company, len(companies))
	for _, company := range companies {
		mapCompanies[company.ID] = company
	}
	buyers, _, _, err := q.buyerRepository.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(buyerIDs...),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return err
	}
	mapBuyers := make(map[uuid.UUID]*models.Buyer, len(buyers))
	for _, buyer := range buyers {
		mapBuyers[buyer.ID] = buyer
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
		company, ok := mapCompanies[transaction.CompanyID]
		if !ok || !transaction.ShippedAt.Before(now.Add(-company.MaximumDeliveryTime.Duration)) {
			continue
		}
		transaction.Status = models.TransactionStatusCompleted
		transaction.CompletedAt = &now
		response, err := q.transactionRepository.UpdateTransaction(ctx, tx, transaction)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
			return err
		}
		if buyer, ok := mapBuyers[response.BuyerID]; ok && buyer.ReferrerID == nil && response.ReferrerID != nil {
			buyer.ReferrerID = response.ReferrerID
			buyer.ReferralPoints = 0
			buyer.RewardPoints = 0
			buyer.Wallet = max(0, response.RefundAmount)
			if _, err = q.buyerRepository.UpdateBuyer(ctx, tx, buyer); err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
				return err
			}
		}
		if response.ReferrerID != nil {
			if buyer, ok := mapBuyers[*response.ReferrerID]; ok {
				buyer.ReferralPoints = -response.ReferralPoints
				buyer.RewardPoints = response.RewardPointsDiscount
				buyer.Wallet = 0
				if _, err = q.buyerRepository.UpdateBuyer(ctx, tx, buyer); err != nil {
					helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
					return err
				}
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return err
	}
	return nil
}
