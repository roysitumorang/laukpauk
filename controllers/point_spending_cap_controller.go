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

type pointSpendingCapController struct {
	pointSpendingCapService services.PointSpendingCapService
	buyerService            services.BuyerService
	jsonWebTokenService     services.JsonWebTokenService
	userService             services.UserService
}

func NewPointSpendingCapController(
	pointSpendingCapService services.PointSpendingCapService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
	userService services.UserService,
) *pointSpendingCapController {
	return &pointSpendingCapController{
		pointSpendingCapService: pointSpendingCapService,
		buyerService:            buyerService,
		jsonWebTokenService:     jsonWebTokenService,
		userService:             userService,
	}
}

func (q *pointSpendingCapController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex)
}

func (q *pointSpendingCapController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-V1SellerIndex"
	request, err := sanitizers.FindPointSpendingCaps(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.pointSpendingCapService.FindPointSpendingCaps(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
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

func (q *pointSpendingCapController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidatePointSpendingCap(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointSpendingCap")
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
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if pointSpendingCap != nil {
		if pointSpendingCap.RecordStatus == models.RecordStatusActive {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrPointSpendingCapShoppingCostOverlapping.Error()).
				WriteResponse(c)
		}
		request.ID = pointSpendingCap.ID
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
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
	saved, err := q.pointSpendingCapService.SavePointSpendingCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSavePointSpendingCap")
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
		SetMessage("point spending cap created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointSpendingCapController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-V1SellerShow"
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointSpendingCapNotFound.Error()).
			WriteResponse(c)
	}
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if pointSpendingCap == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointSpendingCapNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pointSpendingCap).
		WriteResponse(c)
}

func (q *pointSpendingCapController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointSpendingCapNotFound.Error()).
			WriteResponse(c)
	}
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if pointSpendingCap == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointSpendingCapNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidatePointSpendingCap(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointSpendingCap")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if pointSpendingCap.MinimumShoppingCost != request.MinimumShoppingCost ||
		pointSpendingCap.MaximumShoppingCost != request.MaximumShoppingCost {
		otherPointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
			ctx,
			models.NewPointSpendingCapFilter().
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost).
				WithExcludedPointSpendingCapIDs(pointSpendingCap.ID),
		)
		if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherPointSpendingCap != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrPointSpendingCapShoppingCostOverlapping.Error()).
				WriteResponse(c)
		}
	}
	request.ID = pointSpendingCap.ID
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
	saved, err := q.pointSpendingCapService.UpdatePointSpendingCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointSpendingCap")
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
		SetMessage("point spending cap updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointSpendingCapController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointSpendingCapNotFound.Error()).
			WriteResponse(c)
	}
	request, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointSpendingCapNotFound.Error()).
			WriteResponse(c)
	}
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentUser.ID
	request.DeletedAt = new(time.Now())
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
	saved, err := q.pointSpendingCapService.UpdatePointSpendingCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointSpendingCap")
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
		SetMessage("point spending cap deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointSpendingCapController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-V1BuyerIndex"
	response, _, err := q.pointSpendingCapService.FindPointSpendingCaps(
		ctx,
		models.NewPointSpendingCapFilter().
			WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
