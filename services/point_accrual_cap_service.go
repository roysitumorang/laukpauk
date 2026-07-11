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
	PointAccrualCapService interface {
		FindPointAccrualCaps(
			ctx context.Context,
			request *models.PointAccrualCapFilter,
		) ([]*models.PointAccrualCap, *models.Pagination, error)
		FindPointAccrualCap(
			ctx context.Context,
			request *models.PointAccrualCapFilter,
		) (*models.PointAccrualCap, error)
		CreatePointAccrualCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointAccrualCap,
		) (*models.PointAccrualCap, error)
		UpdatePointAccrualCap(
			ctx context.Context,
			tx pgx.Tx,
			request *models.PointAccrualCap,
		) (*models.PointAccrualCap, error)
	}

	pointAccrualCapService struct {
		pointRepository repositories.PointAccrualCapRepository
	}
)

func NewPointAccrualCapService(
	pointRepository repositories.PointAccrualCapRepository,
) PointAccrualCapService {
	return new(pointAccrualCapService{
		pointRepository: pointRepository,
	})
}

func (q *pointAccrualCapService) FindPointAccrualCaps(
	ctx context.Context,
	request *models.PointAccrualCapFilter,
) ([]*models.PointAccrualCap, *models.Pagination, error) {
	ctxt := "PointAccrualCapService-FindPointAccrualCaps"
	points, total, pages, err := q.pointRepository.FindPointAccrualCaps(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCaps")
		return nil, nil, err
	}
	n := len(points)
	rows := make([]*models.PointAccrualCap, n)
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

func (q *pointAccrualCapService) FindPointAccrualCap(
	ctx context.Context,
	request *models.PointAccrualCapFilter,
) (*models.PointAccrualCap, error) {
	ctxt := "PointAccrualCapService-FindPointAccrualCap"
	response, err := q.pointRepository.FindPointAccrualCap(ctx, request)
	if err != nil && !errors.Is(err, models.ErrPointAccrualCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCap")
	}
	return response, err
}

func (q *pointAccrualCapService) CreatePointAccrualCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointAccrualCap,
) (*models.PointAccrualCap, error) {
	ctxt := "PointAccrualCapService-CreatePointAccrualCap"
	response, err := q.pointRepository.CreatePointAccrualCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointAccrualCap")
	}
	return response, nil
}

func (q *pointAccrualCapService) UpdatePointAccrualCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointAccrualCap,
) (*models.PointAccrualCap, error) {
	ctxt := "PointAccrualCapService-UpdatePointAccrualCap"
	response, err := q.pointRepository.UpdatePointAccrualCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointAccrualCap")
	}
	return response, nil
}
