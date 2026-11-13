package controllers

import (
	"context"
	"encoding/gob"
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type productController struct {
	sellerService          services.SellerService
	productCategoryService services.ProductCategoryService
	productService         services.ProductService
	productVariantService  services.ProductVariantService
	recipeItemService      services.RecipeItemService
	notificationService    services.NotificationService
	storage                fiber.Storage
}

func NewProductController(
	sellerService services.SellerService,
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
	recipeItemService services.RecipeItemService,
	notificationService services.NotificationService,
	storage fiber.Storage,
) *productController {
	return &productController{
		sellerService:          sellerService,
		productCategoryService: productCategoryService,
		productService:         productService,
		productVariantService:  productVariantService,
		recipeItemService:      recipeItemService,
		notificationService:    notificationService,
		storage:                storage,
	}
}

func (q *productController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id", q.SellerShow).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete).
		Get("/:id/variants/new", q.SellerNewVariant).
		Post("/:id/variants", q.SellerCreateVariant).
		Get("/:id/variants/:product_variant_id/edit", q.SellerEditVariant).
		Post("/:id/variants/:product_variant_id", q.SellerUpdateVariant).
		Post("/:id/variants/:product_variant_id/delete", q.SellerDeleteVariant)
	r.Group("/v1/buyer").
		Get("", q.V1BuyerIndex).
		Get("/:id", q.V1BuyerShow)
}

func (q *productController) SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerIndex"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := &models.ProductFilter{}
	rows := []*models.Product{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("seller/product_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.productService.FindProducts(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return c.Render("seller/product_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/product_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *productController) SellerNew(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerNew"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _, err := q.prepareProduct(ctx, c, flash, currentSeller, csrfToken, notifications, true, true)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/product_new", params, "seller/layouts/authenticated")
}

func (q *productController) SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerCreate"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, mapProductCategoryNames, err := q.prepareProduct(ctx, c, flash, currentSeller, csrfToken, notifications, true, true)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
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
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "seller/layouts/authenticated")
	}
	productCategoryName, ok := mapProductCategoryNames[request.ProductCategoryID]
	if !ok {
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger("product_category_id: not found"),
			"request", request,
			"product", request,
		), "seller/layouts/authenticated")
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithNames(request.Name),
	)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.ProductCategoryName = productCategoryName
	request.Hits = 0
	request.CreatedBy = currentSeller.ID
	request.CreatedAt = now
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = now
	var productVariants []*models.ProductVariant
	var recipeItems []*models.RecipeItem
	if product != nil {
		if product.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return c.Render("seller/product_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductNameViolation.Error()),
				"request", request,
				"product", request,
			), "seller/layouts/authenticated")
		}
		request.ID = product.ID
		request.Name = product.Name
		request.CreatedBy = product.CreatedBy
		request.CreatedAt = product.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
		if photo == nil {
			request.SetPhoto(product.Photo)
		}
		if productVariants, _, err = q.productVariantService.FindProductVariants(
			ctx,
			models.NewProductVariantFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductCategoryIDs(request.ProductCategoryID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return c.Render("seller/product_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"product", request,
			), "seller/layouts/authenticated")
		}
		if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
			ctx,
			models.NewRecipeItemFilter().
				WithProductIDs(product.ID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return c.Render("seller/product_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"product", request,
			), "seller/layouts/authenticated")
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
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
	response, err := q.productService.SaveProduct(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProduct")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "seller/layouts/authenticated")
	}
	for _, productVariant := range productVariants {
		productVariant.ProductCategoryID = response.ProductCategoryID
		productVariant.ProductCategoryName = response.ProductCategoryName
		productVariant.ProductName = response.Name
		productVariant.ProductPhoto = response.Photo
		productVariant.ProductReadyStock = response.ReadyStock
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return c.Render("seller/product_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"product", request,
			), "seller/layouts/authenticated")
		}
	}
	for _, recipeItem := range recipeItems {
		recipeItem.ProductName = response.Name
		recipeItem.ProductPhoto = response.Photo
		recipeItem.ProductReadyStock = response.ReadyStock
		if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return c.Render("seller/product_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"product", request,
			), "seller/layouts/authenticated")
		}
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "seller/layouts/authenticated")
	}
	if product != nil &&
		product.Photo != os.Getenv("PRODUCT_DEFAULT_PHOTO") &&
		product.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(product.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(product.Photo)
			return c.Render("seller/product_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request,
				"product", request,
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return c.Render("seller/product_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
			"product", request,
		), "seller/layouts/authenticated")
	}
	return flash.Success("product created successfully").Redirect(c, sess.Session, "/seller/products")
}

