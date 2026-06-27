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
	blockController struct {
		userService    services.UserService
		blockService   services.BlockService
		blogService    services.BlogService
		messageService services.MessageService
	}
)

func NewBlockController(
	userService services.UserService,
	blockService services.BlockService,
	blogService services.BlogService,
	messageService services.MessageService,
) *blockController {
	return new(blockController{
		userService:    userService,
		blockService:   blockService,
		blogService:    blogService,
		messageService: messageService,
	})
}

func (q *blockController) Mount(r fiber.Router) {
	r.Get("", q.Index).
		Get("/blogs", q.BlogIndex).
		Get("/blogs/:id", q.BlogShow).
		Get("/contact-us", q.NewMessage).
		Post("/contact-us", q.CreateMessage).
		Get("/faq", q.FaqIndex).
		Get("/privacy-policy", q.PrivacyPolicy)
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	admin := r.Group("/admin", adminSessionAuth)
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

func (q *blockController) Index(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-Index"
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	var hero,
		feature,
		why,
		freeShopping,
		commitment,
		kitchenHero *models.Block
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
	blocks, _, err := q.blockService.FindBlocks(
		c,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("block/index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "block/_layout")
	}
	mapBlocks := map[models.BlockName]*models.Block{}
	for _, row := range blocks {
		mapBlocks[row.Name] = row
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		c,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("block/index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "block/_layout")
	}
	for _, row := range blockItems {
		block := mapBlocks[row.BlockName]
		block.Items = append(block.Items, row)
		mapBlocks[row.BlockName] = block
	}
	return c.Render("block/index", params.Push(
		"hero", mapBlocks[models.BlockNameHero],
		"feature", mapBlocks[models.BlockNameFeature],
		"why", mapBlocks[models.BlockNameWhy],
		"freeShopping", mapBlocks[models.BlockNameFreeShopping],
		"commitment", mapBlocks[models.BlockNameCommitment],
		"kitchenHero", mapBlocks[models.BlockNameKitchenHero],
	), "block/_layout")
}

func (q *blockController) BlogIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-BlogIndex"
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
		return c.Render("block/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "block/_layout")
	}
	request = request.WithRecordStatus(models.RecordStatusActive)
	if rows, pagination, err = q.blogService.FindBlogs(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("block/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "block/_layout")
	}
	return c.Render("block/blog_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "block/_layout")
}

func (q *blockController) BlogShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-BlogShow"
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
	return c.Render("block/blog_show", params.Push(
		"pageTitle", helper.Sprintf("%s | Belanja Sayur Online", blog.Subject),
		"blog", blog,
	), "block/_layout")
}

func (q *blockController) NewMessage(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Message)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Hubungi Kami | Belanja Sayur Online",
		"request", request,
	)
	return c.Render("block/message", params, "block/_layout")
}

func (q *blockController) CreateMessage(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-CreateMessage"
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Message)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Hubungi Kami | Belanja Sayur Online",
		"request", request,
	)
	request, statusCode, err := sanitizers.ValidateMessage(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateMessage")
		c.Response().SetStatusCode(statusCode)
		return c.Render("block/message", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "block/_layout")
	}
	request.CreatedAt = time.Now()
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("block/message", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "block/_layout")
	}
	if _, err = q.messageService.CreateMessage(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateMessage")
		return c.Render("block/message", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "block/_layout")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("block/message", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "block/_layout")
	}
	return c.Render("block/message", params.Push(
		"flash", flash.Success("your message submitted successfully, we will contact you shortly"),
		"request", request,
	), "block/_layout")
}

func (q *blockController) FaqIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-FaqIndex"
	flash := helper.NewFlashMessage()
	csrfToken := csrf.TokenFromContext(c)
	var rows []*models.BlockItem
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"pageTitle", "Bantuan | Belanja Sayur Online",
		"rows", rows,
	)
	if rows, _, err = q.blockService.FindBlockItems(
		c,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameFAQ),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("block/faq", params.Push(
			"flash", flash.Danger(err.Error()),
		), "block/_layout")
	}
	return c.Render("block/faq", params.Push(
		"rows", rows,
	))
}

func (q *blockController) PrivacyPolicy(c fiber.Ctx) (err error) {
	return c.Render("block/privacy_policy", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"pageTitle", "Kebijakan Privasi | Belanja Sayur Online",
	), "block/_layout")
}

