package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type productCategoryController struct {
	productCategoryService services.ProductCategoryService
	productService         services.ProductService
	productVariantService  services.ProductVariantService
	userService            services.UserService
	jsonWebTokenService    services.JsonWebTokenService
}

func NewProductCategoryController(
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *productCategoryController {
	return &productCategoryController{
		productCategoryService: productCategoryService,
		productService:         productService,
		productVariantService:  productVariantService,
		userService:            userService,
		jsonWebTokenService:    jsonWebTokenService,
	}
}

func (q *productCategoryController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
	r.Group("/v1/buyer").
		Get("", q.V1BuyerIndex)
}

func (q *productCategoryController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-V1SellerIndex"
	ctx := c.Context()
	request, err := sanitizers.FindProductCategories(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.productCategoryService.FindProductCategories(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategories")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	pagination.Rows = make([]any, len(rows))
	for i := range rows {
		pagination.Rows[i] = rows[i]
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}

func (q *productCategoryController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-V1SellerCreate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
	if photo == nil {
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductCategory")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithNames(request.Name),
	)
	if err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentUser.ID
	request.CreatedAt = now
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = now
	var products []*models.Product
	var productVariants []*models.ProductVariant
	if productCategory != nil {
		if productCategory.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueProductCategoryNameViolation.Error()).
				WriteResponse(c)
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
				WithRecordStatuses(models.RecordStatusActive).
				WithProductCategoryIDs(productCategory.ID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if productVariants, _, err = q.productVariantService.FindProductVariants(
			ctx,
			models.NewProductVariantFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductCategoryIDs(productCategory.ID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
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
	response, err := q.productCategoryService.SaveProductCategory(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveProductCategory")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	for _, product := range products {
		product.ProductCategoryName = response.Name
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	for _, productVariant := range productVariants {
		productVariant.ProductCategoryName = response.Name
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
			request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if productCategory != nil &&
		productCategory.Photo != os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO") &&
		productCategory.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(productCategory.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(productCategory.Photo)
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("product category created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *productCategoryController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-V1SellerShow"
	ctx := c.Context()
	productCategory, err := q.prepareProductCategory(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(productCategory).
		WriteResponse(c)
}

func (q *productCategoryController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-V1SellerUpdate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	productCategory, err := q.prepareProductCategory(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProductCategory(ctx, c)
	if photo == nil {
		request.SetPhoto(productCategory.Photo)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProductCategory")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
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
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherProductCategory != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueProductCategoryNameViolation.Error()).
				WriteResponse(c)
		}
	}
	request.ID = productCategory.ID
	request.RecordStatus = productCategory.RecordStatus
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	products, _, err := q.productService.FindProducts(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(productCategory.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productVariants, _, err := q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(productCategory.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
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
	response, err := q.productCategoryService.UpdateProductCategory(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	for _, product := range products {
		product.ProductCategoryName = response.Name
		if _, err = q.productService.UpdateProduct(ctx, tx, product); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	for _, productVariant := range productVariants {
		productVariant.ProductCategoryName = response.Name
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if productCategory.Photo != os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO") &&
		productCategory.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(productCategory.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("product category updated successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *productCategoryController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "ProductCategoryController-V1SellerDelete"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	productCategory, err := q.prepareProductCategory(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareProductCategory")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productCategory.RecordStatus = models.RecordStatusDeleted
	productCategory.DeletedBy = &currentUser.ID
	productCategory.DeletedAt = new(time.Now())
	productCategory.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
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
	saved, err := q.productCategoryService.UpdateProductCategory(ctx, tx, productCategory)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductCategory")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("product category deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *productCategoryController) prepareProductCategory(
	ctx context.Context,
	c fiber.Ctx,
) (
	*models.ProductCategory,
	error,
) {
	productCategory := &models.ProductCategory{}
	productCategory.SetPhoto(os.Getenv("PRODUCT_CATEGORY_DEFAULT_PHOTO"))
	if c.Params("id") != "" {
		productCategoryID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return nil, models.ErrProductCategoryNotFound
		}
		if productCategory, err = q.productCategoryService.FindProductCategory(
			ctx,
			models.NewProductCategoryFilter().
				WithRecordStatuses(models.RecordStatusActive).
				WithProductCategoryIDs(productCategoryID),
		); err != nil && !errors.Is(err, models.ErrProductCategoryNotFound) {
			return nil, err
		}
		if productCategory == nil {
			return nil, models.ErrProductCategoryNotFound
		}
	}
	return productCategory, nil
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
