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
	blogController struct {
		adminService services.AdminService
		blogService  services.BlogService
	}
)

func NewBlogController(
	userService services.AdminService,
	blogService services.BlogService,
) *blogController {
	return new(blogController{
		adminService: userService,
		blogService:  blogService,
	})
}

func (q *blogController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *blogController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.BlogFilter{}
	rows := []*models.Blog{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Blog Laukpauk",
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
				Field:     models.BlogSortingFieldSubject,
				Direction: models.SortingDirectionAsc,
				Label:     "Subject A-Z",
			},
			{
				Field:     models.BlogSortingFieldSubject,
				Direction: models.SortingDirectionDesc,
				Label:     "Subject Z-A",
			},
		},
	)
	request, err = sanitizers.FindBlogs(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("admin/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.blogService.FindBlogs(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("admin/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/blog_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *blogController) AdminNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Blog{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Blog",
		"request", request,
		"blog", request,
	)
	return c.Render("admin/blog_new", params, "admin/layouts/authenticated")
}

func (q *blogController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Blog)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Blog",
		"request", request,
		"blog", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBlog(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlog")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithSlugs(request.Slug),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if blog != nil {
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueBlogSlugViolation.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.DeletedSlug = nil
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("BLOG_DEFAULT_PHOTO")
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/blog_new", params.Push(
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
	if _, err = q.blogService.CreateBlog(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlog")
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/blog_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("blog created successfully").
		Redirect(c, sess.Session, "/admin/blogs")
}

func (q *blogController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if blog == nil {
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Blog|%s", blog.Subject),
		"request", blog,
		"blog", blog,
	)
	return c.Render("admin/blog_edit", params, "admin/layouts/authenticated")
}

func (q *blogController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if blog == nil {
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Blog|%s", blog.Subject),
		"request", blog,
		"blog", blog,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBlog(ctx, c)
	if photo == nil {
		request.Photo = blog.Photo
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlog")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if blog.Slug != request.Slug {
		otherBlog, err := q.blogService.FindBlog(
			ctx,
			models.NewBlogFilter().
				WithSlugs(request.Slug).
				WithExcludedBlogIDs(blog.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
			return c.Render("admin/blog_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherBlog != nil {
			return c.Render("admin/blog_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueBlogSlugViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = blog.ID
	request.RecordStatus = blog.RecordStatus
	request.Slug = blog.Slug
	request.DeletedSlug = nil
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/blog_edit", params.Push(
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
	if _, err = q.blogService.UpdateBlog(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlog")
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/blog_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("blog updated successfully").
		Redirect(c, sess.Session, "/admin/blogs")
}

func (q *blogController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if blog == nil {
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	blog.RecordStatus = models.RecordStatusDeleted
	blog.DeletedSlug = blog.Slug
	blog.Slug = nil
	blog.DeletedBy = &currentAdmin.ID
	blog.DeletedAt = new(time.Now())
	blog.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.blogService.UpdateBlog(ctx, tx, blog); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	return flash.Clear(c, sess.Session).
		Success("blog deleted successfully").
		Redirect(c, sess.Session, "/admin/blogs")
}
