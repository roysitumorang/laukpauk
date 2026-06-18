package controllers

import (
	"fmt"
	"mime/multipart"
	"os"
	"strings"
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
	landingPageBlockController struct {
		userService             services.UserService
		landingPageBlockService services.LandingPageBlockService
	}
)

func NewLandingPageBlockController(
	userService services.UserService,
	landingPageBlockService services.LandingPageBlockService,
) *landingPageBlockController {
	return new(landingPageBlockController{
		userService:             userService,
		landingPageBlockService: landingPageBlockService,
	})
}

func (q *landingPageBlockController) Mount(r fiber.Router) {
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

func (q *landingPageBlockController) AdminHeroIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockFilter)
	rows := []*models.LandingPageBlock{}
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
	if request, err = sanitizers.FindLandingPageBlocks(ctx, c); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive).
		WithParentNames(models.LandingPageBlockNameHero)
	if rows, pagination, err = q.landingPageBlockService.FindLandingPageBlocks(
		c,
		request,
	); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_block_hero_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminHeroNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlock{
		Subject:  new(""),
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
	return c.Render("admin/landing_page_block_hero_new", params, "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminHeroCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlock)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, models.LandingPageBlockNameHero)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateLandingPageBlock")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
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
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.CreateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateLandingPageBlock")
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameHero, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider created successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
}

func (q *landingPageBlockController) AdminHeroEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameHero).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	request := landingPageBlocks[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"landingPageBlockID", request.ID,
	)
	return c.Render("admin/landing_page_block_hero_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminHeroUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameHero).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	landingPageBlock := landingPageBlocks[0]
	request := new(models.LandingPageBlock)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"landingPageBlockID", landingPageBlock.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, models.LandingPageBlockNameHero, landingPageBlock.ID)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateLandingPageBlock")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageBlock.ID
	request.RecordStatus = landingPageBlock.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = landingPageBlock.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameHero, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider updated successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
}

func (q *landingPageBlockController) AdminHeroDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameHero).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("hero slider not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	request := landingPageBlocks[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	return flash.Success("hero slider deleted successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
}

func (q *landingPageBlockController) saveFile(c fiber.Ctx, file *multipart.FileHeader, parentName models.LandingPageBlockName, filename *string) error {
	if file == nil ||
		filename == nil {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	var sb strings.Builder
	_, _ = sb.WriteString(cwd)
	_, _ = sb.WriteString("/")
	_, _ = fmt.Fprintf(&sb, models.LandingPageBlockPhotoPath, parentName, *filename)
	return c.SaveFile(file, sb.String())
}
