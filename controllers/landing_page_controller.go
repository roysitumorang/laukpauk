package controllers

import (
	"context"
	"errors"
	"log/slog"
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
	landingPageController struct {
		adminService           services.AdminService
		landingPageService     services.LandingPageService
		landingPageItemService services.LandingPageItemService
		blogService            services.BlogService
		messageService         services.MessageService
	}
)

func NewLandingPageController(
	adminService services.AdminService,
	landingPageService services.LandingPageService,
	landingPageItemService services.LandingPageItemService,
	blogService services.BlogService,
	messageService services.MessageService,
) *landingPageController {
	return &landingPageController{
		adminService:           adminService,
		landingPageService:     landingPageService,
		landingPageItemService: landingPageItemService,
		blogService:            blogService,
		messageService:         messageService,
	}
}

func (q *landingPageController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	admin := r.Group("/admin", adminSessionAuth)
	admin.Group("/hero-slider").
		Get("", q.AdminHeroSliderItemIndex).
		Get("/items/new", q.AdminHeroSliderItemNew).
		Post("/items", q.AdminHeroSliderItemCreate).
		Get("/items/:id/edit", q.AdminHeroSliderItemEdit).
		Post("/items/:id", q.AdminHeroSliderItemUpdate).
		Post("/items/:id/delete", q.AdminHeroSliderItemDelete)
	admin.Group("/feature").
		Get("", q.AdminFeatureItemIndex).
		Post("/custom_attributes", q.AdminFeatureUpdate).
		Get("/items/new", q.AdminFeatureItemNew).
		Post("/items", q.AdminFeatureItemCreate).
		Get("/items/:id/edit", q.AdminFeatureItemEdit).
		Post("/items/:id", q.AdminFeatureItemUpdate).
		Post("/items/:id/delete", q.AdminFeatureItemDelete)
	admin.Group("/why").
		Get("", q.AdminWhyItemIndex).
		Post("/custom_attributes", q.AdminWhyUpdate).
		Get("/items/new", q.AdminWhyItemNew).
		Post("/items", q.AdminWhyItemCreate).
		Get("/items/:id/edit", q.AdminWhyItemEdit).
		Post("/items/:id", q.AdminWhyItemUpdate).
		Post("/items/:id/delete", q.AdminWhyItemDelete)
	admin.Group("/commitment").
		Get("", q.AdminCommitmentIndex).
		Post("/custom_attributes", q.AdminCommitmentUpdate)
	admin.Group("/free-shopping").
		Get("", q.AdminFreeShoppingIndex).
		Post("/custom_attributes", q.AdminFreeShoppingUpdate)
	admin.Group("/kitchen-hero").
		Get("", q.AdminKitchenHeroIndex).
		Post("/custom_attributes", q.AdminKitchenHeroUpdate)
	admin.Group("/faq").
		Get("", q.AdminFaqItemIndex).
		Get("/items/new", q.AdminFaqItemNew).
		Post("/items", q.AdminFaqItemCreate).
		Get("/items/:id/edit", q.AdminFaqItemEdit).
		Post("/items/:id", q.AdminFaqItemUpdate).
		Post("/items/:id/delete", q.AdminFaqItemDelete)
}

func (q *landingPageController) AdminHeroSliderItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroSliderItemIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageItemFilter)
	rows := []*models.LandingPageItem{}
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
		},
	)
	request, err = sanitizers.FindLandingPageItems(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/landing_page_hero_slider_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.WithRecordStatus(models.RecordStatusActive).
		WithLandingPageNames(models.LandingPageNameHeroSlider)
	params.Push("request", request)
	if rows, pagination, err = q.landingPageItemService.FindLandingPageItems(ctx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("admin/landing_page_hero_slider_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_hero_slider_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminHeroSliderItemNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{
		LinkText: new(""),
		LinkURL:  new(""),
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
		"landingPageItem", request,
	)
	return c.Render("admin/landing_page_hero_slider_new", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminHeroSliderItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroSliderItemCreate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{
		LinkText: new(""),
		LinkURL:  new(""),
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
		"landingPageItem", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameHeroSlider)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_hero_slider_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	now := time.Now().In(time.Local)
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_hero_slider_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
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
	if _, err = q.landingPageItemService.CreateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageItem")
		return c.Render("admin/landing_page_hero_slider_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_hero_slider_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(nil),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_hero_slider_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page hero slider item created successfully").Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
}

func (q *landingPageController) AdminHeroSliderItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroSliderItemEdit"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameHeroSlider).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	return c.Render("admin/landing_page_hero_slider_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminHeroSliderItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroSliderItemUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameHeroSlider).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameHeroSlider)
	if photo == nil {
		request.Photo = landingPageItem.Photo
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_hero_slider_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageItem.ID
	request.RecordStatus = landingPageItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_hero_slider_edit", params.Push(
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
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return c.Render("admin/landing_page_hero_slider_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_hero_slider_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_hero_slider_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page hero slider item updated successfully").Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
}

