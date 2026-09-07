package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type recipeController struct {
	sellerService      services.SellerService
	buyerService       services.BuyerService
	recipeService      services.RecipeService
	recipeGuideService services.RecipeGuideService
	recipeItemService  services.RecipeItemService
	productService     services.ProductService
}

func NewRecipeController(
	sellerService services.SellerService,
	buyerService services.BuyerService,
	recipeService services.RecipeService,
	recipeGuideService services.RecipeGuideService,
	recipeItemService services.RecipeItemService,
	productService services.ProductService,
) *recipeController {
	return &recipeController{
		sellerService:      sellerService,
		buyerService:       buyerService,
		recipeService:      recipeService,
		recipeGuideService: recipeGuideService,
		recipeItemService:  recipeItemService,
		productService:     productService,
	}
}

func (q *recipeController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id", q.SellerShow).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete).
		Get("/:id/guides/new", q.SellerNewGuide).
		Post("/:id/guides", q.SellerCreateGuide).
		Get("/:id/guides/:guide_id/edit", q.SellerEditGuide).
		Post("/:id/guides/:guide_id", q.SellerUpdateGuide).
		Post("/:id/guides/:guide_id/delete", q.SellerDeleteGuide).
		Get("/:id/items/new", q.SellerNewItem).
		Post("/:id/items", q.SellerCreateItem).
		Get("/:id/items/:item_id/edit", q.SellerEditItem).
		Post("/:id/items/:item_id", q.SellerUpdateItem).
		Post("/:id/items/:item_id/delete", q.SellerDeleteItem)
}

func (q *recipeController) SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerIndex"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := &models.RecipeFilter{}
	rows := []*models.Recipe{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Data Resep",
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
			{
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	request, err = sanitizers.FindRecipes(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipes")
		return c.Render("seller/recipe_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.recipeService.FindRecipes(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipes")
		return c.Render("seller/recipe_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/recipe_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *recipeController) SellerNew(c fiber.Ctx) (err error) {
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _ := q.prepareRecipe(ctx, c, flash, currentSeller, csrfToken, true)
	return c.Render("seller/recipe_new", params, "seller/layouts/authenticated")
}

func (q *recipeController) SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerCreate"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _ := q.prepareRecipe(ctx, c, flash, currentSeller, csrfToken, true)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateRecipe(ctx, c)
	if photo == nil {
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
	}
	params.Push(
		"request", request,
		"recipe", request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipe")
		c.Response().SetStatusCode(statusCode)
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return c.Render("seller/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "seller/layouts/authenticated")
	}
	var recipe *models.Recipe
	if request.Name != nil {
		recipe, err = q.recipeService.FindRecipe(
			ctx,
			models.NewRecipeFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithNames(*request.Name),
		)
		if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
			request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
			return c.Render("seller/recipe_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"recipe", request,
			), "seller/layouts/authenticated")
		}
	}
	now := time.Now()
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentSeller.ID
	request.CreatedAt = now
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = now
	if recipe != nil {
		if recipe.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
			return c.Render("seller/recipe_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueRecipeNameViolation.Error()),
				"request", request,
				"recipe", request,
			), "seller/layouts/authenticated")
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
		return c.Render("seller/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "seller/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeService.SaveRecipe(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipe")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return c.Render("seller/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "seller/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return c.Render("seller/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "seller/layouts/authenticated")
	}
	if recipe != nil &&
		recipe.Photo != os.Getenv("RECIPE_DEFAULT_PHOTO") &&
		recipe.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(recipe.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(recipe.Photo)
			return c.Render("seller/recipe_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"recipe", request,
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return c.Render("seller/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "seller/layouts/authenticated")
	}
	return flash.Success("recipe created successfully").Redirect(c, sess.Session, "/seller/recipes")
}

func (q *recipeController) SellerShow(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerShow"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	return c.Render("seller/recipe_show", params, "seller/layouts/authenticated")
}

func (q *recipeController) SellerEdit(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerEdit"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	return c.Render("seller/recipe_edit", params, "seller/layouts/authenticated")
}

func (q *recipeController) SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerUpdate"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, params, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateRecipe(ctx, c)
	if photo == nil {
		request.SetPhoto(recipe.Photo)
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipe")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "seller/layouts/authenticated")
	}
	if recipe.Name != nil && request.Name != nil && *recipe.Name != *request.Name {
		otherRecipe, err := q.recipeService.FindRecipe(
			ctx,
			models.NewRecipeFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithNames(*request.Name).
				WithExcludedRecipeIDs(recipe.ID),
		)
		if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
			return c.Render("seller/recipe_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(recipe.Photo),
			), "seller/layouts/authenticated")
		}
		if otherRecipe != nil {
			return c.Render("seller/recipe_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueRecipeNameViolation.Error()),
				"request", request.SetPhoto(recipe.Photo),
			), "seller/layouts/authenticated")
		}
	}
	request.ID = recipe.ID
	request.CompanyID = recipe.CompanyID
	request.RecordStatus = recipe.RecordStatus
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "seller/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeService.UpdateRecipe(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipe")
		return c.Render("seller/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "seller/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("seller/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "seller/layouts/authenticated")
	}
	if recipe.Photo != os.Getenv("RECIPE_DEFAULT_PHOTO") &&
		recipe.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(recipe.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			return c.Render("seller/recipe_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(recipe.Photo),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "seller/layouts/authenticated")
	}
	return flash.Success("recipe updated successfully").Redirect(c, sess.Session, "/seller/recipes")
}

func (q *recipeController) SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerDelete"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, _, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	recipe.RecordStatus = models.RecordStatusDeleted
	if recipe.Name != nil {
		recipe.DeletedName = new(*recipe.Name)
	}
	recipe.Name = nil
	recipe.DeletedBy = &currentSeller.ID
	recipe.DeletedAt = new(time.Now())
	recipe.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeService.UpdateRecipe(ctx, tx, recipe); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	return flash.Success("recipe deleted successfully").Redirect(c, sess.Session, "/seller/recipes")
}

