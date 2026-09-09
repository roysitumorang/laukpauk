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

type OneTimePasswordService interface {
	FindOneTimePasswords(
		ctx context.Context,
		request *models.OneTimePasswordFilter,
	) ([]*models.OneTimePassword, *models.Pagination, error)
	FindOneTimePassword(
		ctx context.Context,
		request *models.OneTimePasswordFilter,
	) (*models.OneTimePassword, error)
	CreateOneTimePassword(
		ctx context.Context,
		tx pgx.Tx,
		request *models.OneTimePassword,
	) (*models.OneTimePassword, error)
	UpdateOneTimePassword(
		ctx context.Context,
		tx pgx.Tx,
		request *models.OneTimePassword,
	) (*models.OneTimePassword, error)
}

type oneTimePasswordService struct {
	oneTimePasswordRepository repositories.OneTimePasswordRepository
}

func NewOneTimePasswordService(
	oneTimePasswordRepository repositories.OneTimePasswordRepository,
) OneTimePasswordService {
	return &oneTimePasswordService{
		oneTimePasswordRepository: oneTimePasswordRepository,
	}
}

func (q *oneTimePasswordService) FindOneTimePasswords(
	ctx context.Context,
	request *models.OneTimePasswordFilter,
) ([]*models.OneTimePassword, *models.Pagination, error) {
	ctxt := "OneTimePasswordService-FindOneTimePasswords"
	oneTimePasswords, total, pages, err := q.oneTimePasswordRepository.FindOneTimePasswords(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindOneTimePasswords")
		return nil, nil, err
	}
	n := len(oneTimePasswords)
	rows := make([]*models.OneTimePassword, n)
	if n > 0 {
		copy(rows, oneTimePasswords)
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

func (q *oneTimePasswordService) FindOneTimePassword(
	ctx context.Context,
	request *models.OneTimePasswordFilter,
) (*models.OneTimePassword, error) {
	ctxt := "OneTimePasswordService-FindOneTimePassword"
	response, err := q.oneTimePasswordRepository.FindOneTimePassword(ctx, request)
	if err != nil && !errors.Is(err, models.ErrOneTimePasswordNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindOneTimePassword")
	}
	return response, err
}

func (q *oneTimePasswordService) CreateOneTimePassword(
	ctx context.Context,
	tx pgx.Tx,
	request *models.OneTimePassword,
) (*models.OneTimePassword, error) {
	ctxt := "OneTimePasswordService-CreateOneTimePassword"
	response, err := q.oneTimePasswordRepository.CreateOneTimePassword(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateOneTimePassword")
	}
	return response, err
}

func (q *oneTimePasswordService) UpdateOneTimePassword(
	ctx context.Context,
	tx pgx.Tx,
	request *models.OneTimePassword,
) (*models.OneTimePassword, error) {
	ctxt := "OneTimePasswordService-UpdateOneTimePassword"
	response, err := q.oneTimePasswordRepository.UpdateOneTimePassword(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrOneTimePasswordNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateOneTimePassword")
	}
	return response, err
}