func (q *landingPageController) AdminHeroSliderItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroSliderItemDelete"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameHeroSlider).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	landingPageItem.RecordStatus = models.RecordStatusDeleted
	landingPageItem.DeletedBy = &currentAdmin.ID
	landingPageItem.DeletedAt = new(time.Now())
	landingPageItem.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, landingPageItem); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
	}
	return flash.Clear(c, sess.Session).Success("landing page hero slider item deleted successfully").Redirect(c, sess.Session, "/admin/landing_pages/hero-slider")
}

func (q *landingPageController) AdminFeatureItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := &models.LandingPageItemFilter{}
	rows := []*models.LandingPageItem{}
	request := &models.LandingPage{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Fitur Laukpauk",
		"limits", models.Limits,
		"request", request,
		"landingPage", request,
		"rows", rows,
		"paginationRequest", paginationRequest,
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
		},
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameFeature),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_feature_index", params.Push(
			"flash", models.ErrLandingPageNotFound.Error(),
		), "admin/layouts/authenticated")
	}
	params.Push(
		"request", landingPage,
		"landingPage", landingPage,
	)
	paginationRequest, err = sanitizers.FindLandingPageItems(ctx, c)
	params.Push("paginationRequest", paginationRequest)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("admin/landing_page_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.landingPageItemService.FindLandingPageItems(
		ctx,
		paginationRequest.WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFeature),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("admin/landing_page_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_feature_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFeatureUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameFeature),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	if landingPage == nil {
		return flash.Danger(models.ErrLandingPageNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	request, _, _, err := sanitizers.ValidateLandingPageCustomAttributes(ctx, c, models.LandingPageNameFeature)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	landingPage.CustomAttributes = request
	landingPage.UpdatedBy = currentAdmin.ID
	landingPage.UpdatedAt = time.Now()
	landingPage.DeletedBy = nil
	landingPage.DeletedAt = nil
	landingPage.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.landingPageService.UpdateLandingPage(ctx, tx, landingPage); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPage")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	return flash.Clear(c, sess.Session).Success("landing page feature updated successfully").Redirect(c, sess.Session, "/admin/landing_pages/feature")
}

func (q *landingPageController) AdminFeatureItemNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Fitur",
		"request", request,
		"landingPageItem", request,
	)
	return c.Render("admin/landing_page_feature_new", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFeatureItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemCreate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Fitur",
		"request", request,
		"landingPageItem", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameFeature)
	params.Push(
		"request", request,
		"landingPageItem", request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_feature_new", params.Push(
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
		return c.Render("admin/landing_page_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
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
	if _, err = q.landingPageItemService.CreateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageItem")
		return c.Render("admin/landing_page_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_feature_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(nil),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page feature item created successfully").Redirect(c, sess.Session, "/admin/landing_pages/feature")
}

func (q *landingPageController) AdminFeatureItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemEdit"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFeature).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	return c.Render("admin/landing_page_feature_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFeatureItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFeature).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameFeature)
	if photo == nil {
		request.Photo = landingPageItem.Photo
	}
	params.Push(
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageItem.ID
	request.RecordStatus = landingPageItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_feature_edit", params.Push(
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
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return c.Render("admin/landing_page_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_feature_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page feature item updated successfully").Redirect(c, sess.Session, "/admin/landing_pages/feature")
}

