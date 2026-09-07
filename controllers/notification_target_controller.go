package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
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
		sellerService             services.SellerService
		buyerService              services.BuyerService
		notificationTargetService services.NotificationTargetService
		jsonWebTokenService       services.JsonWebTokenService
	}
)

func NewNotificationTargetController(
	adminService services.AdminService,
	sellerService services.SellerService,
	userService services.BuyerService,
	notificationTargetService services.NotificationTargetService,
	jsonWebTokenService services.JsonWebTokenService,
) *notificationTargetController {
	return &notificationTargetController{
		adminService:              adminService,
		sellerService:             sellerService,
		buyerService:              userService,
		notificationTargetService: notificationTargetService,
		jsonWebTokenService:       jsonWebTokenService,
	}
}

func (q *notificationTargetController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/admin", adminSessionAuth).
		Post("", q.AdminCreate)
	r.Group("/seller", sellerSessionAuth).
		Post("", q.SellerCreate)
	r.Group("/v1/buyer").
		Post("", buyerKeyAuth, q.V1BuyerSave)
}

func (q *notificationTargetController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationTargetController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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

func (q *notificationTargetController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationTargetController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	request, statusCode, err := sanitizers.ValidateNotificationTarget(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNotificationTarget")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.AdminID = uuid.Nil()
	request.SellerID = currentSeller.ID
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
