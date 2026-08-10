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

type PointEarningCapService interface {
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
	SavePointEarningCap(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointEarningCap,
	) (*models.PointEarningCap, error)
}

type pointEarningCapService struct {
	pointEarningCapRepository repositories.PointEarningCapRepository
}

func NewPointEarningCapService(
	pointEarningCapRepository repositories.PointEarningCapRepository,
) PointEarningCapService {
	return &pointEarningCapService{
		pointEarningCapRepository: pointEarningCapRepository,
	}
}

func (q *pointEarningCapService) FindPointEarningCaps(
	ctx context.Context,
	request *models.PointEarningCapFilter,
) ([]*models.PointEarningCap, *models.Pagination, error) {
	ctxt := "PointEarningCapService-FindPointEarningCaps"
	pointEarningCaps, total, pages, err := q.pointEarningCapRepository.FindPointEarningCaps(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return nil, nil, err
	}
	n := len(pointEarningCaps)
	rows := make([]*models.PointEarningCap, n)
	if n > 0 {
		copy(rows, pointEarningCaps)
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
	response, err := q.pointEarningCapRepository.FindPointEarningCap(ctx, request)
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
	response, err := q.pointEarningCapRepository.CreatePointEarningCap(ctx, tx, request)
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
	response, err := q.pointEarningCapRepository.UpdatePointEarningCap(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
	}
	return response, nil
}

func (q *pointEarningCapService) SavePointEarningCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointEarningCap,
) (response *models.PointEarningCap, err error) {
	ctxt := "PointEarningCapService-SavePointEarningCap"
	if request.New() {
		if response, err = q.pointEarningCapRepository.CreatePointEarningCap(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointEarningCap")
		}
	} else if response, err = q.pointEarningCapRepository.UpdatePointEarningCap(ctx, tx, request); err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
	}
	return response, err
}
