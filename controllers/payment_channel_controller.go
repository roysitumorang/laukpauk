package controllers

import (
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	paymentChannelController struct {
		userService           services.UserService
		paymentChannelService services.PaymentChannelService
		jsonWebTokenService   services.JsonWebTokenService
	}
)

func NewPaymentChannelController(
	userService services.UserService,
	paymentChannelService services.PaymentChannelService,
	jsonWebTokenService services.JsonWebTokenService,
) *paymentChannelController {
	return &paymentChannelController{
		userService:           userService,
		paymentChannelService: paymentChannelService,
		jsonWebTokenService:   jsonWebTokenService,
	}
}

func (q *paymentChannelController) Mount(r fiber.Router) {
	buyerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleBuyer, q.userService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex)
}

func (q *paymentChannelController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PaymentChannelController-V1BuyerIndex"
	amount, _ := strconv.ParseInt(c.Query("amount"), 10, 64)
	response, _, err := q.paymentChannelService.FindPaymentChannels(
		c,
		models.NewPaymentChannelFilter().
			WithAmount(max(0, amount)),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPaymentChannels")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			JSON(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		JSON(c)
}
