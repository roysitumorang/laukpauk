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

type productController struct {
	adminService           services.AdminService
	productCategoryService services.ProductCategoryService
	productService         services.ProductService
	productVariantService  services.ProductVariantService
}

func NewProductController(
	adminService services.AdminService,
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
) *productController {
	return &productController{
		adminService:           adminService,
		productCategoryService: productCategoryService,
		productService:         productService,
		productVariantService:  productVariantService,
	}
}

func (q *productController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id", q.AdminShow).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete).
		Get("/:id/variants/new", q.AdminNewVariant).
		Post("/:id/variants", q.AdminCreateVariant).
		Get("/:id/variants/:product_variant_id/edit", q.AdminEditVariant).
		Post("/:id/variants/:product_variant_id", q.AdminUpdateVariant).
		Post("/:id/variants/:product_variant_id/delete", q.AdminDeleteVariant)
	r.Group("/v1/buyer").
		Get("", q.V1BuyerIndex).
		Get("/:id", q.V1BuyerShow)
}

func (q *productController) AdminIndex(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminIndex"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := new(models.ProductFilter)
	rows := []*models.Product{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Produk",
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
	if request, err = sanitizers.FindProducts(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return c.Render("admin/product_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.productService.FindProducts(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return c.Render("admin/product_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/product_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *productController) AdminNew(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminNew"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _, err := q.prepareProduct(ctx, c, flash, currentAdmin, csrfToken, true, true)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/product_new", params, "admin/layouts/authenticated")
}

func (q *productController) AdminCreate(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminCreate"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request, params, mapProductCategoryNames, err := q.prepareProduct(ctx, c, flash, currentAdmin, csrfToken, true, true)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if photo == nil {
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
	}
	params.Push(
		"request", request,
		"product", request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		c.Response().SetStatusCode(statusCode)
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "admin/layouts/authenticated")
	}
	productCategoryName, ok := mapProductCategoryNames[request.ProductCategoryID]
	if !ok {
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger("product_category_id: not found"),
			"request", request,
			"product", request,
		), "admin/layouts/authenticated")
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithNames(request.Name),
	)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.ProductCategoryName = productCategoryName
	request.Hits = 0
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = now
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = now
	if product != nil {
		if product.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return c.Render("admin/product_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductNameViolation.Error()),
				"request", request,
				"product", request,
			), "admin/layouts/authenticated")
		}
		request.ID = product.ID
		request.Name = product.DeletedName
		request.DeletedName = nil
		request.CreatedBy = product.CreatedBy
		request.CreatedAt = product.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
		if photo == nil {
			request.SetPhoto(product.Photo)
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
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
	if _, err = q.productService.SaveProduct(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProduct")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "admin/layouts/authenticated")
	}
	if product != nil &&
		product.Photo != os.Getenv("PRODUCT_DEFAULT_PHOTO") &&
		product.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(product.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(product.Photo)
			return c.Render("admin/product_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"product", request,
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("product created successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) AdminShow(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminShow"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, err := q.prepareProductByID(ctx, c)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductByID")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if product == nil {
		return flash.Danger(models.ErrProductNotFound.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	var productVariants []*models.ProductVariant
	var productName string
	if product.Name != nil {
		productName = *product.Name
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Detail Produk|%s", productName),
		"product", product,
		"productVariants", productVariants,
		"discountTypePercentage", models.DiscountTypePercentage,
	)
	if productVariants, _, err = q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(product.ID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return c.Render("admin/product_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("productVariants", productVariants)
	return c.Render("admin/product_show", params, "admin/layouts/authenticated")
}

func (q *productController) AdminEdit(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminEdit"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _, err := q.prepareProduct(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	return c.Render("admin/product_edit", params, "admin/layouts/authenticated")
}

func (q *productController) AdminUpdate(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminUpdate"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, params, mapProductCategoryNames, err := q.prepareProduct(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if photo == nil {
		request.SetPhoto(product.Photo)
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "admin/layouts/authenticated")
	}
	params.Push("request", request)
	productCategoryName, ok := mapProductCategoryNames[request.ProductCategoryID]
	if !ok {
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger("product_category_id: not found"),
			"request", request.SetPhoto(product.Photo),
		), "admin/layouts/authenticated")
	}
	if product.Name != request.Name {
		otherProduct, err := q.productService.FindProduct(
			ctx,
			models.NewProductFilter().
				WithNames(request.Name).
				WithExcludedProductIDs(product.ID),
		)
		if err != nil && !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
			return c.Render("admin/product_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(product.Photo),
			), "admin/layouts/authenticated")
		}
		if otherProduct != nil {
			return c.Render("admin/product_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductNameViolation.Error()),
				"request", request.SetPhoto(product.Photo),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = product.ID
	request.RecordStatus = product.RecordStatus
	request.ProductCategoryName = productCategoryName
	request.Hits = product.Hits
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
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
	if _, err = q.productService.UpdateProduct(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "admin/layouts/authenticated")
	}
	if product.Photo != os.Getenv("PRODUCT_DEFAULT_PHOTO") &&
		product.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(product.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			return c.Render("admin/product_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(product.Photo),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "admin/layouts/authenticated")
	}
	return flash.Success("product updated successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) AdminDelete(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminDelete"
	ctx, sess, flash, currentAdmin, _ := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request, err := q.prepareProductByID(ctx, c)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductByID")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if request == nil {
		return flash.Danger(models.ErrProductNotFound.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request.RecordStatus = models.RecordStatusDeleted
	if request.Name != nil {
		request.DeletedName = new(*request.Name)
	}
	request.Name = nil
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.productService.UpdateProduct(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	return flash.Success("product deleted successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) AdminNewVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminNewVariant"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	return c.Render("admin/product_variant_new", params, "admin/layouts/authenticated")
}

func (q *productController) AdminCreateVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminCreateVariant"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, _, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_variant_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithProductIDs(product.ID).
			WithStockUnits(request.StockUnit),
	)
	if err != nil && !errors.Is(err, models.ErrProductVariantNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.ProductCategoryID = product.ProductCategoryID
	request.ProductCategoryName = product.ProductCategoryName
	request.ProductID = product.ID
	if product.Name != nil {
		request.ProductName = *product.Name
	}
	request.ProductReadyStock = product.ReadyStock
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = now
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = now
	if productVariant != nil {
		if productVariant.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/product_variant_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductVariantStockUnitViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = productVariant.ID
		request.DeletedStockUnit = nil
		request.CreatedBy = productVariant.ID
		request.CreatedAt = productVariant.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_variant_new", params.Push(
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
	if _, err = q.productVariantService.SaveProductVariant(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProductVariant")
		return c.Render("admin/product_variant_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_variant_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("product variant created successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
}

func (q *productController) AdminEditVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminEditVariant"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	return c.Render("admin/product_variant_edit", params, "admin/layouts/authenticated")
}

func (q *productController) AdminUpdateVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminUpdateVariant"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, productVariant, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if productVariant.StockUnit != request.StockUnit {
		otherProductVariant, err := q.productVariantService.FindProductVariant(
			ctx,
			models.NewProductVariantFilter().
				WithProductIDs(product.ID).
				WithStockUnits(request.StockUnit).
				WithExcludedProductVariantIDs(productVariant.ID),
		)
		if err != nil && !errors.Is(err, models.ErrProductVariantNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
			return c.Render("admin/product_variant_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherProductVariant != nil {
			return c.Render("admin/product_variant_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductVariantStockUnitViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = productVariant.ID
	request.RecordStatus = productVariant.RecordStatus
	request.ProductCategoryID = product.ProductCategoryID
	request.ProductCategoryName = product.ProductCategoryName
	if product.Name != nil {
		request.ProductName = *product.Name
	}
	request.ProductReadyStock = productVariant.ProductReadyStock
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_variant_edit", params.Push(
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
	if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
		return c.Render("admin/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("product variant updated successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
}

func (q *productController) AdminDeleteVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-AdminDeleteVariant"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, productVariant, _, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if productVariant.StockUnit != nil {
		productVariant.DeletedStockUnit = new(*productVariant.StockUnit)
	}
	productVariant.StockUnit = nil
	productVariant.RecordStatus = models.RecordStatusDeleted
	productVariant.DeletedBy = new(currentAdmin.ID)
	productVariant.DeletedAt = new(time.Now())
	productVariant.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
	}
	return flash.Success("product variant deleted successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
}

func (q *productController) prepare(c fiber.Ctx) (
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

func (q *productController) prepareProductByID(
	ctx context.Context,
	c fiber.Ctx,
) (
	*models.Product,
	error,
) {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, models.ErrProductNotFound
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		return nil, err
	}
	if product == nil {
		return nil, models.ErrProductNotFound
	}
	return product, nil
}

func (q *productController) prepareProduct(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentAdmin *models.Admin,
	csrfToken string,
	withParams,
	withProductCategories bool,
) (
	product *models.Product,
	params models.Map,
	mapProductCategoryNames map[uuid.UUID]string,
	err error,
) {
	product = &models.Product{}
	product.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
	if withParams {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentAdmin", currentAdmin,
			"pageTitle", "Produk Baru",
			"request", product,
			"product", product,
			"productTypes", models.ProductTypes,
		)
	}
	if c.Params("id") != "" {
		if product, err = q.prepareProductByID(ctx, c); err != nil {
			return nil, nil, nil, err
		}
		if withParams && params.Size() > 0 && product != nil && product.Name != nil {
			params.Push(
				"pageTitle", helper.Sprintf("Edit Data Produk|%s", *product.Name),
				"request", product,
				"product", product,
			)
		}
	}
	if withProductCategories {
		var productCategories []*models.ProductCategory
		if withParams && params.Size() > 0 {
			params.Push("productCategories", productCategories)
		}
		productCategories, _, err := q.productCategoryService.FindProductCategories(
			ctx,
			models.NewProductCategoryFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithSortings(models.Sorting{
					Field:     models.SortProductCategoriesByName,
					Direction: models.SortingDirectionAsc,
				}),
		)
		if err != nil {
			return nil, nil, nil, err
		}
		if withParams && params.Size() > 0 {
			params.Push("productCategories", productCategories)
		}
		mapProductCategoryNames = map[uuid.UUID]string{}
		for _, productCategory := range productCategories {
			if productCategory.Name != nil {
				mapProductCategoryNames[productCategory.ID] = *productCategory.Name
			}
		}
	}
	return product, params, mapProductCategoryNames, nil
}

func (q *productController) prepareProductVariant(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentAdmin *models.Admin,
	csrfToken string,
	withParams bool,
) (
	*models.Product,
	*models.ProductVariant,
	models.Map,
	error,
) {
	product, err := q.prepareProductByID(ctx, c)
	if err != nil {
		return nil, nil, nil, err
	}
	productVariant := &models.ProductVariant{}
	var params models.Map
	if withParams && product.Name != nil {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentAdmin", currentAdmin,
			"pageTitle", helper.Sprintf("Detail Produk|%s", *product.Name),
			"product", product,
			"productVariant", productVariant,
			"request", productVariant,
			"discountTypes", models.DiscountTypes,
		)
	}
	if c.Params("product_variant_id") != "" {
		productVariantID, err := uuid.Parse(c.Params("product_variant_id"))
		if err != nil {
			return nil, nil, nil, models.ErrProductVariantNotFound
		}
		if productVariant, err = q.productVariantService.FindProductVariant(
			ctx,
			models.NewProductVariantFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductIDs(product.ID).
				WithProductVariantIDs(productVariantID),
		); err != nil {
			return nil, nil, nil, err
		}
		if productVariant == nil {
			return nil, nil, nil, models.ErrProductVariantNotFound
		}
		if withParams && params.Size() > 0 {
			params.Push(
				"productVariant", productVariant,
				"request", productVariant,
			)
		}
	}
	return product, productVariant, params, nil
}

func (q *productController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-V1BuyerIndex"
	request, err := sanitizers.FindProducts(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.productService.FindProducts(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
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

func (q *productController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-V1BuyerShow"
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrProductNotFound.Error()).
			WriteResponse(c)
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.productService.IncreaseProductHits(
		ctx,
		tx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if errors.Is(err, models.ErrProductNotFound) || response == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrIncreaseProductHits")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	variants, _, err := q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(response.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response.Photo = response.PhotoPath(c.BaseURL())
	response.Variants = make([]*models.ProductVariant, len(variants))
	copy(response.Variants, variants)
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
