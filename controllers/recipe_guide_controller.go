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

type recipeGuideController struct {
	recipeService       services.RecipeService
	recipeGuideService  services.RecipeGuideService
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewRecipeGuideController(
	recipeService services.RecipeService,
	recipeGuideService services.RecipeGuideService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *recipeGuideController {
	return &recipeGuideController{
		recipeService:       recipeService,
		recipeGuideService:  recipeGuideService,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *recipeGuideController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Put("/:id", q.V1SellerUpdate).
		Delete("/:id", q.V1SellerDelete)
}

func (q *recipeGuideController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "RecipeGuideController-V1SellerCreate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.recipeGuideService.CreateRecipeGuide(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRecipeGuide")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("recipe guide created successfully").
		SetData(response).
		JSON(c)
}

func (q *recipeGuideController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "RecipeGuideController-V1SellerShow"
	ctx := c.Context()
	recipeGuideID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeGuideNotFound.Error()).
			JSON(c)
	}
	response, err := q.recipeGuideService.FindRecipeGuide(
		ctx,
		models.NewRecipeGuideFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeGuideIDs(recipeGuideID),
	)
	if errors.Is(err, models.ErrRecipeGuideNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeGuide")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		JSON(c)
}

func (q *recipeGuideController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "RecipeGuideController-V1SellerUpdate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipeGuideID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeGuideNotFound.Error()).
			JSON(c)
	}
	recipeGuide, err := q.recipeGuideService.FindRecipeGuide(
		ctx,
		models.NewRecipeGuideFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeGuideIDs(recipeGuideID),
	)
	if errors.Is(err, models.ErrRecipeGuideNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeGuide")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.recipeGuideService.UpdateRecipeGuide(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("recipe guide updated successfully").
		SetData(response).
		JSON(c)
}

func (q *recipeGuideController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "RecipeGuideController-V1SellerDelete"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	recipeGuideID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrRecipeGuideNotFound.Error()).
			JSON(c)
	}
	recipeGuide, err := q.recipeGuideService.FindRecipeGuide(
		ctx,
		models.NewRecipeGuideFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRecipeGuideIDs(recipeGuideID),
	)
	if errors.Is(err, models.ErrRecipeGuideNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeGuide")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err := q.recipeGuideService.UpdateRecipeGuide(ctx, tx, recipeGuide); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusNoContent).
		SendStatus(c)
}
