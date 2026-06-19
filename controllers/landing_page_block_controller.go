package controllers

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"os"
	"strings"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
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
	admin := r.Group("/admin", adminSessionAuth)
	admin.Group("/hero").
		Get("", q.AdminHeroItemIndex).
		Get("/new", q.AdminHeroItemNew).
		Post("", q.AdminHeroItemCreate).
		Get("/:id/edit", q.AdminHeroItemEdit).
		Post("/:id", q.AdminHeroItemUpdate).
		Post("/:id/delete", q.AdminHeroItemDelete)
	admin.Group("/feature").
		Get("", q.AdminFeatureItemIndex).
		Post("/custom_attributes", q.AdminFeatureUpdate).
		Get("/new", q.AdminFeatureItemNew).
		Post("", q.AdminFeatureItemCreate).
		Get("/:id/edit", q.AdminFeatureItemEdit).
		Post("/:id", q.AdminFeatureItemUpdate).
		Post("/:id/delete", q.AdminFeatureItemDelete)
	admin.Group("/why").
		Get("", q.AdminWhyItemIndex).
		Post("/custom_attributes", q.AdminWhyUpdate).
		Get("/new", q.AdminWhyItemNew).
		Post("", q.AdminWhyItemCreate).
		Get("/:id/edit", q.AdminWhyItemEdit).
		Post("/:id", q.AdminWhyItemUpdate).
		Post("/:id/delete", q.AdminWhyItemDelete)
}

func (q *landingPageBlockController) AdminHeroItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroItemIndex"
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
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

func (q *landingPageBlockController) AdminHeroItemNew(c fiber.Ctx) (err error) {
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

func (q *landingPageBlockController) AdminHeroItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroItemCreate"
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
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, new(models.LandingPageBlockNameHero))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlock")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.CreateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlock")
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameHero, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider created successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
}

func (q *landingPageBlockController) AdminHeroItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
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

func (q *landingPageBlockController) AdminHeroItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
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
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, new(models.LandingPageBlockNameHero), landingPageBlock.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlock")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameHero, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("hero slider updated successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
}

func (q *landingPageBlockController) AdminHeroItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminHeroItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
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

func (q *landingPageBlockController) AdminFeatureItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminFeatureItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.LandingPageBlockFilter)
	rows := []*models.LandingPageBlock{}
	request := new(models.LandingPageBlock)
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Fitur Laukpauk",
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
	)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameFeature),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(landingPageBlocks) == 0 {
		return c.Render("admin/landing_page_block_feature_index", params.Push(
			"flash", "landing page block not found",
		), "admin/layouts/authenticated")
	}
	request = landingPageBlocks[0]
	params.Push("request", request)
	if paginationRequest, err = sanitizers.FindLandingPageBlocks(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithParentNames(models.LandingPageBlockNameFeature)
	if rows, pagination, err = q.landingPageBlockService.FindLandingPageBlocks(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_block_feature_index", params.Push(
		"rows", rows,
		"paginationRequest", paginationRequest,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminFeatureUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminFeatureUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameFeature),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	landingPageBlock := landingPageBlocks[0]
	request, err := sanitizers.ValidateLandingPageBlockCustomAttributes(ctx, c, models.LandingPageBlockNameFeature)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	landingPageBlock.CustomAttributes = request
	landingPageBlock.UpdatedBy = currentAdmin.ID
	landingPageBlock.UpdatedAt = time.Now()
	landingPageBlock.DeletedBy = nil
	landingPageBlock.DeletedAt = nil
	landingPageBlock.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, landingPageBlock); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	return flash.Success("feature subject updated successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
}

