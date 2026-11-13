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

type pointEarningCapController struct {
	pointEarningCapService services.PointEarningCapService
	jsonWebTokenService    services.JsonWebTokenService
	userService            services.UserService
}

func NewPointEarningCapController(
	pointEarningCapService services.PointEarningCapService,
	jsonWebTokenService services.JsonWebTokenService,
	userService services.UserService,
) *pointEarningCapController {
	return &pointEarningCapController{
		pointEarningCapService: pointEarningCapService,
		jsonWebTokenService:    jsonWebTokenService,
		userService:            userService,
	}
}

func (q *pointEarningCapController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
	buyerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleBuyer, q.userService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex)
}

func (q *pointEarningCapController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-V1SellerIndex"
	request, err := sanitizers.FindPointEarningCaps(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.pointEarningCapService.FindPointEarningCaps(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
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

func (q *pointEarningCapController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidatePointEarningCap(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointEarningCap")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	createdBy := currentUser.ID
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	if request.MinimumShoppingCost > 0 && request.MaximumShoppingCost > 0 {
		pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
			ctx,
			models.NewPointEarningCapFilter().
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost),
		)
		if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if pointEarningCap != nil {
			if pointEarningCap.RecordStatus == models.RecordStatusActive {
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(models.ErrPointEarningCapShoppingCostOverlapping.Error()).
					WriteResponse(c)
			}
			request.ID = pointEarningCap.ID
			request.DeletedBy = nil
			request.DeletedAt = nil
			request.DeletionReason = nil
		}
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
	saved, err := q.pointEarningCapService.SavePointEarningCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSavePointEarningCap")
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
		SetMessage("point earning cap created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointEarningCapController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-V1SellerShow"
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointEarningCapNotFound.Error()).
			WriteResponse(c)
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if pointEarningCap == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointEarningCapNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pointEarningCap).
		WriteResponse(c)
}

func (q *pointEarningCapController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointEarningCapNotFound.Error()).
			WriteResponse(c)
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if pointEarningCap == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointEarningCapNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidatePointEarningCap(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointEarningCap")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if (pointEarningCap.MinimumShoppingCost != request.MinimumShoppingCost) &&
		(pointEarningCap.MaximumShoppingCost != request.MaximumShoppingCost) {
		otherPointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
			ctx,
			models.NewPointEarningCapFilter().
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost).
				WithExcludedPointEarningCapIDs(pointEarningCap.ID),
		)
		if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherPointEarningCap != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrPointEarningCapShoppingCostOverlapping.Error()).
				WriteResponse(c)
		}
	}
	request.ID = pointEarningCap.ID
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
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
	saved, err := q.pointEarningCapService.UpdatePointEarningCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
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
		SetMessage("point earning cap updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointEarningCapController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointEarningCapNotFound.Error()).
			WriteResponse(c)
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if pointEarningCap == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointEarningCapNotFound.Error()).
			WriteResponse(c)
	}
	pointEarningCap.RecordStatus = models.RecordStatusDeleted
	pointEarningCap.DeletedBy = &currentUser.ID
	pointEarningCap.DeletedAt = new(time.Now())
	pointEarningCap.DeletionReason = nil
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
	saved, err := q.pointEarningCapService.UpdatePointEarningCap(ctx, tx, pointEarningCap)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
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
		SetMessage("point earning cap deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointEarningCapController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-V1BuyerIndex"
	response, _, err := q.pointEarningCapService.FindPointEarningCaps(
		ctx,
		models.NewPointEarningCapFilter().
			WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
