package repositories

import (
	"context"
	"database/sql"
	"log/slog"

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
	dbWrite *sql.DB
}

func NewTransactionCounterRepository(
	dbRead,
	dbWrite *sql.DB,
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
	if err := q.dbWrite.QueryRowContext(
		ctx,
		`INSERT INTO transaction_counters (
			transaction_window
			, hits
			, created_at
			, updated_at
		) VALUES (?, 1, ?, ?)
		ON DUPLICATE KEY UPDATE
			hits = hits + 1
			, updated_at = VALUES(updated_at)
		RETURNING id
			, transaction_window
			, hits
			, created_at
			, updated_at`,
		helper.SQLArgs(
			request.TransactionWindow,
			request.CreatedAt,
			request.UpdatedAt,
		)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.TransactionWindow,
			&response.Hits,
			&response.CreatedAt,
			&response.UpdatedAt,
		)...,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
