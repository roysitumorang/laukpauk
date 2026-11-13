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

// sellerController manages seller accounts: users with role seller. Only an
// owner can use it; the accounts it creates are staff.
type sellerController struct {
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewSellerController(
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *sellerController {
	return &sellerController{
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
	request, err := sanitizers.FindSellerStaff(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSellerStaff")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.userService.FindUsers(
		ctx,
		request.
			WithRoles(models.RoleSeller).
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

func (q *sellerController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateSellerStaff(ctx, c, true)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSellerStaff")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if statusCode, err = q.checkUniqueLogins(ctx, request, uuid.Nil()); err != nil {
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	encryptedPassword, err := helper.HashPassword(request.Password)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	createdBy := currentUser.ID
	user := &models.User{
		Role:              models.RoleSeller,
		Status:            models.UserStatusActive,
		Name:              request.Name,
		MobilePhone:       request.MobilePhone,
		Username:          &request.Username,
		Email:             request.Email,
		EncryptedPassword: &encryptedPassword,
		PasswordChangedAt: &now,
		Photo:             request.Photo,
		UserLevel:         models.UserLevelSellerStaff,
		CreatedBy:         &createdBy,
		CreatedAt:         now,
		UpdatedBy:         &createdBy,
		UpdatedAt:         now,
	}
	if photo == nil {
		user.Photo = os.Getenv("SELLER_DEFAULT_PHOTO")
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
		SetMessage("seller created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *sellerController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerShow"
	user, err := q.prepareSeller(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareSeller")
		return helper.NewResponse(sellerStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(user).
		WriteResponse(c)
}

func (q *sellerController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	user, err := q.prepareSeller(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareSeller")
		return helper.NewResponse(sellerStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateSellerStaff(ctx, c, false)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSellerStaff")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if statusCode, err = q.checkUniqueLogins(ctx, request, user.ID); err != nil {
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
		SetMessage("seller updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *sellerController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	user, err := q.prepareSeller(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareSeller")
		return helper.NewResponse(sellerStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if user.ID.Compare(currentUser.ID) == 0 {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUserSelfDeletingProhibited.Error()).
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
		SetMessage("seller deactivated successfully").
		SetData(saved).
		WriteResponse(c)
}

// prepareSeller loads the active seller account named by the :id parameter.
func (q *sellerController) prepareSeller(ctx context.Context, c fiber.Ctx) (*models.User, error) {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, models.ErrUserNotFound
	}
	user, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithRoles(models.RoleSeller).
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

// checkUniqueLogins rejects a username, mobile phone or email that another
// user (seller or buyer) already has. excludedUserID is the account being
// updated, or uuid.Nil() when creating one.
func (q *sellerController) checkUniqueLogins(
	ctx context.Context,
	request *models.SellerStaff,
	excludedUserID uuid.UUID,
) (int, error) {
	checks := []struct {
		filter *models.UserFilter
		err    error
	}{
		{models.NewUserFilter().WithUsername(request.Username), models.ErrUniqueUserUsernameViolation},
		{models.NewUserFilter().WithMobilePhone(request.MobilePhone), models.ErrUniqueUserMobilePhoneViolation},
	}
	if request.Email != nil {
		checks = append(checks, struct {
			filter *models.UserFilter
			err    error
		}{models.NewUserFilter().WithEmail(*request.Email), models.ErrUniqueUserEmailViolation})
	}
	for _, check := range checks {
		if excludedUserID.Compare(uuid.Nil()) != 0 {
			check.filter.WithExcludedUserIDs(excludedUserID)
		}
		other, err := q.userService.FindUser(ctx, check.filter)
		if err != nil && !errors.Is(err, models.ErrUserNotFound) {
			return fiber.StatusUnprocessableEntity, err
		}
		if other != nil {
			return fiber.StatusUnprocessableEntity, check.err
		}
	}
	return fiber.StatusOK, nil
}

func sellerStatusCode(err error) int {
	if errors.Is(err, models.ErrUserNotFound) {
		return fiber.StatusNotFound
	}
	return fiber.StatusUnprocessableEntity
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
