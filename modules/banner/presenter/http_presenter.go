package presenter

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/laukpauk/helper"
	bannerUseCase "github.com/roysitumorang/laukpauk/modules/banner/usecase"
	"go.uber.org/zap"
)

type (
	bannerHTTPHandler struct {
		bannerUseCase bannerUseCase.BannerUseCase
	}
)

func NewBannerHTTPHandler(accountUseCase bannerUseCase.BannerUseCase) *bannerHTTPHandler {
	return &bannerHTTPHandler{
		bannerUseCase: accountUseCase,
	}
}

func (q *bannerHTTPHandler) Mount(r fiber.Router) {
	r.Get("", q.FindBanners)
}

func (q *bannerHTTPHandler) FindBanners(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "BannerPresenter-FindBanners"
	response, err := q.bannerUseCase.FindBanners(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBanners")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}
