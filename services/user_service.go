package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	UserService interface {
		FindAdmins(
			ctx context.Context,
			request *models.UserFilter,
		) ([]*models.Admin, *models.Pagination, error)
		FindAdmin(
			ctx context.Context,
			request *models.UserFilter,
		) (*models.Admin, error)
		FindBuyers(
			ctx context.Context,
			request *models.UserFilter,
		) ([]*models.Buyer, *models.Pagination, error)
		FindBuyer(
			ctx context.Context,
			request *models.UserFilter,
		) (*models.Buyer, error)
	}

	userService struct {
		userRepository repositories.UserRepository
	}
)

func NewUserService(
	userRepository repositories.UserRepository,
) UserService {
	return new(userService{
		userRepository: userRepository,
	})
}

func (q *userService) FindAdmins(
	ctx context.Context,
	request *models.UserFilter,
) ([]*models.Admin, *models.Pagination, error) {
	ctxt := "UserService-FindAdmins"
	admins, total, pages, err := q.userRepository.FindAdmins(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmins")
		return nil, nil, err
	}
	n := len(admins)
	rows := make([]*models.Admin, n)
	if n > 0 {
		copy(rows, admins)
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

func (q *userService) FindAdmin(
	ctx context.Context,
	request *models.UserFilter,
) (*models.Admin, error) {
	ctxt := "UserService-FindAdmin"
	response, err := q.userRepository.FindAdmin(ctx, request)
	if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
	}
	return response, err
}

func (q *userService) FindBuyers(
	ctx context.Context,
	request *models.UserFilter,
) ([]*models.Buyer, *models.Pagination, error) {
	ctxt := "UserService-FindBuyers"
	buyers, total, pages, err := q.userRepository.FindBuyers(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return nil, nil, err
	}
	n := len(buyers)
	rows := make([]*models.Buyer, n)
	if n > 0 {
		copy(rows, buyers)
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

func (q *userService) FindBuyer(
	ctx context.Context,
	request *models.UserFilter,
) (*models.Buyer, error) {
	ctxt := "UserService-FindBuyer"
	response, err := q.userRepository.FindBuyer(ctx, request)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
	}
	return response, err
}
