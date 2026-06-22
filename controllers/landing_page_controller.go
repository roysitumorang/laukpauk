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
	landingPageController struct {
		userService             services.UserService
		landingPageBlockService services.LandingPageBlockService
		blogService             services.BlogService
	}
)

func NewLandingPageController(
	userService services.UserService,
	landingPageBlockService services.LandingPageBlockService,
	blogService services.BlogService,
) *landingPageController {
	return new(landingPageController{
		userService:             userService,
		landingPageBlockService: landingPageBlockService,
		blogService:             blogService,
	})
}

func (q *landingPageController) Mount(r fiber.Router) {
	r.Get("", q.Index).
		Get("/blogs", q.BlogIndex).
		Get("/blogs/:id", q.BlogShow).
		Get("/contact-us", q.FormContactUs).
		Post("/contact-us", q.SaveContactUs).
		Get("/faq", q.FaqIndex).
		Get("/privacy-policy", q.PrivacyPolicy)
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	admin := r.Group("/admin/blocks", adminSessionAuth)
	admin.Group("/hero").
		Get("", q.AdminHeroItemIndex).
		Get("/items/new", q.AdminHeroItemNew).
		Post("/items", q.AdminHeroItemCreate).
		Get("/items/:id/edit", q.AdminHeroItemEdit).
		Post("/items/:id", q.AdminHeroItemUpdate).
		Post("/items/:id/delete", q.AdminHeroItemDelete)
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

func (q *landingPageController) Index(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-Index"
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	var hero,
		feature,
		why,
		freeShopping,
		commitment,
		kitchenHero *models.LandingPageBlock
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Laukpauk | Belanja Sayur Online",
		"hero", hero,
		"feature", feature,
		"why", why,
		"freeShopping", freeShopping,
		"commitment", commitment,
		"kitchenHero", kitchenHero,
	)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		c,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("landing_page/index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	mapLandingPageBlocks := map[models.LandingPageBlockName]*models.LandingPageBlock{}
	for _, row := range landingPageBlocks {
		mapLandingPageBlocks[row.Name] = row
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		c,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("landing_page/index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	for _, row := range landingPageBlockItems {
		landingPageBlock := mapLandingPageBlocks[row.LandingPageBlockName]
		landingPageBlock.Items = append(landingPageBlock.Items, row)
		mapLandingPageBlocks[row.LandingPageBlockName] = landingPageBlock
	}
	return c.Render("landing_page/index", params.Push(
		"hero", mapLandingPageBlocks[models.LandingPageBlockNameHero],
		"feature", mapLandingPageBlocks[models.LandingPageBlockNameFeature],
		"why", mapLandingPageBlocks[models.LandingPageBlockNameWhy],
		"freeShopping", mapLandingPageBlocks[models.LandingPageBlockNameFreeShopping],
		"commitment", mapLandingPageBlocks[models.LandingPageBlockNameCommitment],
		"kitchenHero", mapLandingPageBlocks[models.LandingPageBlockNameKitchenHero],
	), "landing_page/_layout")
}

func (q *landingPageController) BlogIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-BlogIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlogFilter)
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
	if request, err = sanitizers.FindBlogs(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("landing_page/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	request = request.WithRecordStatus(models.RecordStatusActive)
	if rows, pagination, err = q.blogService.FindBlogs(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("landing_page/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	return c.Render("landing_page/blog_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "landing_page/_layout")
}

func (q *landingPageController) BlogShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-BlogShow"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
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
		WithRecordStatus(models.RecordStatusActive)
	slug := c.Params("id")
	if blogID, err := uuid.Parse(slug); err == nil {
		request = request.WithBlogIDs(blogID)
	} else {
		request = request.WithSlugs(slug)
	}
	blogs, _, err := q.blogService.FindBlogs(
		c,
		request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/blogs")
	}
	if len(blogs) == 0 {
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/blogs")
	}
	blog = blogs[0]
	return c.Render("landing_page/blog_show", params.Push(
		"pageTitle", fmt.Sprintf("%s | Belanja Sayur Online", blog.Subject),
		"blog", blog,
	), "landing_page/_layout")
}

func (q *landingPageController) FormContactUs(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.ContactUs)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Hubungi Kami | Belanja Sayur Online",
		"request", request,
	)
	return c.Render("landing_page/contact_us", params, "landing_page/_layout")
}

