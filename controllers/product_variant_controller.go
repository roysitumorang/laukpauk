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

type productVariantController struct {
	productCategoryService services.ProductCategoryService
	productService         services.ProductService
	productVariantService  services.ProductVariantService
	recipeItemService      services.RecipeItemService
	userService            services.UserService
	jsonWebTokenService    services.JsonWebTokenService
}

func NewProductVariantController(
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
	recipeItemService services.RecipeItemService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *productVariantController {
	return &productVariantController{
		productCategoryService: productCategoryService,
		productService:         productService,
		productVariantService:  productVariantService,
		recipeItemService:      recipeItemService,
		userService:            userService,
		jsonWebTokenService:    jsonWebTokenService,
	}
}

func (q *productVariantController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Put("/:id", q.V1SellerUpdate).
		Delete("/:id", q.V1SellerDelete)
}

func (q *productVariantController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "ProductVariantController-V1SellerCreate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(request.ProductID),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage("product_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(product.ProductCategoryID),
	)
	if errors.Is(err, models.ErrProductCategoryNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage("product_category_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithProductIDs(product.ID).
			WithStockUnits(request.StockUnit),
	)
	if errors.Is(err, models.ErrProductVariantNotFound) {
		err = nil
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.ProductCategoryID = productCategory.ID
	request.ProductCategoryName = productCategory.Name
	request.ProductID = product.ID
	request.ProductName = product.Name
	request.ProductPhoto = product.Photo
	request.ProductReadyStock = product.ReadyStock
	request.CreatedBy = currentUser.ID
	request.CreatedAt = now
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = now
	if productVariant != nil {
		if productVariant.RecordStatus == models.RecordStatusActive {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueProductVariantStockUnitViolation.Error()).
				WriteResponse(c)
		}
		request.ID = productVariant.ID
		request.StockUnit = productVariant.StockUnit
		request.CreatedBy = productVariant.CreatedBy
		request.CreatedAt = productVariant.CreatedAt
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
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.productVariantService.SaveProductVariant(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProductVariant")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request.SellingPrice < product.LowestSellingPrice {
		product.LowestSellingPrice = request.SellingPrice
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("product variant created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *productVariantController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "ProductVariantController-V1SellerShow"
	ctx := c.Context()
	productVariantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrProductVariantNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithProductVariantIDs(productVariantID).
			WithRecordStatuses(models.RecordStatusActive),
	)
	if errors.Is(err, models.ErrProductVariantNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *productVariantController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "ProductVariantController-V1SellerUpdate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	productVariantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrProductVariantNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithProductVariantIDs(productVariantID).
			WithRecordStatuses(models.RecordStatusActive),
	)
	if errors.Is(err, models.ErrProductVariantNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
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
			WithProductIDs(productVariant.ProductID),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage("product_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(product.ProductCategoryID),
	)
	if errors.Is(err, models.ErrProductCategoryNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage("product_category_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if productVariant.StockUnit != request.StockUnit {
		otherProductVariant, err := q.productVariantService.FindProductVariant(
			ctx,
			models.NewProductVariantFilter().
				WithProductIDs(product.ID).
				WithStockUnits(request.StockUnit).
				WithExcludedProductVariantIDs(productVariant.ID),
		)
		if errors.Is(err, models.ErrProductVariantNotFound) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherProductVariant != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueProductVariantStockUnitViolation.Error()).
				WriteResponse(c)
		}
	}
	recipeItems, _, err := q.recipeItemService.FindRecipeItems(
		ctx,
		models.NewRecipeItemFilter().
			WithProductVariantIDs(productVariant.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.ID = productVariant.ID
	request.RecordStatus = productVariant.RecordStatus
	request.ProductCategoryID = productCategory.ID
	request.ProductCategoryName = productCategory.Name
	request.ProductID = product.ID
	request.ProductName = product.Name
	request.ProductPhoto = product.Photo
	request.ProductReadyStock = productVariant.ProductReadyStock
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
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.productVariantService.UpdateProductVariant(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request.SellingPrice < product.LowestSellingPrice {
		product.LowestSellingPrice = request.SellingPrice
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	for _, recipeItem := range recipeItems {
		recipeItem.ProductVariantStockUnit = response.StockUnit
		if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("product variant updated successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *productVariantController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "ProductVariantController-V1SellerDelete"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	productVariantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrProductVariantNotFound.Error()).
			WriteResponse(c)
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithProductVariantIDs(productVariantID).
			WithRecordStatuses(models.RecordStatusActive),
	)
	if errors.Is(err, models.ErrProductVariantNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productVariant.RecordStatus = models.RecordStatusDeleted
	productVariant.DeletedBy = new(currentUser.ID)
	productVariant.DeletedAt = new(time.Now())
	productVariant.DeletionReason = nil
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
	if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
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
	return helper.NewResponse(fiber.StatusNoContent).
		WriteResponse(c)
}
