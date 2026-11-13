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

type buyerController struct {
	buyerService        services.BuyerService
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewBuyerController(
	buyerService services.BuyerService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *buyerController {
	return &buyerController{
		buyerService:        buyerService,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *buyerController) Mount(r fiber.Router) {
	r.Group("/v1/buyer").
		Post("", q.BuyerCreate)
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)

}

func (q *buyerController) BuyerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-BuyerCreate"
	request, statusCode, err := sanitizers.ValidateBuyerRegistration(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBuyerRegistration")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithUsername(request.Username),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueBuyerUsernameViolation.Error()).
			WriteResponse(c)
	}
	if buyer, err = q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithMobilePhone(request.MobilePhone),
	); err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueBuyerMobilePhoneViolation.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	encryptedPasword, err := helper.HashPassword(request.Password)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer = &models.Buyer{
		Name:              request.Name,
		Username:          request.Username,
		MobilePhone:       request.MobilePhone,
		EncryptedPassword: &encryptedPasword,
		Photo:             os.Getenv("BUYER_DEFAULT_PHOTO"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if request.ReferralCode != "" {
		referrer, err := q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithReferralCodes(request.ReferralCode),
		)
		if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if referrer != nil {
			buyer.ReferrerID = &referrer.ID
		}
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
	response, err := q.buyerService.CreateBuyer(ctx, tx, buyer)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBuyer")
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
	return helper.NewResponse(fiber.StatusCreated).
		SetData(response).
		WriteResponse(c)
}

func (q *buyerController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-V1SellerIndex"
	request, err := sanitizers.FindBuyers(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.buyerService.FindBuyers(
		ctx,
		request.WithStatuses(models.BuyerStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
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

func (q *buyerController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-V1SellerCreate"
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBuyer(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBuyer")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithUsername(request.Username),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueBuyerUsernameViolation.Error()).
			WriteResponse(c)
	}
	if buyer, err = q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithMobilePhone(request.MobilePhone),
	); err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueBuyerMobilePhoneViolation.Error()).
			WriteResponse(c)
	}
	if request.Email != nil {
		if buyer, err = q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithEmail(*request.Email),
		); err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if buyer != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueBuyerEmailViolation.Error()).
				WriteResponse(c)
		}
	}
	now := time.Now().In(time.Local)
	request.ReferralPoints = 0
	request.RewardPoints = 0
	request.Wallet = 0
	request.TransactionVolume = 0
	request.CreatedAt = now
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("BUYER_DEFAULT_PHOTO")
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
	saved, err := q.buyerService.CreateBuyer(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBuyer")
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
		SetMessage("buyer created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *buyerController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-V1SellerShow"
	buyerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBuyerNotFound.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithBuyerIDs(buyerID),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBuyerNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(buyer).
		WriteResponse(c)
}

func (q *buyerController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-V1SellerUpdate"
	buyerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBuyerNotFound.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithBuyerIDs(buyerID),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBuyerNotFound.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBuyer(ctx, c)
	if photo == nil {
		request.Photo = buyer.Photo
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBuyer")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer.Username != request.Username {
		otherBuyer, err := q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithUsername(request.Username).
				WithExcludedBuyerIDs(buyer.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherBuyer != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueBuyerUsernameViolation.Error()).
				WriteResponse(c)
		}
	}
	if buyer.MobilePhone != request.MobilePhone {
		otherBuyer, err := q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithMobilePhone(request.MobilePhone).
				WithExcludedBuyerIDs(buyer.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherBuyer != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueBuyerMobilePhoneViolation.Error()).
				WriteResponse(c)
		}
	}
	if request.Email != nil &&
		(buyer.Email == nil ||
			*buyer.Email != *request.Email) {
		otherBuyer, err := q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithEmail(*request.Email).
				WithExcludedBuyerIDs(buyer.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherBuyer != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueBuyerEmailViolation.Error()).
				WriteResponse(c)
		}
	}
	now := time.Now().In(time.Local)
	buyer.Name = request.Name
	buyer.Username = request.Username
	buyer.Photo = request.Photo
	buyer.Address = request.Address
	buyer.ReferralPoints = 0
	buyer.RewardPoints = 0
	buyer.Wallet = 0
	buyer.TransactionVolume = 0
	buyer.ReferrerID = request.ReferrerID
	buyer.UpdatedAt = now
	buyer.DeactivatedBy = nil
	buyer.DeactivatedAt = nil
	buyer.DeactivationReason = nil
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
	saved, err := q.buyerService.UpdateBuyer(ctx, tx, buyer)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
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
		SetMessage("buyer updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *buyerController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	buyerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBuyerNotFound.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithBuyerIDs(buyerID),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBuyerNotFound.Error()).
			WriteResponse(c)
	}
	buyer.Status = models.BuyerStatusDeactivated
	buyer.DeactivatedBy = &currentUser.ID
	buyer.DeactivatedAt = new(time.Now())
	buyer.DeactivationReason = nil
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
	saved, err := q.buyerService.UpdateBuyer(ctx, tx, buyer)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
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
		SetMessage("buyer deactivated successfully").
		SetData(saved).
		WriteResponse(c)
}
