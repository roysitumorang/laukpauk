package controllers

import (
	"fmt"
	"mime/multipart"
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
		Post("/:id/deactivate", q.AdminDeactivate)
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
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
	)
	if request, err = sanitizers.FindArticles(ctx, c); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindArticles")
		return c.Render("admin/article_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.articleService.FindArticles(
		c,
		request,
	); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindArticles")
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
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
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
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateArticle(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateArticle")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.articleService.CreateArticle(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateArticle")
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/article_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
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
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("article not found").Redirect(c, sess.Session, "/admin/articles")
	}
	articles, _, err := q.articleService.FindArticles(ctx, models.NewArticleFilter().WithArticleIDs(articleID))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindArticles")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if len(articles) == 0 || articles[0].DeactivatedBy != nil {
		return flash.Danger("article not found").Redirect(c, sess.Session, "/admin/articles")
	}
	request := articles[0]
	params := models.NewMap().Push(
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"request", request,
		"articleID", request.ID,
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
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("article not found").Redirect(c, sess.Session, "/admin/articles")
	}
	articles, _, err := q.articleService.FindArticles(ctx, models.NewArticleFilter().WithArticleIDs(articleID))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindArticles")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if len(articles) == 0 || articles[0].DeactivatedBy != nil {
		return flash.Danger("article not found").Redirect(c, sess.Session, "/admin/articles")
	}
	article := articles[0]
	request := new(models.Article)
	params := models.NewMap().Push(
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"request", request,
		"articleID", article.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateArticle(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateArticle")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = article.ID
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeactivatedBy = nil
	request.DeactivatedAt = nil
	request.DeactivationReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.articleService.UpdateArticle(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateArticle")
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/article_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("article updated successfully").Redirect(c, sess.Session, "/admin/articles")
}

func (q *articleController) AdminDeactivate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ArticleController-AdminDeactivate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	articleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("article not found").Redirect(c, sess.Session, "/admin/articles")
	}
	articles, _, err := q.articleService.FindArticles(ctx, models.NewArticleFilter().WithArticleIDs(articleID))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindArticles")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if len(articles) == 0 || articles[0].DeactivatedBy != nil {
		return flash.Danger("article not found").Redirect(c, sess.Session, "/admin/articles")
	}
	request := articles[0]
	request.DeactivatedBy = &currentAdmin.ID
	request.DeactivatedAt = new(time.Now())
	request.DeactivationReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if _, err = q.articleService.UpdateArticle(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateArticle")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/articles")
	}
	return flash.Success("article deactivated successfully").Redirect(c, sess.Session, "/admin/articles")
}

func (q *articleController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	if err := c.SaveFile(file, fmt.Sprintf(models.ArticlePhotoPath, filename)); err != nil {
		return err
	}
	return c.SaveFile(file, fmt.Sprintf(models.ArticlePhotoCompressPath, filename))
}