func (q *productController) SellerShow(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerShow"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, err := q.prepareProductByID(ctx, c)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductByID")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	if product == nil {
		return flash.Danger(models.ErrProductNotFound.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	var productVariants []*models.ProductVariant
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Detail Produk|%s", product.Name),
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
		return c.Render("seller/product_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	params.Push("productVariants", productVariants)
	return c.Render("seller/product_show", params, "seller/layouts/authenticated")
}

func (q *productController) SellerEdit(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerEdit"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, _, err := q.prepareProduct(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		notifications,
		true,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	return c.Render("seller/product_edit", params, "seller/layouts/authenticated")
}

func (q *productController) SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerUpdate"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, params, mapProductCategoryNames, err := q.prepareProduct(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		notifications,
		true,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProduct")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if photo == nil {
		request.SetPhoto(product.Photo)
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "seller/layouts/authenticated")
	}
	params.Push("request", request)
	productCategoryName, ok := mapProductCategoryNames[request.ProductCategoryID]
	if !ok {
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger("product_category_id: not found"),
			"request", request.SetPhoto(product.Photo),
		), "seller/layouts/authenticated")
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
			return c.Render("seller/product_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(product.Photo),
			), "seller/layouts/authenticated")
		}
		if otherProduct != nil {
			return c.Render("seller/product_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductNameViolation.Error()),
				"request", request.SetPhoto(product.Photo),
			), "seller/layouts/authenticated")
		}
	}
	request.ID = product.ID
	request.RecordStatus = product.RecordStatus
	request.ProductCategoryName = productCategoryName
	request.Hits = product.Hits
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	productVariants, _, err := q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(request.ProductCategoryID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "seller/layouts/authenticated")
	}
	recipeItems, _, err := q.recipeItemService.FindRecipeItems(
		ctx,
		models.NewRecipeItemFilter().
			WithProductIDs(product.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "seller/layouts/authenticated")
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
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
	response, err := q.productService.UpdateProduct(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "seller/layouts/authenticated")
	}
	for _, productVariant := range productVariants {
		productVariant.ProductCategoryID = response.ProductCategoryID
		productVariant.ProductCategoryName = response.ProductCategoryName
		productVariant.ProductName = response.Name
		productVariant.ProductPhoto = response.Photo
		productVariant.ProductReadyStock = response.ReadyStock
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
			return c.Render("seller/product_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(product.Photo),
			), "seller/layouts/authenticated")
		}
	}
	for _, recipeItem := range recipeItems {
		recipeItem.ProductName = response.Name
		recipeItem.ProductPhoto = response.Photo
		recipeItem.ProductReadyStock = response.ReadyStock
		if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
			return c.Render("seller/product_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(product.Photo),
			), "seller/layouts/authenticated")
		}
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "seller/layouts/authenticated")
	}
	if product.Photo != os.Getenv("PRODUCT_DEFAULT_PHOTO") &&
		product.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(product.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			return c.Render("seller/product_edit", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(product.Photo),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/product_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(product.Photo),
		), "seller/layouts/authenticated")
	}
	return flash.Success("product updated successfully").Redirect(c, sess.Session, "/seller/products")
}

