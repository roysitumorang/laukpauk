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

type RecipeGuideService interface {
	FindRecipeGuides(
		ctx context.Context,
		request *models.RecipeGuideFilter,
	) ([]*models.RecipeGuide, *models.Pagination, error)
	FindRecipeGuide(
		ctx context.Context,
		request *models.RecipeGuideFilter,
	) (*models.RecipeGuide, error)
	CreateRecipeGuide(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeGuide,
	) (*models.RecipeGuide, error)
	UpdateRecipeGuide(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeGuide,
	) (*models.RecipeGuide, error)
}

type recipeGuideService struct {
	recipeGuideRepository repositories.RecipeGuideRepository
}

func NewRecipeGuideService(
	recipeGuideRepository repositories.RecipeGuideRepository,
) RecipeGuideService {
	return &recipeGuideService{
		recipeGuideRepository: recipeGuideRepository,
	}
}

func (q *recipeGuideService) FindRecipeGuides(
	ctx context.Context,
	request *models.RecipeGuideFilter,
) ([]*models.RecipeGuide, *models.Pagination, error) {
	ctxt := "RecipeGuideService-FindRecipeGuides"
	recipeGuides, total, pages, err := q.recipeGuideRepository.FindRecipeGuides(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeGuides")
		return nil, nil, err
	}
	n := len(recipeGuides)
	rows := make([]*models.RecipeGuide, n)
	if n > 0 {
		copy(rows, recipeGuides)
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

func (q *recipeGuideService) FindRecipeGuide(
	ctx context.Context,
	request *models.RecipeGuideFilter,
) (*models.RecipeGuide, error) {
	ctxt := "RecipeGuideService-FindRecipeGuide"
	response, err := q.recipeGuideRepository.FindRecipeGuide(ctx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeGuideNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeGuide")
	}
	return response, err
}

func (q *recipeGuideService) CreateRecipeGuide(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeGuide,
) (*models.RecipeGuide, error) {
	ctxt := "RecipeGuideService-CreateRecipeGuide"
	response, err := q.recipeGuideRepository.CreateRecipeGuide(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRecipeGuide")
	}
	return response, err
}

func (q *recipeGuideService) UpdateRecipeGuide(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeGuide,
) (*models.RecipeGuide, error) {
	ctxt := "RecipeGuideService-UpdateRecipeGuide"
	response, err := q.recipeGuideRepository.UpdateRecipeGuide(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeGuideNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeGuide")
	}
	return response, err
}
