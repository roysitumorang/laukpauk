package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	SmsMessageService interface {
		FindSmsMessages(
			ctx context.Context,
			request *models.SmsMessageFilter,
		) ([]*models.SmsMessage, *models.Pagination, error)
		FindSmsMessage(
			ctx context.Context,
			request *models.SmsMessageFilter,
		) (*models.SmsMessage, error)
		CreateSmsMessage(
			ctx context.Context,
			tx pgx.Tx,
			request *models.SmsMessage,
		) (*models.SmsMessage, error)
		UpdateSmsMessage(
			ctx context.Context,
			tx pgx.Tx,
			request *models.SmsMessage,
		) (*models.SmsMessage, error)
		ConsumeMessage(
			ctx context.Context,
			topic models.Topic,
			message models.TopicMessage,
		) error
	}

	smsMessageService struct {
		smsMessageRepository repositories.SmsMessageRepository
		kafkaClient          *externals.KafkaClient
		zenzivaClient        *externals.ZenzivaClient
	}
)

func NewSmsMessageService(
	smsMessageRepository repositories.SmsMessageRepository,
	kafkaClient *externals.KafkaClient,
	zenzivaClient *externals.ZenzivaClient,
) SmsMessageService {
	return new(smsMessageService{
		smsMessageRepository: smsMessageRepository,
		kafkaClient:          kafkaClient,
		zenzivaClient:        zenzivaClient,
	})
}

func (q *smsMessageService) FindSmsMessages(
	ctx context.Context,
	request *models.SmsMessageFilter,
) ([]*models.SmsMessage, *models.Pagination, error) {
	ctxt := "SmsMessageService-FindSmsMessages"
	smsMessages, total, pages, err := q.smsMessageRepository.FindSmsMessages(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSmsMessages")
		return nil, nil, err
	}
	n := len(smsMessages)
	rows := make([]*models.SmsMessage, n)
	if n > 0 {
		copy(rows, smsMessages)
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

func (q *smsMessageService) FindSmsMessage(
	ctx context.Context,
	request *models.SmsMessageFilter,
) (*models.SmsMessage, error) {
	ctxt := "SmsMessageService-FindSmsMessage"
	response, err := q.smsMessageRepository.FindSmsMessage(ctx, request)
	if err != nil && !errors.Is(err, models.ErrSmsMessageNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSmsMessage")
	}
	return response, err
}

func (q *smsMessageService) CreateSmsMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.SmsMessage,
) (*models.SmsMessage, error) {
	ctxt := "SmsMessageService-CreateSmsMessage"
	response, err := q.smsMessageRepository.CreateSmsMessage(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSmsMessage")
		return nil, err
	}
	return response, nil
}

func (q *smsMessageService) UpdateSmsMessage(
	ctx context.Context,
	tx pgx.Tx,
	request *models.SmsMessage,
) (*models.SmsMessage, error) {
	ctxt := "SmsMessageService-UpdateSmsMessage"
	response, err := q.smsMessageRepository.UpdateSmsMessage(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSmsMessage")
		return nil, err
	}
	return response, nil
}

func (q *smsMessageService) ConsumeMessage(
	ctx context.Context,
	topic models.Topic,
	message models.TopicMessage,
) error {
	ctxt := "SmsMessageService-ConsumeMessage"
	if topic != models.TopicSms {
		return nil
	}
	switch message.Event {
	case models.TopicSmsEventSend:
		smsMessage, err := q.smsMessageRepository.FindSmsMessage(
			ctx,
			models.NewSmsMessageFilter().
				WithSmsMessageIDs(message.ID).
				WithDeliveryStatusBlank(),
		)
		if err != nil && errors.Is(err, models.ErrSmsMessageNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSmsMessage")
			return err
		}
		if smsMessage == nil {
			return nil
		}
		var messageID *int64
		if smsMessage.MessageType == models.SmsMessageTypeOtp {
			messageID, err = q.zenzivaClient.SendOTP(ctx, smsMessage.MobilePhone, smsMessage.Body)
		} else {
			messageID, err = q.zenzivaClient.SendSMS(ctx, smsMessage.MobilePhone, smsMessage.Body)
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSend")
			return err
		}
		if messageID == nil {
			return nil
		}
		smsMessage.MessageID = messageID
		smsMessage.SentAt = new(time.Now())
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
		if _, err = q.smsMessageRepository.UpdateSmsMessage(ctx, tx, smsMessage); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSmsMessage")
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return err
		}
	case models.TopicSmsEventCheckReport:
		smsMessage, err := q.smsMessageRepository.FindSmsMessage(
			ctx,
			models.NewSmsMessageFilter().
				WithSmsMessageIDs(message.ID).
				WithDeliveryStatusBlank(),
		)
		if err != nil && errors.Is(err, models.ErrSmsMessageNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSmsMessage")
			return err
		}
		if smsMessage == nil ||
			smsMessage.MessageID == nil {
			return nil
		}
		deliveryStatus, err := q.zenzivaClient.CheckReport(ctx, *smsMessage.MessageID)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCheckReport")
			return err
		}
		if deliveryStatus == nil {
			return nil
		}
		smsMessage.DeliveryStatus = deliveryStatus
		smsMessage.UpdatedAt = time.Now()
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
		if _, err = q.smsMessageRepository.UpdateSmsMessage(ctx, tx, smsMessage); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSmsMessage")
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return err
		}
	}
	return nil
}
