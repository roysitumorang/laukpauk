package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	cartController struct {
		buyerService          services.BuyerService
		productVariantService services.ProductVariantService
		cartService           services.CartService
		jsonWebTokenService   services.JsonWebTokenService
	}
)

func NewCartController(
	userService services.BuyerService,
	productVariantService services.ProductVariantService,
	cartService services.CartService,
	jsonWebTokenService services.JsonWebTokenService,
) *cartController {
	return &cartController{
		buyerService:          userService,
		productVariantService: productVariantService,
		cartService:           cartService,
		jsonWebTokenService:   jsonWebTokenService,
	}
}

func (q *cartController) Mount(r fiber.Router) {
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex).
		Post("", q.V1BuyerSave)
}

func (q *cartController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CartController-V1BuyerIndex"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	response, err := q.cartService.FindCart(
		ctx,
		models.NewCartFilter().
			WithBuyerIDs(currentBuyer.ID),
	)
	if err != nil && !errors.Is(err, models.ErrCartNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCart")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		response = &models.Cart{
			BuyerID: &currentBuyer.ID,
		}
		return helper.NewResponse(fiber.StatusOK).
			SetData(response).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *cartController) V1BuyerSave(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CartController-V1BuyerSave"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateCart(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateCart")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productVariantIDs := make([]uuid.UUID, len(request.Items))
	for i := range request.Items {
		productVariantIDs[i] = request.Items[i].ProductVariantID
	}
	productVariants, _, err := q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductVariantIDs(productVariantIDs...),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	mapProductVariantIDs := make(map[uuid.UUID]bool, len(productVariants))
	for _, productVariant := range productVariants {
		mapProductVariantIDs[productVariant.ID] = productVariant.ProductReadyStock
	}
	for i, cartItem := range request.Items {
		if readyStock, ok := mapProductVariantIDs[cartItem.ProductVariantID]; !ok || !readyStock {
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(helper.Sprintf("items[%d].product_variant_id: not found or out of stock", i)).
				WriteResponse(c)
		}
	}
	now := time.Now().In(time.Local)
	request.BuyerID = &currentBuyer.ID
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
	response, err := q.cartService.SaveCart(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveCart")
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
		SetData(response).
		WriteResponse(c)
}
