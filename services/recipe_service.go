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

type RecipeService interface {
	FindRecipes(
		ctx context.Context,
		request *models.RecipeFilter,
	) ([]*models.Recipe, *models.Pagination, error)
	FindRecipe(
		ctx context.Context,
		request *models.RecipeFilter,
	) (*models.Recipe, error)
	CreateRecipe(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Recipe,
	) (*models.Recipe, error)
	UpdateRecipe(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Recipe,
	) (*models.Recipe, error)
	SaveRecipe(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Recipe,
	) (*models.Recipe, error)
}

type recipeService struct {
	recipeRepository repositories.RecipeRepository
}

func NewRecipeService(
	recipeRepository repositories.RecipeRepository,
) RecipeService {
	return &recipeService{
		recipeRepository: recipeRepository,
	}
}

func (q *recipeService) FindRecipes(
	ctx context.Context,
	request *models.RecipeFilter,
) ([]*models.Recipe, *models.Pagination, error) {
	ctxt := "RecipeService-FindRecipes"
	recipes, total, pages, err := q.recipeRepository.FindRecipes(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipes")
		return nil, nil, err
	}
	n := len(recipes)
	rows := make([]*models.Recipe, n)
	if n > 0 {
		copy(rows, recipes)
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

func (q *recipeService) FindRecipe(
	ctx context.Context,
	request *models.RecipeFilter,
) (*models.Recipe, error) {
	ctxt := "RecipeService-FindRecipe"
	response, err := q.recipeRepository.FindRecipe(ctx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipe")
	}
	return response, err
}

func (q *recipeService) CreateRecipe(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Recipe,
) (*models.Recipe, error) {
	ctxt := "RecipeService-CreateRecipe"
	response, err := q.recipeRepository.CreateRecipe(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRecipe")
	}
	return response, err
}

func (q *recipeService) UpdateRecipe(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Recipe,
) (*models.Recipe, error) {
	ctxt := "RecipeService-UpdateRecipe"
	response, err := q.recipeRepository.UpdateRecipe(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipe")
	}
	return response, err
}

func (q *recipeService) SaveRecipe(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Recipe,
) (*models.Recipe, error) {
	ctxt := "RecipeService-SaveRecipe"
	response, err := q.recipeRepository.SaveRecipe(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipe")
	}
	return response, err
}
