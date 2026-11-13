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
	deliveryPlanController struct {
		deliveryPlanService services.DeliveryPlanService
		buyerService        services.BuyerService
		jsonWebTokenService services.JsonWebTokenService
		userService         services.UserService
	}
)

func NewDeliveryPlanController(
	deliveryPlanService services.DeliveryPlanService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
	userService services.UserService,
) *deliveryPlanController {
	return &deliveryPlanController{
		deliveryPlanService: deliveryPlanService,
		buyerService:        buyerService,
		jsonWebTokenService: jsonWebTokenService,
		userService:         userService,
	}
}

func (q *deliveryPlanController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("/now", q.V1BuyerShow)
}

func (q *deliveryPlanController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-V1SellerIndex"
	pagination := new(models.Pagination)
	request, err := sanitizers.FindDeliveryPlans(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.deliveryPlanService.FindDeliveryPlans(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	pagination.Rows = make([]any, len(rows))
	for i := range rows {
		pagination.Rows[i] = rows[i]
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}

func (q *deliveryPlanController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidateDeliveryPlan(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryPlan")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithDeliveryWeekdays(request.Weekday).
			WithTransactionWindow(request.TransactionWindowMin, request.TransactionWindowMax),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryPlan != nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrUniqueDeliveryPlanTransactionWindowViolation.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	createdBy := currentUser.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
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
	saved, err := q.deliveryPlanService.CreateDeliveryPlan(ctx, tx, request)
	if err != nil {
		if !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateDeliveryPlan")
		}
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
		SetMessage("delivery plan created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *deliveryPlanController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-V1SellerShow"
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryPlan == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(deliveryPlan).
		WriteResponse(c)
}

func (q *deliveryPlanController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryPlan == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateDeliveryPlan(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryPlan")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryPlan.Weekday != request.Weekday ||
		deliveryPlan.TransactionWindowMin != request.TransactionWindowMin ||
		deliveryPlan.TransactionWindowMax != request.TransactionWindowMax {
		otherDeliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
			ctx,
			models.NewDeliveryPlanFilter().
				WithDeliveryWeekdays(request.Weekday).
				WithTransactionWindow(request.TransactionWindowMin, request.TransactionWindowMax).
				WithExcludedDeliveryPlanIDs(deliveryPlan.ID),
		)
		if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherDeliveryPlan != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueDeliveryPlanTransactionWindowViolation.Error()).
				WriteResponse(c)
		}
	}
	request.ID = deliveryPlan.ID
	request.RecordStatus = deliveryPlan.RecordStatus
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
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
	saved, err := q.deliveryPlanService.UpdateDeliveryPlan(ctx, tx, request)
	if err != nil {
		if !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
		}
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
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("delivery plan updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *deliveryPlanController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryPlan == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	deliveryPlan.RecordStatus = models.RecordStatusDeleted
	deliveryPlan.DeletedBy = &currentUser.ID
	deliveryPlan.DeletedAt = new(time.Now())
	deliveryPlan.DeletionReason = nil
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
	saved, err := q.deliveryPlanService.UpdateDeliveryPlan(ctx, tx, deliveryPlan)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
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
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("delivery plan deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *deliveryPlanController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-V1BuyerShow"
	now := time.Now().In(time.Local)
	response, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithTransactionDatetime(now),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