func (q *landingPageBlockController) AdminFeatureItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlock{
		Subject: new(""),
		Body:    new(""),
	})
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Fitur",
		"request", request,
	)
	return c.Render("admin/landing_page_block_feature_new", params, "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminFeatureItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminFeatureItemCreate"
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
		"pageTitle", "Add New Fitur",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, new(models.LandingPageBlockNameFeature))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlock")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_feature_new", params.Push(
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.CreateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlock")
		return c.Render("admin/landing_page_block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameFeature, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("feature created successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
}

func (q *landingPageBlockController) AdminFeatureItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminFeatureItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameFeature).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	request := landingPageBlocks[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", request,
		"landingPageBlockID", request.ID,
	)
	return c.Render("admin/landing_page_block_feature_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminFeatureItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminFeatureItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameFeature).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	landingPageBlock := landingPageBlocks[0]
	request := new(models.LandingPageBlock)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", request,
		"landingPageBlockID", landingPageBlock.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, new(models.LandingPageBlockNameFeature), landingPageBlock.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlock")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_feature_edit", params.Push(
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return c.Render("admin/landing_page_block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameFeature, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("feature updated successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
}

func (q *landingPageBlockController) AdminFeatureItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminFeatureItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameFeature).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("feature not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	request := landingPageBlocks[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
	}
	return flash.Success("feature deleted successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/feature")
}

func (q *landingPageBlockController) AdminWhyItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminWhyItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.LandingPageBlockFilter)
	rows := []*models.LandingPageBlock{}
	request := new(models.LandingPageBlock)
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Mengapa Harus Pilih Laukpauk",
		"limits", models.Limits,
		"paginationRequest", paginationRequest,
		"request", request,
		"rows", rows,
		"pagination", pagination,
	)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameWhy),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(landingPageBlocks) == 0 {
		return c.Render("admin/landing_page_block_why_index", params.Push(
			"flash", "landing page block not found",
		), "admin/layouts/authenticated")
	}
	request = landingPageBlocks[0]
	params.Push("request", request)
	if paginationRequest, err = sanitizers.FindLandingPageBlocks(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithParentNames(models.LandingPageBlockNameWhy)
	if rows, pagination, err = q.landingPageBlockService.FindLandingPageBlocks(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_block_why_index", params.Push(
		"rows", rows,
		"paginationRequest", paginationRequest,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminWhyUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminWhyUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameWhy),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	landingPageBlock := landingPageBlocks[0]
	request, err := sanitizers.ValidateLandingPageBlockCustomAttributes(ctx, c, models.LandingPageBlockNameWhy)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	landingPageBlock.CustomAttributes = request
	landingPageBlock.UpdatedBy = currentAdmin.ID
	landingPageBlock.UpdatedAt = time.Now()
	landingPageBlock.DeletedBy = nil
	landingPageBlock.DeletedAt = nil
	landingPageBlock.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, landingPageBlock); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	return flash.Success("why subject updated successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
}

func (q *landingPageBlockController) AdminWhyItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlock{
		Subject: new(""),
		Body:    new(""),
	})
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Penjelasan Baru",
		"request", request,
	)
	return c.Render("admin/landing_page_block_why_new", params, "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminWhyItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminWhyItemCreate"
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
		"pageTitle", "Penjelasan Baru",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, new(models.LandingPageBlockNameWhy))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlock")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_why_new", params.Push(
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.CreateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlock")
		return c.Render("admin/landing_page_block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameWhy, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("why created successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
}

func (q *landingPageBlockController) AdminWhyItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminWhyItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameWhy).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	request := landingPageBlocks[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", request,
		"landingPageBlockID", request.ID,
	)
	return c.Render("admin/landing_page_block_why_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageBlockController) AdminWhyItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminWhyItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameWhy).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	landingPageBlock := landingPageBlocks[0]
	request := new(models.LandingPageBlock)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", request,
		"landingPageBlockID", landingPageBlock.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlock(ctx, c, new(models.LandingPageBlockNameWhy), landingPageBlock.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlock")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_why_edit", params.Push(
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return c.Render("admin/landing_page_block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameWhy, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("why updated successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
}

func (q *landingPageBlockController) AdminWhyItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageBlockController-AdminWhyItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithParentNames(models.LandingPageBlockNameWhy).
			WithLandingPageBlockIDs(landingPageBlockID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("why not found").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	request := landingPageBlocks[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
	}
	return flash.Success("why deleted successfully").Redirect(c, sess.Session, "/admin/landing_page_blocks/why")
}
