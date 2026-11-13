package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
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

type recipeController struct {
	recipeService         services.RecipeService
	recipeGuideService    services.RecipeGuideService
	recipeItemService     services.RecipeItemService
	productService        services.ProductService
	productVariantService services.ProductVariantService
	userService           services.UserService
	jsonWebTokenService   services.JsonWebTokenService
}

func NewRecipeController(
	recipeService services.RecipeService,
	recipeGuideService services.RecipeGuideService,
	recipeItemService services.RecipeItemService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *recipeController {
	return &recipeController{
		recipeService:         recipeService,
		recipeGuideService:    recipeGuideService,
		recipeItemService:     recipeItemService,
		productService:        productService,
		productVariantService: productVariantService,
		userService:           userService,
		jsonWebTokenService:   jsonWebTokenService,
	}
}

func (q *recipeController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete).
		Post("/:id/guides", q.V1SellerCreateGuide).
		Get("/:id/guides/:guide_id", q.V1SellerShowGuide).
		Post("/:id/guides/:guide_id", q.V1SellerUpdateGuide).
		Post("/:id/guides/:guide_id/delete", q.V1SellerDeleteGuide).
		Post("/:id/items", q.V1SellerCreateItem).
		Get("/:id/items/:item_id", q.V1SellerShowItem).
		Post("/:id/items/:item_id", q.V1SellerUpdateItem).
		Post("/:id/items/:item_id/delete", q.V1SellerDeleteItem)
}

