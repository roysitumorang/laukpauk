package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
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
		adminService        services.AdminService
		buyerService        services.BuyerService
		notificationService services.NotificationService
		kafkaClient         *externals.KafkaClient
		userService         services.UserService
		jsonWebTokenService services.JsonWebTokenService
	}
)

func NewNotificationController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	notificationService services.NotificationService,
	kafkaClient *externals.KafkaClient,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *notificationController {
	return &notificationController{
		adminService:        adminService,
		buyerService:        buyerService,
		notificationService: notificationService,
		kafkaClient:         kafkaClient,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *notificationController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate)
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Post("/:id/read", q.V1SellerRead)
}

func (q *notificationController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.NotificationFilter{}
	rows := []*models.Notification{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Push Notifikasi",
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
		"sortings", []models.Sorting{
			{
				Field:     models.SortingFieldID,
				Direction: models.SortingDirectionDesc,
				Label:     "Newest-Oldest",
			},
			{
				Field:     models.SortingFieldID,
				Direction: models.SortingDirectionAsc,
				Label:     "Oldest-Newest",
			},
		},
	)
	request, err = sanitizers.FindNotifications(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotifications")
		return c.Render("/admin/notification_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.notificationService.FindNotifications(
		ctx,
		request.WithAdminIDs(currentAdmin.ID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotifications")
		return c.Render("/admin/notification_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/notification_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *notificationController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-AdminNew"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Notification{}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Kirim Notifikasi",
		"request", request,
		"notification", request,
		"buyers", buyers,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithSortings(models.Sorting{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/notification_new", params.Push("buyers", buyers), "admin/layouts/authenticated")
}

func (q *notificationController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Notification{}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Kirim Notifikasi",
		"request", request,
		"notification", request,
		"buyers", buyers,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithSortings(models.Sorting{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	mapBuyerIDs := map[uuid.UUID]*models.Buyer{}
	for _, buyer := range buyers {
		mapBuyerIDs[buyer.ID] = buyer
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateNotification(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNotification")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	buyer, ok := mapBuyerIDs[request.BuyerID]
	if !ok {
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger("buyer_id: not found"),
		), "admin/layouts/authenticated")
	}
	request.NotificationType = models.NotificationTypeOther
	request.Params = map[string]string{}
	request.CreatedByAdminID = currentAdmin.ID
	request.CreatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
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
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	buyer.UnreadNotifications = 1
	if _, err = q.buyerService.UpdateBuyer(ctx, tx, buyer); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/notification_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(nil),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = q.kafkaClient.Publish(ctx, models.TopicMessage{
		Topic: models.TopicNotification,
		Event: models.TopicNotificationEventSend,
		ID:    notification.ID,
	}); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
	}
	return flash.Clear(c, sess.Session).
		Success("push notification created successfully").
		Redirect(c, sess.Session, "/admin/notifications")
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
		request.WithSellerIDs(currentUser.ID),
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
	var buyers []*models.Buyer
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithSortings(models.Sorting{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	mapBuyerIDs := map[uuid.UUID]*models.Buyer{}
	for _, buyer := range buyers {
		mapBuyerIDs[buyer.ID] = buyer
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateNotification(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNotification")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer, ok := mapBuyerIDs[request.BuyerID]
	if !ok {
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage("buyer_id: not found").
			WriteResponse(c)
	}
	request.NotificationType = models.NotificationTypeOther
	request.Params = map[string]string{}
	request.CreatedBySellerID = currentUser.ID
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
	buyer.UnreadNotifications = 1
	saved, err := q.buyerService.UpdateBuyer(ctx, tx, buyer)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
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
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("push notification created successfully").
		SetData(saved).
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
			WithSellerIDs(currentUser.ID).
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