func (q *landingPageController) SaveContactUs(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-SaveContactUs"
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.ContactUs)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Hubungi Kami | Belanja Sayur Online",
		"request", request,
	)
	request, statusCode, err := sanitizers.ValidateContactUs(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateContactUs")
		c.Response().SetStatusCode(statusCode)
		return c.Render("landing_page/contact_us", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "landing_page/_layout")
	}
	return c.Render("landing_page/contact_us", params.Push(
		"flash", flash.Success("your message submitted successfully, we will contact you shortly"),
		"request", request,
	), "landing_page/_layout")
}

func (q *landingPageController) FaqIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-FaqIndex"
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	var rows []*models.LandingPageBlockItem
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Bantuan | Belanja Sayur Online",
		"rows", rows,
	)
	if rows, _, err = q.landingPageBlockService.FindLandingPageBlockItems(
		c,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameFAQ),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return c.Render("landing_page/faq", params.Push(
			"flash", flash.Danger(err.Error()),
		), "landing_page/_layout")
	}
	return c.Render("landing_page/faq", params.Push(
		"rows", rows,
	))
}

func (q *landingPageController) PrivacyPolicy(c fiber.Ctx) (err error) {
	return c.Render("landing_page/privacy_policy", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"pageTitle", "Kebijakan Privasi | Belanja Sayur Online",
	), "landing_page/_layout")
}

func (q *landingPageController) AdminHeroItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItemFilter)
	rows := []*models.LandingPageBlockItem{}
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
	if request, err = sanitizers.FindLandingPageBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return c.Render("admin/landing_page_block_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive).
		WithLandingPageBlockNames(models.LandingPageBlockNameHero)
	if rows, pagination, err = q.landingPageBlockService.FindLandingPageBlockItems(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
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

func (q *landingPageController) AdminHeroItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem{
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

func (q *landingPageController) AdminHeroItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameHero)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
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
	if _, err = q.landingPageBlockService.CreateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlockItem")
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
	return flash.Success("landing page block item created successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
}

func (q *landingPageController) AdminHeroItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameHero).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	request := landingPageBlockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"landingPageBlockItemID", request.ID,
	)
	return c.Render("admin/landing_page_block_hero_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminHeroItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameHero).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	landingPageBlockItem := landingPageBlockItems[0]
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"landingPageBlockItemID", landingPageBlockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameHero, landingPageBlockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageBlockItem.ID
	request.RecordStatus = landingPageBlockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = landingPageBlockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
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
	return flash.Success("landing page block item updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
}

func (q *landingPageController) AdminHeroItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminHeroItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameHero).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	request := landingPageBlockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
	}
	return flash.Success("landing page block item deleted successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/hero")
}

func (q *landingPageController) saveFile(c fiber.Ctx, file *multipart.FileHeader, parentName models.LandingPageBlockName, filename *string) error {
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

func (q *landingPageController) AdminFeatureItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.LandingPageBlockItemFilter)
	rows := []*models.LandingPageBlockItem{}
	request := new(models.LandingPageBlock)
	existing := request
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
		"existing", existing,
		"rows", rows,
		"paginationRequest", paginationRequest,
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
	existing = request
	params.Push(
		"request", request,
		"existing", existing,
	)
	if paginationRequest, err = sanitizers.FindLandingPageBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return c.Render("admin/landing_page_block_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithLandingPageBlockNames(models.LandingPageBlockNameFeature)
	if rows, pagination, err = q.landingPageBlockService.FindLandingPageBlockItems(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
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

func (q *landingPageController) AdminFeatureUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureUpdate"
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
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	landingPageBlock := landingPageBlocks[0]
	request, _, err := sanitizers.ValidateLandingPageBlockCustomAttributes(ctx, c, models.LandingPageBlockNameFeature)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, landingPageBlock); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	return flash.Success("landing page block updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
}

func (q *landingPageController) AdminFeatureItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem)
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

func (q *landingPageController) AdminFeatureItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Fitur",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameFeature)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
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
	if _, err = q.landingPageBlockService.CreateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlockItem")
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
	return flash.Success("landing page block item created successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
}

func (q *landingPageController) AdminFeatureItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameFeature).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	request := landingPageBlockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", request,
		"landingPageBlockItemID", request.ID,
	)
	return c.Render("admin/landing_page_block_feature_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFeatureItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameFeature).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	landingPageBlockItem := landingPageBlockItems[0]
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", request,
		"landingPageBlockItemID", landingPageBlockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameFeature, landingPageBlockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageBlockItem.ID
	request.RecordStatus = landingPageBlockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = landingPageBlockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
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
	return flash.Success("landing page block item updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
}

func (q *landingPageController) AdminFeatureItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFeatureItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameFeature).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	request := landingPageBlockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
	}
	return flash.Success("landing page block item deleted successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/feature")
}

