package controllers

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	deliveryRouteController struct {
		adminService         services.AdminService
		buyerService         services.BuyerService
		deliveryRouteService services.DeliveryRouteService
		settingService       services.SettingService
		jsonWebTokenService  services.JsonWebTokenService
	}
)

func NewDeliveryRouteController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	deliveryRouteService services.DeliveryRouteService,
	settingService services.SettingService,
	jsonWebTokenService services.JsonWebTokenService,
) *deliveryRouteController {
	return &deliveryRouteController{
		adminService:         adminService,
		buyerService:         buyerService,
		deliveryRouteService: deliveryRouteService,
		settingService:       settingService,
		jsonWebTokenService:  jsonWebTokenService,
	}
}

func (q *deliveryRouteController) Mount(r fiber.Router) {
	buyerKeyAuth := middleware.BuyerKeyAuth(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Post("", q.V1BuyerSave)
}

func (q *deliveryRouteController) V1BuyerSave(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryRouteController-V1BuyerSave"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateDeliveryRoute(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryRoute")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	settings, _, err := q.settingService.FindSettings(
		ctx,
		models.NewSettingFilter().
			WithSettingNames(
				models.SettingNameDeliveryOriginLongitude,
				models.SettingNameDeliveryOriginLatitude,
				models.SettingNameDeliveryMaxDistanceKM,
			),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSettings")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	var (
		originLongitude,
		originLatitude decimal.Decimal
		deliveryMaxDistanceKM int64
	)
	for _, setting := range settings {
		switch setting.Name {
		case models.SettingNameDeliveryOriginLongitude.String():
			if originLongitude, err = decimal.Parse(setting.Value); err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(err.Error()).
					WriteResponse(c)
			}
		case models.SettingNameDeliveryOriginLatitude.String():
			if originLatitude, err = decimal.Parse(setting.Value); err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(err.Error()).
					WriteResponse(c)
			}
		case models.SettingNameDeliveryMaxDistanceKM.String():
			if deliveryMaxDistanceKM, err = strconv.ParseInt(setting.Value, 10, 64); err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
				return helper.NewResponse(fiber.StatusUnprocessableEntity).
					SetMessage(err.Error()).
					WriteResponse(c)
			}
		}
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
	now := time.Now()
	request.OriginLongitude = originLongitude
	request.OriginLatitude = originLatitude
	request.DistanceKM, _, _ = distanceKM.Ceil(0).Int64(0)
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
	response, err := q.deliveryRouteService.SaveDeliveryRoute(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveDeliveryRoute")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request.DistanceKM <= deliveryMaxDistanceKM {
		currentBuyer.DeliveryRouteID = &response.ID
		if _, err = q.buyerService.UpdateBuyer(ctx, tx, currentBuyer); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request.DistanceKM > deliveryMaxDistanceKM {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("your coordinates is out of our coverage area").
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
