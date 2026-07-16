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
	JsonWebTokenService interface {
		FindJsonWebTokens(
			ctx context.Context,
			request *models.JsonWebTokenFilter,
		) ([]*models.JsonWebToken, *models.Pagination, error)
		FindJsonWebToken(
			ctx context.Context,
			request *models.JsonWebTokenFilter,
		) (*models.JsonWebToken, error)
		CreateJsonWebToken(
			ctx context.Context,
			tx pgx.Tx,
			request *models.JsonWebToken,
		) (*models.JsonWebToken, error)
		DeleteJsonWebToken(
			ctx context.Context,
			tx pgx.Tx,
			request *models.JsonWebTokenFilter,
		) (int64, error)
	}

	jsonWebTokenService struct {
		jsonWebTokenRepository repositories.JsonWebTokenRepository
	}
)

func NewJsonWebTokenService(
	jsonWebTokenRepository repositories.JsonWebTokenRepository,
) JsonWebTokenService {
	return &jsonWebTokenService{
		jsonWebTokenRepository: jsonWebTokenRepository,
	}
}

func (q *jsonWebTokenService) FindJsonWebTokens(
	ctx context.Context,
	request *models.JsonWebTokenFilter,
) ([]*models.JsonWebToken, *models.Pagination, error) {
	ctxt := "JsonWebTokenService-FindJsonWebTokens"
	jsonWebTokens, total, pages, err := q.jsonWebTokenRepository.FindJsonWebTokens(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindJsonWebTokens")
		return nil, nil, err
	}
	n := len(jsonWebTokens)
	rows := make([]*models.JsonWebToken, n)
	if n > 0 {
		copy(rows, jsonWebTokens)
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

func (q *jsonWebTokenService) FindJsonWebToken(
	ctx context.Context,
	request *models.JsonWebTokenFilter,
) (*models.JsonWebToken, error) {
	ctxt := "JsonWebTokenService-FindJsonWebToken"
	response, err := q.jsonWebTokenRepository.FindJsonWebToken(ctx, request)
	if err != nil && !errors.Is(err, models.ErrJsonWebTokenNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindJsonWebToken")
	}
	return response, err
}

func (q *jsonWebTokenService) CreateJsonWebToken(
	ctx context.Context,
	tx pgx.Tx,
	request *models.JsonWebToken,
) (*models.JsonWebToken, error) {
	ctxt := "JsonWebTokenService-CreateJsonWebToken"
	response, err := q.jsonWebTokenRepository.CreateJsonWebToken(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateJsonWebToken")
	}
	return response, err
}

func (q *jsonWebTokenService) DeleteJsonWebToken(
	ctx context.Context,
	tx pgx.Tx,
	request *models.JsonWebTokenFilter,
) (int64, error) {
	ctxt := "JsonWebTokenService-DeleteJsonWebToken"
	response, err := q.jsonWebTokenRepository.DeleteJsonWebToken(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteJsonWebToken")
	}
	return response, nil
}
