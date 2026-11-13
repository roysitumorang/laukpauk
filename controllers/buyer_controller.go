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

// buyerController lets sellers manage buyer accounts: users with role buyer.
type buyerController struct {
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewBuyerController(
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *buyerController {
	return &buyerController{
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *buyerController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
}

func (q *buyerController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-V1SellerIndex"
	request, err := sanitizers.FindUserAccounts(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUserAccounts")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.userService.FindUsers(
		ctx,
		request.
			WithRoles(models.RoleBuyer).
			WithStatuses(models.UserStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUsers")
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
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateUserAccount(ctx, c, false, true)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateUserAccount")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if statusCode, err = checkUniqueLogins(ctx, q.userService, request, uuid.Nil()); err != nil {
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if statusCode, err = q.checkReferrer(ctx, request.ReferrerID, uuid.Nil()); err != nil {
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	createdBy := currentUser.ID
	user := &models.User{
		Role:        models.RoleBuyer,
		Status:      models.UserStatusActive,
		Name:        request.Name,
		MobilePhone: request.MobilePhone,
		Username:    &request.Username,
		Email:       request.Email,
		Photo:       request.Photo,
		Address:     request.Address,
		ReferrerID:  request.ReferrerID,
		CreatedBy:   &createdBy,
		CreatedAt:   now,
		UpdatedBy:   &createdBy,
		UpdatedAt:   now,
	}
	if request.Password != "" {
		encryptedPassword, err := helper.HashPassword(request.Password)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		user.EncryptedPassword = &encryptedPassword
		user.PasswordChangedAt = &now
	}
	if photo == nil {
		user.Photo = os.Getenv("BUYER_DEFAULT_PHOTO")
	}
	if err = user.Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return helper.NewResponse(fiber.StatusBadRequest).
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
	saved, err := q.userService.CreateUser(ctx, tx, user)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, user.Photo); err != nil {
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
	user, err := q.prepareBuyer(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareBuyer")
		return helper.NewResponse(userNotFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(user).
		WriteResponse(c)
}

func (q *buyerController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	user, err := q.prepareBuyer(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareBuyer")
		return helper.NewResponse(userNotFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateUserAccount(ctx, c, false, true)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateUserAccount")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if statusCode, err = checkUniqueLogins(ctx, q.userService, request, user.ID); err != nil {
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if statusCode, err = q.checkReferrer(ctx, request.ReferrerID, user.ID); err != nil {
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	if request.Password != "" {
		encryptedPassword, err := helper.HashPassword(request.Password)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		user.EncryptedPassword = &encryptedPassword
		user.PasswordChangedAt = &now
	}
	user.Name = request.Name
	user.Username = &request.Username
	user.MobilePhone = request.MobilePhone
	user.Email = request.Email
	user.Address = request.Address
	user.ReferrerID = request.ReferrerID
	if photo != nil {
		user.Photo = request.Photo
	}
	user.UpdatedBy = &currentUser.ID
	user.UpdatedAt = now
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
	saved, err := q.userService.UpdateUser(ctx, tx, user)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, user.Photo); err != nil {
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
	user, err := q.prepareBuyer(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareBuyer")
		return helper.NewResponse(userNotFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	user.Status = models.UserStatusDeactivated
	user.UpdatedBy = &currentUser.ID
	user.UpdatedAt = now
	user.DeactivatedBy = &currentUser.ID
	user.DeactivatedAt = &now
	user.DeactivationReason = nil
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
	saved, err := q.userService.UpdateUser(ctx, tx, user)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateUser")
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

// prepareBuyer loads the active buyer account named by the :id parameter.
func (q *buyerController) prepareBuyer(ctx context.Context, c fiber.Ctx) (*models.User, error) {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, models.ErrUserNotFound
	}
	user, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithRoles(models.RoleBuyer).
			WithStatuses(models.UserStatusActive).
			WithUserIDs(userID),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return nil, err
	}
	if user == nil {
		return nil, models.ErrUserNotFound
	}
	return user, nil
}

// checkReferrer accepts no referrer, or an active buyer other than the
// account itself (buyerID, or uuid.Nil() when creating one).
func (q *buyerController) checkReferrer(ctx context.Context, referrerID *uuid.UUID, buyerID uuid.UUID) (int, error) {
	if referrerID == nil {
		return fiber.StatusOK, nil
	}
	if referrerID.Compare(buyerID) == 0 {
		return fiber.StatusUnprocessableEntity, errors.New("referrer_id: can't refer yourself")
	}
	referrer, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithRoles(models.RoleBuyer).
			WithStatuses(models.UserStatusActive).
			WithUserIDs(*referrerID),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return fiber.StatusUnprocessableEntity, err
	}
	if referrer == nil {
		return fiber.StatusUnprocessableEntity, errors.New("referrer_id: not found")
	}
	return fiber.StatusOK, nil
}