func (q *recipeController) SellerNewGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerNewGuide"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	return c.Render("seller/recipe_guide_new", params, "seller/layouts/authenticated")
}

func (q *recipeController) SellerCreateGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerCreateGuide"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, _, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/recipe_guide_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = models.RecordStatusActive
	request.RecipeID = recipe.ID
	request.CreatedBy = currentSeller.ID
	request.CreatedAt = now
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/recipe_guide_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeGuideService.CreateRecipeGuide(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRecipeGuide")
		return c.Render("seller/recipe_guide_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/recipe_guide_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Success("recipe guide created successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
}

func (q *recipeController) SellerEditGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerEditGuide"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	return c.Render("seller/recipe_guide_edit", params, "seller/layouts/authenticated")
}

func (q *recipeController) SellerUpdateGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerUpdateGuide"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeGuide, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/recipe_guide_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	request.ID = recipeGuide.ID
	request.CompanyID = recipeGuide.CompanyID
	request.RecordStatus = recipeGuide.RecordStatus
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/recipe_guide_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeGuideService.UpdateRecipeGuide(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
		return c.Render("seller/recipe_guide_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/recipe_guide_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Success("recipe guide updated successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
}

func (q *recipeController) SellerDeleteGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerDeleteGuide"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeGuide, _, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	recipeGuide.RecordStatus = models.RecordStatusDeleted
	recipeGuide.DeletedBy = new(currentSeller.ID)
	recipeGuide.DeletedAt = new(time.Now())
	recipeGuide.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeGuideService.UpdateRecipeGuide(ctx, tx, recipeGuide); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
	}
	return flash.Success("recipe guide deleted successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
}

func (q *recipeController) SellerNewItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerNewItem"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, _, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	return c.Render("seller/recipe_item_new", params, "seller/layouts/authenticated")
}

func (q *recipeController) SellerCreateItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerCreateItem"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, _, products, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	mapProducts := make(map[uuid.UUID]*models.Product, len(products))
	for _, product := range products {
		mapProducts[product.ID] = product
	}
	var recipeItem *models.RecipeItem
	product, ok := mapProducts[request.ProductID]
	if !ok {
		return c.Render("seller/recipe_item_new", params.Push(
			"flash", flash.Danger("product_id: not found"),
		), "seller/layouts/authenticated")
	}
	request.ProductName = product.Name
	request.ProductPhoto = product.Photo
	if recipeItem, err = q.recipeItemService.FindRecipeItem(
		ctx,
		models.NewRecipeItemFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecipeIDs(recipe.ID).
			WithProductIDs(request.ProductID),
	); err != nil && !errors.Is(err, models.ErrRecipeItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItem")
		return c.Render("seller/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = models.RecordStatusActive
	request.RecipeID = recipe.ID
	request.CreatedBy = currentSeller.ID
	request.CreatedAt = now
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = now
	if recipeItem != nil {
		if recipeItem.RecordStatus == models.RecordStatusActive {
			return c.Render("seller/recipe_item_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueRecipeItemProductIdViolation.Error()),
			), "seller/layouts/authenticated")
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
		return c.Render("seller/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeItemService.SaveRecipeItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipeItem")
		return c.Render("seller/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Success("recipe item created successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
}

func (q *recipeController) SellerEditItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerEditItem"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, _, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	return c.Render("seller/recipe_item_edit", params, "seller/layouts/authenticated")
}

func (q *recipeController) SellerUpdateItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerUpdateItem"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeItem, products, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/recipe_item_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	mapProducts := make(map[uuid.UUID]*models.Product, len(products))
	for _, product := range products {
		mapProducts[product.ID] = product
	}
	if product, ok := mapProducts[recipeItem.ProductID]; ok {
		request.ProductName = product.Name
		request.ProductPhoto = product.Photo
	}
	request.ID = recipeItem.ID
	request.CompanyID = recipeItem.CompanyID
	request.ProductID = recipeItem.ProductID
	request.RecordStatus = recipeItem.RecordStatus
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/recipe_item_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
		return c.Render("seller/recipe_item_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/recipe_item_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Success("recipe item updated successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
}

func (q *recipeController) SellerDeleteItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-SellerDeleteItem"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeItem, _, _, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/recipes")
	}
	recipeItem.RecordStatus = models.RecordStatusDeleted
	recipeItem.DeletedBy = new(currentSeller.ID)
	recipeItem.DeletedAt = new(time.Now())
	recipeItem.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
	}
	return flash.Success("recipe item deleted successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/recipes/%s", recipe.ID))
}

