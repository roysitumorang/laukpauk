package controllers

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	paymentChannelController struct {
		buyerService          services.BuyerService
		paymentChannelService services.PaymentChannelService
		jsonWebTokenService   services.JsonWebTokenService
	}
)

func NewPaymentChannelController(
	buyerService services.BuyerService,
	paymentChannelService services.PaymentChannelService,
	jsonWebTokenService services.JsonWebTokenService,
) *paymentChannelController {
	return &paymentChannelController{
		buyerService:          buyerService,
		paymentChannelService: paymentChannelService,
		jsonWebTokenService:   jsonWebTokenService,
	}
}

func (q *paymentChannelController) Mount(r fiber.Router) {
	buyerKeyAuth := middleware.BuyerKeyAuth(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Get("", buyerKeyAuth, q.V1BuyerIndex)
}

func (q *paymentChannelController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PaymentChannelController-V1BuyerIndex"
	request, err := sanitizers.FindPaymentChannels(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPaymentChannels")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.paymentChannelService.FindPaymentChannels(c, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPaymentChannels")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	pagination.Rows = make([]any, len(rows))
	for i, row := range rows {
		pagination.Rows[i] = row
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}