func (q *blockController) AdminHeroItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminHeroItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItemFilter)
	rows := []*models.BlockItem{}
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
	if request, err = sanitizers.FindBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive).
		WithBlockNames(models.BlockNameHero)
	if rows, pagination, err = q.blockService.FindBlockItems(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/block_hero_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *blockController) AdminHeroItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem{
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
	return c.Render("admin/block_hero_new", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminHeroItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminHeroItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Hero Slider",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameHero)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_hero_new", params.Push(
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
		return c.Render("admin/block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.CreateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlockItem")
		return c.Render("admin/block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameHero, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_hero_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item created successfully").Redirect(c, sess.Session, "/admin/blocks/hero")
}

func (q *blockController) AdminHeroItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminHeroItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameHero).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	request := blockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"blockItemID", request.ID,
	)
	return c.Render("admin/block_hero_edit", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminHeroItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminHeroItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameHero).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	blockItem := blockItems[0]
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Hero Slider",
		"request", request,
		"blockItemID", blockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameHero, blockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = blockItem.ID
	request.RecordStatus = blockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = blockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return c.Render("admin/block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameHero, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_hero_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item updated successfully").Redirect(c, sess.Session, "/admin/blocks/hero")
}

func (q *blockController) AdminHeroItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminHeroItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameHero).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	request := blockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/hero")
	}
	return flash.Success("block item deleted successfully").Redirect(c, sess.Session, "/admin/blocks/hero")
}

func (q *blockController) saveFile(c fiber.Ctx, file *multipart.FileHeader, parentName models.BlockName, filename *string) error {
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
	_, _ = fmt.Fprintf(&sb, models.BlockPhotoPath, parentName, *filename)
	return c.SaveFile(file, sb.String())
}

func (q *blockController) AdminFeatureItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFeatureItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.BlockItemFilter)
	rows := []*models.BlockItem{}
	request := new(models.Block)
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
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameFeature),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("admin/block_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(blocks) == 0 {
		return c.Render("admin/block_feature_index", params.Push(
			"flash", "block not found",
		), "admin/layouts/authenticated")
	}
	request = blocks[0]
	existing = request
	params.Push(
		"request", request,
		"existing", existing,
	)
	if paginationRequest, err = sanitizers.FindBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithBlockNames(models.BlockNameFeature)
	if rows, pagination, err = q.blockService.FindBlockItems(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_feature_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/block_feature_index", params.Push(
		"rows", rows,
		"paginationRequest", paginationRequest,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *blockController) AdminFeatureUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFeatureUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameFeature),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if len(blocks) == 0 {
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	block := blocks[0]
	request, _, err := sanitizers.ValidateBlockCustomAttributes(ctx, c, models.BlockNameFeature)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	block.CustomAttributes = request
	block.UpdatedBy = currentAdmin.ID
	block.UpdatedAt = time.Now()
	block.DeletedBy = nil
	block.DeletedAt = nil
	block.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if _, err = q.blockService.UpdateBlock(ctx, tx, block); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	return flash.Success("block updated successfully").Redirect(c, sess.Session, "/admin/blocks/feature")
}

func (q *blockController) AdminFeatureItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Fitur",
		"request", request,
	)
	return c.Render("admin/block_feature_new", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminFeatureItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFeatureItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Fitur",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameFeature)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_feature_new", params.Push(
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
		return c.Render("admin/block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.CreateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlockItem")
		return c.Render("admin/block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameFeature, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_feature_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item created successfully").Redirect(c, sess.Session, "/admin/blocks/feature")
}

func (q *blockController) AdminFeatureItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFeatureItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameFeature).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	request := blockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", request,
		"blockItemID", request.ID,
	)
	return c.Render("admin/block_feature_edit", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminFeatureItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFeatureItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameFeature).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	blockItem := blockItems[0]
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Fitur Laukpauk",
		"request", request,
		"blockItemID", blockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameFeature, blockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = blockItem.ID
	request.RecordStatus = blockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = blockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return c.Render("admin/block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameFeature, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_feature_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item updated successfully").Redirect(c, sess.Session, "/admin/blocks/feature")
}

func (q *blockController) AdminFeatureItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFeatureItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameFeature).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	request := blockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/feature")
	}
	return flash.Success("block item deleted successfully").Redirect(c, sess.Session, "/admin/blocks/feature")
}

