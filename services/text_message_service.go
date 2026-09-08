package services

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type TextMessageService interface {
	FindTextMessages(
		ctx context.Context,
		request *models.TextMessageFilter,
	) ([]*models.TextMessage, *models.Pagination, error)
	FindTextMessage(
		ctx context.Context,
		request *models.TextMessageFilter,
	) (*models.TextMessage, error)
	CreateTextMessage(
		ctx context.Context,
		tx pgx.Tx,
		request *models.TextMessage,
	) (*models.TextMessage, error)
	UpdateTextMessage(
		ctx context.Context,
		tx pgx.Tx,
		request *models.TextMessage,
	) (*models.TextMessage, error)
	DeleteTextMessage(
		ctx context.Context,
		tx pgx.Tx,
		textMessageID uuid.UUID,
	) (int64, error)
	ConsumeMessage(
		ctx context.Context,
		topic models.Topic,
		message models.TopicMessage,
	) error
}

type textMessageService struct {
	textMessageRepository repositories.TextMessageRepository
	kafkaClient           *externals.KafkaClient
	zenzivaClient         *externals.ZenzivaClient
}

func NewTextMessageService(
	textMessageRepository repositories.TextMessageRepository,
	kafkaClient *externals.KafkaClient,
	zenzivaClient *externals.ZenzivaClient,
) TextMessageService {
	return &textMessageService{
		textMessageRepository: textMessageRepository,
		kafkaClient:           kafkaClient,
		zenzivaClient:         zenzivaClient,
	}
}

func (q *textMessageService) FindTextMessages(
	ctx context.Context,
	request *models.TextMessageFilter,
) ([]*models.TextMessage, *models.Pagination, error) {
	ctxt := "TextMessageService-FindTextMessages"
	textMessages, total, pages, err := q.textMessageRepository.FindTextMessages(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTextMessages")
		return nil, nil, err
	}
	n := len(textMessages)
	rows := make([]*models.TextMessage, n)
	if n > 0 {
		copy(rows, textMessages)
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

func (q *textMessageService) FindTextMessage(
	ctx context.Context,
	request *models.TextMessageFilter,
) (*models.TextMessage, error) {
	ctxt := "TextMessageService-FindTextMessage"
	response, err := q.textMessageRepository.FindTextMessage(ctx, request)
	if err != nil && !errors.Is(err, models.ErrTextMessageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTextMessage")
	}
	return response, err
}

func (q *textMessageService) CreateTextMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TextMessage,
) (*models.TextMessage, error) {
	ctxt := "TextMessageService-CreateTextMessage"
	response, err := q.textMessageRepository.CreateTextMessage(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTextMessage")
		return nil, err
	}
	return response, nil
}

func (q *textMessageService) UpdateTextMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TextMessage,
) (*models.TextMessage, error) {
	ctxt := "TextMessageService-UpdateTextMessage"
	response, err := q.textMessageRepository.UpdateTextMessage(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTextMessage")
		return nil, err
	}
	return response, nil
}

func (q *textMessageService) DeleteTextMessage(
	ctx context.Context,
	tx pgx.Tx,
	textMessageID uuid.UUID,
) (int64, error) {
	ctxt := "TextMessageService-DeleteTextMessage"
	response, err := q.textMessageRepository.DeleteTextMessage(ctx, tx, textMessageID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteTextMessage")
	}
	return response, err
}

func (q *textMessageService) ConsumeMessage(
	ctx context.Context,
	topic models.Topic,
	message models.TopicMessage,
) error {
	ctxt := "TextMessageService-ConsumeMessage"
	if topic != models.TopicTextMessage {
		return nil
	}
	switch message.Event {
	case models.TopicTextMessageEventSend:
		textMessage, err := q.textMessageRepository.FindTextMessage(
			ctx,
			models.NewTextMessageFilter().
				WithTextMessageIDs(message.ID).
				WithDeliveryStatusBlank(),
		)
		if err != nil && errors.Is(err, models.ErrTextMessageNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTextMessage")
			return err
		}
		if textMessage == nil {
			return nil
		}
		var messageID *int64
		if textMessage.MessageType == models.TextMessageTypeOtp {
			messageID, err = q.zenzivaClient.SendOTP(ctx, textMessage.MobilePhone, textMessage.Body)
		} else {
			messageID, err = q.zenzivaClient.SendSMS(ctx, textMessage.MobilePhone, textMessage.Body)
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSend")
			return err
		}
		if messageID == nil {
			return nil
		}
		textMessage.MessageID = messageID
		textMessage.SentAt = new(time.Now())
		tx, err := repositories.BeginTx(ctx)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
			return err
		}
		defer func(ctx context.Context) {
			if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
				err = nil
			}
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
			}
		}(ctx)
		if _, err = q.textMessageRepository.UpdateTextMessage(ctx, tx, textMessage); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTextMessage")
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return err
		}
	case models.TopicTextMessageEventCheckReport:
		textMessage, err := q.textMessageRepository.FindTextMessage(
			ctx,
			models.NewTextMessageFilter().
				WithTextMessageIDs(message.ID).
				WithDeliveryStatusBlank(),
		)
		if err != nil && errors.Is(err, models.ErrTextMessageNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTextMessage")
			return err
		}
		if textMessage == nil ||
			textMessage.MessageID == nil {
			return nil
		}
		deliveryStatus, err := q.zenzivaClient.CheckReport(ctx, *textMessage.MessageID)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCheckReport")
			return err
		}
		if deliveryStatus == nil {
			return nil
		}
		textMessage.DeliveryStatus = deliveryStatus
		textMessage.UpdatedAt = time.Now()
		tx, err := repositories.BeginTx(ctx)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
			return err
		}
		defer func(ctx context.Context) {
			if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
				err = nil
			}
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
			}
		}(ctx)
		if _, err = q.textMessageRepository.UpdateTextMessage(ctx, tx, textMessage); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTextMessage")
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return err
		}
	}
	return nil
}
