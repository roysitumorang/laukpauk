package controllers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	settingController struct {
		adminService        services.AdminService
		buyerService        services.BuyerService
		settingService      services.SettingService
		jsonWebTokenService services.JsonWebTokenService
	}
)

func NewSettingController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	settingService services.SettingService,
	jsonWebTokenService services.JsonWebTokenService,
) *settingController {
	return &settingController{
		adminService:        adminService,
		buyerService:        buyerService,
		settingService:      settingService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *settingController) Mount(r fiber.Router) {
	buyerKeyAuth := middleware.BuyerKeyAuth(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerShow)
}

func (q *settingController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CartController-V1BuyerShow"
	response, err := q.settingService.FindSetting(ctx)
	if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCart")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrSettingNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
