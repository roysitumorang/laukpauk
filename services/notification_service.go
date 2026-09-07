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

type NotificationService interface {
	FindNotifications(
		ctx context.Context,
		request *models.NotificationFilter,
	) ([]*models.Notification, *models.Pagination, error)
	FindNotification(
		ctx context.Context,
		request *models.NotificationFilter,
	) (*models.Notification, error)
	CreateNotification(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Notification,
	) (*models.Notification, error)
	UpdateNotification(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Notification,
	) (*models.Notification, error)
	ConsumeMessage(
		ctx context.Context,
		topic models.Topic,
		message models.TopicMessage,
	) error
}

type notificationService struct {
	notificationRepository       repositories.NotificationRepository
	notificationTargetRepository repositories.NotificationTargetRepository
	kafkaClient                  *externals.KafkaClient
	fcmClient                    *externals.FcmClient
}

func NewNotificationService(
	notificationRepository repositories.NotificationRepository,
	notificationTargetRepository repositories.NotificationTargetRepository,
	kafkaClient *externals.KafkaClient,
	fcmClient *externals.FcmClient,
) NotificationService {
	return &notificationService{
		notificationRepository:       notificationRepository,
		notificationTargetRepository: notificationTargetRepository,
		kafkaClient:                  kafkaClient,
		fcmClient:                    fcmClient,
	}
}

func (q *notificationService) FindNotifications(
	ctx context.Context,
	request *models.NotificationFilter,
) ([]*models.Notification, *models.Pagination, error) {
	ctxt := "NotificationService-FindNotifications"
	notifications, total, pages, err := q.notificationRepository.FindNotifications(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotifications")
		return nil, nil, err
	}
	n := len(notifications)
	rows := make([]*models.Notification, n)
	if n > 0 {
		copy(rows, notifications)
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

func (q *notificationService) FindNotification(
	ctx context.Context,
	request *models.NotificationFilter,
) (*models.Notification, error) {
	ctxt := "NotificationService-FindNotification"
	response, err := q.notificationRepository.FindNotification(ctx, request)
	if err != nil && !errors.Is(err, models.ErrNotificationNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotification")
	}
	return response, err
}

func (q *notificationService) CreateNotification(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Notification,
) (*models.Notification, error) {
	ctxt := "NotificationService-CreateNotification"
	response, err := q.notificationRepository.CreateNotification(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateNotification")
		return nil, err
	}
	return response, nil
}

func (q *notificationService) UpdateNotification(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Notification,
) (*models.Notification, error) {
	ctxt := "NotificationService-UpdateNotification"
	response, err := q.notificationRepository.UpdateNotification(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateNotification")
		return nil, err
	}
	return response, nil
}

func (q *notificationService) ConsumeMessage(
	ctx context.Context,
	topic models.Topic,
	message models.TopicMessage,
) error {
	ctxt := "NotificationService-ConsumeMessage"
	if topic != models.TopicNotification {
		return nil
	}
	switch message.Event {
	case models.TopicNotificationEventSend:
		notification, err := q.notificationRepository.FindNotification(
			ctx,
			models.NewNotificationFilter().
				WithNotificationIDs(message.ID).
				WithSentAtBlank(new(true)),
		)
		if err != nil && errors.Is(err, models.ErrNotificationNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotification")
			return err
		}
		if notification == nil {
			return nil
		}
		notificationTargetFilter := models.NewNotificationTargetFilter()
		if notification.AdminID.Compare(uuid.Nil()) != 0 {
			notificationTargetFilter.WithAdminIDs(notification.AdminID)
		} else if notification.SellerID.Compare(uuid.Nil()) != 0 {
			notificationTargetFilter.WithAdminIDs(notification.SellerID)
		} else if notification.BuyerID.Compare(uuid.Nil()) != 0 {
			notificationTargetFilter.WithAdminIDs(notification.BuyerID)
		}
		notificationTargets, _, _, err := q.notificationTargetRepository.FindNotificationTargets(
			ctx,
			notificationTargetFilter,
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotificationTargets")
			return err
		}
		if len(notificationTargets) == 0 {
			return nil
		}
		var photo string
		if notification.Photo != nil {
			photo = *notification.Photo
		}
		for _, notificationTarget := range notificationTargets {
			if _, err = q.fcmClient.Send(
				ctx,
				notificationTarget.FirebaseInstallationID,
				notification.Subject,
				notification.Body,
				photo,
				notification.Params,
			); err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSend")
				return err
			}
		}
		notification.SentAt = new(time.Now())
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
		if _, err = q.notificationRepository.UpdateNotification(ctx, tx, notification); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateNotification")
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return err
		}
	}
	return nil
}
