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
		adminService        services.AdminService
		sellerService       services.SellerService
		messageService      services.MessageService
		notificationService services.NotificationService
	}
)

func NewHomeController(
	adminService services.AdminService,
	sellerService services.SellerService,
	messageService services.MessageService,
	notificationService services.NotificationService,
) *homeController {
	return &homeController{
		adminService:        adminService,
		sellerService:       sellerService,
		messageService:      messageService,
		notificationService: notificationService,
	}
}

func (q *homeController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Get("/privacy-policy", q.PrivacyPolicy)
	r.Group("/admin").
		Get("", adminSessionAuth, q.AdminIndex)
	r.Group("/seller").
		Get("", sellerSessionAuth, q.SellerIndex)
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

func (q *homeController) SellerIndex(c fiber.Ctx) error {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	csrfToken := csrf.TokenFromContext(c)
	return c.Render("seller/home_index", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Dashboard Lauk Pauk",
	), "seller/layouts/authenticated")
}
