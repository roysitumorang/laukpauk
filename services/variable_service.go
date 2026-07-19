package services

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	VariableService interface {
		FindVariables(
			ctx context.Context,
			request *models.VariableFilter,
		) ([]*models.Variable, *models.Pagination, error)
		FindVariable(
			ctx context.Context,
			request *models.VariableFilter,
		) (*models.Variable, error)
		FindVariableSetting(
			ctx context.Context,
		) (*models.VariableSetting, error)
		SaveVariable(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Variable,
		) (*models.Variable, error)
	}

	variableService struct {
		variableRepository repositories.VariableRepository
	}
)

func NewVariableService(
	variableRepository repositories.VariableRepository,
) VariableService {
	return &variableService{
		variableRepository: variableRepository,
	}
}

func (q *variableService) FindVariables(
	ctx context.Context,
	request *models.VariableFilter,
) ([]*models.Variable, *models.Pagination, error) {
	ctxt := "VariableService-FindVariables"
	variables, total, pages, err := q.variableRepository.FindVariables(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVariables")
		return nil, nil, err
	}
	n := len(variables)
	rows := make([]*models.Variable, n)
	if n > 0 {
		copy(rows, variables)
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

func (q *variableService) FindVariable(
	ctx context.Context,
	request *models.VariableFilter,
) (*models.Variable, error) {
	ctxt := "VariableService-FindVariable"
	response, err := q.variableRepository.FindVariable(ctx, request)
	if err != nil && !errors.Is(err, models.ErrVariableNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVariable")
	}
	return response, err
}

func (q *variableService) FindVariableSetting(
	ctx context.Context,
) (*models.VariableSetting, error) {
	ctxt := "VariableService-FindVariableSetting"
	variable, err := q.variableRepository.FindVariable(
		ctx,
		models.NewVariableFilter().
			WithNames(models.VariableNameSetting),
	)
	if err != nil && !errors.Is(err, models.ErrVariableNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVariable")
		return nil, err
	}
	if variable == nil {
		return nil, models.ErrVariableNotFound
	}
	valueByte, err := json.Marshal(variable.Value)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrMarshal")
		return nil, err
	}
	var response models.VariableSetting
	if err = json.Unmarshal(valueByte, &response); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	return &response, err
}

func (q *variableService) SaveVariable(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Variable,
) (*models.Variable, error) {
	ctxt := "VariableService-SaveVariable"
	response, err := q.variableRepository.SaveVariable(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveVariable")
	}
	return response, err
}
