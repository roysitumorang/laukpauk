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
	homeController struct {
		adminService           services.AdminService
		landingPageService     services.LandingPageService
		landingPageItemService services.LandingPageItemService
		blogService            services.BlogService
		messageService         services.MessageService
	}
)

func NewHomeController(
	adminService services.AdminService,
	landingPageService services.LandingPageService,
	landingPageItemService services.LandingPageItemService,
	blogService services.BlogService,
	messageService services.MessageService,
) *homeController {
	return &homeController{
		adminService:           adminService,
		landingPageService:     landingPageService,
		landingPageItemService: landingPageItemService,
		blogService:            blogService,
		messageService:         messageService,
	}
}

func (q *homeController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Get("", q.Index).
		Get("/blogs", q.BlogIndex).
		Get("/blogs/:id", q.BlogShow).
		Get("/contact-us", q.NewMessage).
		Post("/contact-us", q.CreateMessage).
		Get("/faq", q.FaqIndex).
		Get("/privacy-policy", q.PrivacyPolicy)
	r.Group("/admin").
		Get("", adminSessionAuth, q.AdminIndex)
}

func (q *homeController) Index(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "HomeController-Index"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	var heroSlider,
		feature,
		why,
		freeShopping,
		commitment,
		kitchenHero *models.LandingPage
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Laukpauk | Belanja Sayur Online",
		"heroSlider", heroSlider,
		"feature", feature,
		"why", why,
		"freeShopping", freeShopping,
		"commitment", commitment,
		"kitchenHero", kitchenHero,
	)
	landingPages, _, err := q.landingPageService.FindLandingPages(
		c,
		models.NewLandingPageFilter().
			WithRecordStatus(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPages")
		return c.Render("landing_page/index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	mapLandingPages := make(map[models.LandingPageName]*models.LandingPage, len(landingPages))
	for _, row := range landingPages {
		mapLandingPages[row.Name] = row
	}
	landingPageItems, _, err := q.landingPageItemService.FindLandingPageItems(
		c,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("landing_page/index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	for _, landingPageItem := range landingPageItems {
		landingPage := mapLandingPages[landingPageItem.LandingPageName]
		landingPage.Items = append(landingPage.Items, landingPageItem)
		mapLandingPages[landingPageItem.LandingPageName] = landingPage
	}
	return c.Render("landing_page/index", params.Push(
		"heroSlider", mapLandingPages[models.LandingPageNameHeroSlider],
		"feature", mapLandingPages[models.LandingPageNameFeature],
		"why", mapLandingPages[models.LandingPageNameWhy],
		"freeShopping", mapLandingPages[models.LandingPageNameFreeShopping],
		"commitment", mapLandingPages[models.LandingPageNameCommitment],
		"kitchenHero", mapLandingPages[models.LandingPageNameKitchenHero],
	), "landing_page/_layout")
}

func (q *homeController) BlogIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "HomeController-BlogIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.BlogFilter{}
	rows := []*models.Blog{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Blog | Belanja Sayur Online",
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
	)
	request, err = sanitizers.FindBlogs(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("landing_page/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	if rows, pagination, err = q.blogService.FindBlogs(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("landing_page/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	return c.Render("landing_page/blog_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "landing_page/_layout")
}

func (q *homeController) BlogShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "HomeController-BlogShow"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	var blog *models.Blog
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Blog | Belanja Sayur Online",
		"blog", blog,
	)
	request := models.NewBlogFilter().
		WithRecordStatuses(models.RecordStatusActive)
	slug := c.Params("id")
	if blogID, err := uuid.Parse(slug); err == nil {
		request.WithBlogIDs(blogID)
	} else {
		request.WithSlugs(&slug)
	}
	if blog, err = q.blogService.FindBlog(c, request); err != nil &&
		!errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/blogs")
	}
	if blog == nil {
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/blogs")
	}
	return c.Render("landing_page/blog_show", params.Push(
		"pageTitle", helper.Sprintf("%s | Belanja Sayur Online", blog.Subject),
		"blog", blog,
	), "landing_page/_layout")
}

func (q *homeController) NewMessage(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Message{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Hubungi Kami | Belanja Sayur Online",
		"request", request,
	)
	return c.Render("landing_page/message", params, "landing_page/_layout")
}

func (q *homeController) CreateMessage(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "HomeController-CreateMessage"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Message{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Hubungi Kami | Belanja Sayur Online",
		"request", request,
	)
	request, statusCode, err := sanitizers.ValidateMessage(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateMessage")
		c.Response().SetStatusCode(statusCode)
		return c.Render("landing_page/message", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	request.CreatedAt = time.Now()
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("landing_page/message", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.messageService.CreateMessage(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateMessage")
		return c.Render("landing_page/message", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("landing_page/message", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	return c.Render("landing_page/message", params.Push(
		"flash", flash.Success("your message submitted successfully, we will contact you shortly"),
	), "landing_page/_layout")
}

func (q *homeController) FaqIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "HomeController-FaqIndex"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	var rows []*models.LandingPageItem
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Bantuan | Belanja Sayur Online",
		"rows", rows,
	)
	if rows, _, err = q.landingPageItemService.FindLandingPageItems(
		c,
		models.NewLandingPageItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageNames(models.LandingPageNameFAQ),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageItems")
		return c.Render("landing_page/faq", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	return c.Render("landing_page/faq", params.Push(
		"rows", rows,
	))
}

func (q *homeController) PrivacyPolicy(c fiber.Ctx) (err error) {
	return c.Render("landing_page/privacy_policy", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"pageTitle", "Kebijakan Privasi | Belanja Sayur Online",
	), "landing_page/_layout")
}

func (q *homeController) AdminIndex(c fiber.Ctx) error {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	return c.Render("admin/home_index", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Dashboard Lauk Pauk",
	), "admin/layouts/authenticated")
}
