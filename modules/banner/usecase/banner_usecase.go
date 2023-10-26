package usecase

import (
	"context"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/banner/model"
	bannerQuery "github.com/roysitumorang/laukpauk/modules/banner/query"
	"go.uber.org/zap"
)

type (
	bannerUseCaseImplementation struct {
		bannerQuery bannerQuery.BannerQuery
	}
)

func NewBannerUseCase(
	bannerQuery bannerQuery.BannerQuery,
) BannerUseCase {
	return &bannerUseCaseImplementation{
		bannerQuery: bannerQuery,
	}
}

func (q *bannerUseCaseImplementation) FindBanners(ctx context.Context) (response []model.Banner, err error) {
	ctxt := "BannerUseCase-FindBanners"
	if response, err = q.bannerQuery.FindBanners(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBanners")
	}
	return
}
