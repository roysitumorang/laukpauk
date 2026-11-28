package repositories

import (
	"context"
	"log/slog"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type TransactionCounterRepository interface {
	SaveTransactionCounter(
		ctx context.Context,
		request *models.TransactionCounter,
	) (*models.TransactionCounter, error)
}

type transactionCounterRepository struct {
	dbRead,
	dbWrite *database.Pool
}

func NewTransactionCounterRepository(
	dbRead,
	dbWrite *database.Pool,
) TransactionCounterRepository {
	return &transactionCounterRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *transactionCounterRepository) SaveTransactionCounter(
	ctx context.Context,
	request *models.TransactionCounter,
) (*models.TransactionCounter, error) {
	ctxt := "TransactionCounterRepository-SaveTransactionCounter"
	var response models.TransactionCounter
	if err := q.dbWrite.QueryRow(
		ctx,
		`INSERT INTO transaction_counters (
			transaction_window
			, hits
			, created_at
			, updated_at
		) VALUES ($1, 1, $2, $3)
		ON DUPLICATE KEY UPDATE
			hits = hits + 1
			, updated_at = $3
		RETURNING id
			, transaction_window
			, hits
			, created_at
			, updated_at`,
		request.TransactionWindow,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.TransactionWindow,
		&response.Hits,
		&response.CreatedAt,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
