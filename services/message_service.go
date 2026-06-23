package services

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	MessageService interface {
		CreateMessage(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Message,
		) (*models.Message, error)
	}

	messageService struct {
		messageRepository repositories.MessageRepository
	}
)

func NewMessageService(
	messageRepository repositories.MessageRepository,
) MessageService {
	return new(messageService{
		messageRepository: messageRepository,
	})
}

func (q *messageService) CreateMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Message,
) (*models.Message, error) {
	ctxt := "MessageService-CreateMessage"
	response, err := q.messageRepository.CreateMessage(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateMessage")
		return nil, err
	}
	return response, nil
}
