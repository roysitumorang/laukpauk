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
	adminService       services.AdminService
	buyerService       services.BuyerService
	recipeService      services.RecipeService
	recipeGuideService services.RecipeGuideService
	recipeItemService  services.RecipeItemService
	productService     services.ProductService
}

func NewRecipeController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	recipeService services.RecipeService,
	recipeGuideService services.RecipeGuideService,
	recipeItemService services.RecipeItemService,
	productService services.ProductService,
) *recipeController {
	return &recipeController{
		adminService:       adminService,
		buyerService:       buyerService,
		recipeService:      recipeService,
		recipeGuideService: recipeGuideService,
		recipeItemService:  recipeItemService,
		productService:     productService,
	}
}

func (q *recipeController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id", q.AdminShow).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete).
		Get("/:id/guides/new", q.AdminNewGuide).
		Post("/:id/guides", q.AdminCreateGuide).
		Get("/:id/guides/:guide_id/edit", q.AdminEditGuide).
		Post("/:id/guides/:guide_id", q.AdminUpdateGuide).
		Post("/:id/guides/:guide_id/delete", q.AdminDeleteGuide).
		Get("/:id/items/new", q.AdminNewItem).
		Post("/:id/items", q.AdminCreateItem).
		Get("/:id/items/:item_id/edit", q.AdminEditItem).
		Post("/:id/items/:item_id", q.AdminUpdateItem).
		Post("/:id/items/:item_id/delete", q.AdminDeleteItem)
}

