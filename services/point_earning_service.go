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

type PointEarningService interface {
	FindPointEarnings(
		ctx context.Context,
		request *models.PointEarningFilter,
	) ([]*models.PointEarning, *models.Pagination, error)
	FindPointEarning(
		ctx context.Context,
		request *models.PointEarningFilter,
	) (*models.PointEarning, error)
	CreatePointEarning(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointEarning,
	) (*models.PointEarning, error)
	UpdatePointEarning(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointEarning,
	) (*models.PointEarning, error)
	SavePointEarning(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointEarning,
	) (*models.PointEarning, error)
}

type pointEarningService struct {
	pointEarningRepository repositories.PointEarningRepository
}

func NewPointEarningService(
	pointEarningRepository repositories.PointEarningRepository,
) PointEarningService {
	return &pointEarningService{
		pointEarningRepository: pointEarningRepository,
	}
}

func (q *pointEarningService) FindPointEarnings(
	ctx context.Context,
	request *models.PointEarningFilter,
) ([]*models.PointEarning, *models.Pagination, error) {
	ctxt := "PointEarningService-FindPointEarnings"
	pointEarnings, total, pages, err := q.pointEarningRepository.FindPointEarnings(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarnings")
		return nil, nil, err
	}
	n := len(pointEarnings)
	rows := make([]*models.PointEarning, n)
	if n > 0 {
		copy(rows, pointEarnings)
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

func (q *pointEarningService) FindPointEarning(
	ctx context.Context,
	request *models.PointEarningFilter,
) (*models.PointEarning, error) {
	ctxt := "PointEarningService-FindPointEarning"
	response, err := q.pointEarningRepository.FindPointEarning(ctx, request)
	if err != nil && !errors.Is(err, models.ErrPointEarningNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarning")
	}
	return response, err
}

func (q *pointEarningService) CreatePointEarning(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarning,
) (*models.PointEarning, error) {
	ctxt := "PointEarningService-CreatePointEarning"
	response, err := q.pointEarningRepository.CreatePointEarning(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointEarning")
	}
	return response, nil
}

func (q *pointEarningService) UpdatePointEarning(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarning,
) (*models.PointEarning, error) {
	ctxt := "PointEarningService-UpdatePointEarning"
	response, err := q.pointEarningRepository.UpdatePointEarning(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrPointEarningNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarning")
	}
	return response, nil
}

func (q *pointEarningService) SavePointEarning(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarning,
) (response *models.PointEarning, err error) {
	ctxt := "PointEarningService-SavePointEarning"
	if request.New() {
		if response, err = q.pointEarningRepository.CreatePointEarning(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointEarning")
		}
	} else if response, err = q.pointEarningRepository.UpdatePointEarning(ctx, tx, request); err != nil && !errors.Is(err, models.ErrPointEarningNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarning")
	}
	return response, err
}
