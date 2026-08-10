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

type PointSpendingCapService interface {
	FindPointSpendingCaps(
		ctx context.Context,
		request *models.PointSpendingCapFilter,
	) ([]*models.PointSpendingCap, *models.Pagination, error)
	FindPointSpendingCap(
		ctx context.Context,
		request *models.PointSpendingCapFilter,
	) (*models.PointSpendingCap, error)
	CreatePointSpendingCap(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointSpendingCap,
	) (*models.PointSpendingCap, error)
	UpdatePointSpendingCap(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointSpendingCap,
	) (*models.PointSpendingCap, error)
	SavePointSpendingCap(
		ctx context.Context,
		tx pgx.Tx,
		request *models.PointSpendingCap,
	) (*models.PointSpendingCap, error)
}

type pointSpendingCapService struct {
	pointSpendingCapRepository repositories.PointSpendingCapRepository
}

func NewPointSpendingCapService(
	pointSpendingCapRepository repositories.PointSpendingCapRepository,
) PointSpendingCapService {
	return &pointSpendingCapService{
		pointSpendingCapRepository: pointSpendingCapRepository,
	}
}

func (q *pointSpendingCapService) FindPointSpendingCaps(
	ctx context.Context,
	request *models.PointSpendingCapFilter,
) ([]*models.PointSpendingCap, *models.Pagination, error) {
	ctxt := "PointSpendingCapService-FindPointSpendingCaps"
	pointSpendingCaps, total, pages, err := q.pointSpendingCapRepository.FindPointSpendingCaps(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
		return nil, nil, err
	}
	n := len(pointSpendingCaps)
	rows := make([]*models.PointSpendingCap, n)
	if n > 0 {
		copy(rows, pointSpendingCaps)
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

func (q *pointSpendingCapService) FindPointSpendingCap(
	ctx context.Context,
	request *models.PointSpendingCapFilter,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapService-FindPointSpendingCap"
	response, err := q.pointSpendingCapRepository.FindPointSpendingCap(ctx, request)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
	}
	return response, err
}

func (q *pointSpendingCapService) CreatePointSpendingCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointSpendingCap,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapService-CreatePointSpendingCap"
	response, err := q.pointSpendingCapRepository.CreatePointSpendingCap(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointSpendingCap")
	}
	return response, nil
}

func (q *pointSpendingCapService) UpdatePointSpendingCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointSpendingCap,
) (*models.PointSpendingCap, error) {
	ctxt := "PointSpendingCapService-UpdatePointSpendingCap"
	response, err := q.pointSpendingCapRepository.UpdatePointSpendingCap(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointSpendingCap")
	}
	return response, nil
}

func (q *pointSpendingCapService) SavePointSpendingCap(
	ctx context.Context,
	tx pgx.Tx,
	request *models.PointSpendingCap,
) (response *models.PointSpendingCap, err error) {
	ctxt := "PointSpendingCapService-SavePointSpendingCap"
	if request.New() {
		if response, err = q.pointSpendingCapRepository.CreatePointSpendingCap(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointSpendingCap")
		}
	} else if response, err = q.pointSpendingCapRepository.UpdatePointSpendingCap(ctx, tx, request); err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointSpendingCap")
	}
	return response, err
}
