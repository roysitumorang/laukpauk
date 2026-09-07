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

type productCategoryController struct {
	adminService           services.AdminService
	sellerService          services.SellerService
	productCategoryService services.ProductCategoryService
	productService         services.ProductService
	productVariantService  services.ProductVariantService
}

func NewProductCategoryController(
	adminService services.AdminService,
	sellerService services.SellerService,
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
) *productCategoryController {
	return &productCategoryController{
		adminService:           adminService,
		sellerService:          sellerService,
		productCategoryService: productCategoryService,
		productService:         productService,
		productVariantService:  productVariantService,
	}
}

func (q *productCategoryController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
	r.Group("/v1/buyer").
		Get("", q.V1BuyerIndex)
}

func (q *productCategoryController) SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-SellerIndex"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := &models.ProductCategoryFilter{}
	rows := []*models.ProductCategory{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
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
	request, err = sanitizers.FindProductCategories(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("seller/product_category_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.productCategoryService.FindProductCategories(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return c.Render("seller/product_category_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/product_category_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *productCategoryController) SellerNew(c fiber.Ctx) (err error) {
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _ := q.prepareProductCategory(ctx, c, flash, currentSeller, csrfToken, true)
	return c.Render("seller/product_category_new", params, "seller/layouts/authenticated")
}

func (q *productCategoryController) SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-SellerCreate"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _ := q.prepareProductCategory(ctx, c, flash, currentSeller, csrfToken, true)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
	if photo == nil {
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
	}
	params.Push(
		"request", request,
		"productCategory", request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductCategory")
		c.Response().SetStatusCode(statusCode)
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return c.Render("seller/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"productCategory", request,
		), "seller/layouts/authenticated")
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithNames(request.Name),
	)
	if err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return c.Render("seller/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"productCategory", request,
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentSeller.ID
	request.CreatedAt = now
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = now
	var products []*models.Product
	var productVariants []*models.ProductVariant
	if productCategory != nil {
		if productCategory.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return c.Render("seller/product_category_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductCategoryNameViolation.Error()),
				"request", request,
				"productCategory", request,
			), "seller/layouts/authenticated")
		}
		request.ID = productCategory.ID
		request.Name = productCategory.Name
		request.CreatedBy = productCategory.CreatedBy
		request.CreatedAt = productCategory.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
		if photo == nil {
			request.SetPhoto(productCategory.Photo)
		}
		if products, _, err = q.productService.FindProducts(
			ctx,
			models.NewProductFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecordStatuses(models.RecordStatusActive).
				WithProductCategoryIDs(productCategory.ID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return c.Render("seller/product_category_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"productCategory", request,
			), "seller/layouts/authenticated")
		}
		if productVariants, _, err = q.productVariantService.FindProductVariants(
			ctx,
			models.NewProductVariantFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithRecordStatuses(models.RecordStatusActive).
				WithProductCategoryIDs(productCategory.ID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return c.Render("seller/product_category_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"productCategory", request,
			), "seller/layouts/authenticated")
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return c.Render("seller/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"productCategory", request,
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
	response, err := q.productCategoryService.SaveProductCategory(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProductCategory")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return c.Render("seller/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"productCategory", request,
		), "seller/layouts/authenticated")
	}
	for _, product := range products {
		product.ProductCategoryName = response.Name
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return c.Render("seller/product_category_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"productCategory", request,
			), "seller/layouts/authenticated")
		}
	}
	for _, productVariant := range productVariants {
		productVariant.ProductCategoryName = response.Name
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return c.Render("seller/product_category_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"productCategory", request,
			), "seller/layouts/authenticated")
		}
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return c.Render("seller/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"productCategory", request,
		), "seller/layouts/authenticated")
	}
	if productCategory != nil &&
		productCategory.Photo != os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO") &&
		productCategory.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(productCategory.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(productCategory.Photo)
			return c.Render("seller/product_category_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"productCategory", request,
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return c.Render("seller/product_category_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"productCategory", request,
		), "seller/layouts/authenticated")
	}
	return flash.Success("product category created successfully").Redirect(c, sess.Session, "/seller/product_categories")
}

