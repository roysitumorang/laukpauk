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

type sellerController struct {
	sellerService       services.SellerService
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewSellerController(
	sellerService services.SellerService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *sellerController {
	return &sellerController{
		sellerService:       sellerService,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *sellerController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService, models.UserLevelSellerOwner)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
}

func (q *sellerController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerIndex"
	request, err := sanitizers.FindSellers(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSellers")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.sellerService.FindSellers(
		ctx,
		request.WithStatuses(models.SellerStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSellers")
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

func (q *sellerController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateSeller(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSeller")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithUsername(request.Username),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if seller != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueSellerUsernameViolation.Error()).
			WriteResponse(c)
	}
	if seller, err = q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithMobilePhone(request.MobilePhone),
	); err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if seller != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueSellerMobilePhoneViolation.Error()).
			WriteResponse(c)
	}
	if request.Email != nil {
		if seller, err = q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithEmail(*request.Email),
		); err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if seller != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueSellerEmailViolation.Error()).
				WriteResponse(c)
		}
	}
	now := time.Now().In(time.Local)
	createdBy := currentUser.ID
	request.SellerLevel = models.SellerLevelStaff
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedBy = &createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("SELLER_DEFAULT_PHOTO")
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
	saved, err := q.sellerService.CreateSeller(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSeller")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
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
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("seller created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *sellerController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerShow"
	sellerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrSellerNotFound.Error()).
			WriteResponse(c)
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithStatuses(models.SellerStatusActive).
			WithSellerIDs(sellerID),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if seller == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrSellerNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(seller).
		WriteResponse(c)
}

func (q *sellerController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	sellerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrSellerNotFound.Error()).
			WriteResponse(c)
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithStatuses(models.SellerStatusActive).
			WithSellerIDs(sellerID),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if seller == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrSellerNotFound.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateSeller(ctx, c)
	if photo == nil {
		request.Photo = seller.Photo
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSeller")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if seller.Username != request.Username {
		otherSeller, err := q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithUsername(request.Username).
				WithExcludedSellerIDs(seller.ID),
		)
		if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherSeller != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueSellerUsernameViolation.Error()).
				WriteResponse(c)
		}
	}
	if seller.MobilePhone != request.MobilePhone {
		otherSeller, err := q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithMobilePhone(request.MobilePhone).
				WithExcludedSellerIDs(seller.ID),
		)
		if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherSeller != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueSellerMobilePhoneViolation.Error()).
				WriteResponse(c)
		}
	}
	if request.Email != nil && (seller.Email == nil ||
		*seller.Email != *request.Email) {
		otherSeller, err := q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithEmail(*request.Email).
				WithExcludedSellerIDs(seller.ID),
		)
		if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherSeller != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueSellerEmailViolation.Error()).
				WriteResponse(c)
		}
	}
	now := time.Now().In(time.Local)
	seller.Name = request.Name
	seller.Username = request.Username
	seller.MobilePhone = request.MobilePhone
	seller.Email = request.Email
	seller.Photo = request.Photo
	seller.UpdatedBy = &currentUser.ID
	seller.UpdatedAt = now
	seller.DeactivatedBy = nil
	seller.DeactivatedAt = nil
	seller.DeactivationReason = nil
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
	saved, err := q.sellerService.UpdateSeller(ctx, tx, seller)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSeller")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
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
		SetMessage("seller updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *sellerController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	sellerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrSellerNotFound.Error()).
			WriteResponse(c)
	}
	if sellerID.Compare(currentUser.ID) == 0 {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrSellerSelfDeletingProhibited.Error()).
			WriteResponse(c)
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithStatuses(models.SellerStatusActive).
			WithSellerIDs(sellerID),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if seller == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrSellerNotFound.Error()).
			WriteResponse(c)
	}
	seller.Status = models.SellerStatusDeactivated
	seller.DeactivatedBy = &currentUser.ID
	seller.DeactivatedAt = new(time.Now())
	seller.DeactivationReason = nil
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
	saved, err := q.sellerService.UpdateSeller(ctx, tx, seller)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSeller")
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
		SetMessage("seller deactivated successfully").
		SetData(saved).
		WriteResponse(c)
}

// notFoundStatusCode answers 404 for the errors seller lookups return when a
// record does not exist, and 422 for anything else.
func notFoundStatusCode(err error) int {
	for _, notFound := range []error{
		models.ErrProductNotFound,
		models.ErrProductVariantNotFound,
		models.ErrProductCategoryNotFound,
		models.ErrRecipeNotFound,
		models.ErrRecipeGuideNotFound,
		models.ErrRecipeItemNotFound,
		models.ErrTransactionNotFound,
	} {
		if errors.Is(err, notFound) {
			return fiber.StatusNotFound
		}
	}
	return fiber.StatusUnprocessableEntity
}