func (q *recipeController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerIndex"
	ctx := c.Context()
	request, err := sanitizers.FindRecipes(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipes")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.recipeService.FindRecipes(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipes")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
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

func (q *recipeController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerCreate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateRecipe(ctx, c)
	if photo == nil {
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipe")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	var recipe *models.Recipe
	if request.Name != nil {
		recipe, err = q.recipeService.FindRecipe(
			ctx,
			models.NewRecipeFilter().
				WithNames(*request.Name),
		)
		if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
			request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentUser.ID
	request.CreatedAt = now
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = now
	if recipe != nil {
		if recipe.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueRecipeNameViolation.Error()).
				WriteResponse(c)
		}
		request.ID = recipe.ID
		request.Name = recipe.DeletedName
		request.DeletedName = nil
		request.CreatedBy = recipe.CreatedBy
		request.CreatedAt = recipe.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
		if photo == nil {
			request.SetPhoto(recipe.Photo)
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
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
	saved, err := q.recipeService.SaveRecipe(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipe")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if recipe != nil &&
		recipe.Photo != os.Getenv("RECIPE_DEFAULT_PHOTO") &&
		recipe.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(recipe.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(recipe.Photo)
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("recipe created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerShow"
	ctx := c.Context()
	recipe, err := q.prepareRecipe(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if recipe.Guides, _, err = q.recipeGuideService.FindRecipeGuides(
		ctx,
		models.NewRecipeGuideFilter().
			WithRecipeIDs(recipe.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.DefaultAscendingSortBy),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeGuides")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if recipe.Items, _, err = q.recipeItemService.FindRecipeItems(
		ctx,
		models.NewRecipeItemFilter().
			WithRecipeIDs(recipe.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.DefaultAscendingSortBy),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(recipe).
		WriteResponse(c)
}

func (q *recipeController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerUpdate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipe, err := q.prepareRecipe(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateRecipe(ctx, c)
	if photo == nil {
		request.SetPhoto(recipe.Photo)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipe")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if recipe.Name != nil && request.Name != nil && *recipe.Name != *request.Name {
		otherRecipe, err := q.recipeService.FindRecipe(
			ctx,
			models.NewRecipeFilter().
				WithNames(*request.Name).
				WithExcludedRecipeIDs(recipe.ID),
		)
		if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherRecipe != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueRecipeNameViolation.Error()).
				WriteResponse(c)
		}
	}
	request.ID = recipe.ID
	request.RecordStatus = recipe.RecordStatus
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
	saved, err := q.recipeService.UpdateRecipe(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipe")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if recipe.Photo != os.Getenv("RECIPE_DEFAULT_PHOTO") &&
		recipe.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(recipe.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
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
		SetMessage("recipe updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerDelete"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipe, err := q.prepareRecipe(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	recipe.RecordStatus = models.RecordStatusDeleted
	if recipe.Name != nil {
		recipe.DeletedName = new(*recipe.Name)
	}
	recipe.Name = nil
	recipe.DeletedBy = &currentUser.ID
	recipe.DeletedAt = new(time.Now())
	recipe.DeletionReason = nil
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
	saved, err := q.recipeService.UpdateRecipe(ctx, tx, recipe)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipe")
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
		SetMessage("recipe deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) V1SellerCreateGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerCreateGuide"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipe, _, err := q.prepareRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		return helper.NewResponse(statusCode).
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
	saved, err := q.recipeGuideService.CreateRecipeGuide(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRecipeGuide")
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
		SetMessage("recipe guide created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) V1SellerShowGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerShowGuide"
	ctx := c.Context()
	_, recipeGuide, err := q.prepareRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(recipeGuide).
		WriteResponse(c)
}

func (q *recipeController) V1SellerUpdateGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerUpdateGuide"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	_, recipeGuide, err := q.prepareRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.ID = recipeGuide.ID
	request.RecordStatus = recipeGuide.RecordStatus
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
	saved, err := q.recipeGuideService.UpdateRecipeGuide(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
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
		SetMessage("recipe guide updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) V1SellerDeleteGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerDeleteGuide"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	_, recipeGuide, err := q.prepareRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	recipeGuide.RecordStatus = models.RecordStatusDeleted
	recipeGuide.DeletedBy = new(currentUser.ID)
	recipeGuide.DeletedAt = new(time.Now())
	recipeGuide.DeletionReason = nil
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
	saved, err := q.recipeGuideService.UpdateRecipeGuide(ctx, tx, recipeGuide)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
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
		SetMessage("recipe guide deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) V1SellerCreateItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerCreateItem"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipe, _, products, err := q.prepareRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	mapVariants := map[uuid.UUID]*models.ProductVariant{}
	for _, product := range products {
		for _, variant := range product.Variants {
			mapVariants[variant.ID] = variant
		}
	}
	var recipeItem *models.RecipeItem
	variant, ok := mapVariants[request.ProductVariantID]
	if !ok {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("product_variant_id: not found").
			WriteResponse(c)
	}
	request.ProductID = variant.ProductID
	request.ProductName = variant.ProductName
	request.ProductPhoto = variant.ProductPhoto
	request.ProductVariantStockUnit = variant.StockUnit
	if recipeItem, err = q.recipeItemService.FindRecipeItem(
		ctx,
		models.NewRecipeItemFilter().
			WithRecipeIDs(recipe.ID).
			WithProductIDs(request.ProductID),
	); err != nil && !errors.Is(err, models.ErrRecipeItemNotFound) {
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
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
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
	if err = tx.Commit(ctx); err != nil {
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

func (q *recipeController) V1SellerShowItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerShowItem"
	ctx := c.Context()
	_, recipeItem, products, err := q.prepareRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(models.NewMap().Push(
			"recipe_item", recipeItem,
			"products", products,
		)).
		WriteResponse(c)
}

func (q *recipeController) V1SellerUpdateItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerUpdateItem"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	_, recipeItem, products, err := q.prepareRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	mapVariants := map[uuid.UUID]*models.ProductVariant{}
	for _, product := range products {
		for _, variant := range product.Variants {
			mapVariants[variant.ID] = variant
		}
	}
	if variant, ok := mapVariants[recipeItem.ProductVariantID]; ok {
		request.ProductID = variant.ProductID
		request.ProductName = variant.ProductName
		request.ProductPhoto = variant.ProductPhoto
		request.ProductVariantStockUnit = variant.StockUnit
	}
	request.ID = recipeItem.ID
	request.ProductVariantID = recipeItem.ProductVariantID
	request.RecordStatus = recipeItem.RecordStatus
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
	saved, err := q.recipeItemService.UpdateRecipeItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
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
		SetMessage("recipe item updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) V1SellerDeleteItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerDeleteItem"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	_, recipeItem, _, err := q.prepareRecipeItem(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return helper.NewResponse(notFoundStatusCode(err)).
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
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	saved, err := q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
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
		SetMessage("recipe item deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *recipeController) prepareRecipe(
	ctx context.Context,
	c fiber.Ctx,
) (
	*models.Recipe,
	error,
) {
	recipe := &models.Recipe{}
	recipe.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
	if c.Params("id") != "" {
		var err error
		if recipe, err = q.prepareRecipeByID(ctx, c); err != nil {
			return nil, err
		}
	}
	return recipe, nil
}

func (q *recipeController) prepareRecipeByID(
	ctx context.Context,
	c fiber.Ctx,
) (
	*models.Recipe,
	error,
) {
	recipeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, models.ErrRecipeNotFound
	}
	recipe, err := q.recipeService.FindRecipe(
		ctx,
		models.NewRecipeFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeIDs(recipeID),
	)
	if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
		return nil, err
	}
	if recipe == nil {
		return nil, models.ErrRecipeNotFound
	}
	return recipe, nil
}

func (q *recipeController) prepareRecipeGuide(
	ctx context.Context,
	c fiber.Ctx,
) (
	*models.Recipe,
	*models.RecipeGuide,
	error,
) {
	recipe, err := q.prepareRecipeByID(ctx, c)
	if err != nil {
		return nil, nil, err
	}
	recipeGuide := &models.RecipeGuide{}
	if c.Params("guide_id") != "" {
		recipeGuideID, err := uuid.Parse(c.Params("guide_id"))
		if err != nil {
			return nil, nil, models.ErrRecipeGuideNotFound
		}
		if recipeGuide, err = q.recipeGuideService.FindRecipeGuide(
			ctx,
			models.NewRecipeGuideFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithRecipeIDs(recipe.ID).
				WithRecipeGuideIDs(recipeGuideID),
		); err != nil && !errors.Is(err, models.ErrRecipeGuideNotFound) {
			return nil, nil, err
		}
		if recipeGuide == nil {
			return nil, nil, models.ErrRecipeGuideNotFound
		}
	}
	return recipe, recipeGuide, nil
}

func (q *recipeController) prepareRecipeItem(
	ctx context.Context,
	c fiber.Ctx,
) (
	*models.Recipe,
	*models.RecipeItem,
	[]*models.Product,
	error,
) {
	recipe, err := q.prepareRecipeByID(ctx, c)
	if err != nil {
		return nil, nil, nil, err
	}
	recipeItem := &models.RecipeItem{}
	recipeItems, _, err := q.recipeItemService.FindRecipeItems(
		ctx,
		models.NewRecipeItemFilter().
			WithRecipeIDs(recipe.ID).
			WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	var productIDs []uuid.UUID
	for _, recipeItem := range recipeItems {
		productIDs = append(productIDs, recipeItem.ProductID)
	}
	if c.Params("item_id") != "" {
		recipeItemID, err := uuid.Parse(c.Params("item_id"))
		if err != nil {
			return nil, nil, nil, models.ErrRecipeItemNotFound
		}
		if recipeItem, err = q.recipeItemService.FindRecipeItem(
			ctx,
			models.NewRecipeItemFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithRecipeIDs(recipe.ID).
				WithRecipeItemIDs(recipeItemID),
		); err != nil && !errors.Is(err, models.ErrRecipeItemNotFound) {
			return nil, nil, nil, err
		}
		if recipeItem == nil {
			return nil, nil, nil, models.ErrRecipeItemNotFound
		}
		oldProductIDs := make([]uuid.UUID, len(productIDs))
		copy(oldProductIDs, productIDs)
		productIDs = productIDs[:0]
		for _, oldProductID := range oldProductIDs {
			if recipeItem.ProductID.Compare(oldProductID) != 0 {
				productIDs = append(productIDs, oldProductID)
			}
		}
	}
	products, _, err := q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithExcludedProductIDs(productIDs...).
			WithSortings(models.Sorting{
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	productIDs = make([]uuid.UUID, len(products))
	for i := range products {
		productIDs[i] = products[i].ID
	}
	variants, _, err := q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productIDs...),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	mapProductVariants := map[uuid.UUID][]*models.ProductVariant{}
	for _, variant := range variants {
		mapProductVariants[variant.ProductID] = append(mapProductVariants[variant.ProductID], variant)
	}
	for i := range products {
		products[i].Variants = mapProductVariants[products[i].ID]
	}
	return recipe, recipeItem, products, nil
}
