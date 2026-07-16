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
	CartItemService interface {
		FindCartItems(
			ctx context.Context,
			request *models.CartItemFilter,
		) ([]*models.CartItem, *models.Pagination, error)
		FindCartItem(
			ctx context.Context,
			request *models.CartItemFilter,
		) (*models.CartItem, error)
		SaveCartItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.CartItem,
		) (*models.CartItem, error)
		DeleteCartItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.CartItemFilter,
		) (int64, error)
	}

	cartItemService struct {
		cartItemRepository repositories.CartItemRepository
	}
)

func NewCartItemService(
	cartItemRepository repositories.CartItemRepository,
) CartItemService {
	return &cartItemService{
		cartItemRepository: cartItemRepository,
	}
}

func (q *cartItemService) FindCartItems(
	ctx context.Context,
	request *models.CartItemFilter,
) ([]*models.CartItem, *models.Pagination, error) {
	ctxt := "CartItemService-FindCartItems"
	cartItems, total, pages, err := q.cartItemRepository.FindCartItems(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCartItems")
		return nil, nil, err
	}
	n := len(cartItems)
	rows := make([]*models.CartItem, n)
	if n > 0 {
		copy(rows, cartItems)
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

func (q *cartItemService) FindCartItem(
	ctx context.Context,
	request *models.CartItemFilter,
) (*models.CartItem, error) {
	ctxt := "CartItemService-FindCartItem"
	response, err := q.cartItemRepository.FindCartItem(ctx, request)
	if err != nil && !errors.Is(err, models.ErrCartItemNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCartItem")
	}
	return response, err
}

func (q *cartItemService) SaveCartItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.CartItem,
) (*models.CartItem, error) {
	ctxt := "CartItemService-CreateCartItem"
	response, err := q.cartItemRepository.SaveCartItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateCartItem")
	}
	return response, err
}

func (q *cartItemService) DeleteCartItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.CartItemFilter,
) (int64, error) {
	ctxt := "CartItemService-DeleteCartItem"
	response, err := q.cartItemRepository.DeleteCartItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteCartItem")
	}
	return response, nil
}
