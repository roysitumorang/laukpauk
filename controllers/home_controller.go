package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	homeController struct {
		adminService services.AdminService
	}
)

func NewHomeController(
	adminService services.AdminService,
) *homeController {
	return new(homeController{
		adminService: adminService,
	})
}

func (q *homeController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	admin := r.Group("/admin")
	admin.Get("", adminSessionAuth, q.AdminIndex)
}

func (q *homeController) AdminIndex(c fiber.Ctx) error {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