func (q *blockController) AdminWhyItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminWhyItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.BlockItemFilter)
	rows := []*models.BlockItem{}
	request := new(models.Block)
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
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameWhy),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("admin/block_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(blocks) == 0 {
		return c.Render("admin/block_why_index", params.Push(
			"flash", "block not found",
		), "admin/layouts/authenticated")
	}
	request = blocks[0]
	params.Push("request", request)
	if paginationRequest, err = sanitizers.FindBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithBlockNames(models.BlockNameWhy)
	if rows, pagination, err = q.blockService.FindBlockItems(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_why_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/block_why_index", params.Push(
		"rows", rows,
		"paginationRequest", paginationRequest,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *blockController) AdminWhyUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminWhyUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameWhy),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if len(blocks) == 0 {
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	block := blocks[0]
	request, _, err := sanitizers.ValidateBlockCustomAttributes(ctx, c, models.BlockNameWhy)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	block.CustomAttributes = request
	block.UpdatedBy = currentAdmin.ID
	block.UpdatedAt = time.Now()
	block.DeletedBy = nil
	block.DeletedAt = nil
	block.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if _, err = q.blockService.UpdateBlock(ctx, tx, block); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	return flash.Success("block updated successfully").Redirect(c, sess.Session, "/admin/blocks/why")
}

func (q *blockController) AdminWhyItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Penjelasan Baru",
		"request", request,
	)
	return c.Render("admin/block_why_new", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminWhyItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminWhyItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Penjelasan Baru",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameWhy)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_why_new", params.Push(
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
		return c.Render("admin/block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.CreateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlockItem")
		return c.Render("admin/block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameWhy, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_why_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item created successfully").Redirect(c, sess.Session, "/admin/blocks/why")
}

func (q *blockController) AdminWhyItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminWhyItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameWhy).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	request := blockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", request,
		"blockItemID", request.ID,
	)
	return c.Render("admin/block_why_edit", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminWhyItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminWhyItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameWhy).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	blockItem := blockItems[0]
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Penjelasan",
		"request", request,
		"blockItemID", blockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameWhy, blockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = blockItem.ID
	request.RecordStatus = blockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = blockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return c.Render("admin/block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameWhy, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_why_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item updated successfully").Redirect(c, sess.Session, "/admin/blocks/why")
}

func (q *blockController) AdminWhyItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminWhyItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameWhy).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/why")
	}
	request := blockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/why")
	}
	return flash.Success("block item deleted successfully").Redirect(c, sess.Session, "/admin/blocks/why")
}

func (q *blockController) AdminCommitmentIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminCommitmentIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	existing := new(models.Block)
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
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameCommitment),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("admin/block_commitment_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(blocks) == 0 {
		return c.Render("admin/block_commitment_index", params.Push(
			"flash", "block not found",
		), "admin/layouts/authenticated")
	}
	existing = blocks[0]
	request = existing
	return c.Render("admin/block_commitment_index", params.Push(
		"request", request,
		"existing", existing,
	), "admin/layouts/authenticated")
}

func (q *blockController) AdminCommitmentUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminCommitmentUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameCommitment),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/commitment")
	}
	if len(blocks) == 0 {
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/commitment")
	}
	block := blocks[0]
	request, _, err := sanitizers.ValidateBlockCustomAttributes(ctx, c, models.BlockNameCommitment)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/commitment")
	}
	block.CustomAttributes = request
	block.UpdatedBy = currentAdmin.ID
	block.UpdatedAt = time.Now()
	block.DeletedBy = nil
	block.DeletedAt = nil
	block.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/commitment")
	}
	if _, err = q.blockService.UpdateBlock(ctx, tx, block); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/commitment")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/commitment")
	}
	return flash.Success("block updated successfully").Redirect(c, sess.Session, "/admin/blocks/commitment")
}

func (q *blockController) AdminFreeShoppingIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFreeShoppingIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	existing := new(models.Block)
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
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameFreeShopping),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("admin/block_free_shopping_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(blocks) == 0 {
		return c.Render("admin/block_free_shopping_index", params.Push(
			"flash", "block not found",
		), "admin/layouts/authenticated")
	}
	existing = blocks[0]
	request = existing
	return c.Render("admin/block_free_shopping_index", params.Push(
		"request", request,
		"existing", existing,
	), "admin/layouts/authenticated")
}

func (q *blockController) AdminFreeShoppingUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFreeShoppingUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameFreeShopping),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/free-shopping")
	}
	if len(blocks) == 0 {
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/free-shopping")
	}
	block := blocks[0]
	request, photo, err := sanitizers.ValidateBlockCustomAttributes(ctx, c, models.BlockNameFreeShopping)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/free-shopping")
	}
	if photo == nil {
		request["photo"] = block.CustomAttributes["photo"]
	}
	block.CustomAttributes = request
	block.UpdatedBy = currentAdmin.ID
	block.UpdatedAt = time.Now()
	block.DeletedBy = nil
	block.DeletedAt = nil
	block.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/free-shopping")
	}
	if _, err = q.blockService.UpdateBlock(ctx, tx, block); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/free-shopping")
	}
	if err = q.saveFile(c, photo, models.BlockNameFreeShopping, new(request["photo"])); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/free-shopping")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/free-shopping")
	}
	return flash.Success("block updated successfully").Redirect(c, sess.Session, "/admin/blocks/free-shopping")
}

