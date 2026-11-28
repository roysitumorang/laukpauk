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
	jsonWebTokenService services.JsonWebTokenService
	textMessageService  services.TextMessageService
	notificationService services.NotificationService
	kafkaClient         *externals.KafkaClient
}

func NewUserController(
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
	textMessageService services.TextMessageService,
	notificationService services.NotificationService,
	kafkaClient *externals.KafkaClient,
) *userController {
	return &userController{
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
		textMessageService:  textMessageService,
		notificationService: notificationService,
		kafkaClient:         kafkaClient,
	}
}

func (q *userController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller").
		Post("/sign_in", q.V1SellerSignIn).
		Get("/me", sellerKeyAuth, q.V1SellerMe)
	buyerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleBuyer, q.userService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Post("", q.V1BuyerSignUp).
		Post("/sign_in", q.V1BuyerSignIn).
		Get("/me", buyerKeyAuth, q.V1BuyerMe).
		Post("/mobile_phone/confirm", buyerKeyAuth, q.V1BuyerConfirmMobilePhone)
}

func (q *userController) V1SellerSignIn(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "UserController-V1SellerSignIn"
	request, statusCode, err := sanitizers.ValidateSignIn(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSignIn")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			JSON(c)
	}
	user, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithRoles(models.RoleSeller).
			WithLogin(request.Login).
			WithStatuses(models.UserStatusActive),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			JSON(c)
	}
	if user == nil || user.EncryptedPassword == nil {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginFailed.Error()).
			JSON(c)
	}
	now := time.Now().In(time.Local)
	matched, err := helper.VerifyPassword(request.Password, *user.EncryptedPassword)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
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
			JSON(c)
	}
	if !matched {
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				JSON(c)
		}
		err = models.ErrLoginFailed
		if response.Status == models.UserStatusLocked {
			err = models.ErrUserLocked
		}
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
	}
	tokenString, err := helper.GenerateAccessToken(user.ID.String(), jsonWebToken.ID.String(), user.MobilePhone, jsonWebToken.CreatedAt, jsonWebToken.ExpiredAt)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrGenerateAccessToken")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetData(models.SignInResponse{
			IDToken:   tokenString,
			ExpiredAt: jsonWebToken.ExpiredAt,
			User:      response,
		}).
		JSON(c)
}

func (q *userController) V1SellerMe(c fiber.Ctx) error {
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	return helper.NewResponse(fiber.StatusOK).
		SetData(currentUser).
		JSON(c)
}

func (q *userController) V1BuyerSignUp(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "UserController-V1BuyerSignUp"
	request, statusCode, err := sanitizers.ValidateSignUp(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSignUp")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			JSON(c)
	}
	user, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithMobilePhone(request.MobilePhone),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if user != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueUserMobilePhoneViolation.Error()).
			JSON(c)
	}
	if user, err = q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithUsername(request.Username),
	); err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if user != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueUserUsernameViolation.Error()).
			JSON(c)
	}
	now := time.Now().In(time.Local)
	encryptedPasword, err := helper.HashPassword(request.Password)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	mobilePhoneConfirmationToken := helper.RandomNumber(6)
	encryptedMobilePhoneConfirmationToken, err := helper.HashPassword(mobilePhoneConfirmationToken)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	user = &models.User{
		Role:                          models.RoleBuyer,
		Status:                        models.UserStatusActive,
		Name:                          request.Name,
		Username:                      &request.Username,
		MobilePhoneConfirmationSentAt: &now,
		MobilePhoneConfirmationToken:  &encryptedMobilePhoneConfirmationToken,
		MobilePhone:                   request.MobilePhone,
		EncryptedPassword:             &encryptedPasword,
		Photo:                         os.Getenv("BUYER_DEFAULT_PHOTO"),
		CreatedAt:                     now,
		UpdatedAt:                     now,
	}
	if request.ReferralCode != "" {
		referrer, err := q.userService.FindUser(
			ctx,
			models.NewUserFilter().
				WithRoles(models.RoleBuyer).
				WithReferralCodes(request.ReferralCode),
		)
		if err != nil && !errors.Is(err, models.ErrUserNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				JSON(c)
		}
		if referrer != nil {
			user.ReferrerID = &referrer.ID
		}
	}
	if err = user.Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
	}
	textMessage, err := q.textMessageService.CreateTextMessage(
		ctx,
		tx,
		&models.TextMessage{
			UserID:      response.ID,
			MobilePhone: response.MobilePhone,
			MessageType: models.TextMessageTypeOtp,
			Body:        helper.Sprintf("Use %s to confirm your mobile number. If you did not sign up, ignore this message.", mobilePhoneConfirmationToken),
		},
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTextMessage")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	go func() {
		if err = q.kafkaClient.Publish(
			ctx,
			models.TopicMessage{
				Topic: models.TopicTextMessage,
				Event: models.TopicTextMessageEventSend,
				ID:    textMessage.ID,
			},
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
		}
	}()
	return helper.NewResponse(fiber.StatusCreated).
		SetData(response).
		JSON(c)
}

func (q *userController) V1BuyerSignIn(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "UserController-V1BuyerSignIn"
	request, statusCode, err := sanitizers.ValidateSignIn(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSignIn")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			JSON(c)
	}
	user, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithRoles(models.RoleBuyer).
			WithLogin(request.Login).
			WithStatuses(models.UserStatusActive),
	)
	if errors.Is(err, models.ErrUserNotFound) {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrLoginFailed.Error()).
			JSON(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if user.EncryptedPassword == nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrLoginFailed.Error()).
			JSON(c)
	}
	now := time.Now().In(time.Local)
	matched, err := helper.VerifyPassword(request.Password, *user.EncryptedPassword)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
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
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if !matched {
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				JSON(c)
		}
		err = models.ErrLoginFailed
		if response.Status == models.UserStatusLocked {
			err = models.ErrUserLocked
		}
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
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
			JSON(c)
	}
	tokenString, err := helper.GenerateAccessToken(user.ID.String(), jsonWebToken.ID.String(), user.MobilePhone, jsonWebToken.CreatedAt, jsonWebToken.ExpiredAt)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrGenerateAccessToken")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetData(models.SignInResponse{
			IDToken:   tokenString,
			ExpiredAt: jsonWebToken.ExpiredAt,
			User:      response,
		}).
		JSON(c)
}

func (q *userController) V1BuyerMe(c fiber.Ctx) error {
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	return helper.NewResponse(fiber.StatusOK).
		SetData(currentUser).
		JSON(c)
}

func (q *userController) V1BuyerConfirmMobilePhone(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "UserController-V1BuyerConfirmMobilePhone"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ConfirmMobilePhone(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrConfirmMobilePhone")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			JSON(c)
	}
	if currentUser.MobilePhoneConfirmationToken == nil {
		return helper.NewResponse(fiber.StatusOK).
			SetMessage("mobile phone already confirmed").
			JSON(c)
	}
	matched, err := helper.VerifyPassword(request.Token, *currentUser.MobilePhoneConfirmationToken)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if !matched {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("mobile phone confirmation failed").
			JSON(c)
	}
	currentUser.MobilePhoneConfirmationToken = nil
	currentUser.MobilePhoneConfirmedAt = new(time.Now())
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
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
		currentUser,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		JSON(c)
}
