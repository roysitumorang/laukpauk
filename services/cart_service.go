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
	CartService interface {
		FindCarts(
			ctx context.Context,
			request *models.CartFilter,
		) ([]*models.Cart, *models.Pagination, error)
		FindCart(
			ctx context.Context,
			request *models.CartFilter,
		) (*models.Cart, error)
		SaveCart(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Cart,
		) (*models.Cart, error)
	}

	cartService struct {
		cartRepository repositories.CartRepository
	}
)

func NewCartService(
	cartRepository repositories.CartRepository,
) CartService {
	return &cartService{
		cartRepository: cartRepository,
	}
}

func (q *cartService) FindCarts(
	ctx context.Context,
	request *models.CartFilter,
) ([]*models.Cart, *models.Pagination, error) {
	ctxt := "CartService-FindCarts"
	carts, total, pages, err := q.cartRepository.FindCarts(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCarts")
		return nil, nil, err
	}
	n := len(carts)
	rows := make([]*models.Cart, n)
	if n > 0 {
		copy(rows, carts)
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

func (q *cartService) FindCart(
	ctx context.Context,
	request *models.CartFilter,
) (*models.Cart, error) {
	ctxt := "CartService-FindCart"
	response, err := q.cartRepository.FindCart(ctx, request)
	if err != nil && !errors.Is(err, models.ErrCartNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCart")
	}
	return response, err
}

func (q *cartService) SaveCart(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Cart,
) (*models.Cart, error) {
	ctxt := "CartService-CreateCart"
	response, err := q.cartRepository.SaveCart(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateCart")
	}
	return response, err
}
