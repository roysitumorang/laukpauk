package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	fcmRegistrationController struct {
		buyerService           services.BuyerService
		fcmRegistrationService services.FcmRegistrationService
		jsonWebTokenService    services.JsonWebTokenService
	}
)

func NewFcmRegistrationController(
	userService services.BuyerService,
	fcmRegistrationService services.FcmRegistrationService,
	jsonWebTokenService services.JsonWebTokenService,
) *fcmRegistrationController {
	return &fcmRegistrationController{
		buyerService:           userService,
		fcmRegistrationService: fcmRegistrationService,
		jsonWebTokenService:    jsonWebTokenService,
	}
}

func (q *fcmRegistrationController) Mount(r fiber.Router) {
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Post("", buyerKeyAuth, q.V1BuyerSave)
}

func (q *fcmRegistrationController) V1BuyerSave(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "FcmRegistrationController-V1BuyerSave"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateFcmRegistration(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateFcmRegistration")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	duration, err := time.ParseDuration(os.Getenv("FCM_TOKEN_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.BuyerID = currentBuyer.ID
	request.CreatedAt = now
	request.UpdatedAt = now
	request.ExpiredAt = now.Add(duration + time.Nanosecond)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.fcmRegistrationService.SaveFcmRegistration(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFcmRegistration")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetData(response).
		WriteResponse(c)
}
