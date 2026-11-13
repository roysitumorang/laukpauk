package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

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
	notificationTargetController struct {
		adminService              services.AdminService
		buyerService              services.BuyerService
		notificationTargetService services.NotificationTargetService
		jsonWebTokenService       services.JsonWebTokenService
		userService               services.UserService
	}
)

func NewNotificationTargetController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	notificationTargetService services.NotificationTargetService,
	jsonWebTokenService services.JsonWebTokenService,
	userService services.UserService,
) *notificationTargetController {
	return &notificationTargetController{
		adminService:              adminService,
		buyerService:              buyerService,
		notificationTargetService: notificationTargetService,
		jsonWebTokenService:       jsonWebTokenService,
		userService:               userService,
	}
}

func (q *notificationTargetController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/admin", adminSessionAuth).
		Post("", q.AdminCreate)
	r.Group("/v1/seller", sellerKeyAuth).
		Post("", q.V1SellerCreate)
	r.Group("/v1/buyer", buyerKeyAuth).
		Post("", q.V1BuyerSave)
}

func (q *notificationTargetController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationTargetController-AdminCreate"
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	request, statusCode, err := sanitizers.ValidateNotificationTarget(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNotificationTarget")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.AdminID = currentAdmin.ID
	request.SellerID = uuid.Nil()
	request.BuyerID = uuid.Nil()
	request.CreatedAt = now
	request.UpdatedAt = now
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
	response, err := q.notificationTargetService.SaveNotificationTarget(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveNotificationTarget")
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

func (q *notificationTargetController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationTargetController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidateNotificationTarget(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNotificationTarget")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.AdminID = uuid.Nil()
	request.SellerID = currentUser.ID
	request.BuyerID = uuid.Nil()
	request.CreatedAt = now
	request.UpdatedAt = now
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
	response, err := q.notificationTargetService.SaveNotificationTarget(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveNotificationTarget")
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

func (q *notificationTargetController) V1BuyerSave(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationTargetController-V1BuyerSave"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateNotificationTarget(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNotificationTarget")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.AdminID = uuid.Nil()
	request.SellerID = uuid.Nil()
	request.BuyerID = currentBuyer.ID
	request.CreatedAt = now
	request.UpdatedAt = now
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
	response, err := q.notificationTargetService.SaveNotificationTarget(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveNotificationTarget")
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