func (q *productController) SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerDelete"
	ctx, sess, flash, currentSeller, _, _ := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request, err := q.prepareProductByID(ctx, c)
	if err != nil && !errors.Is(err, models.ErrProductNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductByID")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	if request == nil {
		return flash.Danger(models.ErrProductNotFound.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentSeller.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	return flash.Success("product deleted successfully").Redirect(c, sess.Session, "/seller/products")
}

func (q *productController) SellerNewVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerNewVariant"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		notifications,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	return c.Render("seller/product_variant_new", params, "seller/layouts/authenticated")
}

func (q *productController) SellerCreateVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerCreateVariant"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, _, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		notifications,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/product_variant_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	productVariant, err := q.productVariantService.FindProductVariant(
		ctx,
		models.NewProductVariantFilter().
			WithProductIDs(product.ID).
			WithStockUnits(request.StockUnit),
	)
	if err != nil && !errors.Is(err, models.ErrProductVariantNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/products/%s", product.ID))
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.ProductCategoryID = product.ProductCategoryID
	request.ProductCategoryName = product.ProductCategoryName
	request.ProductID = product.ID
	request.ProductName = product.Name
	request.ProductPhoto = product.Photo
	request.ProductReadyStock = product.ReadyStock
	request.CreatedBy = currentSeller.ID
	request.CreatedAt = now
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = now
	if productVariant != nil {
		if productVariant.RecordStatus == models.RecordStatusActive {
			return c.Render("seller/product_variant_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductVariantStockUnitViolation.Error()),
			), "seller/layouts/authenticated")
		}
		request.ID = productVariant.ID
		request.StockUnit = productVariant.StockUnit
		request.CreatedBy = productVariant.CreatedBy
		request.CreatedAt = productVariant.CreatedAt
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/product_variant_new", params.Push(
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
	if _, err = q.productVariantService.SaveProductVariant(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProductVariant")
		return c.Render("seller/product_variant_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if request.SellingPrice < product.LowestSellingPrice {
		product.LowestSellingPrice = request.SellingPrice
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			return c.Render("seller/product_variant_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/product_variant_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Success("product variant created successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/products/%s", product.ID))
}

func (q *productController) SellerEditVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerEditVariant"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, _, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		notifications,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	return c.Render("seller/product_variant_edit", params, "seller/layouts/authenticated")
}

func (q *productController) SellerUpdateVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerUpdateVariant"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, productVariant, params, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		notifications,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	request, statusCode, err := sanitizers.ValidateProductVariant(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductVariant")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
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
			return c.Render("seller/product_variant_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherProductVariant != nil {
			return c.Render("seller/product_variant_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueProductVariantStockUnitViolation.Error()),
			), "seller/layouts/authenticated")
		}
	}
	recipeItems, _, err := q.recipeItemService.FindRecipeItems(
		ctx,
		models.NewRecipeItemFilter().
			WithProductVariantIDs(productVariant.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
		return c.Render("seller/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	request.ID = productVariant.ID
	request.RecordStatus = productVariant.RecordStatus
	request.ProductCategoryID = product.ProductCategoryID
	request.ProductCategoryName = product.ProductCategoryName
	request.ProductID = product.ID
	request.ProductName = product.Name
	request.ProductPhoto = product.Photo
	request.ProductReadyStock = productVariant.ProductReadyStock
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/product_variant_edit", params.Push(
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
	response, err := q.productVariantService.UpdateProductVariant(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
		return c.Render("seller/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if request.SellingPrice < product.LowestSellingPrice {
		product.LowestSellingPrice = request.SellingPrice
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			return c.Render("seller/product_variant_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
	}
	for _, recipeItem := range recipeItems {
		recipeItem.ProductVariantStockUnit = response.StockUnit
		if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
			return c.Render("seller/product_variant_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/product_variant_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Success("product variant updated successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/products/%s", product.ID))
}

func (q *productController) SellerDeleteVariant(c fiber.Ctx) (err error) {
	ctxt := "ProductController-SellerDeleteVariant"
	ctx, sess, flash, currentSeller, csrfToken, notifications := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	product, productVariant, _, err := q.prepareProductVariant(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		notifications,
		false,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductVariant")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/products")
	}
	productVariant.RecordStatus = models.RecordStatusDeleted
	productVariant.DeletedBy = new(currentSeller.ID)
	productVariant.DeletedAt = new(time.Now())
	productVariant.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/products/%s", product.ID))
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/products/%s", product.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/products/%s", product.ID))
	}
	return flash.Success("product variant deleted successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/products/%s", product.ID))
}

func (q *productController) prepare(c fiber.Ctx) (
	context.Context,
	*session.Middleware,
	*helper.FlashMessage,
	*models.Seller,
	string,
	[]*models.Notification,
) {
	ctx := c.Context()
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications := c.Locals(models.CurrentNotifications).([]*models.Notification)
	csrfToken := csrf.TokenFromContext(c)
	return ctx, sess, flash, currentSeller, csrfToken, notifications
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
	currentSeller *models.Seller,
	csrfToken string,
	notifications []*models.Notification,
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
			"currentSeller", currentSeller,
			"notifications", notifications,
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
		if withParams && params.Size() > 0 && product != nil {
			params.Push(
				"pageTitle", helper.Sprintf("Edit Data Produk|%s", product.Name),
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
			mapProductCategoryNames[productCategory.ID] = productCategory.Name
		}
	}
	return product, params, mapProductCategoryNames, nil
}

func (q *productController) prepareProductVariant(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentSeller *models.Seller,
	csrfToken string,
	notifications []*models.Notification,
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
	if withParams {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentSeller", currentSeller,
			"notifications", notifications,
			"pageTitle", helper.Sprintf("Detail Produk|%s", product.Name),
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
	infiniteScroll, err := q.productService.InfiniteScrollProducts(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive).
			WithReadyStock(new(true)),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrInfiniteScrollProducts")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	for i := range infiniteScroll.Rows {
		row, ok := infiniteScroll.Rows[i].(*models.Product)
		if ok {
			row.Photo = row.PhotoPath(c.BaseURL())
		}
		infiniteScroll.Rows[i] = row
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(infiniteScroll).
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
	productCache, err := q.storage.GetWithContext(ctx, c.Path())
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrGetWithContext")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buffer := pools.BufferPool.Get()
	defer func() {
		buffer.Reset()
		pools.BufferPool.Put(buffer)
	}()
	response := &models.Product{}
	if productCache == nil {
		if response, err = q.productService.FindProduct(
			ctx,
			models.NewProductFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductIDs(productID),
		); errors.Is(err, models.ErrProductNotFound) || response == nil {
			return helper.NewResponse(fiber.StatusNotFound).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
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
	} else {
		_, _ = buffer.Write(productCache)
		if err = gob.NewDecoder(buffer).Decode(&response); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDecode")
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	hitsKey := helper.Sprintf("%s-hits", c.Path())
	hitsCache, err := q.storage.GetWithContext(ctx, hitsKey)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrGetWithContext")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if hitsCache != nil {
		buffer.Reset()
		_, _ = buffer.Write(hitsCache)
		if err = gob.NewDecoder(buffer).Decode(&response.Hits); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDecode")
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	var counter atomic.Int64
	counter.Store(response.Hits)
	counter.Add(1)
	if hits := counter.Load(); hits%10 != 0 {
		response.Hits = hits
	} else {
		product, err := q.productService.FindProduct(
			ctx,
			models.NewProductFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductIDs(productID),
		)
		if errors.Is(err, models.ErrProductNotFound) || product == nil {
			return helper.NewResponse(fiber.StatusNotFound).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
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
		product.Hits = hits
		response, err = q.productService.UpdateProduct(ctx, tx, product)
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
	}
	buffer.Reset()
	if err = gob.NewEncoder(buffer).Encode(response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrEncode")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productCache = buffer.Bytes()
	if err = q.storage.SetWithContext(ctx, c.Path(), productCache, 0); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSetWithContext")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buffer.Reset()
	if err = gob.NewEncoder(buffer).Encode(response.Hits); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrEncode")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	hitsCache = buffer.Bytes()
	if err = q.storage.SetWithContext(ctx, hitsKey, hitsCache, 0); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSetWithContext")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
