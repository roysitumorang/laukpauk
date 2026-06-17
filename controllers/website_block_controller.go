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
	websiteBlockController struct {
		userService             services.UserService
		websiteBlockService     services.WebsiteBlockService
		websiteBlockItemService services.WebsiteBlockItemService
	}
)

func NewWebsiteBlockController(
	userService services.UserService,
	websiteBlockService services.WebsiteBlockService,
	websiteBlockItemService services.WebsiteBlockItemService,
) *websiteBlockController {
	return new(websiteBlockController{
		userService:             userService,
		websiteBlockService:     websiteBlockService,
		websiteBlockItemService: websiteBlockItemService,
	})
}

func (q *websiteBlockController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	r.Group("/admin", adminSessionAuth).
		Group("/hero").
		Get("", q.AdminHeroIndex).
		Get("/new", q.AdminHeroNew).
		Post("", q.AdminHeroCreate).
		Get("/:id/edit", q.AdminHeroEdit).
		Post("/:id", q.AdminHeroUpdate).
		Post("/:id/delete", q.AdminHeroDelete)
}

func (q *websiteBlockController) AdminHeroIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteBlockController-AdminHeroIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.WebsiteBlockItemFilter)
	rows := []*models.WebsiteBlockItem{}
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
	if request, err = sanitizers.FindWebsiteBlockItems(ctx, c); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlockItems")
		return c.Render("admin/website_block_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive).
		WithWebsiteBlockNames(models.WebsiteBlockNameHero)
	if rows, pagination, err = q.websiteBlockItemService.FindWebsiteBlockItems(
		c,
		request,
	); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlockItems")
		return c.Render("admin/website_block_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/website_block_hero_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *websiteBlockController) AdminHeroNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.WebsiteBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
	)
	return c.Render("admin/website_block_hero_new", params, "admin/layouts/authenticated")
}

func (q *websiteBlockController) AdminHeroCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteBlockController-AdminHeroCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.WebsiteBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateWebsiteBlockItem(ctx, c, models.WebsiteBlockNameHero)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateWebsiteBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/website_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	websiteBlocks, _, err := q.websiteBlockService.FindWebsiteBlocks(
		c,
		models.NewWebsiteBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.WebsiteBlockNameHero),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlocks")
		return c.Render("admin/website_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if len(websiteBlocks) == 0 {
		return c.Render("admin/website_block_hero_new", params.Push(
			"flash", flash.Danger("website block not found"),
			"request", request,
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.WebsiteBlockID = websiteBlocks[0].ID
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/website_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.websiteBlockItemService.CreateWebsiteBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateWebsiteBlockItem")
		return c.Render("admin/website_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/website_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/website_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider created successfully").Redirect(c, sess.Session, "/admin/website_blocks/hero")
}

func (q *websiteBlockController) AdminHeroEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteBlockController-AdminHeroEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	websiteBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	websiteBlockItems, _, err := q.websiteBlockItemService.FindWebsiteBlockItems(
		ctx,
		models.NewWebsiteBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithWebsiteBlockNames(models.WebsiteBlockNameHero).
			WithWebsiteBlockItemIDs(websiteBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	if len(websiteBlockItems) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	request := websiteBlockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"websiteBlockItemID", request.ID,
	)
	return c.Render("admin/website_block_hero_edit", params, "admin/layouts/authenticated")
}

func (q *websiteBlockController) AdminHeroUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteBlockController-AdminHeroUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	websiteBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	websiteBlockItems, _, err := q.websiteBlockItemService.FindWebsiteBlockItems(
		ctx,
		models.NewWebsiteBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithWebsiteBlockNames(models.WebsiteBlockNameHero).
			WithWebsiteBlockItemIDs(websiteBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlockItems")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	if len(websiteBlockItems) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	websiteBlockItem := websiteBlockItems[0]
	request := new(models.WebsiteBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"websiteBlockItemID", websiteBlockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateWebsiteBlockItem(ctx, c, models.WebsiteBlockNameHero, websiteBlockItem.ID)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateWebsiteBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/website_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = websiteBlockItem.ID
	request.RecordStatus = websiteBlockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = websiteBlockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/website_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.websiteBlockItemService.UpdateWebsiteBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateWebsiteBlockItem")
		return c.Render("admin/website_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/website_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/website_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider updated successfully").Redirect(c, sess.Session, "/admin/website_blocks/hero")
}

func (q *websiteBlockController) AdminHeroDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "WebsiteBlockController-AdminHeroDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	websiteBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	websiteBlockItems, _, err := q.websiteBlockItemService.FindWebsiteBlockItems(
		ctx,
		models.NewWebsiteBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithWebsiteBlockNames(models.WebsiteBlockNameHero).
			WithWebsiteBlockItemIDs(websiteBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindWebsiteBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	if len(websiteBlockItems) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	request := websiteBlockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	if _, err = q.websiteBlockItemService.UpdateWebsiteBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateWebsiteBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/website_blocks/hero")
	}
	return flash.Success("hero slider deleted successfully").Redirect(c, sess.Session, "/admin/website_blocks/hero")
}

func (q *websiteBlockController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil ||
		filename == "" {
		return nil
	}
	return c.SaveFile(file, fmt.Sprintf(models.WebsiteBlockHeroPhotoPath, filename))
}
