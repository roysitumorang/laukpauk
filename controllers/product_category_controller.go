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
	productCategoryController struct {
		adminService           services.AdminService
		productCategoryService services.ProductCategoryService
	}
)

func NewProductCategoryController(
	adminService services.AdminService,
	productCategoryService services.ProductCategoryService,
) *productCategoryController {
	return &productCategoryController{
		adminService:           adminService,
		productCategoryService: productCategoryService,
	}
}

func (q *productCategoryController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
	r.Group("/v1/buyer").
		Get("", q.V1BuyerIndex)
}

func (q *productCategoryController) AdminIndex(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-AdminIndex"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := &models.ProductCategoryFilter{}
	rows := []*models.ProductCategory{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Kategori Produk",
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
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	if request, err = sanitizers.FindProductCategories(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_category_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.productCategoryService.FindProductCategories(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_category_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/product_category_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *productCategoryController) AdminNew(c fiber.Ctx) (err error) {
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _ := q.prepareProductCategory(ctx, c, flash, currentAdmin, csrfToken, true)
	return c.Render("admin/product_category_new", params, "admin/layouts/authenticated")
}

func (q *productCategoryController) AdminCreate(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-AdminCreate"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request, params, _ := q.prepareProductCategory(ctx, c, flash, currentAdmin, csrfToken, true)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
	params.Push(
		"request", request,
		"productCategory", request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductCategory")
		c.Response().SetStatusCode(statusCode)
		request.SetPhoto("")
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"productCategory", request,
		), "admin/layouts/authenticated")
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithNames(request.Name),
	)
	if err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = now
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO")
	}
	if productCategory != nil {
		if productCategory.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/product_category_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductCategoryNameViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = productCategory.ID
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
		if photo == nil {
			request.Photo = productCategory.Photo
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_category_new", params.Push(
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
	if _, err = q.productCategoryService.SaveProductCategory(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProductCategory")
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("product category created successfully").Redirect(c, sess.Session, "/admin/product_categories")
}

func (q *productCategoryController) AdminEdit(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-AdminEdit"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareProductCategory(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	return c.Render("admin/product_category_edit", params, "admin/layouts/authenticated")
}

func (q *productCategoryController) AdminUpdate(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-AdminUpdate"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	productCategory, params, err := q.prepareProductCategory(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductCategory")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if productCategory.Name != request.Name {
		otherProductCategory, err := q.productCategoryService.FindProductCategory(
			ctx,
			models.NewProductCategoryFilter().
				WithNames(request.Name).
				WithExcludedProductCategoryIDs(productCategory.ID),
		)
		if err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
			return c.Render("admin/product_category_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherProductCategory != nil {
			return c.Render("admin/product_category_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductCategoryNameViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = productCategory.ID
	request.RecordStatus = productCategory.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = productCategory.Photo
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_category_edit", params.Push(
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
	if _, err = q.productCategoryService.UpdateProductCategory(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return c.Render("admin/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if request.Photo != productCategory.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(productCategory.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			return c.Render("admin/product_category_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("product category updated successfully").Redirect(c, sess.Session, "/admin/product_categories")
}

func (q *productCategoryController) AdminDelete(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-AdminDelete"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	productCategory, _, err := q.prepareProductCategory(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	request := productCategory
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.productCategoryService.UpdateProductCategory(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	return flash.Success("product category deleted successfully").Redirect(c, sess.Session, "/admin/product_categories")
}

func (q *productCategoryController) prepare(c fiber.Ctx) (
	context.Context,
	*session.Middleware,
	*helper.FlashMessage,
	*models.Admin,
	string,
) {
	ctx := c.Context()
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	return ctx, sess, flash, currentAdmin, csrfToken
}

func (q *productCategoryController) prepareProductCategory(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentAdmin *models.Admin,
	csrfToken string,
	withParams bool,
) (
	*models.ProductCategory,
	models.Map,
	error,
) {
	productCategory := &models.ProductCategory{}
	var params models.Map
	if withParams {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentAdmin", currentAdmin,
			"pageTitle", "Kategori Produk Baru",
			"request", productCategory,
			"productCategory", productCategory,
		)
	}
	if c.Params("id") != "" {
		productCategoryID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return nil, nil, models.ErrProductCategoryNotFound
		}
		if productCategory, err = q.productCategoryService.FindProductCategory(
			ctx,
			models.NewProductCategoryFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductCategoryIDs(productCategoryID),
		); err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
			return nil, nil, err
		}
		if productCategory == nil {
			return nil, nil, models.ErrProductCategoryNotFound
		}
		if withParams && params.Size() > 0 {
			params.Push(
				"pageTitle", helper.Sprintf("Edit Data Kategori Produk|%s", productCategory.Name),
				"request", productCategory,
				"productCategory", productCategory,
			)
		}
	}
	return productCategory, params, nil
}

func (q *productCategoryController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductCategoryController-V1BuyerIndex"
	request, err := sanitizers.FindProductCategories(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.productCategoryService.FindProductCategories(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	pagination.Rows = make([]any, len(rows))
	for i, row := range rows {
		row.Photo = row.PhotoPath(c.BaseURL())
		pagination.Rows[i] = row
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}
