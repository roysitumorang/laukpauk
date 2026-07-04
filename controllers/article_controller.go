package controllers

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"
	"os"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	articleController struct {
		userService    services.UserService
		articleService services.ArticleService
	}
)

func NewArticleController(
	userService services.UserService,
	articleService services.ArticleService,
) *articleController {
	return new(articleController{
		userService:    userService,
		articleService: articleService,
	})
}

func (q *articleController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *articleController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-AdminIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.ArticleFilter)
	rows := []*models.Article{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
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
	request.WithRecordStatuses(models.RecordStatusActive)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticles")
		return c.Render("admin/article_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.articleService.FindArticles(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticles")
		return c.Render("admin/article_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/article_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *articleController) AdminNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Article)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Artikel",
		"request", request,
	)
	return c.Render("admin/article_new", params, "admin/layouts/authenticated")
}

func (q *articleController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-AdminCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Article)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Artikel",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateArticle(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateArticle")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/article_new", params.Push(
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
		return c.Render("admin/article_new", params.Push(
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
	if _, err = q.articleService.CreateArticle(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateArticle")
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("article created successfully").Redirect(c, sess.Session, "/admin/articles")
}

func (q *articleController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-AdminEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	articleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	article, err := q.articleService.FindArticle(
		ctx,
		models.NewArticleFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithArticleIDs(articleID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrArticleNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticle")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if article == nil {
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Artikel|%s", article.Subject),
		"request", article,
		"article", article,
	)
	return c.Render("admin/article_edit", params, "admin/layouts/authenticated")
}

func (q *articleController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-AdminUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	articleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	article, err := q.articleService.FindArticle(
		ctx,
		models.NewArticleFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithArticleIDs(articleID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrArticleNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticle")
		}
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if article == nil {
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	request := new(models.Article)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Artikel|%s", article.Subject),
		"request", request,
		"article", article,
	)
	request, photo, statusCode, err := sanitizers.ValidateArticle(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateArticle")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = article.ID
	request.RecordStatus = article.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = article.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/article_edit", params.Push(
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
	if _, err = q.articleService.UpdateArticle(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateArticle")
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("article updated successfully").Redirect(c, sess.Session, "/admin/articles")
}

func (q *articleController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-AdminDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	articleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("article not found").Redirect(c, sess.Session, "/admin/articles")
	}
	article, err := q.articleService.FindArticle(
		ctx,
		models.NewArticleFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithArticleIDs(articleID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrArticleNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindArticle")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if article == nil {
		return flash.Danger(models.ErrArticleNotFound.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	request := article
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	return flash.Success("article deleted successfully").Redirect(c, sess.Session, "/admin/articles")
}

func (q *articleController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	if err := c.SaveFile(file, helper.Sprintf(models.ArticlePhotoPath, filename)); err != nil {
		return err
	}
	return c.SaveFile(file, helper.Sprintf(models.ArticlePhotoCompressPath, filename))
}
