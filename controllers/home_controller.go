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
		userService services.UserService
	}
)

func NewHomeController(
	userService services.UserService,
) *homeController {
	return new(homeController{
		userService: userService,
	})
}

func (q *homeController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	admin := r.Group("/admin")
	admin.Get("", adminSessionAuth, q.AdminIndex)
}

func (q *homeController) AdminIndex(c fiber.Ctx) error {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	return c.Render("admin/home_index", models.NewMap().Push(
		"baseURL", c.Req().BaseURL(),
		"flash", flash,
		"path", c.Route().Path,
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
	), "admin/layouts/authenticated")
}
