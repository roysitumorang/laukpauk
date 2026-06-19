package controllers

import (
	"fmt"
	"mime/multipart"
	"os"
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
	"go.uber.org/zap"
)

type (
	blogController struct {
		userService services.UserService
		blogService services.BlogService
	}
)

func NewBlogController(
	userService services.UserService,
	blogService services.BlogService,
) *blogController {
	return new(blogController{
		userService: userService,
		blogService: blogService,
	})
}

func (q *blogController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
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
		"pageTitle", "Blog Laukpauk",
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
	)
	if request, err = sanitizers.FindBlogs(ctx, c); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("admin/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive)
	if rows, pagination, err = q.blogService.FindBlogs(
		c,
		request,
	); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("admin/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/blog_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *blogController) AdminNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
	)
	return c.Render("admin/blog_new", params, "admin/layouts/authenticated")
}

func (q *blogController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
	)
	request, photo, statusCode, err := sanitizers.ValidateBlog(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateBlog")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
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
		request.Photo = os.Getenv("BLOG_DEFAULT_PHOTO")
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.blogService.CreateBlog(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateBlog")
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("blog created successfully").Redirect(c, sess.Session, "/admin/blogs")
}

func (q *blogController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/admin/blogs")
	}
	blogs, _, err := q.blogService.FindBlogs(
		ctx,
		models.NewBlogFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBlogs")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if len(blogs) == 0 {
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/admin/blogs")
	}
	request := blogs[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", fmt.Sprintf("Edit Data Blog|%s", request.Subject),
		"request", request,
		"blogID", request.ID,
	)
	return c.Render("admin/blog_edit", params, "admin/layouts/authenticated")
}

func (q *blogController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/admin/blogs")
	}
	blogs, _, err := q.blogService.FindBlogs(
		ctx,
		models.NewBlogFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBlogs")
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/admin/blogs")
	}
	if len(blogs) == 0 {
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/admin/blogs")
	}
	blog := blogs[0]
	request := new(models.Blog)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", fmt.Sprintf("Edit Data Blog|%s", blog.Subject),
		"request", request,
		"blogID", blog.ID,
	)
	request, photo, statusCode, err := sanitizers.ValidateBlog(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateBlog")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = blog.ID
	request.RecordStatus = blog.RecordStatus
	request.Slug = blog.Slug
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = blog.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.blogService.UpdateBlog(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateBlog")
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("blog updated successfully").Redirect(c, sess.Session, "/admin/blogs")
}

func (q *blogController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-AdminDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/admin/blogs")
	}
	blogs, _, err := q.blogService.FindBlogs(
		ctx,
		models.NewBlogFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBlogs")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if len(blogs) == 0 {
		return flash.Danger("blog not found").Redirect(c, sess.Session, "/admin/blogs")
	}
	request := blogs[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if _, err = q.blogService.UpdateBlog(ctx, tx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/blogs")
	}
	return flash.Success("blog deactivated successfully").Redirect(c, sess.Session, "/admin/blogs")
}

func (q *blogController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	if err := c.SaveFile(file, fmt.Sprintf(models.BlogPhotoPath, filename)); err != nil {
		return err
	}
	return c.SaveFile(file, fmt.Sprintf(models.BlogPhotoCompressPath, filename))
}
