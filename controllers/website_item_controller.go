package controllers

import (
	"fmt"
	"mime/multipart"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
	"go.uber.org/zap"
)

type (
	websiteItemController struct {
		userService        services.UserService
		websiteItemService services.WebsiteItemService
	}
)

func NewWebsiteItemController(
	userService services.UserService,
	websiteItemService services.WebsiteItemService,
) *websiteItemController {
	return new(websiteItemController{
		userService:        userService,
		websiteItemService: websiteItemService,
	})
}

func (q *websiteItemController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	admin := r.Group("/admin", adminSessionAuth)
	admin.Group("/hero").
		Get("", q.AdminHeroIndex).
		Get("/new", q.AdminHeroNew).
		Post("", q.AdminHeroCreate).
		Get("/:id/edit", q.AdminHeroEdit).
		Post("/:id", q.AdminHeroUpdate).
		Post("/:id/delete", q.AdminHeroDelete)
}

func (q *websiteItemController) AdminHeroIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteItemController-AdminHeroIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.WebsiteItemFilter)
	rows := []*models.WebsiteItem{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Hero Slider",
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
	)
	if request, err = sanitizers.FindWebsiteItems(ctx, c); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteItems")
		return c.Render("admin/website_item_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive).
		WithWebsiteItemTypes(models.WebsiteItemTypeHero)
	if rows, pagination, err = q.websiteItemService.FindWebsiteItems(
		c,
		request,
	); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteItems")
		return c.Render("admin/website_item_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/website_item_hero_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *websiteItemController) AdminHeroNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	var emptyString string
	request := new(models.WebsiteItem{
		Body:     new(emptyString),
		LinkText: new(emptyString),
		LinkURL:  new(emptyString),
	})
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
	)
	return c.Render("admin/website_item_hero_new", params, "admin/layouts/authenticated")
}

func (q *websiteItemController) AdminHeroCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteItemController-AdminHeroCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.WebsiteItem{
		Body:     new(""),
		LinkText: new(""),
		LinkURL:  new(""),
	})
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateWebsiteItem(ctx, c, models.WebsiteItemTypeHero)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateWebsiteItemHero")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/website_item_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/website_item_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.websiteItemService.CreateWebsiteItem(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateWebsiteItem")
		return c.Render("admin/website_item_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/website_item_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/website_item_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider created successfully").Redirect(c, sess.Session, "/admin/website_items/hero")
}

func (q *websiteItemController) AdminHeroEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteItemController-AdminHeroEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	websiteItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	websiteItems, _, err := q.websiteItemService.FindWebsiteItems(
		ctx,
		models.NewWebsiteItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithWebsiteItemIDs(websiteItemID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	if len(websiteItems) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	request := websiteItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"websiteItemID", request.ID,
	)
	return c.Render("admin/website_item_hero_edit", params, "admin/layouts/authenticated")
}

func (q *websiteItemController) AdminHeroUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteItemController-AdminHeroUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	websiteItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	websiteItems, _, err := q.websiteItemService.FindWebsiteItems(
		ctx,
		models.NewWebsiteItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithWebsiteItemIDs(websiteItemID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteItems")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	if len(websiteItems) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	websiteItem := websiteItems[0]
	request := new(models.WebsiteItem{
		Body:     new(""),
		LinkText: new(""),
		LinkURL:  new(""),
	})
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"websiteItemID", websiteItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateWebsiteItem(ctx, c, models.WebsiteItemTypeHero)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateWebsiteItemHero")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/website_item_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = websiteItem.ID
	request.RecordStatus = websiteItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = websiteItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/website_item_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.websiteItemService.UpdateWebsiteItem(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateWebsiteItem")
		return c.Render("admin/website_item_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/website_item_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/website_item_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider updated successfully").Redirect(c, sess.Session, "/admin/website_items/hero")
}

func (q *websiteItemController) AdminHeroDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteItemController-AdminHeroDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	websiteItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	websiteItems, _, err := q.websiteItemService.FindWebsiteItems(
		ctx,
		models.NewWebsiteItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithWebsiteItemIDs(websiteItemID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	if len(websiteItems) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	request := websiteItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	if _, err = q.websiteItemService.UpdateWebsiteItem(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateWebsiteItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_items/hero")
	}
	return flash.Success("hero slider deleted successfully").Redirect(c, sess.Session, "/admin/website_items/hero")
}

func (q *websiteItemController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename *string) error {
	if file == nil ||
		filename == nil {
		return nil
	}
	return c.SaveFile(file, fmt.Sprintf(models.WebsiteItemHeroPhotoPath, *filename))
}
