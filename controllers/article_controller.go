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

type articleController struct {
	sellerService       services.SellerService
	articleService      services.ArticleService
	notificationService services.NotificationService
}

func NewArticleController(
	sellerService services.SellerService,
	articleService services.ArticleService,
	notificationService services.NotificationService,
) *articleController {
	return &articleController{
		sellerService:       sellerService,
		articleService:      articleService,
		notificationService: notificationService,
	}
}

func (q *articleController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
}

func (q *articleController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.ArticleFilter{}
	rows := []*models.Article{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Data Artikel",
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
				Field:     models.ArticleSortingFieldSubject,
				Direction: models.SortingDirectionAsc,
				Label:     "Subject A-Z",
			},
			{
				Field:     models.ArticleSortingFieldSubject,
				Direction: models.SortingDirectionDesc,
				Label:     "Subject Z-A",
			},
		},
	)
	request, err = sanitizers.FindArticles(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticles")
		return c.Render("seller/article_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.articleService.FindArticles(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticles")
		return c.Render("seller/article_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/article_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *articleController) SellerNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Article{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Add New Artikel",
		"request", request,
		"article", request,
	)
	return c.Render("seller/article_new", params, "seller/layouts/authenticated")
}

func (q *articleController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Article{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Add New Artikel",
		"request", request,
		"article", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateArticle(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateArticle")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentSeller.ID
	request.CompanyID = currentSeller.CompanyID
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
		return c.Render("seller/article_new", params.Push(
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
	if _, err = q.articleService.CreateArticle(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateArticle")
		return c.Render("seller/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/article_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("article created successfully").
		Redirect(c, sess.Session, "/seller/articles")
}

func (q *articleController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	articleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	article, err := q.articleService.FindArticle(
		ctx,
		models.NewArticleFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithArticleIDs(articleID),
	)
	if err != nil && !errors.Is(err, models.ErrArticleNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticle")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	if article == nil {
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Data Artikel|%s", article.Subject),
		"request", article,
		"article", article,
	)
	return c.Render("seller/article_edit", params, "seller/layouts/authenticated")
}

func (q *articleController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	articleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	article, err := q.articleService.FindArticle(
		ctx,
		models.NewArticleFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithArticleIDs(articleID),
	)
	if err != nil && !errors.Is(err, models.ErrArticleNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticle")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	if article == nil {
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Data Artikel|%s", article.Subject),
		"request", article,
		"article", article,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateArticle(ctx, c)
	if photo == nil {
		request.Photo = article.Photo
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateArticle")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	request.ID = article.ID
	request.CompanyID = article.CompanyID
	request.RecordStatus = article.RecordStatus
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/article_edit", params.Push(
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
	if _, err = q.articleService.UpdateArticle(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateArticle")
		return c.Render("seller/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/article_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("article updated successfully").
		Redirect(c, sess.Session, "/seller/articles")
}

func (q *articleController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	articleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	article, err := q.articleService.FindArticle(
		ctx,
		models.NewArticleFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithArticleIDs(articleID),
	)
	if err != nil && !errors.Is(err, models.ErrArticleNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticle")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	if article == nil {
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	article.RecordStatus = models.RecordStatusDeleted
	article.DeletedBy = &currentSeller.ID
	article.DeletedAt = new(time.Now())
	article.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.articleService.UpdateArticle(ctx, tx, article); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateArticle")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/articles")
	}
	return flash.Clear(c, sess.Session).
		Success("article deleted successfully").
		Redirect(c, sess.Session, "/seller/articles")
}