func (q *productCategoryController) SellerEdit(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-SellerEdit"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareProductCategory(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/product_categories")
	}
	return c.Render("seller/product_category_edit", params, "seller/layouts/authenticated")
}

func (q *productCategoryController) SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-SellerUpdate"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	productCategory, params, err := q.prepareProductCategory(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/product_categories")
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
	if photo == nil {
		request.SetPhoto(productCategory.Photo)
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductCategory")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(productCategory.Photo),
		), "seller/layouts/authenticated")
	}
	if productCategory.Name != request.Name {
		otherProductCategory, err := q.productCategoryService.FindProductCategory(
			ctx,
			models.NewProductCategoryFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithNames(request.Name).
				WithExcludedProductCategoryIDs(productCategory.ID),
		)
		if err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
			return c.Render("seller/product_category_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(productCategory.Photo),
			), "seller/layouts/authenticated")
		}
		if otherProductCategory != nil {
			return c.Render("seller/product_category_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductCategoryNameViolation.Error()),
				"request", request.SetPhoto(productCategory.Photo),
			), "seller/layouts/authenticated")
		}
	}
	request.ID = productCategory.ID
	request.CompanyID = productCategory.CompanyID
	request.RecordStatus = productCategory.RecordStatus
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	products, _, err := q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(productCategory.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return c.Render("seller/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(productCategory.Photo),
		), "seller/layouts/authenticated")
	}
	productVariants, _, err := q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(productCategory.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return c.Render("seller/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(productCategory.Photo),
		), "seller/layouts/authenticated")
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(productCategory.Photo),
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
	response, err := q.productCategoryService.UpdateProductCategory(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return c.Render("seller/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(productCategory.Photo),
		), "seller/layouts/authenticated")
	}
	for _, product := range products {
		product.ProductCategoryName = response.Name
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			return c.Render("seller/product_category_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(productCategory.Photo),
			), "seller/layouts/authenticated")
		}
	}
	for _, productVariant := range productVariants {
		productVariant.ProductCategoryName = response.Name
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
			return c.Render("seller/product_category_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(productCategory.Photo),
			), "seller/layouts/authenticated")
		}
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("seller/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(productCategory.Photo),
		), "seller/layouts/authenticated")
	}
	if productCategory.Photo != os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO") &&
		productCategory.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(productCategory.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			return c.Render("seller/product_category_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(productCategory.Photo),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/product_category_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(productCategory.Photo),
		), "seller/layouts/authenticated")
	}
	return flash.Success("product category updated successfully").Redirect(c, sess.Session, "/seller/product_categories")
}

func (q *productCategoryController) SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-SellerDelete"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	productCategory, _, err := q.prepareProductCategory(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/product_categories")
	}
	productCategory.RecordStatus = models.RecordStatusDeleted
	productCategory.DeletedBy = &currentSeller.ID
	productCategory.DeletedAt = new(time.Now())
	productCategory.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/product_categories")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.productCategoryService.UpdateProductCategory(ctx, tx, productCategory); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/product_categories")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/product_categories")
	}
	return flash.Success("product category deleted successfully").Redirect(c, sess.Session, "/seller/product_categories")
}

func (q *productCategoryController) prepare(c fiber.Ctx) (
	context.Context,
	*session.Middleware,
	*helper.FlashMessage,
	*models.Seller,
	string,
) {
	ctx := c.Context()
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	csrfToken := csrf.TokenFromContext(c)
	return ctx, sess, flash, currentSeller, csrfToken
}

func (q *productCategoryController) prepareProductCategory(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentSeller *models.Seller,
	csrfToken string,
	withParams bool,
) (
	*models.ProductCategory,
	models.Map,
	error,
) {
	productCategory := &models.ProductCategory{}
	productCategory.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
	var params models.Map
	if withParams {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentSeller", currentSeller,
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
				WithCompanyIDs(currentSeller.CompanyID).
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
		ctx,
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