func (q *landingPageController) AdminFeatureItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemDelete"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFeature).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	landingPageItem.RecordStatus = models.RecordStatusDeleted
	landingPageItem.DeletedBy = &currentAdmin.ID
	landingPageItem.DeletedAt = new(time.Now())
	landingPageItem.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, landingPageItem); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/feature")
	}
	return flash.Clear(c, sess.Session).Success("landing page feature item deleted successfully").Redirect(c, sess.Session, "/admin/landing_pages/feature")
}

func (q *landingPageController) AdminWhyItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := &models.LandingPageItemFilter{}
	rows := []*models.LandingPageItem{}
	request := &models.LandingPage{}
	pagination := &models.Pagination{}
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
		"landingPage", request,
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
		},
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameWhy),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_why_index", params.Push(
			"flash", models.ErrLandingPageNotFound.Error(),
		), "admin/layouts/authenticated")
	}
	params.Push(
		"request", landingPage,
		"landingPage", landingPage,
	)
	paginationRequest, err = sanitizers.FindLandingPageItems(ctx, c)
	params.Push("paginationRequest", paginationRequest)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("admin/landing_page_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}

	if rows, pagination, err = q.landingPageItemService.FindLandingPageItems(
		ctx,
		paginationRequest.WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameWhy),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("admin/landing_page_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_why_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminWhyUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameWhy),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	if landingPage == nil {
		return flash.Danger(models.ErrLandingPageNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	request, _, _, err := sanitizers.ValidateLandingPageCustomAttributes(ctx, c, models.LandingPageNameWhy)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	landingPage.CustomAttributes = request
	landingPage.UpdatedBy = currentAdmin.ID
	landingPage.UpdatedAt = time.Now()
	landingPage.DeletedBy = nil
	landingPage.DeletedAt = nil
	landingPage.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.landingPageService.UpdateLandingPage(ctx, tx, landingPage); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPage")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	return flash.Clear(c, sess.Session).Success("landing page why updated successfully").Redirect(c, sess.Session, "/admin/landing_pages/why")
}

func (q *landingPageController) AdminWhyItemNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Penjelasan Baru",
		"request", request,
		"landingPageItem", request,
	)
	return c.Render("admin/landing_page_why_new", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminWhyItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemCreate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Penjelasan Baru",
		"request", request,
		"landingPageItem", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameWhy)
	params.Push(
		"request", request,
		"landingPageItem", request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_why_new", params.Push(
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
		return c.Render("admin/landing_page_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
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
	if _, err = q.landingPageItemService.CreateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageItem")
		return c.Render("admin/landing_page_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_why_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(nil),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page why item created successfully").Redirect(c, sess.Session, "/admin/landing_pages/why")
}

func (q *landingPageController) AdminWhyItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemEdit"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameWhy).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	return c.Render("admin/landing_page_why_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminWhyItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameWhy).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameWhy)
	if photo == nil {
		request.Photo = landingPageItem.Photo
	}
	params.Push(
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageItem.ID
	request.RecordStatus = landingPageItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_why_edit", params.Push(
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
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return c.Render("admin/landing_page_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_why_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page why item updated successfully").Redirect(c, sess.Session, "/admin/landing_pages/why")
}

func (q *landingPageController) AdminWhyItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemDelete"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameWhy).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	landingPageItem.RecordStatus = models.RecordStatusDeleted
	landingPageItem.DeletedBy = &currentAdmin.ID
	landingPageItem.DeletedAt = new(time.Now())
	landingPageItem.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, landingPageItem); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/why")
	}
	return flash.Clear(c, sess.Session).Success("landing page why item deleted successfully").Redirect(c, sess.Session, "/admin/landing_pages/why")
}

