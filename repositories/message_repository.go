package repositories

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type MessageRepository interface {
	CreateMessage(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Message,
	) (*models.Message, error)
}

type messageRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewMessageRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) MessageRepository {
	return &messageRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *messageRepository) CreateMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Message,
) (*models.Message, error) {
	ctxt := "MessageRepository-CreateMessage"
	var response models.Message
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO messages (
			name
			, mobile_phone
			, body
			, created_at
		) VALUES ($1, $2, $3, $4)
		RETURNING id
			, name
			, mobile_phone
			, body
			, created_at`,
		request.Name,
		request.MobilePhone,
		request.Body,
		request.CreatedAt,
	).Scan(
		&response.ID,
		&response.Name,
		&response.MobilePhone,
		&response.Body,
		&response.CreatedAt,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
