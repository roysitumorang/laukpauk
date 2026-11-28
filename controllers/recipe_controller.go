package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/database"
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
		Put("/:id", q.V1SellerUpdate).
		Delete("/:id", q.V1SellerDelete)
	r.Group("/v1/buyer").
		Get("", q.V1BuyerIndex).
		Get("/:id", q.V1BuyerShow)
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
		if errors.Is(err, models.ErrRecipeNotFound) {
			err = nil
		}
		if err != nil {
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
		if err = tx.Rollback(ctx); errors.Is(err, database.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.recipeService.SaveRecipe(ctx, tx, request)
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
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("recipe created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *recipeController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerShow"
	ctx := c.Context()
	recipeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	recipe, err := q.recipeService.FindRecipe(
		ctx,
		models.NewRecipeFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeIDs(recipeID),
	)
	if errors.Is(err, models.ErrRecipeNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
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
	recipeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	recipe, err := q.recipeService.FindRecipe(
		ctx,
		models.NewRecipeFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeIDs(recipeID),
	)
	if errors.Is(err, models.ErrRecipeNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
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
		if err = tx.Rollback(ctx); errors.Is(err, database.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.recipeService.UpdateRecipe(ctx, tx, request)
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
		SetData(response).
		WriteResponse(c)
}

func (q *recipeController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1SellerDelete"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	recipe, err := q.recipeService.FindRecipe(
		ctx,
		models.NewRecipeFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeIDs(recipeID),
	)
	if errors.Is(err, models.ErrRecipeNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
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
		if err = tx.Rollback(ctx); errors.Is(err, database.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeService.UpdateRecipe(ctx, tx, recipe); err != nil {
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
	return helper.NewResponse(fiber.StatusNoContent).
		WriteResponse(c)
}

func (q *recipeController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1BuyerIndex"
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
		rows[i].Photo = rows[i].PhotoPath(c.BaseURL())
		pagination.Rows[i] = rows[i]
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}

func (q *recipeController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-V1BuyerShow"
	ctx := c.Context()
	recipeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.recipeService.FindRecipe(
		ctx,
		models.NewRecipeFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeIDs(recipeID),
	)
	if errors.Is(err, models.ErrRecipeNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeNotFound.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response.Guides, _, err = q.recipeGuideService.FindRecipeGuides(
		ctx,
		models.NewRecipeGuideFilter().
			WithRecipeIDs(response.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.DefaultAscendingSortBy),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeGuides")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response.Items, _, err = q.recipeItemService.FindRecipeItems(
		ctx,
		models.NewRecipeItemFilter().
			WithRecipeIDs(response.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.DefaultAscendingSortBy),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response.Photo = response.PhotoPath(c.BaseURL())
	for i := range response.Items {
		response.Items[i].ProductPhoto = response.Items[i].ProductPhotoPath(c.BaseURL())
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
