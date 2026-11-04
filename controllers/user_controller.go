package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type userController struct {
	userService         services.UserService
	companyService      services.CompanyService
	jsonWebTokenService services.JsonWebTokenService
	textMessageService  services.TextMessageService
	notificationService services.NotificationService
	kafkaClient         *externals.KafkaClient
}

func NewUserController(
	userService services.UserService,
	companyService services.CompanyService,
	jsonWebTokenService services.JsonWebTokenService,
	textMessageService services.TextMessageService,
	notificationService services.NotificationService,
	kafkaClient *externals.KafkaClient,
) *userController {
	return &userController{
		userService:         userService,
		companyService:      companyService,
		jsonWebTokenService: jsonWebTokenService,
		textMessageService:  textMessageService,
		notificationService: notificationService,
		kafkaClient:         kafkaClient,
	}
}

func (q *userController) Mount(r fiber.Router) {
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	keyAuth := middlewares.KeyAuthMiddleware(q.userService, q.jsonWebTokenService)
	r.Group("/v1", companyVerify).
		Post("", q.V1SignUp).
		Post("/sign_in", q.V1SignIn).
		Get("/me", keyAuth, q.V1Me)
}

func (q *userController) V1SignUp(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "UserController-V1SignUp"
	company, _ := c.Locals(models.CurrentCompany).(*models.Company)
	request, statusCode, err := sanitizers.ValidateSignUp(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSignUp")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	user, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithCompanyIDs(company.ID).
			WithMobilePhone(request.MobilePhone),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if user != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueUserMobilePhoneViolation.Error()).
			WriteResponse(c)
	}
	if user, err = q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithCompanyIDs(company.ID).
			WithUsername(request.Username),
	); err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if user != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueUserUsernameViolation.Error()).
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
	user = &models.User{
		Role:                          models.RoleBuyer,
		Status:                        models.UserStatusActive,
		Name:                          request.Name,
		Username:                      &request.Username,
		MobilePhoneConfirmationSentAt: &now,
		MobilePhoneConfirmationToken:  new(helper.RandomNumber(6)),
		MobilePhone:                   request.MobilePhone,
		EncryptedPassword:             &encryptedPasword,
		Photo:                         os.Getenv("BUYER_DEFAULT_PHOTO"),
		CompanyID:                     &company.ID,
		CreatedAt:                     now,
		UpdatedAt:                     now,
	}
	if request.ReferralCode != "" {
		referrer, err := q.userService.FindUser(
			ctx,
			models.NewUserFilter().
				WithCompanyIDs(company.ID).
				WithRoles(models.RoleBuyer).
				WithReferralCodes(request.ReferralCode),
		)
		if err != nil && !errors.Is(err, models.ErrUserNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if referrer != nil {
			user.ReferrerID = &referrer.ID
		}
	}
	if err = user.Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
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
	response, err := q.userService.CreateUser(ctx, tx, user)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateUser")
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

func (q *userController) V1SignIn(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "UserController-V1SignIn"
	company, _ := c.Locals(models.CurrentCompany).(*models.Company)
	request, statusCode, err := sanitizers.ValidateSignIn(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateOtpLogin")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	user, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithCompanyIDs(company.ID).
			WithLogin(request.Login).
			WithStatuses(models.UserStatusActive),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if user == nil || user.EncryptedPassword == nil {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	matched, err := helper.VerifyPassword(request.Password, *user.EncryptedPassword)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if !matched {
		user.FailedAttempts++
		loginMaxFailedAttempts, _ := strconv.Atoi(os.Getenv("LOGIN_MAX_FAILED_ATTEMPTS"))
		if user.FailedAttempts >= loginMaxFailedAttempts {
			user.Status = models.UserStatusLocked
			user.UnlockToken = new(helper.RandomNumber(6))
			user.LockedAt = &now
		}
	} else {
		user.SignInCount++
		if user.CurrentSignInAt != nil {
			user.LastSignInAt = new(*user.CurrentSignInAt)
		}
		if user.CurrentSignInIP != nil {
			user.LastSignInIP = new(*user.CurrentSignInIP)
		}
		user.CurrentSignInAt = &now
		user.CurrentSignInIP = new(c.IP())
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
	response, err := q.userService.UpdateUser(
		ctx,
		tx,
		user,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateUser")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if !matched {
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		err = models.ErrLoginFailed
		if response.Status == models.UserStatusLocked {
			err = models.ErrUserLocked
		}
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	expiredAt := now.Add(helper.GetAccessTokenAge()).Add(time.Nanosecond)
	jsonWebTokenRequest := &models.JsonWebToken{
		UserID:    user.ID,
		CreatedAt: now,
		ExpiredAt: expiredAt,
	}
	jsonWebToken, err := q.jsonWebTokenService.CreateJsonWebToken(ctx, tx, jsonWebTokenRequest)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateJsonWebToken")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	tokenString, err := helper.GenerateAccessToken(user.ID.String(), jsonWebToken.ID.String(), user.MobilePhone, jsonWebToken.CreatedAt, jsonWebToken.ExpiredAt)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrGenerateAccessToken")
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
		SetData(models.SignInResponse{
			IDToken:   tokenString,
			ExpiredAt: jsonWebToken.ExpiredAt,
			User:      response,
		}).
		WriteResponse(c)
}

func (q *userController) V1Me(c fiber.Ctx) error {
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	return helper.NewResponse(fiber.StatusOK).
		SetData(currentUser).
		WriteResponse(c)
}
