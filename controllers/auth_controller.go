package controllers

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type authController struct {
	adminService services.AdminService
}

func NewAuthController(
	adminService services.AdminService,
) *authController {
	return &authController{
		adminService: adminService,
	}
}

func (q *authController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin").
		Get("/sessions/new", q.AdminNew).
		Post("/sessions", q.AdminCreate).
		Get("/me", adminSessionAuth, q.AdminShow).
		Get("/logout", q.AdminLogout)
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
