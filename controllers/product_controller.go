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
	productController struct {
		userService            services.UserService
		productCategoryService services.ProductCategoryService
		productService         services.ProductService
		productVariantService  services.ProductVariantService
	}
)

func NewProductController(
	userService services.UserService,
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
) *productController {
	return new(productController{
		userService:            userService,
		productCategoryService: productCategoryService,
		productService:         productService,
		productVariantService:  productVariantService,
	})
}

func (q *productController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id", q.AdminShow).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete).
		Get("/:product_id/variants/new", q.AdminNewVariant).
		Post("/:product_id/variants", q.AdminCreateVariant).
		Get("/:product_id/variants/:product_variant_id/edit", q.AdminEditVariant).
		Post("/:product_id/variants/:product_variant_id", q.AdminUpdateVariant).
		Post("/:product_id/variants/:product_variant_id/delete", q.AdminDeleteVariant)
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
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SortProductCategoriesByName,
				Direction: models.SortingDirectionAsc,
			}),
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
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SortProductCategoriesByName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("productCategories", productCategories)
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
			"flash", flash.Danger("product_category_id: not found"),
		), "admin/layouts/authenticated")
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithNames(request.Name),
	)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
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
	if product != nil {
		if product.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/product_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductNameViolation.Error()),
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
	params.Push(
		"request", request,
		"product", request,
	)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
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
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("product created successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) AdminShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminShow"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	var productVariants []*models.ProductVariant
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Detail Produk|%s", product.Name),
		"product", product,
		"productVariants", productVariants,
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
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request := product
	var productCategories []*models.ProductCategory
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Produk|%s", product.Name),
		"request", request,
		"product", product,
		"productCategories", productCategories,
		"productTypes", models.ProductTypes,
		"stockStatuses", models.StockStatuses,
	)
	if productCategories, _, err = q.productCategoryService.FindProductCategories(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SortProductCategoriesByName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("productCategories", productCategories)
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
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request := new(models.Product)
	var productCategories []*models.ProductCategory
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Produk|%s", product.Name),
		"request", request,
		"product", product,
		"productCategories", productCategories,
		"productTypes", models.ProductTypes,
		"stockStatuses", models.StockStatuses,
	)
	if productCategories, _, err = q.productCategoryService.FindProductCategories(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SortProductCategoriesByName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("productCategories", productCategories)
	request, photo, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if photo == nil {
		request.Photo = product.Photo
	}
	params.Push("request", request)
	mapProductCategoryIDs := map[uuid.UUID]struct{}{}
	for _, productCategory := range productCategories {
		mapProductCategoryIDs[productCategory.ID] = struct{}{}
	}
	if _, ok := mapProductCategoryIDs[request.ProductCategoryID]; !ok {
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger("product_category_id: not found"),
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
			), "admin/layouts/authenticated")
		}
		if otherProduct != nil {
			return c.Render("admin/product_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductNameViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = product.ID
	request.RecordStatus = product.RecordStatus
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
		), "admin/layouts/authenticated")
	}
	if _, err = q.productService.UpdateProduct(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
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
	request, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
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
	return flash.Success("product deleted successfully").Redirect(c, sess.Session, "/admin/products")
}

func (q *productController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	if err := c.SaveFile(file, helper.Sprintf(models.ProductPhotoPath, filename)); err != nil {
		return err
	}
	return c.SaveFile(file, helper.Sprintf(models.ProductPhotoCompressPath, filename))
}

func (q *productController) AdminNewVariant(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminNewVariant"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request := new(models.ProductVariant)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Detail Produk|%s", product.Name),
		"product", product,
		"request", request,
		"priceTypes", models.PriceTypes,
		"discountTypes", models.DiscountTypes,
	)
	return c.Render("admin/product_variant_new", params, "admin/layouts/authenticated")
}

func (q *productController) AdminCreateVariant(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminCreateVariant"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/products")
	}
	request := new(models.ProductVariant)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Detail Produk|%s", product.Name),
		"product", product,
		"request", request,
		"priceTypes", models.PriceTypes,
		"discountTypes", models.DiscountTypes,
	)
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_variant_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	params.Push("request", request)
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithProductIDs(product.ID).
			WithMeasureUnits(request.MeasureUnit),
	)
	if err != nil && !errors.Is(err, models.ErrProductVariantNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", product.ID))
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.ProductID = product.ID
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = now
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = now
	if productVariant != nil {
		if productVariant.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/product_variant_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductVariantMeasureUnitViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = productVariant.ID
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
	if request.ID.Compare(uuid.Nil()) == 0 {
		if _, err = q.productVariantService.CreateProductVariant(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateProductVariant")
		}
	} else {
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
		}
	}
	if err != nil {
		return c.Render("admin/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
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
	ctx := c.Context()
	ctxt := "ProductController-AdminEditVariant"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	productVariantID, err := uuid.Parse(c.Params("product_variant_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product variant not found").Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID).
			WithProductVariantIDs(productVariantID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductVariantNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	product := productVariant.Product
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Detail Produk|%s", product.Name),
		"product", product,
		"productVariant", productVariant,
		"request", productVariant,
		"priceTypes", models.PriceTypes,
		"discountTypes", models.DiscountTypes,
	)
	return c.Render("admin/product_variant_edit", params, "admin/layouts/authenticated")
}

func (q *productController) AdminUpdateVariant(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "ProductController-AdminUpdateVariant"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	productVariantID, err := uuid.Parse(c.Params("product_variant_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product variant not found").Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID).
			WithProductVariantIDs(productVariantID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductVariantNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	product := productVariant.Product
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Detail Produk|%s", product.Name),
		"product", product,
		"productVariant", productVariant,
		"request", productVariant,
		"priceTypes", models.PriceTypes,
		"discountTypes", models.DiscountTypes,
	)
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	params.Push("request", request)
	if productVariant.MeasureUnit != request.MeasureUnit {
		otherProductVariant, err := q.productVariantService.FindProductVariant(
			ctx,
			models.NewProductVariantFilter().
				WithMeasureUnits(request.MeasureUnit).
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
				"flash", flash.Danger(models.ErrUniqueProductVariantMeasureUnitViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = productVariant.ID
	request.RecordStatus = productVariant.RecordStatus
	request.ProductID = productVariant.ID
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
	ctx := c.Context()
	ctxt := "ProductController-AdminDeleteVariant"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product not found").Redirect(c, sess.Session, "/admin/products")
	}
	productVariantID, err := uuid.Parse(c.Params("product_variant_id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("product variant not found").Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID).
			WithProductVariantIDs(productVariantID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrProductVariantNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	if productVariant == nil {
		return flash.Danger(models.ErrProductVariantNotFound.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	request := productVariant
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = new(currentAdmin.ID)
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
	}
	return flash.Success("product variant deleted successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/products/%s", productID))
}
