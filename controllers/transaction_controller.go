package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	transactionController struct {
		adminService              services.AdminService
		buyerService              services.BuyerService
		transactionService        services.TransactionService
		transactionCounterService services.TransactionCounterService
		cartService               services.CartService
		cartItemService           services.CartItemService
		deliveryPlanService       services.DeliveryPlanService
		deliveryRouteService      services.DeliveryRouteService
		settingService            services.SettingService
		voucherService            services.VoucherService
		pointSpendingCapService   services.PointSpendingCapService
		pointService              services.PointService
		productVariantService     services.ProductVariantService
		jsonWebTokenService       services.JsonWebTokenService
	}
)

func NewTransactionController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	transactionService services.TransactionService,
	transactionCounterService services.TransactionCounterService,
	cartService services.CartService,
	cartItemService services.CartItemService,
	deliveryPlanService services.DeliveryPlanService,
	deliveryRouteService services.DeliveryRouteService,
	settingService services.SettingService,
	voucherService services.VoucherService,
	pointSpendingCapService services.PointSpendingCapService,
	pointService services.PointService,
	productVariantService services.ProductVariantService,
	jsonWebTokenService services.JsonWebTokenService,
) *transactionController {
	return &transactionController{
		adminService:              adminService,
		buyerService:              buyerService,
		transactionService:        transactionService,
		transactionCounterService: transactionCounterService,
		cartService:               cartService,
		cartItemService:           cartItemService,
		deliveryPlanService:       deliveryPlanService,
		deliveryRouteService:      deliveryRouteService,
		settingService:            settingService,
		voucherService:            voucherService,
		pointSpendingCapService:   pointSpendingCapService,
		pointService:              pointService,
		productVariantService:     productVariantService,
		jsonWebTokenService:       jsonWebTokenService,
	}
}

func (q *transactionController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	buyerKeyAuth := middleware.BuyerKeyAuth(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Get("", buyerKeyAuth, q.V1BuyerIndex).
		Post("", buyerKeyAuth, q.V1BuyerCreate).
		Get("/:id", buyerKeyAuth, q.V1BuyerShow).
		Post("/:id/pay", q.V1Pay)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/:id", q.AdminShow).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id/process", q.AdminProcess).
		Post("/:id/ship", q.AdminShip).
		Post("/:id/deliver", q.AdminDeliver)

}

