package repositories

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type MessageRepository interface {
	CreateMessage(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Message,
	) (*models.Message, error)
}

type messageRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewMessageRepository(
	dbRead,
	dbWrite *sql.DB,
) MessageRepository {
	return &messageRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *messageRepository) CreateMessage(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Message,
) (*models.Message, error) {
	ctxt := "MessageRepository-CreateMessage"
	var response models.Message
	if err := tx.QueryRowContext(
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
		helper.SQLArgs(
			request.Name,
			request.MobilePhone,
			request.Body,
			request.CreatedAt,
		)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.Name,
			&response.MobilePhone,
			&response.Body,
			&response.CreatedAt,
		)...,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
