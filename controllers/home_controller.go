package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	homeController struct {
		messageService services.MessageService
	}
)

func NewHomeController(
	messageService services.MessageService,
) *homeController {
	return &homeController{
		messageService: messageService,
	}
}

func (q *homeController) Mount(r fiber.Router) {
	r.Get("/privacy-policy", q.PrivacyPolicy)
}

func (q *homeController) PrivacyPolicy(c fiber.Ctx) (err error) {
	return c.Render("landing_page/privacy_policy", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"pageTitle", "Kebijakan Privasi | Belanja Sayur Online",
	), "landing_page/_layout")
}
