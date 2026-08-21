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
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	textMessageController struct {
		adminService       services.AdminService
		buyerService       services.BuyerService
		textMessageService services.TextMessageService
	}
)

func NewTextMessageController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	textMessageService services.TextMessageService,
) *textMessageController {
	return &textMessageController{
		adminService:       adminService,
		buyerService:       buyerService,
		textMessageService: textMessageService,
	}
}

func (q *textMessageController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate)
}

func (q *textMessageController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TextMessageController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.TextMessageFilter{}
	rows := []*models.TextMessage{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Riwayat SMS",
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
	request, err = sanitizers.FindTextMessages(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTextMessages")
		return c.Render("/admin/text_message_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.textMessageService.FindTextMessages(
		ctx,
		request.WithMessageTypes(models.TextMessageTypeSms),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTextMessages")
		return c.Render("/admin/text_message_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/text_message_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *textMessageController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TextMessageController-AdminNew"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.TextMessage{}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Kirim SMS",
		"request", request,
		"textMessage", request,
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
		return c.Render("/admin/text_message_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/text_message_new", params.Push("buyers", buyers), "admin/layouts/authenticated")
}

func (q *textMessageController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TextMessageController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.TextMessage{}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Kirim SMS",
		"request", request,
		"textMessage", request,
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
		return c.Render("/admin/text_message_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	mapBuyerMobilePhones := map[uuid.UUID]string{}
	for _, buyer := range buyers {
		if buyer.MobilePhone != nil {
			mapBuyerMobilePhones[buyer.ID] = *buyer.MobilePhone
		}
	}
	request, statusCode, err := sanitizers.ValidateTextMessage(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTextMessage")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/text_message_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	mobilePhone, ok := mapBuyerMobilePhones[request.BuyerID]
	if !ok {
		return c.Render("/admin/text_message_new", params.Push(
			"flash", flash.Danger("buyer_id: not found"),
		), "admin/layouts/authenticated")
	}
	createdBy := currentAdmin.ID
	now := time.Now()
	request.MobilePhone = mobilePhone
	request.MessageType = models.TextMessageTypeSms
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedAt = now
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/text_message_new", params.Push(
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
	if _, err = q.textMessageService.CreateTextMessage(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTextMessage")
		return c.Render("/admin/text_message_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/text_message_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("sms created successfully").
		Redirect(c, sess.Session, "/admin/text_messages")
}
