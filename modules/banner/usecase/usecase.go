package usecase

import (
	"context"

	"github.com/roysitumorang/laukpauk/modules/banner/model"
)

type (
	BannerUseCase interface {
		FindBanners(ctx context.Context) ([]model.Banner, error)
	}
)
