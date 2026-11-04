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

type UserService interface {
	FindUsers(
		ctx context.Context,
		request *models.UserFilter,
	) ([]*models.User, *models.Pagination, error)
	FindUser(
		ctx context.Context,
		request *models.UserFilter,
	) (*models.User, error)
	CreateUser(
		ctx context.Context,
		tx pgx.Tx,
		request *models.User,
	) (*models.User, error)
	UpdateUser(
		ctx context.Context,
		tx pgx.Tx,
		request *models.User,
	) (*models.User, error)
}

type userService struct {
	userRepository repositories.UserRepository
}

func NewUserService(
	userRepository repositories.UserRepository,
) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (q *userService) FindUsers(
	ctx context.Context,
	request *models.UserFilter,
) ([]*models.User, *models.Pagination, error) {
	ctxt := "UserService-FindUsers"
	users, total, pages, err := q.userRepository.FindUsers(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUsers")
		return nil, nil, err
	}
	n := len(users)
	rows := make([]*models.User, n)
	if n > 0 {
		copy(rows, users)
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

func (q *userService) FindUser(
	ctx context.Context,
	request *models.UserFilter,
) (*models.User, error) {
	ctxt := "UserService-FindUser"
	response, err := q.userRepository.FindUser(ctx, request)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
	}
	return response, err
}

func (q *userService) CreateUser(
	ctx context.Context,
	tx pgx.Tx,
	request *models.User,
) (*models.User, error) {
	ctxt := "UserService-CreateUser"
	response, err := q.userRepository.CreateUser(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateUser")
	}
	return response, err
}

func (q *userService) UpdateUser(
	ctx context.Context,
	tx pgx.Tx,
	request *models.User,
) (*models.User, error) {
	ctxt := "UserService-UpdateUser"
	response, err := q.userRepository.UpdateUser(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateUser")
	}
	return response, err
}
