package repositories

import (
	"context"
	"log/slog"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type MessageRepository interface {
	CreateMessage(
		ctx context.Context,
		tx database.Tx,
		request *models.Message,
	) (*models.Message, error)
}

type messageRepository struct {
	dbRead,
	dbWrite *database.Pool
}

func NewMessageRepository(
	dbRead,
	dbWrite *database.Pool,
) MessageRepository {
	return &messageRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *messageRepository) CreateMessage(
	ctx context.Context,
	tx database.Tx,
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
		) VALUES (?, ?, ?, ?)
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
