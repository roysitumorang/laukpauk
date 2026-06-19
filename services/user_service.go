package services

import (
	"context"
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
