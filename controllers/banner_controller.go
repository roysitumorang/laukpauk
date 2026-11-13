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

type bannerController struct {
	sellerService       services.SellerService
	bannerService       services.BannerService
	notificationService services.NotificationService
}

func NewBannerController(
	sellerService services.SellerService,
	bannerService services.BannerService,
	notificationService services.NotificationService,
) *bannerController {
	return &bannerController{
		sellerService:       sellerService,
		bannerService:       bannerService,
		notificationService: notificationService,
	}
}

func (q *bannerController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/switch", q.SellerSwitchVisibility).
		Post("/:id/delete", q.SellerDelete)
}

func (q *bannerController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.BannerFilter{}
	rows := []*models.Banner{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("seller/banner_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.bannerService.FindBanners(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanners")
		return c.Render("seller/banner_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/banner_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *bannerController) SellerNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Banner{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Add New Banner",
		"request", request,
		"banner", request,
	)
	return c.Render("seller/banner_new", params, "seller/layouts/authenticated")
}

func (q *bannerController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Banner{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Add New Banner",
		"request", request,
		"banner", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBanner(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBanner")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentSeller.ID
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
		return c.Render("seller/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
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
	if _, err = q.bannerService.CreateBanner(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBanner")
		return c.Render("seller/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/banner_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/banner_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("banner created successfully").
		Redirect(c, sess.Session, "/seller/banners")
}

func (q *bannerController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Data Banner|%s", banner.Name),
		"request", banner,
		"banner", banner,
	)
	return c.Render("seller/banner_edit", params, "seller/layouts/authenticated")
}

func (q *bannerController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("seller/banner_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	request.ID = banner.ID
	request.RecordStatus = banner.RecordStatus
	request.Visible = banner.Visible
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/banner_edit", params.Push(
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
	if _, err = q.bannerService.UpdateBanner(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
		return c.Render("seller/banner_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/banner_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/banner_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("banner updated successfully").
		Redirect(c, sess.Session, "/seller/banners")
}

func (q *bannerController) SellerSwitchVisibility(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-SellerSwitchVisibility"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	banner.Visible = !banner.Visible
	banner.UpdatedBy = currentSeller.ID
	banner.UpdatedAt = time.Now()
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	return flash.Clear(c, sess.Session).
		Success("banner visibility switched successfully").
		Redirect(c, sess.Session, "/seller/banners")
}

func (q *bannerController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	if banner == nil {
		return flash.Danger(models.ErrBannerNotFound.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	banner.RecordStatus = models.RecordStatusDeleted
	banner.DeletedBy = &currentSeller.ID
	banner.DeletedAt = new(time.Now())
	banner.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/banners")
	}
	return flash.Clear(c, sess.Session).
		Success("banner deleted successfully").
		Redirect(c, sess.Session, "/seller/banners")
}
