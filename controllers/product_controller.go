package controllers

import (
	"fmt"
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
	productController struct {
		userService            services.UserService
		productCategoryService services.ProductCategoryService
		productService         services.ProductService
	}
)

func NewProductController(
	userService services.UserService,
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
) *productController {
	return new(productController{
		userService:            userService,
		productCategoryService: productCategoryService,
		productService:         productService,
	})
}

func (q *productController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *productController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.ProductFilter)
	rows := []*models.Product{}
	pagination := new(models.Pagination)
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
	)
	if request, err = sanitizers.FindProducts(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return c.Render("admin/product_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive)
	if rows, pagination, err = q.productService.FindProducts(
		c,
		request,
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
	ctx := c.Context()
	ctxt := "ProductController-AdminNew"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Product)
	var productCategories []*models.ProductCategory
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Produk Baru",
		"request", request,
		"product", request,
		"productCategories", productCategories,
		"productTypes", models.ProductTypes,
		"stockStatuses", models.StockStatuses,
	)
	if productCategories, _, err = q.productCategoryService.FindProductCategories(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatus(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/product_new", params.Push(
		"productCategories", productCategories,
	), "admin/layouts/authenticated")
}

func (q *productController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Product)
	var productCategories []*models.ProductCategory
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Produk Baru",
		"request", request,
		"product", request,
		"productCategories", productCategories,
		"productTypes", models.ProductTypes,
		"stockStatuses", models.StockStatuses,
	)
	if productCategories, _, err = q.productCategoryService.FindProductCategories(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatus(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push(
		"productCategories", productCategories,
	)
	request, photo, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		c.Response().SetStatusCode(statusCode)
		request.SetPhoto("")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "admin/layouts/authenticated")
	}
	params.Push(
		"request", request,
		"product", request,
	)
	mapProductCategoryIDs := map[uuid.UUID]struct{}{}
	for _, productCategory := range productCategories {
		mapProductCategoryIDs[productCategory.ID] = struct{}{}
	}
	if _, ok := mapProductCategoryIDs[request.ProductCategoryID]; !ok {
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger("product_category_id: is required"),
		), "admin/layouts/authenticated")
	}
	products, _, err := q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithProductCategoryIDs(request.ProductCategoryID).
			WithProductTypes(request.ProductType).
			WithNames(request.Name),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return c.Render("admin/product_new", params.Push(
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
		request.Photo = os.Getenv("PRODUCT_DEFAULT_PHOTO")
	}
	if len(products) > 0 {
		product := products[0]
		if product.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/product_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductCategoryProductTypeNameViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = product.ID
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
		if photo == nil {
			request.Photo = product.Photo
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if request.ID.Compare(uuid.Nil()) == 0 {
		if _, err = q.productService.CreateProduct(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateProduct")
		}
	} else {
		if _, err = q.productService.UpdateProduct(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		}
	}
	if err != nil {
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("product created successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	products, _, err := q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if len(products) == 0 {
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	product := products[0]
	request := product
	var productCategories []*models.ProductCategory
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", fmt.Sprintf("Edit Data Produk|%s", product.Name),
		"request", request,
		"product", request,
		"productCategories", productCategories,
		"productTypes", models.ProductTypes,
		"stockStatuses", models.StockStatuses,
	)
	if productCategories, _, err = q.productCategoryService.FindProductCategories(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatus(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push(
		"productCategories", productCategories,
	)
	return c.Render("admin/product_edit", params, "admin/layouts/authenticated")
}

func (q *productController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	products, _, err := q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	if len(products) == 0 {
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	product := products[0]
	request := new(models.Product)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", fmt.Sprintf("Edit Data Artikel|%s", product.Name),
		"request", request,
		"product", product,
	)
	request, photo, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = product.ID
	request.RecordStatus = product.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = product.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.productService.UpdateProduct(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("product updated successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	products, _, err := q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if len(products) == 0 {
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	request := products[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if _, err = q.productService.UpdateProduct(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	return flash.Success("product deactivated successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	if err := c.SaveFile(file, fmt.Sprintf(models.ProductPhotoPath, filename)); err != nil {
		return err
	}
	return c.SaveFile(file, fmt.Sprintf(models.ProductPhotoCompressPath, filename))
}
