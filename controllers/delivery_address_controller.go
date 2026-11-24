package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type deliveryAddressController struct {
	userService            services.UserService
	deliveryAddressService services.DeliveryAddressService
	settingService         services.SettingService
	jsonWebTokenService    services.JsonWebTokenService
	routesApiClient        *externals.RoutesApiClient
}

func NewDeliveryAddressController(
	userService services.UserService,
	deliveryAddressService services.DeliveryAddressService,
	settingService services.SettingService,
	jsonWebTokenService services.JsonWebTokenService,
	routesApiClient *externals.RoutesApiClient,
) *deliveryAddressController {
	return &deliveryAddressController{
		userService:            userService,
		deliveryAddressService: deliveryAddressService,
		settingService:         settingService,
		jsonWebTokenService:    jsonWebTokenService,
		routesApiClient:        routesApiClient,
	}
}

func (q *deliveryAddressController) Mount(r fiber.Router) {
	buyerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleBuyer, q.userService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex).
		Post("", q.V1BuyerCreate).
		Get("/:id", q.V1BuyerShow).
		Put("/:id", q.V1BuyerUpdate).
		Put("/:id/main", q.V1BuyerSetAsMain).
		Delete("/:id", q.V1BuyerDelete)
}

func (q *deliveryAddressController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryAddressController-V1BuyerIndex"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, err := sanitizers.FindDeliveryAddresses(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryAddresses")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.deliveryAddressService.FindDeliveryAddresses(
		ctx,
		request.WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithSortings(models.Sorting{
				Field: models.DeliveryAddressSortingFieldMainFirst,
			}),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryAddresses")
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

func (q *deliveryAddressController) V1BuyerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryAddressController-V1BuyerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	setting, err := q.settingService.FindSetting(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateDeliveryAddress(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryAddress")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	deliveryAddress, err := q.deliveryAddressService.FindDeliveryAddress(
		ctx,
		models.NewDeliveryAddressFilter().
			WithUserIDs(currentUser.ID).
			WithDestinationLongitudes(request.DestinationLongitude).
			WithDestinationLatitudes(request.DestinationLatitude),
	)
	if errors.Is(err, models.ErrDeliveryAddressNotFound) {
		err = nil
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	if deliveryAddress != nil {
		if deliveryAddress.RecordStatus == models.RecordStatusActive {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage("delivery address with this longitude and latitude already exists").
				WriteResponse(c)
		}
		deliveryAddress.RecordStatus = models.RecordStatusActive
		deliveryAddress.StreetAddress = request.StreetAddress
		deliveryAddress.StreetAddressNote = request.StreetAddressNote
		deliveryAddress.Label = request.Label
		deliveryAddress.RecipientName = request.RecipientName
		deliveryAddress.RecipientMobilePhone = request.RecipientMobilePhone
		deliveryAddress.Main = request.Main
		deliveryAddress.UpdatedAt = now
		deliveryAddress.DeletedAt = nil
		deliveryAddress.DeletionReason = nil
	} else {
		if request.DistanceMeters, err = q.routesApiClient.ComputeDistanceMeters(
			ctx,
			setting.DeliveryOriginLongitude.String(),
			setting.DeliveryOriginLatitude.String(),
			request.DestinationLongitude.String(),
			request.DestinationLatitude.String(),
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrComputeDistanceMeters")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		distanceMeters, err := decimal.New(request.DistanceMeters, 0)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNew")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		distanceKM, err := distanceMeters.Quo(decimal.Thousand)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrQuo")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		request.RecordStatus = models.RecordStatusActive
		request.UserID = currentUser.ID
		request.OriginLongitude = setting.DeliveryOriginLongitude
		request.OriginLatitude = setting.DeliveryOriginLatitude
		request.DistanceKM, _, _ = distanceKM.Ceil(0).Int64(0)
		request.CreatedAt = now
		request.UpdatedAt = now
		if request.DistanceKM > setting.DeliveryMaxDistanceKM {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage("your coordinates is out of our coverage area").
				WriteResponse(c)
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
	if request.Main {
		if err = q.deliveryAddressService.UnsetMain(ctx, tx, currentUser.ID); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	var response *models.DeliveryAddress
	if deliveryAddress != nil {
		if response, err = q.deliveryAddressService.UpdateDeliveryAddress(ctx, tx, deliveryAddress); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryAddress")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	} else if response, err = q.deliveryAddressService.CreateDeliveryAddress(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateDeliveryAddress")
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

func (q *deliveryAddressController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryAddressController-V1BuyerShow"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	deliveryAddressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryAddressNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.deliveryAddressService.FindDeliveryAddress(
		ctx,
		models.NewDeliveryAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryAddressIDs(deliveryAddressID),
	)
	if errors.Is(err, models.ErrDeliveryAddressNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *deliveryAddressController) V1BuyerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryAddressController-V1BuyerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	deliveryAddressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryAddressNotFound.Error()).
			WriteResponse(c)
	}
	deliveryAddress, err := q.deliveryAddressService.FindDeliveryAddress(
		ctx,
		models.NewDeliveryAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryAddressIDs(deliveryAddressID),
	)
	if errors.Is(err, models.ErrDeliveryAddressNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateDeliveryAddress(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryAddress")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	deliveryAddress.StreetAddress = request.StreetAddress
	deliveryAddress.StreetAddressNote = request.StreetAddressNote
	deliveryAddress.Label = request.Label
	deliveryAddress.RecipientName = request.RecipientName
	deliveryAddress.RecipientMobilePhone = request.RecipientMobilePhone
	deliveryAddress.UpdatedAt = time.Now()
	deliveryAddress.DeletedAt = nil
	deliveryAddress.DeletionReason = nil
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
	if request.Main && !deliveryAddress.Main {
		deliveryAddress.Main = request.Main
		if err = q.deliveryAddressService.UnsetMain(ctx, tx, currentUser.ID); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	response, err := q.deliveryAddressService.UpdateDeliveryAddress(ctx, tx, deliveryAddress)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryAddress")
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
		SetMessage("delivery address updated successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *deliveryAddressController) V1BuyerSetAsMain(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryAddressController-V1BuyerSetAsMain"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	deliveryAddressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryAddressNotFound.Error()).
			WriteResponse(c)
	}
	deliveryAddress, err := q.deliveryAddressService.FindDeliveryAddress(
		ctx,
		models.NewDeliveryAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryAddressIDs(deliveryAddressID),
	)
	if errors.Is(err, models.ErrDeliveryAddressNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if deliveryAddress.Main {
		return helper.NewResponse(fiber.StatusOK).
			SetMessage("address set as primary successfully").
			SetData(deliveryAddress).
			WriteResponse(c)
	}
	deliveryAddress.Main = true
	deliveryAddress.UpdatedAt = time.Now()
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
	if err = q.deliveryAddressService.UnsetMain(ctx, tx, currentUser.ID); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response, err := q.deliveryAddressService.UpdateDeliveryAddress(ctx, tx, deliveryAddress)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryAddress")
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
		SetMessage("delivery address set as primary successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *deliveryAddressController) V1BuyerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryAddressController-V1BuyerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	deliveryAddressID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryAddressNotFound.Error()).
			WriteResponse(c)
	}
	deliveryAddress, err := q.deliveryAddressService.FindDeliveryAddress(
		ctx,
		models.NewDeliveryAddressFilter().
			WithUserIDs(currentUser.ID).
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryAddressIDs(deliveryAddressID),
	)
	if errors.Is(err, models.ErrDeliveryAddressNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	deliveryAddress.RecordStatus = models.RecordStatusDeleted
	deliveryAddress.DeletedAt = new(time.Now())
	deliveryAddress.DeletionReason = nil
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
	if deliveryAddress.Main {
		deliveryAddress.Main = false
	}
	if _, err = q.deliveryAddressService.UpdateDeliveryAddress(ctx, tx, deliveryAddress); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryAddress")
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
	return helper.NewResponse(fiber.StatusNoContent).
		WriteResponse(c)
}
