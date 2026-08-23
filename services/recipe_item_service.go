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

type RecipeItemService interface {
	FindRecipeItems(
		ctx context.Context,
		request *models.RecipeItemFilter,
	) ([]*models.RecipeItem, *models.Pagination, error)
	FindRecipeItem(
		ctx context.Context,
		request *models.RecipeItemFilter,
	) (*models.RecipeItem, error)
	CreateRecipeItem(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeItem,
	) (*models.RecipeItem, error)
	UpdateRecipeItem(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeItem,
	) (*models.RecipeItem, error)
	SaveRecipeItem(
		ctx context.Context,
		tx pgx.Tx,
		request *models.RecipeItem,
	) (*models.RecipeItem, error)
}

type recipeItemService struct {
	recipeItemRepository repositories.RecipeItemRepository
}

func NewRecipeItemService(
	recipeItemRepository repositories.RecipeItemRepository,
) RecipeItemService {
	return &recipeItemService{
		recipeItemRepository: recipeItemRepository,
	}
}

func (q *recipeItemService) FindRecipeItems(
	ctx context.Context,
	request *models.RecipeItemFilter,
) ([]*models.RecipeItem, *models.Pagination, error) {
	ctxt := "RecipeItemService-FindRecipeItems"
	recipeItems, total, pages, err := q.recipeItemRepository.FindRecipeItems(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItems")
		return nil, nil, err
	}
	n := len(recipeItems)
	rows := make([]*models.RecipeItem, n)
	if n > 0 {
		copy(rows, recipeItems)
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

func (q *recipeItemService) FindRecipeItem(
	ctx context.Context,
	request *models.RecipeItemFilter,
) (*models.RecipeItem, error) {
	ctxt := "RecipeItemService-FindRecipeItem"
	response, err := q.recipeItemRepository.FindRecipeItem(ctx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRecipeItem")
	}
	return response, err
}

func (q *recipeItemService) CreateRecipeItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeItem,
) (*models.RecipeItem, error) {
	ctxt := "RecipeItemService-CreateRecipeItem"
	response, err := q.recipeItemRepository.CreateRecipeItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRecipeItem")
	}
	return response, err
}

func (q *recipeItemService) UpdateRecipeItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeItem,
) (*models.RecipeItem, error) {
	ctxt := "RecipeItemService-UpdateRecipeItem"
	response, err := q.recipeItemRepository.UpdateRecipeItem(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRecipeItem")
	}
	return response, err
}

func (q *recipeItemService) SaveRecipeItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RecipeItem,
) (*models.RecipeItem, error) {
	ctxt := "RecipeItemService-SaveRecipeItem"
	response, err := q.recipeItemRepository.SaveRecipeItem(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrRecipeItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveRecipeItem")
	}
	return response, err
}
