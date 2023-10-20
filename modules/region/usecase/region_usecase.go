package usecase

import (
	"context"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/region/model"
	regionQuery "github.com/roysitumorang/laukpauk/modules/region/query"
	"go.uber.org/zap"
)

type (
	regionUseCaseImplementation struct {
		regionQuery regionQuery.RegionQuery
	}
)

func NewRegionUseCase(
	regionQuery regionQuery.RegionQuery,
) RegionUseCase {
	return &regionUseCaseImplementation{
		regionQuery: regionQuery,
	}
}

func (q *regionUseCaseImplementation) FindProvinces(ctx context.Context) (response []model.Region, err error) {
	ctxt := "RegionUseCase-FindProvinces"
	if response, err = q.regionQuery.FindProvinces(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProvinces")
	}
	return
}

func (q *regionUseCaseImplementation) FindCitiesByProvinceID(ctx context.Context, provinceID int64) (response []model.Region, err error) {
	ctxt := "RegionUseCase-FindCitiesByProvinceID"
	if response, err = q.regionQuery.FindCitiesByProvinceID(ctx, provinceID); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindCitiesByProvinceID")
	}
	return
}

func (q *regionUseCaseImplementation) FindSubdistrictsByCityID(ctx context.Context, cityID int64) (response []model.Region, err error) {
	ctxt := "RegionUseCase-FindSubdistrictsByCityID"
	if response, err = q.regionQuery.FindSubdistrictsByCityID(ctx, cityID); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindSubdistrictsByCityID")
	}
	return
}

func (q *regionUseCaseImplementation) FindVillagesBySubdistrictID(ctx context.Context, subdistrictID int64) (response []model.Region, err error) {
	ctxt := "RegionUseCase-FindVillagesBySubdistrictID"
	if response, err = q.regionQuery.FindVillagesBySubdistrictID(ctx, subdistrictID); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVillagesBySubdistrictID")
	}
	return
}
