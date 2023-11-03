package query

import (
	"context"

	"github.com/roysitumorang/laukpauk/modules/banner/model"
)

type (
	BannerQuery interface {
		FindBanners(ctx context.Context) ([]model.Banner, error)
	}
)
