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
	IdempotencyKeyService interface {
		FindIdempotencyKeys(
			ctx context.Context,
			request *models.IdempotencyKeyFilter,
		) ([]*models.IdempotencyKey, *models.Pagination, error)
		FindIdempotencyKey(
			ctx context.Context,
			request *models.IdempotencyKeyFilter,
		) (*models.IdempotencyKey, error)
		CreateIdempotencyKey(
			ctx context.Context,
			tx pgx.Tx,
			request *models.IdempotencyKey,
		) (*models.IdempotencyKey, error)
		UpdateIdempotencyKey(
			ctx context.Context,
			tx pgx.Tx,
			request *models.IdempotencyKey,
		) (*models.IdempotencyKey, error)
		DeleteIdempotencyKey(
			ctx context.Context,
			tx pgx.Tx,
			request *models.IdempotencyKeyFilter,
		) (int64, error)
	}

	idempotencyKeyService struct {
		idempotencyKeyRepository repositories.IdempotencyKeyRepository
	}
)

func NewIdempotencyKeyService(
	idempotencyKeyRepository repositories.IdempotencyKeyRepository,
) IdempotencyKeyService {
	return &idempotencyKeyService{
		idempotencyKeyRepository: idempotencyKeyRepository,
	}
}

func (q *idempotencyKeyService) FindIdempotencyKeys(
	ctx context.Context,
	request *models.IdempotencyKeyFilter,
) ([]*models.IdempotencyKey, *models.Pagination, error) {
	ctxt := "IdempotencyKeyService-FindIdempotencyKeys"
	idempotencyKeys, total, pages, err := q.idempotencyKeyRepository.FindIdempotencyKeys(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindIdempotencyKeys")
		return nil, nil, err
	}
	n := len(idempotencyKeys)
	rows := make([]*models.IdempotencyKey, n)
	if n > 0 {
		copy(rows, idempotencyKeys)
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

func (q *idempotencyKeyService) FindIdempotencyKey(
	ctx context.Context,
	request *models.IdempotencyKeyFilter,
) (*models.IdempotencyKey, error) {
	ctxt := "IdempotencyKeyService-FindIdempotencyKey"
	response, err := q.idempotencyKeyRepository.FindIdempotencyKey(ctx, request)
	if err != nil && !errors.Is(err, models.ErrIdempotencyKeyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindIdempotencyKey")
	}
	return response, err
}

func (q *idempotencyKeyService) CreateIdempotencyKey(
	ctx context.Context,
	tx pgx.Tx,
	request *models.IdempotencyKey,
) (*models.IdempotencyKey, error) {
	ctxt := "IdempotencyKeyService-CreateIdempotencyKey"
	response, err := q.idempotencyKeyRepository.CreateIdempotencyKey(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateIdempotencyKey")
	}
	return response, err
}

func (q *idempotencyKeyService) UpdateIdempotencyKey(
	ctx context.Context,
	tx pgx.Tx,
	request *models.IdempotencyKey,
) (*models.IdempotencyKey, error) {
	ctxt := "IdempotencyKeyService-UpdateIdempotencyKey"
	response, err := q.idempotencyKeyRepository.UpdateIdempotencyKey(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateIdempotencyKey")
	}
	return response, err
}

func (q *idempotencyKeyService) DeleteIdempotencyKey(
	ctx context.Context,
	tx pgx.Tx,
	request *models.IdempotencyKeyFilter,
) (int64, error) {
	ctxt := "IdempotencyKeyService-DeleteIdempotencyKey"
	response, err := q.idempotencyKeyRepository.DeleteIdempotencyKey(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteIdempotencyKey")
	}
	return response, nil
}
