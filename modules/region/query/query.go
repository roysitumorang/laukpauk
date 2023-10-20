package query

import (
	"context"

	"github.com/roysitumorang/laukpauk/modules/region/model"
)

type (
	RegionQuery interface {
		FindProvinces(ctx context.Context) ([]model.Region, error)
		FindCitiesByProvinceID(ctx context.Context, provinceID int64) (response []model.Region, err error)
		FindSubdistrictsByCityID(ctx context.Context, cityID int64) (response []model.Region, err error)
		FindVillagesBySubdistrictID(ctx context.Context, subdistrictID int64) (response []model.Region, err error)
	}
)