func (q *landingPageController) AdminWhyItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.LandingPageBlockItemFilter)
	rows := []*models.LandingPageBlockItem{}
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
	if paginationRequest, err = sanitizers.FindLandingPageBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return c.Render("admin/landing_page_block_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithLandingPageBlockNames(models.LandingPageBlockNameWhy)
	if rows, pagination, err = q.landingPageBlockService.FindLandingPageBlockItems(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
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

func (q *landingPageController) AdminWhyUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyUpdate"
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
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	landingPageBlock := landingPageBlocks[0]
	request, _, err := sanitizers.ValidateLandingPageBlockCustomAttributes(ctx, c, models.LandingPageBlockNameWhy)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, landingPageBlock); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	return flash.Success("landing page block updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
}

func (q *landingPageController) AdminWhyItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem)
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

func (q *landingPageController) AdminWhyItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Penjelasan Baru",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameWhy)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
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
	if _, err = q.landingPageBlockService.CreateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlockItem")
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
	return flash.Success("landing page block item created successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
}

func (q *landingPageController) AdminWhyItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameWhy).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	request := landingPageBlockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", request,
		"landingPageBlockItemID", request.ID,
	)
	return c.Render("admin/landing_page_block_why_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminWhyItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameWhy).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	landingPageBlockItem := landingPageBlockItems[0]
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", request,
		"landingPageBlockItemID", landingPageBlockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameWhy, landingPageBlockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageBlockItem.ID
	request.RecordStatus = landingPageBlockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = landingPageBlockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
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
	return flash.Success("landing page block item updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
}

func (q *landingPageController) AdminWhyItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameWhy).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	request := landingPageBlockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
	}
	return flash.Success("landing page block item deleted successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/why")
}

func (q *landingPageController) AdminCommitmentIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminCommitmentIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	existing := new(models.LandingPageBlock)
	request := existing
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Komitmen",
		"request", request,
		"existing", existing,
	)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameCommitment),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_commitment_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(landingPageBlocks) == 0 {
		return c.Render("admin/landing_page_block_commitment_index", params.Push(
			"flash", "landing page block not found",
		), "admin/layouts/authenticated")
	}
	existing = landingPageBlocks[0]
	request = existing
	return c.Render("admin/landing_page_block_commitment_index", params.Push(
		"request", request,
		"existing", existing,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminCommitmentUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminCommitmentUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameCommitment),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/commitment")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/commitment")
	}
	landingPageBlock := landingPageBlocks[0]
	request, _, err := sanitizers.ValidateLandingPageBlockCustomAttributes(ctx, c, models.LandingPageBlockNameCommitment)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/commitment")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/commitment")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, landingPageBlock); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/commitment")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/commitment")
	}
	return flash.Success("landing page block updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/commitment")
}

func (q *landingPageController) AdminFreeShoppingIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFreeShoppingIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	existing := new(models.LandingPageBlock)
	request := existing
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Uang Belanja Gratis",
		"request", request,
		"existing", existing,
	)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameFreeShopping),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_free_shopping_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(landingPageBlocks) == 0 {
		return c.Render("admin/landing_page_block_free_shopping_index", params.Push(
			"flash", "landing page block not found",
		), "admin/layouts/authenticated")
	}
	existing = landingPageBlocks[0]
	request = existing
	return c.Render("admin/landing_page_block_free_shopping_index", params.Push(
		"request", request,
		"existing", existing,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFreeShoppingUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFreeShoppingUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameFreeShopping),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
	}
	landingPageBlock := landingPageBlocks[0]
	request, photo, err := sanitizers.ValidateLandingPageBlockCustomAttributes(ctx, c, models.LandingPageBlockNameFreeShopping)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
	}
	if photo == nil {
		request["photo"] = landingPageBlock.CustomAttributes["photo"]
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, landingPageBlock); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameFreeShopping, new(request["photo"])); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
	}
	return flash.Success("landing page block updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/free-shopping")
}

