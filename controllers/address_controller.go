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

type addressController struct {
	addressService      services.AddressService
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewAddressController(
	addressService services.AddressService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *addressController {
	return &addressController{
		addressService:      addressService,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *addressController) Mount(r fiber.Router) {
	buyerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleBuyer, q.userService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex).
		Post("", q.V1BuyerCreate).
		Get("/:id", q.V1BuyerShow).
		Post("/:id", q.V1BuyerUpdate).
		Post("/:id/main", q.V1BuyerSetAsMain).
		Post("/:id/delete", q.V1BuyerDelete)
}

func (q *addressController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AddressController-V1BuyerIndex"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, err := sanitizers.FindAddresses(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddresses")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.addressService.FindAddresses(
		ctx,
		request.WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddresses")
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

func (q *addressController) V1BuyerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AddressController-V1BuyerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, statusCode, err := sanitizers.ValidateAddress(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateAddress")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.UserID = currentUser.ID
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
	if request.Main {
		if err = q.addressService.UnsetMain(ctx, tx, currentUser.ID); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	response, err := q.addressService.CreateAddress(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateAddress")
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
		SetMessage("address created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *addressController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AddressController-V1BuyerShow"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	addressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	address, err := q.addressService.FindAddress(
		ctx,
		models.NewAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithAddressIDs(addressID),
	)
	if err != nil && !errors.Is(err, models.ErrAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if address == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(address).
		WriteResponse(c)
}

func (q *addressController) V1BuyerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AddressController-V1BuyerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	addressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	address, err := q.addressService.FindAddress(
		ctx,
		models.NewAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithAddressIDs(addressID),
	)
	if err != nil && !errors.Is(err, models.ErrAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if address == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateAddress(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateAddress")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.ID = address.ID
	request.RecordStatus = address.RecordStatus
	request.UserID = currentUser.ID
	request.CreatedAt = address.CreatedAt
	request.UpdatedAt = time.Now()
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
	if request.Main && !address.Main {
		if err = q.addressService.UnsetMain(ctx, tx, currentUser.ID); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	response, err := q.addressService.UpdateAddress(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateAddress")
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
		SetMessage("address updated successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *addressController) V1BuyerSetAsMain(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AddressController-V1BuyerSetAsMain"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	addressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	address, err := q.addressService.FindAddress(
		ctx,
		models.NewAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithAddressIDs(addressID),
	)
	if err != nil && !errors.Is(err, models.ErrAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if address == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	if address.Main {
		return helper.NewResponse(fiber.StatusOK).
			SetMessage("address set as primary successfully").
			SetData(address).
			WriteResponse(c)
	}
	address.Main = true
	address.UpdatedAt = time.Now()
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
	if err = q.addressService.UnsetMain(ctx, tx, currentUser.ID); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response, err := q.addressService.UpdateAddress(ctx, tx, address)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateAddress")
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
		SetMessage("address set as primary successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *addressController) V1BuyerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AddressController-V1BuyerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	addressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	address, err := q.addressService.FindAddress(
		ctx,
		models.NewAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithAddressIDs(addressID),
	)
	if err != nil && !errors.Is(err, models.ErrAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if address == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrAddressNotFound.Error()).
			WriteResponse(c)
	}
	address.RecordStatus = models.RecordStatusDeleted
	address.DeletedAt = new(time.Now())
	address.DeletionReason = nil
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
	if address.Main {
		address.Main = false
	}
	response, err := q.addressService.UpdateAddress(ctx, tx, address)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateAddress")
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
		SetMessage("address deleted successfully").
		SetData(response).
		WriteResponse(c)
}
