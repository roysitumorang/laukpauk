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
	PointEarningCapService interface {
		FindPointEarningCaps(
			ctx context.Context,
			request *models.PointEarningCapFilter,
		) ([]*models.PointEarningCap, *models.Pagination, error)
		FindPointEarningCap(
			ctx context.Context,
			request *models.PointEarningCapFilter,
		) (*models.PointEarningCap, error)
		CreatePointEarningCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointEarningCap,
		) (*models.PointEarningCap, error)
		UpdatePointEarningCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointEarningCap,
		) (*models.PointEarningCap, error)
	}

	pointEarningCapService struct {
		pointRepository repositories.PointEarningCapRepository
	}
)

func NewPointEarningCapService(
	pointRepository repositories.PointEarningCapRepository,
) PointEarningCapService {
	return new(pointEarningCapService{
		pointRepository: pointRepository,
	})
}

func (q *pointEarningCapService) FindPointEarningCaps(
	ctx context.Context,
	request *models.PointEarningCapFilter,
) ([]*models.PointEarningCap, *models.Pagination, error) {
	ctxt := "PointEarningCapService-FindPointEarningCaps"
	points, total, pages, err := q.pointRepository.FindPointEarningCaps(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return nil, nil, err
	}
	n := len(points)
	rows := make([]*models.PointEarningCap, n)
	if n > 0 {
		copy(rows, points)
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

func (q *pointEarningCapService) FindPointEarningCap(
	ctx context.Context,
	request *models.PointEarningCapFilter,
) (*models.PointEarningCap, error) {
	ctxt := "PointEarningCapService-FindPointEarningCap"
	response, err := q.pointRepository.FindPointEarningCap(ctx, request)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
	}
	return response, err
}

func (q *pointEarningCapService) CreatePointEarningCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarningCap,
) (*models.PointEarningCap, error) {
	ctxt := "PointEarningCapService-CreatePointEarningCap"
	response, err := q.pointRepository.CreatePointEarningCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointEarningCap")
	}
	return response, nil
}

func (q *pointEarningCapService) UpdatePointEarningCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarningCap,
) (*models.PointEarningCap, error) {
	ctxt := "PointEarningCapService-UpdatePointEarningCap"
	response, err := q.pointRepository.UpdatePointEarningCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
	}
	return response, nil
}
