package services

import (
	"context"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	TransactionCounterService interface {
		SaveTransactionCounter(
			ctx context.Context,
			request *models.TransactionCounter,
		) (*models.TransactionCounter, error)
	}

	transactionCounterService struct {
		transactionCounterRepository repositories.TransactionCounterRepository
	}
)

func NewTransactionCounterService(
	transactionCounterRepository repositories.TransactionCounterRepository,
) TransactionCounterService {
	return &transactionCounterService{
		transactionCounterRepository: transactionCounterRepository,
	}
}

func (q *transactionCounterService) SaveTransactionCounter(
	ctx context.Context,
	request *models.TransactionCounter,
) (*models.TransactionCounter, error) {
	ctxt := "TransactionCounterService-SaveTransactionCounter"
	response, err := q.transactionCounterRepository.SaveTransactionCounter(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveTransactionCounter")
	}
	return response, err
}
