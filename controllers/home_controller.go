package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	homeController struct {
		adminService   services.AdminService
		messageService services.MessageService
	}
)

func NewHomeController(
	adminService services.AdminService,
	messageService services.MessageService,
) *homeController {
	return &homeController{
		adminService:   adminService,
		messageService: messageService,
	}
}

func (q *homeController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Get("/privacy-policy", q.PrivacyPolicy)
	r.Group("/admin").
		Get("", adminSessionAuth, q.AdminIndex)
}

func (q *homeController) PrivacyPolicy(c fiber.Ctx) (err error) {
	return c.Render("landing_page/privacy_policy", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"pageTitle", "Kebijakan Privasi | Belanja Sayur Online",
	), "landing_page/_layout")
}

func (q *homeController) AdminIndex(c fiber.Ctx) error {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	return c.Render("admin/home_index", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Dashboard Lauk Pauk",
	), "admin/layouts/authenticated")
}