func (q *recipeController) AdminIndex(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminIndex"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := &models.RecipeFilter{}
	rows := []*models.Recipe{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
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
		return c.Render("admin/recipe_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.recipeService.FindRecipes(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipes")
		return c.Render("admin/recipe_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/recipe_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *recipeController) AdminNew(c fiber.Ctx) (err error) {
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _ := q.prepareRecipe(ctx, c, flash, currentAdmin, csrfToken, true)
	return c.Render("admin/recipe_new", params, "admin/layouts/authenticated")
}

func (q *recipeController) AdminCreate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminCreate"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _ := q.prepareRecipe(ctx, c, flash, currentAdmin, csrfToken, true)
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
		return c.Render("admin/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "admin/layouts/authenticated")
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
			return c.Render("admin/recipe_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"recipe", request,
			), "admin/layouts/authenticated")
		}
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = now
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = now
	if recipe != nil {
		if recipe.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
			return c.Render("admin/recipe_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueRecipeNameViolation.Error()),
				"request", request,
				"recipe", request,
			), "admin/layouts/authenticated")
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
		return c.Render("admin/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
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
	if _, err = q.recipeService.SaveRecipe(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipe")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return c.Render("admin/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return c.Render("admin/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "admin/layouts/authenticated")
	}
	if recipe != nil &&
		recipe.Photo != os.Getenv("RECIPE_DEFAULT_PHOTO") &&
		recipe.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(recipe.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(recipe.Photo)
			return c.Render("admin/recipe_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"recipe", request,
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("RECIPE_DEFAULT_PHOTO"))
		return c.Render("admin/recipe_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"recipe", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("recipe created successfully").Redirect(c, sess.Session, "/admin/recipes")
}

func (q *recipeController) AdminShow(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminShow"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	return c.Render("admin/recipe_show", params, "admin/layouts/authenticated")
}

func (q *recipeController) AdminEdit(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminEdit"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	return c.Render("admin/recipe_edit", params, "admin/layouts/authenticated")
}

func (q *recipeController) AdminUpdate(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminUpdate"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, params, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateRecipe(ctx, c)
	if photo == nil {
		request.SetPhoto(recipe.Photo)
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipe")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "admin/layouts/authenticated")
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
			return c.Render("admin/recipe_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(recipe.Photo),
			), "admin/layouts/authenticated")
		}
		if otherRecipe != nil {
			return c.Render("admin/recipe_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueRecipeNameViolation.Error()),
				"request", request.SetPhoto(recipe.Photo),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = recipe.ID
	request.RecordStatus = recipe.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
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
	if _, err = q.recipeService.UpdateRecipe(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipe")
		return c.Render("admin/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "admin/layouts/authenticated")
	}
	if recipe.Photo != os.Getenv("RECIPE_DEFAULT_PHOTO") &&
		recipe.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(recipe.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			return c.Render("admin/recipe_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(recipe.Photo),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/recipe_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(recipe.Photo),
		), "admin/layouts/authenticated")
	}
	return flash.Success("recipe updated successfully").Redirect(c, sess.Session, "/admin/recipes")
}

func (q *recipeController) AdminDelete(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminDelete"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, _, err := q.prepareRecipe(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipe")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	recipe.RecordStatus = models.RecordStatusDeleted
	if recipe.Name != nil {
		recipe.DeletedName = new(*recipe.Name)
	}
	recipe.Name = nil
	recipe.DeletedBy = &currentAdmin.ID
	recipe.DeletedAt = new(time.Now())
	recipe.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	return flash.Success("recipe deleted successfully").Redirect(c, sess.Session, "/admin/recipes")
}

func (q *recipeController) AdminNewGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminNewGuide"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	return c.Render("admin/recipe_guide_new", params, "admin/layouts/authenticated")
}

func (q *recipeController) AdminCreateGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminCreateGuide"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, _, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/recipe_guide_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.RecipeID = recipe.ID
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = now
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/recipe_guide_new", params.Push(
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
	if _, err = q.recipeGuideService.CreateRecipeGuide(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRecipeGuide")
		return c.Render("admin/recipe_guide_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/recipe_guide_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("recipe guide created successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
}

func (q *recipeController) AdminEditGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminEditGuide"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	return c.Render("admin/recipe_guide_edit", params, "admin/layouts/authenticated")
}

func (q *recipeController) AdminUpdateGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminUpdateGuide"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeGuide, params, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeGuide(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeGuide")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/recipe_guide_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = recipeGuide.ID
	request.RecordStatus = recipeGuide.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/recipe_guide_edit", params.Push(
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
	if _, err = q.recipeGuideService.UpdateRecipeGuide(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
		return c.Render("admin/recipe_guide_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/recipe_guide_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("recipe guide updated successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
}

func (q *recipeController) AdminDeleteGuide(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminDeleteGuide"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeGuide, _, err := q.prepareRecipeGuide(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeGuide")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	recipeGuide.RecordStatus = models.RecordStatusDeleted
	recipeGuide.DeletedBy = new(currentAdmin.ID)
	recipeGuide.DeletedAt = new(time.Now())
	recipeGuide.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
	}
	return flash.Success("recipe guide deleted successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
}

func (q *recipeController) AdminNewItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminNewItem"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, _, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	return c.Render("admin/recipe_item_new", params, "admin/layouts/authenticated")
}

func (q *recipeController) AdminCreateItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminCreateItem"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, _, products, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	mapProductIDs := map[uuid.UUID]struct{}{}
	for i := range products {
		mapProductIDs[products[i].ID] = struct{}{}
	}
	var recipeItem *models.RecipeItem
	if request.ProductID != nil {
		if _, ok := mapProductIDs[*request.ProductID]; !ok {
			return c.Render("admin/recipe_item_new", params.Push(
				"flash", flash.Danger("product_id: not found"),
			), "admin/layouts/authenticated")
		}
		if recipeItem, err = q.recipeItemService.FindRecipeItem(
			ctx,
			models.NewRecipeItemFilter().
				WithRecipeIDs(recipe.ID).
				WithProductIDs(*request.ProductID),
		); err != nil && !errors.Is(err, models.ErrRecipeItemNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItem")
			return c.Render("admin/recipe_item_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.RecipeID = recipe.ID
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = now
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = now
	if recipeItem != nil {
		if recipeItem.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/recipe_item_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueRecipeItemProductIdViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = recipeItem.ID
		request.DeletedProductID = nil
		request.CreatedBy = recipeItem.CreatedBy
		request.CreatedAt = recipeItem.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/recipe_item_new", params.Push(
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
	if _, err = q.recipeItemService.SaveRecipeItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipeItem")
		return c.Render("admin/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/recipe_item_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("recipe item created successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
}

func (q *recipeController) AdminEditItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminEditItem"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, _, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	return c.Render("admin/recipe_item_edit", params, "admin/layouts/authenticated")
}

func (q *recipeController) AdminUpdateItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminUpdateItem"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeItem, _, params, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	request, statusCode, err := sanitizers.ValidateRecipeItem(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRecipeItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/recipe_item_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = recipeItem.ID
	request.RecordStatus = recipeItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/recipe_item_edit", params.Push(
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
	if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
		return c.Render("admin/recipe_item_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/recipe_item_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("recipe item updated successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
}

func (q *recipeController) AdminDeleteItem(c fiber.Ctx) (err error) {
	ctxt := "RecipeController-AdminDeleteItem"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	recipe, recipeItem, _, _, err := q.prepareRecipeItem(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareRecipeItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/recipes")
	}
	recipeItem.RecordStatus = models.RecordStatusDeleted
	if recipeItem.ProductID != nil {
		recipeItem.DeletedProductID = new(*recipeItem.ProductID)
	}
	recipeItem.ProductID = nil
	recipeItem.DeletedBy = new(currentAdmin.ID)
	recipeItem.DeletedAt = new(time.Now())
	recipeItem.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
	}
	return flash.Success("recipe item deleted successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/recipes/%s", recipe.ID))
}

func (q *recipeController) prepare(c fiber.Ctx) (
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

func (q *recipeController) prepareRecipe(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentAdmin *models.Admin,
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
			"currentAdmin", currentAdmin,
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
					WithRecipeIDs(recipe.ID).
					WithRecordStatuses(models.RecordStatusActive).
					WithSortings(models.DefaultAscendingSortBy),
			); err != nil {
				return nil, nil, err
			}
			if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
				ctx,
				models.NewRecipeItemFilter().
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
	flash *helper.FlashMessage,
	currentAdmin *models.Admin,
	csrfToken string,
	withParams bool,
) (
	*models.Recipe,
	*models.RecipeGuide,
	models.Map,
	error,
) {
	recipe, err := q.prepareRecipeByID(ctx, c)
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
			"currentAdmin", currentAdmin,
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
				WithRecipeIDs(recipe.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.DefaultAscendingSortBy),
		); err != nil {
			return nil, nil, nil, err
		}
		if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
			ctx,
			models.NewRecipeItemFilter().
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
	currentAdmin *models.Admin,
	csrfToken string,
	withParams bool,
) (
	*models.Recipe,
	*models.RecipeItem,
	[]*models.Product,
	models.Map,
	error,
) {
	recipe, err := q.prepareRecipeByID(ctx, c)
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
			"currentAdmin", currentAdmin,
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
				WithRecipeIDs(recipe.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.DefaultAscendingSortBy),
		); err != nil {
			return nil, nil, nil, nil, err
		}
		if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
			ctx,
			models.NewRecipeItemFilter().
				WithRecipeIDs(recipe.ID).
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.DefaultAscendingSortBy),
		); err != nil {
			return nil, nil, nil, nil, err
		}
		for _, recipeItem := range recipeItems {
			if recipeItem.ProductID != nil {
				productIDs = append(productIDs, *recipeItem.ProductID)
			}
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
			if recipeItem.ProductID != nil &&
				recipeItem.ProductID.Compare(oldProductID) != 0 {
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
