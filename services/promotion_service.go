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

type PromotionService interface {
	FindPromotions(
		ctx context.Context,
		request *models.PromotionFilter,
	) ([]*models.Promotion, *models.Pagination, error)
	FindPromotion(
		ctx context.Context,
		request *models.PromotionFilter,
	) (*models.Promotion, error)
	CreatePromotion(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Promotion,
	) (*models.Promotion, error)
	UpdatePromotion(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Promotion,
	) (*models.Promotion, error)
}

type promotionService struct {
	promotionRepository repositories.PromotionRepository
}

func NewPromotionService(
	promotionRepository repositories.PromotionRepository,
) PromotionService {
	return &promotionService{
		promotionRepository: promotionRepository,
	}
}

func (q *promotionService) FindPromotions(
	ctx context.Context,
	request *models.PromotionFilter,
) ([]*models.Promotion, *models.Pagination, error) {
	ctxt := "PromotionService-FindPromotions"
	promotions, total, pages, err := q.promotionRepository.FindPromotions(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotions")
		return nil, nil, err
	}
	n := len(promotions)
	rows := make([]*models.Promotion, n)
	if n > 0 {
		copy(rows, promotions)
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

func (q *promotionService) FindPromotion(
	ctx context.Context,
	request *models.PromotionFilter,
) (*models.Promotion, error) {
	ctxt := "PromotionService-FindPromotion"
	response, err := q.promotionRepository.FindPromotion(ctx, request)
	if errors.Is(err, models.ErrPromotionNotFound) {
		return nil, err
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotion")
	}
	return response, err
}

func (q *promotionService) CreatePromotion(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Promotion,
) (*models.Promotion, error) {
	ctxt := "PromotionService-CreatePromotion"
	response, err := q.promotionRepository.CreatePromotion(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePromotion")
	}
	return response, nil
}
func (q *promotionService) UpdatePromotion(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Promotion,
) (*models.Promotion, error) {
	ctxt := "PromotionService-UpdatePromotion"
	response, err := q.promotionRepository.UpdatePromotion(ctx, tx, request)
	if errors.Is(err, models.ErrPromotionNotFound) {
		return nil, err
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePromotion")
	}
	return response, nil
}
