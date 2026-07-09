package controllers

import (
	"errors"
	"log/slog"
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
	return new(productCategoryController{
		adminService:           adminService,
		productCategoryService: productCategoryService,
	})
}

func (q *productCategoryController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *productCategoryController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductCategoryController-AdminIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.ProductCategoryFilter)
	rows := []*models.ProductCategory{}
	pagination := new(models.Pagination)
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
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.ProductCategory)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Kategori Produk Baru",
		"request", request,
		"productCategory", request,
	)
	return c.Render("admin/product_category_new", params, "admin/layouts/authenticated")
}

func (q *productCategoryController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductCategoryController-AdminCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.ProductCategory)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Kategori Produk Baru",
		"request", request,
		"productCategory", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
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
	params.Push(
		"request", request,
		"productCategory", request,
	)
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
			"request", request,
		), "admin/layouts/authenticated")
	}
	if request.ID.Compare(uuid.Nil()) == 0 {
		if _, err = q.productCategoryService.CreateProductCategory(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateProductCategory")
		}
	} else {
		if _, err = q.productCategoryService.UpdateProductCategory(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		}
	}
	if err != nil {
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("product category created successfully").Redirect(c, sess.Session, "/admin/product_categories")
}

func (q *productCategoryController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductCategoryController-AdminEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productCategoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product category not found").Redirect(c, sess.Session, "/admin/product_categories")
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(productCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	request := productCategory
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Kategori Produk|%s", productCategory.Name),
		"request", request,
		"productCategory", productCategory,
	)
	return c.Render("admin/product_category_edit", params, "admin/layouts/authenticated")
}

func (q *productCategoryController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductCategoryController-AdminUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productCategoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product category not found").Redirect(c, sess.Session, "/admin/product_categories")
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(productCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/product_categories")
	}
	request := new(models.ProductCategory)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Artikel|%s", productCategory.Name),
		"request", request,
		"productCategory", productCategory,
	)
	request, photo, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductCategory")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
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
	if _, err = q.productCategoryService.UpdateProductCategory(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return c.Render("admin/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
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
	ctx := c.Context()
	ctxt := "ProductCategoryController-AdminDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	productCategoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product category not found").Redirect(c, sess.Session, "/admin/product_categories")
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(productCategoryID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		}
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

func (q *productCategoryController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	if err := c.SaveFile(file, helper.Sprintf(models.ProductCategoryPhotoPath, filename)); err != nil {
		return err
	}
	return c.SaveFile(file, helper.Sprintf(models.ProductCategoryPhotoCompressPath, filename))
}
