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
	RewardPointService interface {
		FindRewardPoints(
			ctx context.Context,
			request *models.RewardPointFilter,
		) ([]*models.RewardPoint, *models.Pagination, error)
		FindRewardPoint(
			ctx context.Context,
			request *models.RewardPointFilter,
		) (*models.RewardPoint, error)
		CreateRewardPoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.RewardPoint,
		) (*models.RewardPoint, error)
		UpdateRewardPoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.RewardPoint,
		) (*models.RewardPoint, error)
	}

	rewardPointService struct {
		rewardPointRepository repositories.RewardPointRepository
	}
)

func NewRewardPointService(
	rewardPointRepository repositories.RewardPointRepository,
) RewardPointService {
	return new(rewardPointService{
		rewardPointRepository: rewardPointRepository,
	})
}

func (q *rewardPointService) FindRewardPoints(
	ctx context.Context,
	request *models.RewardPointFilter,
) ([]*models.RewardPoint, *models.Pagination, error) {
	ctxt := "RewardPointService-FindRewardPoints"
	points, total, pages, err := q.rewardPointRepository.FindRewardPoints(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRewardPoints")
		return nil, nil, err
	}
	n := len(points)
	rows := make([]*models.RewardPoint, n)
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

func (q *rewardPointService) FindRewardPoint(
	ctx context.Context,
	request *models.RewardPointFilter,
) (*models.RewardPoint, error) {
	ctxt := "RewardPointService-FindRewardPoint"
	response, err := q.rewardPointRepository.FindRewardPoint(ctx, request)
	if err != nil && !errors.Is(err, models.ErrRewardPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRewardPoint")
	}
	return response, err
}

func (q *rewardPointService) CreateRewardPoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RewardPoint,
) (*models.RewardPoint, error) {
	ctxt := "RewardPointService-CreateRewardPoint"
	response, err := q.rewardPointRepository.CreateRewardPoint(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRewardPoint")
	}
	return response, nil
}

func (q *rewardPointService) UpdateRewardPoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RewardPoint,
) (*models.RewardPoint, error) {
	ctxt := "RewardPointService-UpdateRewardPoint"
	response, err := q.rewardPointRepository.UpdateRewardPoint(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRewardPoint")
	}
	return response, nil
}
