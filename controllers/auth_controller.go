package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
	"go.uber.org/zap"
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
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	if sess.Get(models.UserID) != nil {
		return flash.Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.AdminLoginRequest)
	return c.Render("admin/auth_login", models.NewMap().Push(
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"request", request,
	))
}

func (q *authController) AdminLogin(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-AdminLogin"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	if sess.Get(models.UserID) != nil {
		return flash.Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"request", new(models.AdminLoginRequest),
	)
	request, statusCode, err := sanitizers.ValidateAdminLogin(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidateAdminLogin")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	admins, _, err := q.userService.FindAdmins(
		ctx,
		models.NewUserFilter().
			WithEmail(request.Email),
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindAdmins")
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if len(admins) == 0 ||
		admins[0].EncryptedPassword == nil {
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	admin := admins[0]
	encryptedPassword := helper.String2ByteSlice(*admin.EncryptedPassword)
	if !helper.MatchedHashAndPassword(encryptedPassword, helper.String2ByteSlice(request.Password)) {
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	if err = sess.Regenerate(); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRegenerate")
	}
	sess.Set(models.UserID, admin.ID)
	return flash.Redirect(c, sess.Session, "/admin")
}

func (q *authController) AdminProfile(c fiber.Ctx) error {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
	)
	return c.Render("admin/auth_me", params)
}

func (q *authController) AdminLogout(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-AdminLogout"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	if csrfHandler := csrf.HandlerFromContext(c); csrfHandler != nil {
		if err := csrfHandler.DeleteToken(c); err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrDeleteToken")
		}
	}
	if err := sess.Reset(); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrReset")
	}
	return flash.Redirect(c, sess.Session, "/admin/auth/login")
}
