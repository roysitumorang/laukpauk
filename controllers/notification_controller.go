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
		sellerService       services.SellerService
		buyerService        services.BuyerService
		notificationService services.NotificationService
		kafkaClient         *externals.KafkaClient
	}
)

func NewNotificationController(
	adminService services.AdminService,
	sellerService services.SellerService,
	buyerService services.BuyerService,
	notificationService services.NotificationService,
	kafkaClient *externals.KafkaClient,
) *notificationController {
	return &notificationController{
		adminService:        adminService,
		sellerService:       sellerService,
		buyerService:        buyerService,
		notificationService: notificationService,
		kafkaClient:         kafkaClient,
	}
}

func (q *notificationController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate)
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Post("/:id/read", q.SellerRead)
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

func (q *notificationController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
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
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/notification_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.notificationService.FindNotifications(
		ctx,
		request.WithSellerIDs(currentSeller.ID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotifications")
		return c.Render("/seller/notification_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("/seller/notification_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *notificationController) SellerNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-SellerNew"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
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
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("/seller/notification_new", params.Push("buyers", buyers), "seller/layouts/authenticated")
}

func (q *notificationController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
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
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
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
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	buyer, ok := mapBuyerIDs[request.BuyerID]
	if !ok {
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger("buyer_id: not found"),
		), "seller/layouts/authenticated")
	}
	request.NotificationType = models.NotificationTypeOther
	request.Params = map[string]string{}
	request.CreatedBySellerID = currentSeller.ID
	request.CreatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
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
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	buyer.UnreadNotifications = 1
	if _, err = q.buyerService.UpdateBuyer(ctx, tx, buyer); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if photo != nil && request.Photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, *request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/notification_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(nil),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/seller/notification_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
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
		Redirect(c, sess.Session, "/seller/notifications")
}

func (q *notificationController) SellerRead(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NotificationController-SellerRead"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrNotificationNotFound.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	notification, err := q.notificationService.FindNotification(
		ctx,
		models.NewNotificationFilter().
			WithSellerIDs(currentSeller.ID).
			WithNotificationIDs(notificationID).
			WithReadAtBlank(new(true)),
	)
	if err != nil && !errors.Is(err, models.ErrNotificationNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNotification")
		return flash.Danger(err.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	if notification == nil {
		return flash.Danger(models.ErrNotificationNotFound.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	notification.ReadAt = new(time.Now())
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.notificationService.UpdateNotification(ctx, tx, notification); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateNotification")
		return flash.Danger(err.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateNotification")
		return flash.Danger(err.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	currentSeller.UnreadNotifications = -1
	if _, err = q.sellerService.UpdateSeller(ctx, tx, currentSeller); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSeller")
		return flash.Danger(err.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).
			Redirect(c, sess.Session, "/seller/notifications")
	}
	return flash.Clear(c, sess.Session).
		Success("notification marked as read successfully").
		Redirect(c, sess.Session, "/seller/notifications")
}
