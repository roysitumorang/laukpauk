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

type pointController struct {
	pointService        services.PointService
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewPointController(
	pointService services.PointService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *pointController {
	return &pointController{
		pointService:        pointService,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *pointController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
}

func (q *pointController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-V1SellerIndex"
	request, err := sanitizers.FindPoints(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoints")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.pointService.FindPoints(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoints")
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

func (q *pointController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidatePoint(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePoint")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithRoles(models.RoleBuyer).
			WithStatuses(models.UserStatusActive).
			WithUserIDs(request.UserID),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer == nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("user_id: not found").
			WriteResponse(c)
	}
	createdBy := currentUser.ID
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedBy = &createdBy
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
	saved, err := q.pointService.CreatePoint(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePoint")
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
		SetMessage("point created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-V1SellerShow"
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointNotFound.Error()).
			WriteResponse(c)
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if point == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(point).
		WriteResponse(c)
}

func (q *pointController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointNotFound.Error()).
			WriteResponse(c)
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if point == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidatePoint(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePoint")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.ID = point.ID
	request.RecordStatus = point.RecordStatus
	request.PointType = point.PointType
	request.UserID = point.UserID
	request.TransactionID = point.TransactionID
	request.UpdatedBy = &currentUser.ID
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
	saved, err := q.pointService.UpdatePoint(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
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
		SetMessage("point updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *pointController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointNotFound.Error()).
			WriteResponse(c)
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if point == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPointNotFound.Error()).
			WriteResponse(c)
	}
	point.RecordStatus = models.RecordStatusDeleted
	point.DeletedBy = &currentUser.ID
	point.DeletedAt = new(time.Now())
	point.DeletionReason = nil
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
	saved, err := q.pointService.UpdatePoint(ctx, tx, point)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
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
		SetMessage("point deleted successfully").
		SetData(saved).
		WriteResponse(c)
}
