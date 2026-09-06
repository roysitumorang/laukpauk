package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"

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

type (
	deliveryRouteController struct {
		adminService         services.AdminService
		buyerService         services.BuyerService
		deliveryRouteService services.DeliveryRouteService
		companyService       services.CompanyService
		jsonWebTokenService  services.JsonWebTokenService
		routesApiClient      *externals.RoutesApiClient
	}
)

func NewDeliveryRouteController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	deliveryRouteService services.DeliveryRouteService,
	companyService services.CompanyService,
	jsonWebTokenService services.JsonWebTokenService,
	routesApiClient *externals.RoutesApiClient,
) *deliveryRouteController {
	return &deliveryRouteController{
		adminService:         adminService,
		buyerService:         buyerService,
		deliveryRouteService: deliveryRouteService,
		companyService:       companyService,
		jsonWebTokenService:  jsonWebTokenService,
		routesApiClient:      routesApiClient,
	}
}

func (q *deliveryRouteController) Mount(r fiber.Router) {
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
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
	company, err := q.companyService.FindCompany(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(currentBuyer.CompanyID),
	)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if company == nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrCompanyNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.deliveryRouteService.FindDeliveryRoute(
		ctx,
		models.NewDeliveryRouteFilter().
			WithOriginLongitude(&company.DeliveryOriginLongitude).
			WithOriginLatitude(&company.DeliveryOriginLatitude).
			WithDestinationLongitude(&request.DestinationLongitude).
			WithDestinationLatitude(&request.DestinationLatitude),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryRouteNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryRoute")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response != nil {
		return helper.NewResponse(fiber.StatusOK).
			SetData(response).
			WriteResponse(c)
	}
	if request.DistanceMeters, err = q.routesApiClient.ComputeDistanceMeters(
		ctx,
		company.DeliveryOriginLongitude.String(),
		company.DeliveryOriginLatitude.String(),
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
	now := time.Now()
	request.OriginLongitude = company.DeliveryOriginLongitude
	request.OriginLatitude = company.DeliveryOriginLatitude
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
	if response, err = q.deliveryRouteService.SaveDeliveryRoute(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveDeliveryRoute")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request.DistanceKM <= company.DeliveryMaxDistanceKM {
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
	if request.DistanceKM > company.DeliveryMaxDistanceKM {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("your coordinates is out of our coverage area").
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
