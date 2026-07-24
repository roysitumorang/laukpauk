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
	"github.com/roysitumorang/laukpauk/middleware"
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
		cartItemService       services.CartItemService
		jsonWebTokenService   services.JsonWebTokenService
	}
)

func NewCartController(
	userService services.BuyerService,
	productVariantService services.ProductVariantService,
	cartService services.CartService,
	cartItemService services.CartItemService,
	jsonWebTokenService services.JsonWebTokenService,
) *cartController {
	return &cartController{
		buyerService:          userService,
		productVariantService: productVariantService,
		cartService:           cartService,
		cartItemService:       cartItemService,
		jsonWebTokenService:   jsonWebTokenService,
	}
}

func (q *cartController) Mount(r fiber.Router) {
	buyerKeyAuth := middleware.BuyerKeyAuth(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex).
		Post("/items", q.V1BuyerSave).
		Delete("/items/:id", q.V1BuyerDelete)
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
	cartItems, _, err := q.cartItemService.FindCartItems(
		ctx,
		models.NewCartItemFilter().
			WithCartIDs(response.ID).
			WithSortings(models.Sorting{
				Field:     models.SortingFieldID,
				Direction: models.SortingDirectionAsc,
			}),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCartItems")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response.Items = make([]*models.CartItem, len(cartItems))
	for i, cartItem := range cartItems {
		response.Items[i] = cartItem
		response.Quantity += cartItem.Quantity
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *cartController) V1BuyerSave(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CartController-V1BuyerSave"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateCartItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateCartItem")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductVariantIDs(request.ProductVariantID),
	)
	if err != nil && !errors.Is(err, models.ErrProductVariantNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if productVariant == nil {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage("product_variant_id: not found").
			WriteResponse(c)
	}
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
	now := time.Now().In(time.Local)
	if response == nil {
		if response, err = q.cartService.SaveCart(
			ctx,
			tx,
			&models.Cart{
				BuyerID:   &currentBuyer.ID,
				CreatedAt: now,
				UpdatedAt: now,
			},
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveCart")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	request.CartID = response.ID
	request.CreatedAt = now
	request.UpdatedAt = now
	if _, err = q.cartItemService.SaveCartItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveCartItem")
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
	cartItems, _, err := q.cartItemService.FindCartItems(
		ctx,
		models.NewCartItemFilter().
			WithCartIDs(response.ID).
			WithSortings(models.Sorting{
				Field:     models.SortingFieldID,
				Direction: models.SortingDirectionAsc,
			}),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCartItems")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response.Items = make([]*models.CartItem, len(cartItems))
	for i, cartItem := range cartItems {
		response.Items[i] = cartItem
		response.Quantity += cartItem.Quantity
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *cartController) V1BuyerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CartController-V1BuyerDelete"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	cartItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrCartItemNotFound.Error()).
			WriteResponse(c)
	}
	cart, err := q.cartService.FindCart(
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
	if cart == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrCartItemNotFound.Error()).
			WriteResponse(c)
	}
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
	if _, err = q.cartItemService.DeleteCartItem(
		ctx,
		tx,
		models.NewCartItemFilter().
			WithCartIDs(cart.ID).
			WithCartItemIDs(cartItemID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteCartItem")
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
	return helper.NewResponse(fiber.StatusNoContent).WriteResponse(c)
}