func (q *recipeController) prepare(c fiber.Ctx) (
	context.Context,
	*session.Middleware,
	*helper.FlashMessage,
	*models.Seller,
	string,
) {
	ctx := c.Context()
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	csrfToken := csrf.TokenFromContext(c)
	return ctx, sess, flash, currentSeller, csrfToken
}

func (q *recipeController) prepareRecipe(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentSeller *models.Seller,
	csrfToken string,
	withParams bool,
) (
	*models.Recipe,
	models.Map,
	error,
) {
	recipe := &models.Recipe{}
	recipe.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
	recipeGuides := []*models.RecipeGuide{}
	recipeItems := []*models.RecipeItem{}
	var params models.Map
	if withParams {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentSeller", currentSeller,
			"pageTitle", "Add New Resep",
			"request", recipe,
			"recipe", recipe,
			"recipeGuides", recipeGuides,
			"recipeItems", recipeItems,
		)
	}
	if c.Params("id") != "" {
		recipeID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return nil, nil, models.ErrRecipeNotFound
		}
		if recipe, err = q.recipeService.FindRecipe(
			ctx,
			models.NewRecipeFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecordStatuses(models.RecordStatusActive).
				WithRecipeIDs(recipeID),
		); err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
			return nil, nil, err
		}
		if recipe == nil {
			return nil, nil, models.ErrRecipeNotFound
		}
		if withParams && params.Size() > 0 && recipe.Name != nil {
			params.Push(
				"pageTitle", helper.Sprintf("Edit Data Resep|%s", *recipe.Name),
				"request", recipe,
				"recipe", recipe,
			)
			if strings.HasSuffix(c.OriginalURL(), recipe.ID.String()) {
				params.Push(
					"pageTitle", helper.Sprintf("Detail Data Resep|%s", *recipe.Name),
				)
			}
			if recipeGuides, _, err = q.recipeGuideService.FindRecipeGuides(
				ctx,
				models.NewRecipeGuideFilter().
					WithCompanyIDs(currentSeller.CompanyID).
					WithRecipeIDs(recipe.ID).
					WithRecordStatuses(models.RecordStatusActive).
					WithSortings(models.DefaultAscendingSortBy),
			); err != nil {
				return nil, nil, err
			}
			if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
				ctx,
				models.NewRecipeItemFilter().
					WithCompanyIDs(currentSeller.CompanyID).
					WithRecipeIDs(recipe.ID).
					WithRecordStatuses(models.RecordStatusActive).
					WithSortings(models.DefaultAscendingSortBy),
			); err != nil {
				return nil, nil, err
			}
			params.Push(
				"recipeGuides", recipeGuides,
				"recipeItems", recipeItems,
			)
		}
	}
	return recipe, params, nil
}