func (q *landingPageController) AdminCommitmentIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminCommitmentIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPage{
		CustomAttributes: map[string]string{},
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Komitmen",
		"request", request,
		"landingPage", request,
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameCommitment),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_commitment_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_commitment_index", params.Push(
			"flash", models.ErrLandingPageNotFound,
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_commitment_index", params.Push(
		"request", landingPage,
		"existing", landingPage,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminCommitmentUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminCommitmentUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPage{
		CustomAttributes: map[string]string{},
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Komitmen",
		"request", request,
		"landingPage", request,
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameCommitment),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_commitment_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_commitment_index", params.Push(
			"flash", flash.Danger(models.ErrLandingPageNotFound.Error()),
		), "admin/layouts/authenticated")
	}
	request.CustomAttributes, _, _, err = sanitizers.ValidateLandingPageCustomAttributes(ctx, c, models.LandingPageNameCommitment)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return c.Render("admin/landing_page_commitment_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	landingPage.CustomAttributes = request.CustomAttributes
	landingPage.UpdatedBy = currentAdmin.ID
	landingPage.UpdatedAt = time.Now()
	landingPage.DeletedBy = nil
	landingPage.DeletedAt = nil
	landingPage.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_commitment_index", params.Push(
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
	response, err := q.landingPageService.UpdateLandingPage(ctx, tx, landingPage)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPage")
		return c.Render("admin/landing_page_commitment_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_commitment_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_commitment_index", params.Push(
		"flash", flash.Clear(c, sess.Session).Success("landing page commitment updated successfully"),
		"request", response,
		"landingPage", response,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFreeShoppingIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFreeShoppingIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPage{
		CustomAttributes: map[string]string{},
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Uang Belanja Gratis",
		"request", request,
		"landingPage", request,
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameFreeShopping),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
			"flash", models.ErrLandingPageNotFound.Error(),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_free_shopping_index", params.Push(
		"request", landingPage,
		"landingPage", landingPage,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFreeShoppingUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFreeShoppingUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPage{
		CustomAttributes: map[string]string{},
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Uang Belanja Gratis",
		"request", request,
		"landingPage", request,
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameFreeShopping),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
			"flash", flash.Danger(models.ErrLandingPageNotFound.Error()),
		), "admin/layouts/authenticated")
	}
	customAttributes, photo, mimetype, err := sanitizers.ValidateLandingPageCustomAttributes(ctx, c, models.LandingPageNameFreeShopping)
	if photo == nil {
		customAttributes["photo"] = landingPage.CustomAttributes["photo"]
	}
	request.CustomAttributes = customAttributes
	params.Push(
		"request", request,
		"landingPage", landingPage,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	landingPage.CustomAttributes = request.CustomAttributes
	landingPage.UpdatedBy = currentAdmin.ID
	landingPage.UpdatedAt = time.Now()
	landingPage.DeletedBy = nil
	landingPage.DeletedAt = nil
	landingPage.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
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
	response, err := q.landingPageService.UpdateLandingPage(ctx, tx, landingPage)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPage")
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.CustomAttributes["photo"] != "" {
		if err = helper.SaveFile(c, photo, mimetype, request.CustomAttributes["photo"]); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_free_shopping_index", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_free_shopping_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_free_shopping_index", params.Push(
		"flash", flash.Clear(c, sess.Session).Success("landing page free shopping updated successfully"),
		"request", response,
		"landingPage", response,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminKitchenHeroIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminKitchenHeroIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPage{
		CustomAttributes: map[string]string{},
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Pahlawan Dapur",
		"request", request,
		"landingPage", request,
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameKitchenHero),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
			"flash", models.ErrLandingPageNotFound.Error(),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
		"request", landingPage,
		"landingPage", landingPage,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminKitchenHeroUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminKitchenHeroUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPage{
		CustomAttributes: map[string]string{},
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Pahlawan Dapur",
		"request", request,
		"landingPage", request,
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameKitchenHero),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
			"flash", flash.Danger(models.ErrLandingPageNotFound.Error()),
		), "admin/layouts/authenticated")
	}
	customAttributes, photo, mimetype, err := sanitizers.ValidateLandingPageCustomAttributes(ctx, c, models.LandingPageNameKitchenHero)
	if photo == nil {
		customAttributes["photo"] = landingPage.CustomAttributes["photo"]
	}
	request.CustomAttributes = customAttributes
	params.Push(
		"request", request,
		"landingPage", landingPage,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	landingPage.CustomAttributes = request.CustomAttributes
	landingPage.UpdatedBy = currentAdmin.ID
	landingPage.UpdatedAt = time.Now()
	landingPage.DeletedBy = nil
	landingPage.DeletedAt = nil
	landingPage.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
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
	if _, err = q.landingPageService.UpdateLandingPage(ctx, tx, landingPage); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPage")
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.CustomAttributes["photo"] != "" {
		if err = helper.SaveFile(c, photo, mimetype, request.CustomAttributes["photo"]); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_kitchen_hero_index", params.Push(
		"flash", flash.Clear(c, sess.Session).Success("landing page kitchen hero updated successfully"),
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFaqItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := &models.LandingPageItemFilter{}
	rows := []*models.LandingPageItem{}
	request := &models.LandingPage{
		CustomAttributes: map[string]string{},
	}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Halaman Bantuan",
		"limits", models.Limits,
		"paginationRequest", paginationRequest,
		"request", request,
		"landingPage", request,
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
		},
	)
	landingPage, err := q.landingPageService.FindLandingPage(
		ctx,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageNameFAQ),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPage")
		return c.Render("admin/landing_page_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPage == nil {
		return c.Render("admin/landing_page_faq_index", params.Push(
			"flash", models.ErrLandingPageNotFound.Error(),
		), "admin/layouts/authenticated")
	}
	params.Push(
		"request", landingPage,
		"landingPage", landingPage,
	)
	paginationRequest, err = sanitizers.FindLandingPageItems(ctx, c)
	params.Push("paginationRequest", paginationRequest)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("admin/landing_page_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.landingPageItemService.FindLandingPageItems(
		ctx,
		paginationRequest.WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFAQ),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("admin/landing_page_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_faq_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFaqItemNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Bantuan Baru",
		"request", request,
		"landingPageItem", request,
	)
	return c.Render("admin/landing_page_faq_new", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFaqItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemCreate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Bantuan Baru",
		"request", request,
		"landingPageItem", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameFAQ)
	params.Push(
		"request", request,
		"landingPageItem", request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_faq_new", params.Push(
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
		return c.Render("admin/landing_page_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
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
	if _, err = q.landingPageItemService.CreateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageItem")
		return c.Render("admin/landing_page_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_faq_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(nil),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page FAQ item created successfully").Redirect(c, sess.Session, "/admin/landing_pages/faq")
}

func (q *landingPageController) AdminFaqItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemEdit"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Bantuan",
		"request", request,
		"landingPageItem", request,
	)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFAQ).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return c.Render("admin/landing_page_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	params.Push(
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	return c.Render("admin/landing_page_faq_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFaqItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemUpdate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.LandingPageItem{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Bantuan",
		"request", request,
		"landingPageItem", request,
	)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFAQ).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	params.Push(
		"request", landingPageItem,
		"landingPageItem", landingPageItem,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateLandingPageItem(ctx, c, models.LandingPageNameFAQ)
	if photo == nil {
		request.Photo = landingPageItem.Photo
	}
	params.Push(
		"request", request,
		"landingPageItem", landingPageItem,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageItem.ID
	request.RecordStatus = landingPageItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_faq_edit", params.Push(
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
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return c.Render("admin/landing_page_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/landing_page_faq_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).Success("landing page FAQ item updated successfully").Redirect(c, sess.Session, "/admin/landing_pages/faq")
}

func (q *landingPageController) AdminFaqItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemDelete"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	landingPageItem, err := q.landingPageItemService.FindLandingPageItem(
		ctx,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFAQ).
			WithLandingPageItemIDs(landingPageItemID),
	)
	if err != nil && !errors.Is(err, models.ErrLandingPageItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	if landingPageItem == nil {
		return flash.Danger(models.ErrLandingPageItemNotFound.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	landingPageItem.RecordStatus = models.RecordStatusDeleted
	landingPageItem.DeletedBy = &currentAdmin.ID
	landingPageItem.DeletedAt = new(time.Now())
	landingPageItem.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.landingPageItemService.UpdateLandingPageItem(ctx, tx, landingPageItem); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_pages/faq")
	}
	return flash.Clear(c, sess.Session).Success("landing page FAQ item deleted successfully").Redirect(c, sess.Session, "/admin/landing_pages/faq")
}