func (q *transactionController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerIndex"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, err := sanitizers.FindTransactions(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.transactionService.FindTransactions(
		c,
		request.WithBuyerIDs(currentBuyer.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	pagination.Rows = make([]any, len(rows))
	for i := range rows {
		pagination.Rows[i] = rows[i]
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}

func (q *transactionController) V1BuyerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerCreate"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateTransaction(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		c.Response().SetStatusCode(statusCode)
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	cart, err := q.cartService.FindCart(
		ctx,
		models.NewCartFilter().
			WithCartIDs(currentBuyer.ID),
	)
	if err != nil && !errors.Is(err, models.ErrCartNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCart")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if cart == nil || request.CartID == nil || request.CartID.Compare(cart.ID) != 0 {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("cart_id: not found").
			WriteResponse(c)
	}
	cartItems, _, err := q.cartItemService.FindCartItems(
		ctx,
		models.NewCartItemFilter().
			WithCartIDs(cart.ID).
			WithSortings(models.Sorting{
				Field:     models.SortingFieldID,
				Direction: models.SortingDirectionAsc,
			}),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCartItems")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	mapCartItems := map[uuid.UUID]*models.CartItem{}
	for _, item := range cartItems {
		mapCartItems[item.ID] = item
	}
	deliveryRoute, err := q.deliveryRouteService.FindDeliveryRoute(
		ctx,
		models.NewDeliveryRouteFilter().
			WithDeliveryRouteIDs(request.DeliveryRouteID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryRouteNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryRoute")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryRoute == nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("delivery_route_id: not found").
			WriteResponse(c)
	}
	request.DeliveryLongitude = deliveryRoute.DestinationLongitude
	request.DeliveryLatitude = deliveryRoute.DestinationLatitude
	request.DeliveryDistanceKM = deliveryRoute.DistanceKM
	now := time.Now().In(time.Local)
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithTransactionDatetime(now),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryPlan == nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	var deliveryDatetimeIdValid bool
	if deliveryPlan.SameDay {
		for _, deliveryDatetime := range deliveryPlan.SameDayDeliveryDatetimes {
			if deliveryDatetimeIdValid = deliveryDatetimeIdValid ||
				request.DeliveryDatetimeID.Compare(deliveryDatetime.ID) == 0; deliveryDatetimeIdValid {
				request.DeliveryPlanID = deliveryDatetime.DeliveryPlanID
				request.DeliverySpeed = deliveryDatetime.DeliverySpeed
				request.DeliveryTimestamp = deliveryDatetime.Timestamp
				break
			}
		}
	}
	if !deliveryDatetimeIdValid && deliveryPlan.NextDay {
		for _, deliveryDatetime := range deliveryPlan.NextDayDeliveryDatetimes {
			if deliveryDatetimeIdValid = deliveryDatetimeIdValid ||
				request.DeliveryDatetimeID.Compare(deliveryDatetime.ID) == 0; deliveryDatetimeIdValid {
				request.DeliveryPlanID = deliveryDatetime.DeliveryPlanID
				request.DeliverySpeed = deliveryDatetime.DeliverySpeed
				request.DeliveryTimestamp = deliveryDatetime.Timestamp
				break
			}
		}
	}
	if !deliveryDatetimeIdValid && deliveryPlan.NextTwoDay {
		for _, deliveryDatetime := range deliveryPlan.NextTwoDayDeliveryDatetimes {
			if deliveryDatetimeIdValid = deliveryDatetimeIdValid ||
				request.DeliveryDatetimeID.Compare(deliveryDatetime.ID) == 0; deliveryDatetimeIdValid {
				request.DeliveryPlanID = deliveryDatetime.DeliveryPlanID
				request.DeliverySpeed = deliveryDatetime.DeliverySpeed
				request.DeliveryTimestamp = deliveryDatetime.Timestamp
				break
			}
		}
	}
	if !deliveryDatetimeIdValid {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("delivery_datetime_id: not found").
			WriteResponse(c)
	}
	setting, err := q.settingService.FindSetting(ctx)
	if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrSettingNotFound.Error()).
			WriteResponse(c)
	}
	if setting == nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.DeliveryCost = request.DeliveryDistanceKM * setting.DeliveryCostPerKM
	mapCartItemIDs := map[uuid.UUID]struct{}{}
	if n := len(request.Items); n > 0 {
		productVariantIDs := make([]uuid.UUID, len(request.Items))
		for i, item := range request.Items {
			cartItem, ok := mapCartItems[item.CartItemID]
			if !ok {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(helper.Sprintf("items[%d].cart_item_id: not found in your cart", i)).
					WriteResponse(c)
			}
			if _, ok = mapCartItemIDs[item.CartItemID]; ok {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(helper.Sprintf("items[%d].cart_item_id: duplicated", i)).
					WriteResponse(c)
			}
			mapCartItemIDs[item.CartItemID] = struct{}{}
			item.ProductVariantID = cartItem.ProductVariantID
			item.Quantity = cartItem.Quantity
			request.Items[i] = item
			productVariantIDs[i] = item.ProductVariantID
		}
		productVariants, _, err := q.productVariantService.FindProductVariants(
			ctx,
			models.NewProductVariantFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductVariantIDs(productVariantIDs...),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		mapProductVariants := make(map[uuid.UUID]*models.ProductVariant, len(productVariants))
		for _, productVariant := range productVariants {
			mapProductVariants[productVariant.ID] = productVariant
		}
		for i, item := range request.Items {
			productVariant, ok := mapProductVariants[item.ProductVariantID]
			if !ok {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(helper.Sprintf("items[%d].product_variant_id: not found", i)).
					WriteResponse(c)
			}
			if !productVariant.ProductReadyStock {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(helper.Sprintf("items[%d].product_variant_id: out of stock", i)).
					WriteResponse(c)
			}
			item.ProductID = productVariant.ProductID
			item.ProductName = productVariant.ProductName
			item.StockUnit = productVariant.StockUnit
			item.DiscountMinimumQuantity = productVariant.DiscountMinimumQuantity
			item.Price = productVariant.Price
			request.Items[i] = item
		}
	}
	request.CalculateShoppingCost()
	var sb strings.Builder
	if setting.ReferralPointsEnabled {
		if request.ReferralCode != nil {
			if request.ShoppingCost < setting.RefferalPointsMinimumShoppingCost {
				sb.Reset()
				_, _ = fmt.Fprintf(&sb, "referral_code: requires shopping cost minimum %s", humanize.Comma(setting.RefferalPointsMinimumShoppingCost))
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(sb.String()).
					WriteResponse(c)
			}
			referrer, err := q.buyerService.FindBuyer(
				ctx,
				models.NewBuyerFilter().
					WithReferralCodes(*request.ReferralCode),
			)
			if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(err.Error()).
					WriteResponse(c)
			}
			if referrer == nil {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage("referral_code: is invalid").
					WriteResponse(c)
			}
			_, pagination, err := q.transactionService.FindTransactions(
				ctx,
				models.NewTransactionFilter().
					WithBuyerIDs(currentBuyer.ID).
					WithExcludedTransactionStatuses(models.TransactionStatusCancelled).
					WithReferrerIDs(referrer.ID).
					WithReturnCountOnly(true),
			)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(err.Error()).
					WriteResponse(c)
			}
			total := int(pagination.Info.Total)
			if (currentBuyer.ReferrerID != nil &&
				currentBuyer.ReferrerID.Compare(referrer.ID) != 0) ||
				total >= len(setting.ReferralPointsSlots) {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage("referral_code: is invalid").
					WriteResponse(c)
			}
			currentBuyer.ReferrerID = &referrer.ID
			request.ReferrerID = &referrer.ID
			request.ReferralPoints = setting.ReferralPointsSlots[total]
		} else {
			request.ReferrerID = nil
			request.ReferralPoints = 0
		}
	}
	request.RewardPointsDiscount = 0
	if request.RewardPointsDiscounted {
		pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
			ctx,
			models.NewPointSpendingCapFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithShoppingCosts(request.ShoppingCost),
		)
		if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if pointSpendingCap != nil {
			request.RewardPointsDiscount = pointSpendingCap.DisposablePoints
		}
	} else if setting.VoucherEnabled {
		if request.VoucherCode != nil {
			voucher, err := q.voucherService.FindVoucher(
				ctx,
				models.NewVoucherFilter().
					WithRecordStatuses(models.RecordStatusActive).
					WithCodes(*request.VoucherCode).
					WithTransactionDatetime(now).
					WithBuyerIDs(currentBuyer.ID),
			)
			if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(err.Error()).
					WriteResponse(c)
			}
			if voucher == nil {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(models.ErrVoucherNotFound.Error()).
					WriteResponse(c)
			}
			if voucher.QuotaPerBuyer > 0 {
				_, pagination, err := q.transactionService.FindTransactions(
					ctx,
					models.NewTransactionFilter().
						WithBuyerIDs(currentBuyer.ID).
						WithExcludedTransactionStatuses(models.TransactionStatusCancelled).
						WithVoucherIDs(voucher.ID).
						WithReturnCountOnly(true),
				)
				if err != nil {
					helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
					return helper.NewResponse(fiber.StatusUnprocessableEntity).
						SetMessage(err.Error()).
						WriteResponse(c)
				}
				if pagination.Info.Total >= int64(voucher.QuotaPerBuyer) {
					return helper.NewResponse(fiber.StatusUnprocessableEntity).
						SetMessage("voucher quota exceeded").
						WriteResponse(c)
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
	paymentTerms, err := time.ParseDuration(setting.PaymentTerms)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "%s%03d", now.Format("06"), now.YearDay())
	counter, err := q.transactionCounterService.SaveTransactionCounter(
		ctx,
		&models.TransactionCounter{
			TransactionWindow: sb.String(),
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveTransactionCounter")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		models.TransactionReferenceNumberFormat,
		counter.TransactionWindow,
		counter.Hits,
		strings.ToUpper(helper.RandomString(4)),
	)
	request.BuyerID = currentBuyer.ID
	request.ReferenceNumber = sb.String()
	request.PaymentDueDate = now.Add(paymentTerms)
	request.CreatedAt = now
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.transactionService.CreateTransaction(ctx, tx, request, &sb)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTransaction")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	cart.BuyerID = nil
	cart.DeletedBuyerID = &currentBuyer.ID
	cart.UpdatedAt = now
	if _, err = q.cartService.UpdateCart(ctx, tx, cart); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCart")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("transaction created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *transactionController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerShow"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.transactionService.FindTransaction(
		c,
		models.NewTransactionFilter().
			WithBuyerIDs(currentBuyer.ID).
			WithTransactionIDs(transactionID),
	)
	if err != nil && !errors.Is(err, models.ErrTransactionNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *transactionController) AdminIndex(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminIndex"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := new(models.TransactionFilter)
	rows := []*models.Transaction{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Transaction",
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
		"sortings", []models.Sorting{
			{
				Field:     models.SortingFieldID,
				Direction: models.SortingDirectionDesc,
				Label:     "Newest-Oldest",
			},
			{
				Field:     models.SortingFieldID,
				Direction: models.SortingDirectionAsc,
				Label:     "Oldest-Newest",
			},
		},
	)
	if request, err = sanitizers.FindTransactions(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return c.Render("admin/transaction_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.transactionService.FindTransactions(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return c.Render("admin/transaction_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/transaction_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *transactionController) AdminShow(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminShow"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	return c.Render("admin/transaction_show", params, "admin/layouts/authenticated")
}

func (q *transactionController) V1Pay(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1Pay"
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	transaction, err := q.transactionService.FindTransaction(
		ctx,
		models.NewTransactionFilter().
			WithTransactionIDs(transactionID).
			WithTransactionStatuses(models.TransactionStatusPending),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if transaction == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	transaction.TransactionStatus = models.TransactionStatusPaid
	transaction.PaidAt = new(time.Now())
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("transaction paid successfully").
		WriteResponse(c)
}

func (q *transactionController) AdminEdit(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminEdit"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	return c.Render("admin/transaction_edit", params, "admin/layouts/authenticated")
}

func (q *transactionController) AdminProcess(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminProcess"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		models.TransactionStatusPaid,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	request, statusCode, err := sanitizers.ValidateTransaction(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = transaction.ID
	request.TransactionStatus = models.TransactionStatusProcessed
	request.ProcessedBy = &currentAdmin.ID
	request.ProcessedAt = new(time.Now())
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return c.Render("admin/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("transaction processed successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
}

func (q *transactionController) AdminShip(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminShip"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		models.TransactionStatusProcessed,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	transaction.TransactionStatus = models.TransactionStatusShipped
	transaction.ShippedBy = &currentAdmin.ID
	transaction.ShippedAt = new(time.Now())
	params.Push("request", transaction)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("transaction shipped successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
}

func (q *transactionController) AdminDeliver(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminDeliver"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		models.TransactionStatusShipped,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	transaction.TransactionStatus = models.TransactionStatusCompleted
	params.Push("request", transaction)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("transaction delivered successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
}

func (q *transactionController) prepare(c fiber.Ctx) (
	context.Context,
	*session.Middleware,
	*helper.FlashMessage,
	*models.Admin,
	string,
) {
	ctx := c.Context()
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	return ctx, sess, flash, currentAdmin, csrfToken
}

func (q *transactionController) prepareTransaction(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentAdmin *models.Admin,
	csrfToken string,
	withParams bool,
	transactionStatuses ...models.TransactionStatus,
) (
	*models.Transaction,
	models.Map,
	error,
) {
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, models.ErrTransactionNotFound
	}
	request := models.NewTransactionFilter().
		WithTransactionIDs(transactionID)
	if len(transactionStatuses) > 0 {
		request.WithTransactionStatuses(transactionStatuses...)
	}
	transaction, err := q.transactionService.FindTransaction(ctx, request)
	if err != nil && !errors.Is(err, models.ErrTransactionNotFound) {
		return nil, nil, err
	}
	if transaction == nil {
		return nil, nil, models.ErrTransactionNotFound
	}
	var params models.Map
	if withParams {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentAdmin", currentAdmin,
			"pageTitle", helper.Sprintf("Detail Data Transaction|%s", transaction.ReferenceNumber),
			"transaction", transaction,
			"request", transaction,
		)
	}
	return transaction, params, nil
}
