package presenter

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/laukpauk/helper"
	regionUseCase "github.com/roysitumorang/laukpauk/modules/region/usecase"
	"go.uber.org/zap"
)

type (
	regionHTTPHandler struct {
		regionUseCase regionUseCase.RegionUseCase
	}
)

func NewRegionHTTPHandler(accountUseCase regionUseCase.RegionUseCase) *regionHTTPHandler {
	return &regionHTTPHandler{
		regionUseCase: accountUseCase,
	}
}

func (q *regionHTTPHandler) Mount(r fiber.Router) {
	r.Get("/provinces", q.FindProvinces).
		Get("/provinces/:province_id/cities", q.FindCitiesByProvinceID).
		Get("/cities/:city_id/subdistricts", q.FindSubdistrictsByCityID).
		Get("/subdistricts/:subdistrict_id/villages", q.FindVillagesBySubdistrictID)
}

func (q *regionHTTPHandler) FindProvinces(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "RegionPresenter-FindProvinces"
	response, err := q.regionUseCase.FindProvinces(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProvinces")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *regionHTTPHandler) FindCitiesByProvinceID(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "RegionPresenter-FindCitiesByProvinceID"
	provinceID, _ := strconv.ParseInt(c.Params("province_id"), 10, 64)
	response, err := q.regionUseCase.FindCitiesByProvinceID(ctx, provinceID)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindCitiesByProvinceID")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *regionHTTPHandler) FindSubdistrictsByCityID(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "RegionPresenter-FindSubdistrictsByCityID"
	cityID, _ := strconv.ParseInt(c.Params("city_id"), 10, 64)
	response, err := q.regionUseCase.FindSubdistrictsByCityID(ctx, cityID)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindSubdistrictsByCityID")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *regionHTTPHandler) FindVillagesBySubdistrictID(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "RegionPresenter-FindVillagesBySubdistrictID"
	subdistrictID, _ := strconv.ParseInt(c.Params("subdistrict_id"), 10, 64)
	response, err := q.regionUseCase.FindVillagesBySubdistrictID(ctx, subdistrictID)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVillagesBySubdistrictID")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}
