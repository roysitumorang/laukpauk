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
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	idempotencyKeyController struct {
		buyerService          services.BuyerService
		idempotencyKeyService services.IdempotencyKeyService
		jsonWebTokenService   services.JsonWebTokenService
	}
)

func NewIdempotencyKeyController(
	userService services.BuyerService,
	idempotencyKeyService services.IdempotencyKeyService,
	jsonWebTokenService services.JsonWebTokenService,
) *idempotencyKeyController {
	return &idempotencyKeyController{
		buyerService:          userService,
		idempotencyKeyService: idempotencyKeyService,
		jsonWebTokenService:   jsonWebTokenService,
	}
}

func (q *idempotencyKeyController) Mount(r fiber.Router) {
	buyerKeyAuth := middleware.BuyerKeyAuth(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Post("", buyerKeyAuth, q.V1BuyerCreate)
}

func (q *idempotencyKeyController) V1BuyerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "IdempotencyKeyController-V1BuyerCreate"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	now := time.Now()
	duration, err := time.ParseDuration(os.Getenv("IDEMPOTENCY_KEY_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request := &models.IdempotencyKey{
		BuyerID:   currentBuyer.ID,
		CreatedAt: now,
		ExpiredAt: now.Add(duration + time.Nanosecond),
	}
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
	response, err := q.idempotencyKeyService.CreateIdempotencyKey(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateIdempotencyKey")
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
