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

type blogController struct {
	sellerService       services.SellerService
	blogService         services.BlogService
	notificationService services.NotificationService
}

func NewBlogController(
	sellerService services.SellerService,
	blogService services.BlogService,
	notificationService services.NotificationService,
) *blogController {
	return &blogController{
		sellerService:       sellerService,
		blogService:         blogService,
		notificationService: notificationService,
	}
}

func (q *blogController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
}

func (q *blogController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
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
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("seller/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.blogService.FindBlogs(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlogs")
		return c.Render("seller/blog_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/blog_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *blogController) SellerNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
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
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Add New Blog",
		"request", request,
		"blog", request,
	)
	return c.Render("seller/blog_new", params, "seller/layouts/authenticated")
}

func (q *blogController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Blog{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Add New Blog",
		"request", request,
		"blog", request,
	)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBlog(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBlog")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithSlugs(request.Slug),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return c.Render("seller/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if blog != nil {
		return c.Render("seller/blog_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueBlogSlugViolation.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentSeller.ID
	request.CompanyID = currentSeller.CompanyID
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
		return c.Render("seller/blog_new", params.Push(
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
	if _, err = q.blogService.CreateBlog(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlog")
		return c.Render("seller/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/blog_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/blog_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("blog created successfully").
		Redirect(c, sess.Session, "/seller/blogs")
}

func (q *blogController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	if blog == nil {
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Data Blog|%s", blog.Subject),
		"request", blog,
		"blog", blog,
	)
	return c.Render("seller/blog_edit", params, "seller/layouts/authenticated")
}

func (q *blogController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	if blog == nil {
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("seller/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if blog.Slug != request.Slug {
		otherBlog, err := q.blogService.FindBlog(
			ctx,
			models.NewBlogFilter().
				WithCompanyIDs(request.CompanyID).
				WithSlugs(request.Slug).
				WithExcludedBlogIDs(blog.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
			return c.Render("seller/blog_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherBlog != nil {
			return c.Render("seller/blog_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueBlogSlugViolation.Error()),
			), "seller/layouts/authenticated")
		}
	}
	request.ID = blog.ID
	request.CompanyID = blog.CompanyID
	request.RecordStatus = blog.RecordStatus
	request.Slug = blog.Slug
	request.DeletedSlug = nil
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/blog_edit", params.Push(
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
	if _, err = q.blogService.UpdateBlog(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlog")
		return c.Render("seller/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/blog_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/blog_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("blog updated successfully").
		Redirect(c, sess.Session, "/seller/blogs")
}

func (q *blogController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BlogController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	blogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	blog, err := q.blogService.FindBlog(
		ctx,
		models.NewBlogFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithBlogIDs(blogID),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlog")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	if blog == nil {
		return flash.Danger(models.ErrBlogNotFound.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	blog.RecordStatus = models.RecordStatusDeleted
	if blog.Slug != nil {
		blog.DeletedSlug = new(*blog.Slug)
	}
	blog.Slug = nil
	blog.DeletedBy = &currentSeller.ID
	blog.DeletedAt = new(time.Now())
	blog.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/blogs")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/blogs")
	}
	return flash.Clear(c, sess.Session).
		Success("blog deleted successfully").
		Redirect(c, sess.Session, "/seller/blogs")
}