func (q *blockController) AdminKitchenHeroIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminKitchenHeroIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	existing := new(models.Block)
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
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameKitchenHero),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("admin/block_kitchen_hero_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(blocks) == 0 {
		return c.Render("admin/block_kitchen_hero_index", params.Push(
			"flash", "block not found",
		), "admin/layouts/authenticated")
	}
	existing = blocks[0]
	request = existing
	return c.Render("admin/block_kitchen_hero_index", params.Push(
		"request", request,
		"existing", existing,
	), "admin/layouts/authenticated")
}

func (q *blockController) AdminKitchenHeroUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminKitchenHeroUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameKitchenHero),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
	}
	if len(blocks) == 0 {
		return flash.Danger("block not found").Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
	}
	block := blocks[0]
	request, photo, err := sanitizers.ValidateBlockCustomAttributes(ctx, c, models.BlockNameKitchenHero)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockCustomAttributes")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
	}
	if photo == nil {
		request["photo"] = block.CustomAttributes["photo"]
	}
	block.CustomAttributes = request
	block.UpdatedBy = currentAdmin.ID
	block.UpdatedAt = time.Now()
	block.DeletedBy = nil
	block.DeletedAt = nil
	block.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
	}
	if _, err = q.blockService.UpdateBlock(ctx, tx, block); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlock")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
	}
	if err = q.saveFile(c, photo, models.BlockNameKitchenHero, new(request["photo"])); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
	}
	return flash.Success("block updated successfully").Redirect(c, sess.Session, "/admin/blocks/kitchen-hero")
}

func (q *blockController) AdminFaqItemIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFaqItemIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	paginationRequest := new(models.BlockItemFilter)
	rows := []*models.BlockItem{}
	request := new(models.Block)
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
	blocks, _, err := q.blockService.FindBlocks(
		ctx,
		models.NewBlockFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithNames(models.BlockNameFAQ),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return c.Render("admin/block_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if len(blocks) == 0 {
		return c.Render("admin/block_faq_index", params.Push(
			"flash", "block not found",
		), "admin/layouts/authenticated")
	}
	request = blocks[0]
	params.Push("request", request)
	if paginationRequest, err = sanitizers.FindBlockItems(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	paginationRequest = paginationRequest.WithRecordStatus(models.RecordStatusActive).
		WithBlockNames(models.BlockNameFAQ)
	if rows, pagination, err = q.blockService.FindBlockItems(
		c,
		paginationRequest,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return c.Render("admin/block_faq_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/block_faq_index", params.Push(
		"rows", rows,
		"paginationRequest", paginationRequest,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *blockController) AdminFaqItemNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Bantuan Baru",
		"request", request,
	)
	return c.Render("admin/block_faq_new", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminFaqItemCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminWhyItemCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Bantuan Baru",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameFAQ)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_faq_new", params.Push(
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
		return c.Render("admin/block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.CreateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlockItem")
		return c.Render("admin/block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameFAQ, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_faq_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(nil),
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item created successfully").Redirect(c, sess.Session, "/admin/blocks/faq")
}

func (q *blockController) AdminFaqItemEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFaqItemEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameFAQ).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	request := blockItems[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Bantuan",
		"request", request,
		"blockItemID", request.ID,
	)
	return c.Render("admin/block_faq_edit", params, "admin/layouts/authenticated")
}

func (q *blockController) AdminFaqItemUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFaqItemUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameFAQ).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	blockItem := blockItems[0]
	request := new(models.BlockItem)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Bantuan",
		"request", request,
		"blockItemID", blockItem.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlockItem(ctx, c, models.BlockNameFAQ, blockItem.ID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlockItem")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = blockItem.ID
	request.RecordStatus = blockItem.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = blockItem.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return c.Render("admin/block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, models.BlockNameWhy, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/block_faq_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("block item updated successfully").Redirect(c, sess.Session, "/admin/blocks/faq")
}

func (q *blockController) AdminFaqItemDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlockController-AdminFaqItemDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blockItemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	blockItems, _, err := q.blockService.FindBlockItems(
		ctx,
		models.NewBlockItemFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlockNames(models.BlockNameFAQ).
			WithBlockItemIDs(blockItemID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	if len(blockItems) == 0 {
		return flash.Danger("block item not found").Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	request := blockItems[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	if _, err = q.blockService.UpdateBlockItem(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blocks/faq")
	}
	return flash.Success("block item deleted successfully").Redirect(c, sess.Session, "/admin/blocks/faq")
}
