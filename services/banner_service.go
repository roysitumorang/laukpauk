package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	BannerService interface {
		FindBanners(
			ctx context.Context,
			request *models.BannerFilter,
		) ([]*models.Banner, *models.Pagination, error)
		FindBanner(
			ctx context.Context,
			request *models.BannerFilter,
		) (*models.Banner, error)
		CreateBanner(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Banner,
		) (*models.Banner, error)
		UpdateBanner(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Banner,
		) (*models.Banner, error)
	}

	bannerService struct {
		bannerRepository repositories.BannerRepository
	}
)

func NewBannerService(
	bannerRepository repositories.BannerRepository,
) BannerService {
	return &bannerService{
		bannerRepository: bannerRepository,
	}
}

func (q *bannerService) FindBanners(
	ctx context.Context,
	request *models.BannerFilter,
) ([]*models.Banner, *models.Pagination, error) {
	ctxt := "BannerService-FindBanners"
	banners, total, pages, err := q.bannerRepository.FindBanners(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanners")
		return nil, nil, err
	}
	n := len(banners)
	rows := make([]*models.Banner, n)
	if n > 0 {
		copy(rows, banners)
	}
	pagination, err := models.NewPagination(
		total,
		pages,
		request.Limit,
		request.Page,
		request.PaginationURL,
		request.UrlValues,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewPagination")
		return nil, nil, err
	}
	return rows, pagination, nil
}

func (q *bannerService) FindBanner(
	ctx context.Context,
	request *models.BannerFilter,
) (*models.Banner, error) {
	ctxt := "BannerService-FindBanner"
	response, err := q.bannerRepository.FindBanner(ctx, request)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
	}
	return response, err
}

func (q *bannerService) CreateBanner(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Banner,
) (*models.Banner, error) {
	ctxt := "BannerService-CreateBanner"
	response, err := q.bannerRepository.CreateBanner(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBanner")
	}
	return response, err
}
func (q *bannerService) UpdateBanner(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Banner,
) (*models.Banner, error) {
	ctxt := "BannerService-UpdateBanner"
	response, err := q.bannerRepository.UpdateBanner(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
	}
	return response, err
}