func (q *landingPageController) AdminKitchenHeroIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminKitchenHeroIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	existing := new(models.LandingPageBlock)
	request := existing
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Pahlawan Dapur",
		"request", request,
		"existing", existing,
	)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameKitchenHero),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_kitchen_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(landingPageBlocks) == 0 {
		return c.Render("admin/landing_page_block_kitchen_hero_index", params.Push(
			"flash", "landing page block not found",
		), "admin/layouts/authenticated")
	}
	existing = landingPageBlocks[0]
	request = existing
	return c.Render("admin/landing_page_block_kitchen_hero_index", params.Push(
		"request", request,
		"existing", existing,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminKitchenHeroUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminKitchenHeroUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameKitchenHero),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
	}
	if len(landingPageBlocks) == 0 {
		return flash.Danger("landing page block not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
	}
	landingPageBlock := landingPageBlocks[0]
	request, photo, err := sanitizers.ValidateLandingPageBlockCustomAttributes(ctx, c, models.LandingPageBlockNameKitchenHero)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
	}
	if photo == nil {
		request["photo"] = landingPageBlock.CustomAttributes["photo"]
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlock(ctx, tx, landingPageBlock); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameKitchenHero, new(request["photo"])); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
	}
	return flash.Success("landing page block updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/kitchen-hero")
}

func (q *landingPageController) AdminFaqItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.LandingPageBlockItemFilter)
	rows := []*models.LandingPageBlockItem{}
	request := new(models.LandingPageBlock)
	pagination := new(models.Pagination)
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
		"rows", rows,
		"pagination", pagination,
	)
	landingPageBlocks, _, err := q.landingPageBlockService.FindLandingPageBlocks(
		ctx,
		models.NewLandingPageBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.LandingPageBlockNameFAQ),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlocks")
		return c.Render("admin/landing_page_block_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(landingPageBlocks) == 0 {
		return c.Render("admin/landing_page_block_faq_index", params.Push(
			"flash", "landing page block not found",
		), "admin/layouts/authenticated")
	}
	request = landingPageBlocks[0]
	params.Push("request", request)
	if paginationRequest, err = sanitizers.FindLandingPageBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return c.Render("admin/landing_page_block_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithLandingPageBlockNames(models.LandingPageBlockNameFAQ)
	if rows, pagination, err = q.landingPageBlockService.FindLandingPageBlockItems(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return c.Render("admin/landing_page_block_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/landing_page_block_faq_index", params.Push(
		"rows", rows,
		"paginationRequest", paginationRequest,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFaqItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Bantuan Baru",
		"request", request,
	)
	return c.Render("admin/landing_page_block_faq_new", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFaqItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminWhyItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Bantuan Baru",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameFAQ)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_faq_new", params.Push(
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
		return c.Render("admin/landing_page_block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.CreateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateLandingPageBlockItem")
		return c.Render("admin/landing_page_block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameFAQ, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("landing page block item created successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
}

func (q *landingPageController) AdminFaqItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameFAQ).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	request := landingPageBlockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Bantuan",
		"request", request,
		"landingPageBlockItemID", request.ID,
	)
	return c.Render("admin/landing_page_block_faq_edit", params, "admin/layouts/authenticated")
}

func (q *landingPageController) AdminFaqItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameFAQ).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	landingPageBlockItem := landingPageBlockItems[0]
	request := new(models.LandingPageBlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Bantuan",
		"request", request,
		"landingPageBlockItemID", landingPageBlockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateLandingPageBlockItem(ctx, c, models.LandingPageBlockNameFAQ, landingPageBlockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLandingPageBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/landing_page_block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = landingPageBlockItem.ID
	request.RecordStatus = landingPageBlockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = landingPageBlockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/landing_page_block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
		return c.Render("admin/landing_page_block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.LandingPageBlockNameWhy, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/landing_page_block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/landing_page_block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("landing page block item updated successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
}

func (q *landingPageController) AdminFaqItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "LandingPageController-AdminFaqItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	landingPageBlockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	landingPageBlockItems, _, err := q.landingPageBlockService.FindLandingPageBlockItems(
		ctx,
		models.NewLandingPageBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithLandingPageBlockNames(models.LandingPageBlockNameFAQ).
			WithLandingPageBlockItemIDs(landingPageBlockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindLandingPageBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	if len(landingPageBlockItems) == 0 {
		return flash.Danger("landing page block item not found").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	request := landingPageBlockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	if _, err = q.landingPageBlockService.UpdateLandingPageBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateLandingPageBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
	}
	return flash.Success("landing page block item deleted successfully").Redirect(c, sess.Session, "/admin/landing_page/blocks/faq")
}