func (q *recipeController) prepareRecipeByID(
	ctx context.Context,
	c fiber.Ctx,
	companyID uuid.UUID,
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
			WithCompanyIDs(companyID).
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
	flash *helper.FlashMessage,
	currentSeller *models.Seller,
	csrfToken string,
	withParams bool,
) (
	*models.Recipe,
	*models.RecipeGuide,
	models.Map,
	error,
) {
	recipe, err := q.prepareRecipeByID(ctx, c, currentSeller.CompanyID)
	if err != nil {
		return nil, nil, nil, err
	}
	recipeGuide := &models.RecipeGuide{}
	var params models.Map
	if withParams && recipe.Name != nil {
		recipeGuides := []*models.RecipeGuide{}
		recipeItems := []*models.RecipeItem{}
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentSeller", currentSeller,
			"pageTitle", helper.Sprintf("Detail Resep|%s", *recipe.Name),
			"recipe", recipe,
			"recipeGuide", recipeGuide,
			"request", recipeGuide,
			"recipeGuides", recipeGuides,
			"recipeItems", recipeItems,
		)
		if recipeGuides, _, err = q.recipeGuideService.FindRecipeGuides(
			ctx,
			models.NewRecipeGuideFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecipeIDs(recipe.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.DefaultAscendingSortBy),
		); err != nil {
			return nil, nil, nil, err
		}
		if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
			ctx,
			models.NewRecipeItemFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecipeIDs(recipe.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.DefaultAscendingSortBy),
		); err != nil {
			return nil, nil, nil, err
		}
		params.Push(
			"recipeGuides", recipeGuides,
			"recipeItems", recipeItems,
		)
	}
	if c.Params("guide_id") != "" {
		recipeGuideID, err := uuid.Parse(c.Params("guide_id"))
		if err != nil {
			return nil, nil, nil, models.ErrRecipeGuideNotFound
		}
		if recipeGuide, err = q.recipeGuideService.FindRecipeGuide(
			ctx,
			models.NewRecipeGuideFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecordStatuses(models.RecordStatusActive).
				WithRecipeIDs(recipe.ID).
				WithRecipeGuideIDs(recipeGuideID),
		); err != nil {
			return nil, nil, nil, err
		}
		if recipeGuide == nil {
			return nil, nil, nil, models.ErrRecipeGuideNotFound
		}
		if withParams && params.Size() > 0 {
			params.Push(
				"recipeGuide", recipeGuide,
				"request", recipeGuide,
			)
		}
	}
	return recipe, recipeGuide, params, nil
}

func (q *recipeController) prepareRecipeItem(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentSeller *models.Seller,
	csrfToken string,
	withParams bool,
) (
	*models.Recipe,
	*models.RecipeItem,
	[]*models.Product,
	models.Map,
	error,
) {
	recipe, err := q.prepareRecipeByID(ctx, c, currentSeller.CompanyID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	recipeItem := &models.RecipeItem{}
	var productIDs []uuid.UUID
	products := []*models.Product{}
	var params models.Map
	if withParams && recipe.Name != nil {
		recipeGuides := []*models.RecipeGuide{}
		recipeItems := []*models.RecipeItem{}
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentSeller", currentSeller,
			"pageTitle", helper.Sprintf("Detail Resep|%s", *recipe.Name),
			"recipe", recipe,
			"recipeItem", recipeItem,
			"request", recipeItem,
			"recipeGuides", recipeGuides,
			"recipeItems", recipeItems,
			"products", products,
		)
		if recipeGuides, _, err = q.recipeGuideService.FindRecipeGuides(
			ctx,
			models.NewRecipeGuideFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecipeIDs(recipe.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.DefaultAscendingSortBy),
		); err != nil {
			return nil, nil, nil, nil, err
		}
		if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
			ctx,
			models.NewRecipeItemFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecipeIDs(recipe.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.DefaultAscendingSortBy),
		); err != nil {
			return nil, nil, nil, nil, err
		}
		for _, recipeItem := range recipeItems {
			productIDs = append(productIDs, recipeItem.ProductID)
		}
		params.Push(
			"recipeGuides", recipeGuides,
			"recipeItems", recipeItems,
			"products", products,
		)
	}
	if c.Params("item_id") != "" {
		recipeItemID, err := uuid.Parse(c.Params("item_id"))
		if err != nil {
			return nil, nil, nil, nil, models.ErrRecipeItemNotFound
		}
		if recipeItem, err = q.recipeItemService.FindRecipeItem(
			ctx,
			models.NewRecipeItemFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecordStatuses(models.RecordStatusActive).
				WithRecipeIDs(recipe.ID).
				WithRecipeItemIDs(recipeItemID),
		); err != nil {
			return nil, nil, nil, nil, err
		}
		if recipeItem == nil {
			return nil, nil, nil, nil, models.ErrRecipeItemNotFound
		}
		oldProductIDs := make([]uuid.UUID, len(productIDs))
		copy(oldProductIDs, productIDs)
		productIDs = productIDs[:0]
		for _, oldProductID := range oldProductIDs {
			if recipeItem.ProductID.Compare(oldProductID) != 0 {
				productIDs = append(productIDs, oldProductID)
			}
		}
		if withParams && params.Size() > 0 {
			params.Push(
				"recipeItem", recipeItem,
				"request", recipeItem,
			)
		}
	}
	if products, _, err = q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithExcludedProductIDs(productIDs...).
			WithSortings(models.Sorting{
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return nil, nil, nil, nil, err
	}
	if withParams && params.Size() > 0 {
		params.Push("products", products)
	}
	return recipe, recipeItem, products, params, nil
}
