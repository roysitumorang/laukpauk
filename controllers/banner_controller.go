package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
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

type (
	bannerController struct {
		adminService  services.AdminService
		bannerService services.BannerService
	}
)

func NewBannerController(
	userService services.AdminService,
	bannerService services.BannerService,
) *bannerController {
	return &bannerController{
		adminService:  userService,
		bannerService: bannerService,
	}
}

func (q *bannerController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/switch", q.AdminSwitchVisibility).
		Post("/:id/delete", q.AdminDelete)
}

func (q *bannerController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.BannerFilter{}
	rows := []*models.Banner{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Banner",
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
				Field:     models.BannerSortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.BannerSortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	request, err = sanitizers.FindBanners(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanners")
		return c.Render("admin/banner_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.bannerService.FindBanners(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanners")
		return c.Render("admin/banner_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/banner_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *bannerController) AdminNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Banner)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Banner",
		"request", request,
		"banner", request,
	)
	return c.Render("admin/banner_new", params, "admin/layouts/authenticated")
}

func (q *bannerController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Banner)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Banner",
		"request", request,
		"banner", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBanner(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBanner")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("ARTICLE_DEFAULT_PHOTO")
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
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
	if _, err = q.bannerService.CreateBanner(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBanner")
		return c.Render("admin/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/banner_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("banner created successfully").
		Redirect(c, sess.Session, "/admin/banners")
}

func (q *bannerController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Banner|%s", banner.Name),
		"request", banner,
		"banner", banner,
	)
	return c.Render("admin/banner_edit", params, "admin/layouts/authenticated")
}

func (q *bannerController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Banner|%s", banner.Name),
		"request", banner,
		"banner", banner,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBanner(ctx, c)
	if photo == nil {
		request.Photo = banner.Photo
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBanner")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/banner_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = banner.ID
	request.RecordStatus = banner.RecordStatus
	request.Visible = banner.Visible
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/banner_edit", params.Push(
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
	if _, err = q.bannerService.UpdateBanner(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
		return c.Render("admin/banner_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/banner_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/banner_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("banner updated successfully").
		Redirect(c, sess.Session, "/admin/banners")
}

func (q *bannerController) AdminSwitchVisibility(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-AdminSwitchVisibility"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	banner.Visible = !banner.Visible
	banner.UpdatedBy = currentAdmin.ID
	banner.UpdatedAt = time.Now()
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.bannerService.UpdateBanner(ctx, tx, banner); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	return flash.Clear(c, sess.Session).
		Success("banner visibility switched successfully").
		Redirect(c, sess.Session, "/admin/banners")
}

func (q *bannerController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	request := banner
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.bannerService.UpdateBanner(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/banners")
	}
	return flash.Clear(c, sess.Session).
		Success("banner deleted successfully").
		Redirect(c, sess.Session, "/admin/banners")
}
