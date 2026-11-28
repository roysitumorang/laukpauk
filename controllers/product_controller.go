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
	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type productController struct {
	productCategoryService services.ProductCategoryService
	productService         services.ProductService
	productVariantService  services.ProductVariantService
	recipeItemService      services.RecipeItemService
	storage                fiber.Storage
	userService            services.UserService
	jsonWebTokenService    services.JsonWebTokenService
}

func NewProductController(
	productCategoryService services.ProductCategoryService,
	productService services.ProductService,
	productVariantService services.ProductVariantService,
	recipeItemService services.RecipeItemService,
	storage fiber.Storage,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *productController {
	return &productController{
		productCategoryService: productCategoryService,
		productService:         productService,
		productVariantService:  productVariantService,
		recipeItemService:      recipeItemService,
		storage:                storage,
		userService:            userService,
		jsonWebTokenService:    jsonWebTokenService,
	}
}

func (q *productController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Put("/:id", q.V1SellerUpdate).
		Delete("/:id", q.V1SellerDelete)
	r.Group("/v1/buyer").
		Get("", q.V1BuyerIndex).
		Get("/:id", q.V1BuyerShow)
}

func (q *productController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "ProductController-V1SellerIndex"
	ctx := c.Context()
	request, err := sanitizers.FindProducts(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.productService.FindProducts(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProducts")
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

func (q *productController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctxt := "ProductController-V1SellerCreate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if photo == nil {
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(request.ProductCategoryID),
	)
	if errors.Is(err, models.ErrProductCategoryNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage("category_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithNames(request.Name),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		err = nil
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.ProductCategoryName = productCategory.Name
	request.Hits = 0
	request.CreatedBy = currentUser.ID
	request.CreatedAt = now
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = now
	var productVariants []*models.ProductVariant
	var recipeItems []*models.RecipeItem
	if product != nil {
		if product.RecordStatus == models.RecordStatusActive {
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueProductNameViolation.Error()).
				WriteResponse(c)
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
				WithProductIDs(product.ID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if recipeItems, _, err = q.recipeItemService.FindRecipeItems(
			ctx,
			models.NewRecipeItemFilter().
				WithProductIDs(product.ID),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, database.ErrTxClosed) {
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
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
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
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	for _, recipeItem := range recipeItems {
		recipeItem.ProductName = response.Name
		recipeItem.ProductPhoto = response.Photo
		recipeItem.ProductReadyStock = response.ReadyStock
		if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
			request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if product != nil &&
		product.Photo != os.Getenv("PRODUCT_DEFAULT_PHOTO") &&
		product.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(product.Photo)); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRemoveAll")
			request.SetPhoto(product.Photo)
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		request.SetPhoto(os.Getenv("PRODUCT_DEFAULT_PHOTO"))
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("product created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *productController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "ProductController-V1SellerShow"
	ctx := c.Context()
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrProductNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response.Variants, _, err = q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(response.ID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *productController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctxt := "ProductController-V1SellerUpdate"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrProductNotFound.Error()).
			WriteResponse(c)
	}
	product, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateProduct(ctx, c)
	if photo == nil {
		request.SetPhoto(product.Photo)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateProduct")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	productCategory, err := q.productCategoryService.FindProductCategory(
		ctx,
		models.NewProductCategoryFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductCategoryIDs(request.ProductCategoryID),
	)
	if errors.Is(err, models.ErrProductCategoryNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage("category_id: not found").
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductCategory")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
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
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherProduct != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueProductNameViolation.Error()).
				WriteResponse(c)
		}
	}
	request.ID = product.ID
	request.RecordStatus = product.RecordStatus
	request.ProductCategoryName = productCategory.Name
	request.Hits = product.Hits
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	productVariants, _, err := q.productVariantService.FindProductVariants(
		ctx,
		models.NewProductVariantFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(product.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProductVariants")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	recipeItems, _, err := q.recipeItemService.FindRecipeItems(
		ctx,
		models.NewRecipeItemFilter().
			WithProductIDs(product.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
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
		if err = tx.Rollback(ctx); errors.Is(err, database.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.productService.UpdateProduct(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	for _, productVariant := range productVariants {
		productVariant.ProductCategoryID = response.ProductCategoryID
		productVariant.ProductCategoryName = response.ProductCategoryName
		productVariant.ProductName = response.Name
		productVariant.ProductPhoto = response.Photo
		productVariant.ProductReadyStock = response.ReadyStock
		if _, err = q.productVariantService.UpdateProductVariant(ctx, tx, productVariant); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProductVariant")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	for _, recipeItem := range recipeItems {
		recipeItem.ProductName = response.Name
		recipeItem.ProductPhoto = response.Photo
		recipeItem.ProductReadyStock = response.ReadyStock
		if _, err = q.recipeItemService.UpdateRecipeItem(ctx, tx, recipeItem); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
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
	if product.Photo != os.Getenv("PRODUCT_DEFAULT_PHOTO") &&
		product.Photo != request.Photo {
		if err = os.RemoveAll(helper.PhotoLocation(product.Photo)); err != nil {
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
		SetMessage("product updated successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *productController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctxt := "ProductController-V1SellerDelete"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrProductNotFound.Error()).
			WriteResponse(c)
	}
	request, err := q.productService.FindProduct(
		ctx,
		models.NewProductFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithProductIDs(productID),
	)
	if errors.Is(err, models.ErrProductNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindProduct")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentUser.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, database.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.productService.UpdateProduct(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateProduct")
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
	return helper.NewResponse(fiber.StatusNoContent).
		WriteResponse(c)
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
			row.HideBasePrices()
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
			if err = tx.Rollback(ctx); errors.Is(err, database.ErrTxClosed) {
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
		SetData(response.HideBasePrices()).
		WriteResponse(c)
}
