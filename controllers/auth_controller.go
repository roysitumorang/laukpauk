package controllers

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type authController struct {
	adminService        services.AdminService
	buyerService        services.BuyerService
	jsonWebTokenService services.JsonWebTokenService
	textMessageService  services.TextMessageService
	kafkaClient         *externals.KafkaClient
}

func NewAuthController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
	textMessageService services.TextMessageService,
	kafkaClient *externals.KafkaClient,
) *authController {
	return &authController{
		adminService:        adminService,
		buyerService:        buyerService,
		jsonWebTokenService: jsonWebTokenService,
		textMessageService:  textMessageService,
		kafkaClient:         kafkaClient,
	}
}

func (q *authController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin").
		Get("/sessions/new", q.AdminNew).
		Post("/sessions", q.AdminCreate).
		Get("/me", adminSessionAuth, q.AdminShow).
		Get("/logout", q.AdminLogout)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Post("/login", q.V1BuyerCreate).
		Get("/me", buyerKeyAuth, q.V1BuyerShow)
}

func (q *authController) AdminNew(c fiber.Ctx) error {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	if sess.Get(models.AdminID) != nil {
		return flash.Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LoginRequest)
	return c.Render("admin/auth_session_new", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", request,
	))
}

func (q *authController) AdminCreate(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-AdminCreate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	if sess.Get(models.AdminID) != nil {
		return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", new(models.LoginRequest),
	)
	request, statusCode, err := sanitizers.ValidateLogin(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLogin")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/auth_session_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	admin, err := q.adminService.FindAdmin(
		ctx,
		models.NewAdminFilter().
			WithLogin(request.Login).
			WithStatuses(models.AdminStatusActive),
	)
	if err != nil {
		if !errors.Is(err, models.ErrAdminNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
		}
		return c.Render("admin/auth_session_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if admin == nil ||
		admin.EncryptedPassword == nil {
		return c.Render("admin/auth_session_new", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	matched, err := helper.VerifyPassword(request.Password, *admin.EncryptedPassword)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return c.Render("admin/auth_session_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if !matched {
		return c.Render("admin/auth_session_new", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	if err = sess.RegenerateWithContext(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRegenerate")
	}
	sess.Set(models.AdminID, admin.ID)
	next, ok := sess.Get("next").(string)
	if !ok || !strings.HasPrefix(next, "/admin") {
		next = "/admin"
	} else {
		sess.Delete("next")
	}
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, next)
}

func (q *authController) AdminShow(c fiber.Ctx) error {
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
	)
	return c.Render("admin/auth_me", params)
}

func (q *authController) AdminLogout(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-AdminLogout"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	if csrfHandler := csrf.HandlerFromContext(c); csrfHandler != nil {
		if err := csrfHandler.DeleteToken(c); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteToken")
		}
	}
	if err := sess.ResetWithContext(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrReset")
	}
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/admin/auth/sessions/new")
}

func (q *authController) V1BuyerCreate(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-V1BuyerCreate"
	loginRequest, statusCode, err := sanitizers.ValidateOtpLogin(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateOtpLogin")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithLogin(loginRequest.Login).
			WithStatuses(models.BuyerStatusActive),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer == nil || buyer.EncryptedPassword == nil {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
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
	expiredAt := now.Add(helper.GetAccessTokenAge()).Add(time.Nanosecond)
	jsonWebTokenRequest := &models.JsonWebToken{
		UserID:    buyer.ID,
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
	tokenString, err := helper.GenerateAccessToken(buyer.ID.String(), jsonWebToken.ID.String(), buyer.Username, jsonWebToken.CreatedAt, jsonWebToken.ExpiredAt)
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
	response := models.BuyerLoginResponse{
		IDToken:   tokenString,
		ExpiredAt: jsonWebToken.ExpiredAt,
		Buyer:     buyer,
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetData(response).
		WriteResponse(c)
}

func (q *authController) V1BuyerShow(c fiber.Ctx) error {
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	return helper.NewResponse(fiber.StatusOK).
		SetData(currentBuyer).
		WriteResponse(c)
}
