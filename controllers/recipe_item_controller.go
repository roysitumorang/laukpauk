package controllers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type recipeItemController struct {
	recipeService         services.RecipeService
	recipeItemService     services.RecipeItemService
	productService        services.ProductService
	productVariantService services.ProductVariantService
	userService           services.UserService
	jsonWebTokenService   services.JsonWebTokenService
}

func NewRecipeItemController(
	recipeService services.RecipeService,
	recipeItemService services.RecipeItemService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *recipeItemController {
	return &recipeItemController{
		recipeService:         recipeService,
		recipeItemService:     recipeItemService,
		productService:        productService,
		productVariantService: productVariantService,
		userService:           userService,
		jsonWebTokenService:   jsonWebTokenService,
	}
}

func (q *recipeItemController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Put("/:id", q.V1SellerUpdate).
		Delete("/:id", q.V1SellerDelete)
}

func (q *recipeItemController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerCreate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	recipe, err := q.recipeService.FindRecipe(
		ctx,
		models.NewRecipeFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeIDs(request.RecipeID),
	)
	if errors.Is(err, models.ErrRecipeNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("recipe_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	variant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductVariantIDs(request.ProductVariantID),
	)
	if errors.Is(err, models.ErrProductVariantNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("product_variant_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(variant.ProductID),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("product_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.ProductID = product.ID
	request.ProductName = product.Name
	request.ProductPhoto = product.Photo
	request.ProductVariantStockUnit = variant.StockUnit
	recipeItem, err := q.recipeItemService.FindRecipeItem(
		ctx,
		models.NewRecipeItemFilter().
			WithRecipeIDs(recipe.ID).
			WithProductIDs(request.ProductID),
	)
	if errors.Is(err, models.ErrRecipeItemNotFound) {
		err = nil
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItem")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.RecipeID = recipe.ID
	request.CreatedBy = currentUser.ID
	request.CreatedAt = now
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = now
	if recipeItem != nil {
		if recipeItem.RecordStatus == models.RecordStatusActive {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueRecipeItemProductIdViolation.Error()).
				WriteResponse(c)
		}
		request.ID = recipeItem.ID
		request.CreatedBy = recipeItem.CreatedBy
		request.CreatedAt = recipeItem.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(); errors.Is(err, sql.ErrTxDone) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	saved, err := q.recipeItemService.SaveRecipeItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipeItem")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("recipe item created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeItemController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerShow"
	ctx := c.Context()
	recipeItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeItemNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.recipeItemService.FindRecipeItem(
		ctx,
		models.NewRecipeItemFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeItemIDs(recipeItemID),
	)
	if errors.Is(err, models.ErrRecipeItemNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItem")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *recipeItemController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerUpdate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipeItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeItemNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	recipeItem, err := q.recipeItemService.FindRecipeItem(
		ctx,
		models.NewRecipeItemFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeItemIDs(recipeItemID),
	)
	if errors.Is(err, models.ErrRecipeItemNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItem")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	variant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductVariantIDs(recipeItem.ProductVariantID),
	)
	if errors.Is(err, models.ErrProductVariantNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("product_variant_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(variant.ProductID),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("product_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.ID = recipeItem.ID
	request.RecordStatus = recipeItem.RecordStatus
	request.ProductVariantID = recipeItem.ProductVariantID
	request.ProductID = product.ID
	request.ProductName = product.Name
	request.ProductPhoto = product.Photo
	request.ProductVariantStockUnit = variant.StockUnit
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(); errors.Is(err, sql.ErrTxDone) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.recipeItemService.UpdateRecipeItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("recipe item updated successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *recipeItemController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerDelete"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipeItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeItemNotFound.Error()).
			WriteResponse(c)
	}
	recipeItem, err := q.recipeItemService.FindRecipeItem(
		ctx,
		models.NewRecipeItemFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeItemIDs(recipeItemID),
	)
	if errors.Is(err, models.ErrRecipeItemNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItem")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	recipeItem.RecordStatus = models.RecordStatusDeleted
	recipeItem.DeletedBy = new(currentUser.ID)
	recipeItem.DeletedAt = new(time.Now())
	recipeItem.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(); errors.Is(err, sql.ErrTxDone) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusNoContent).
		WriteResponse(c)
}
