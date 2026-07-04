package controllers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	authController struct {
		userService services.UserService
	}
)

func NewAuthController(
	userService services.UserService,
) *authController {
	return new(authController{
		userService: userService,
	})
}

func (q *authController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	admin := r.Group("/admin")
	admin.Get("/login", q.AdminNewLogin).
		Post("/login", q.AdminLogin).
		Get("/me", adminSessionAuth, q.AdminProfile).
		Get("/logout", q.AdminLogout)
}

func (q *authController) AdminNewLogin(c fiber.Ctx) error {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	if sess.Get(models.UserID) != nil {
		return flash.Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.AdminLoginRequest)
	return c.Render("admin/auth_login", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", request,
	))
}

func (q *authController) AdminLogin(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-AdminLogin"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	if sess.Get(models.UserID) != nil {
		return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", new(models.AdminLoginRequest),
	)
	request, statusCode, err := sanitizers.ValidateAdminLogin(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateAdminLogin")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	admin, err := q.userService.FindAdmin(
		ctx,
		models.NewUserFilter().
			WithEmail(request.Email),
	)
	if err != nil {
		if !errors.Is(err, models.ErrAdminNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
		}
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if admin == nil ||
		admin.EncryptedPassword == nil {
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	encryptedPassword := helper.String2ByteSlice(*admin.EncryptedPassword)
	if !helper.MatchedHashAndPassword(encryptedPassword, helper.String2ByteSlice(request.Password)) {
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	if err = sess.RegenerateWithContext(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRegenerate")
	}
	sess.Set(models.UserID, admin.ID)
	next, ok := sess.Get("next").(string)
	if !ok {
		next = "/admin"
	} else {
		sess.Delete("next")
	}
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, next)
}

func (q *authController) AdminProfile(c fiber.Ctx) error {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/admin/auth/login")
}
