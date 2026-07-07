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
	PointService interface {
		FindPoints(
			ctx context.Context,
			request *models.PointFilter,
		) ([]*models.Point, *models.Pagination, error)
		FindPoint(
			ctx context.Context,
			request *models.PointFilter,
		) (*models.Point, error)
		CreatePoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Point,
		) (*models.Point, error)
		UpdatePoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Point,
		) (*models.Point, error)
	}

	pointService struct {
		pointRepository repositories.PointRepository
	}
)

func NewPointService(
	pointRepository repositories.PointRepository,
) PointService {
	return new(pointService{
		pointRepository: pointRepository,
	})
}

func (q *pointService) FindPoints(
	ctx context.Context,
	request *models.PointFilter,
) ([]*models.Point, *models.Pagination, error) {
	ctxt := "PointService-FindPoints"
	points, total, pages, err := q.pointRepository.FindPoints(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoints")
		return nil, nil, err
	}
	n := len(points)
	rows := make([]*models.Point, n)
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

func (q *pointService) FindPoint(
	ctx context.Context,
	request *models.PointFilter,
) (*models.Point, error) {
	ctxt := "PointService-FindPoint"
	response, err := q.pointRepository.FindPoint(ctx, request)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
	}
	return response, err
}

func (q *pointService) CreatePoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointService-CreatePoint"
	response, err := q.pointRepository.CreatePoint(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePoint")
	}
	return response, nil
}

func (q *pointService) UpdatePoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointService-UpdatePoint"
	response, err := q.pointRepository.UpdatePoint(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
	}
	return response, nil
}
