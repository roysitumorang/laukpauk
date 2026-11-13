package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	notificationController struct {
		notificationService services.NotificationService
		kafkaClient         *externals.KafkaClient
		userService         services.UserService
		jsonWebTokenService services.JsonWebTokenService
	}
)

func NewNotificationController(
	notificationService services.NotificationService,
	kafkaClient *externals.KafkaClient,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *notificationController {
	return &notificationController{
		notificationService: notificationService,
		kafkaClient:         kafkaClient,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *notificationController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Post("/:id/read", q.V1SellerRead)
}

func (q *notificationController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-V1SellerIndex"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, err := sanitizers.FindNotifications(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotifications")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.notificationService.FindNotifications(
		ctx,
		request.WithUserIDs(currentUser.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotifications")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	pagination.Rows = make([]any, len(rows))
	for i := range rows {
		pagination.Rows[i] = rows[i]
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}

func (q *notificationController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateNotification(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNotification")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	recipient, err := q.userService.FindUser(
		ctx,
		models.NewUserFilter().
			WithRoles(models.RoleBuyer).
			WithStatuses(models.UserStatusActive).
			WithUserIDs(request.UserID),
	)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindUser")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if recipient == nil {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("user_id: not found").
			WriteResponse(c)
	}
	request.NotificationType = models.NotificationTypeOther
	request.Params = map[string]string{}
	request.CreatedBy = currentUser.ID
	request.CreatedAt = time.Now()
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	notification, err := q.notificationService.CreateNotification(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateNotification")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = q.kafkaClient.Publish(ctx, models.TopicMessage{
		Topic: models.TopicNotification,
		Event: models.TopicNotificationEventSend,
		ID:    notification.ID,
	}); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("push notification created successfully").
		SetData(notification).
		WriteResponse(c)
}

func (q *notificationController) V1SellerRead(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-V1SellerRead"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrNotificationNotFound.Error()).
			WriteResponse(c)
	}
	notification, err := q.notificationService.FindNotification(
		ctx,
		models.NewNotificationFilter().
			WithUserIDs(currentUser.ID).
			WithNotificationIDs(notificationID).
			WithReadAtBlank(new(true)),
	)
	if err != nil && !errors.Is(err, models.ErrNotificationNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotification")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if notification == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrNotificationNotFound.Error()).
			WriteResponse(c)
	}
	notification.ReadAt = new(time.Now())
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	saved, err := q.notificationService.UpdateNotification(ctx, tx, notification)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateNotification")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("notification marked as read successfully").
		SetData(saved).
		WriteResponse(c)
}
