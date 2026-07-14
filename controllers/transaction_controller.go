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
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	transactionController struct {
		adminService       services.AdminService
		transactionService services.TransactionService
	}
)

func NewTransactionController(
	adminService services.AdminService,
	transactionService services.TransactionService,
) *transactionController {
	return new(transactionController{
		adminService:       adminService,
		transactionService: transactionService,
	})
}

func (q *transactionController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Post("/:id/pay", q.Pay)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/:id", q.AdminShow).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id/process", q.AdminProcess).
		Post("/:id/ship", q.AdminShip).
		Post("/:id/deliver", q.AdminDeliver)

}

func (q *transactionController) AdminIndex(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminIndex"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := new(models.TransactionFilter)
	rows := []*models.Transaction{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Transaction",
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
	if request, err = sanitizers.FindTransactions(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return c.Render("admin/transaction_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.transactionService.FindTransactions(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return c.Render("admin/transaction_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/transaction_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *transactionController) AdminShow(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminShow"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	return c.Render("admin/transaction_show", params, "admin/layouts/authenticated")
}

func (q *transactionController) Pay(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-Pay"
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	transaction, err := q.transactionService.FindTransaction(
		ctx,
		models.NewTransactionFilter().
			WithTransactionIDs(transactionID).
			WithTransactionStatuses(models.TransactionStatusPending),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if transaction == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	transaction.TransactionStatus = models.TransactionStatusPaid
	transaction.PaidAt = new(time.Now())
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
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
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
		SetMessage("transaction paid successfully").
		WriteResponse(c)
}

func (q *transactionController) AdminEdit(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminEdit"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	return c.Render("admin/transaction_edit", params, "admin/layouts/authenticated")
}

func (q *transactionController) AdminProcess(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminProcess"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		models.TransactionStatusPaid,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	request, statusCode, err := sanitizers.ValidateTransaction(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = transaction.ID
	request.TransactionStatus = models.TransactionStatusProcessed
	request.ProcessedBy = &currentAdmin.ID
	request.ProcessedAt = new(time.Now())
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/transaction_edit", params.Push(
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
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return c.Render("admin/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("transaction processed successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
}

func (q *transactionController) AdminShip(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminShip"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		models.TransactionStatusProcessed,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	transaction.TransactionStatus = models.TransactionStatusShipped
	transaction.ShippedBy = &currentAdmin.ID
	transaction.ShippedAt = new(time.Now())
	params.Push("request", transaction)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/transaction_show", params.Push(
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
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("transaction shipped successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
}

func (q *transactionController) AdminDeliver(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminDeliver"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		true,
		models.TransactionStatusShipped,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	transaction.TransactionStatus = models.TransactionStatusDelivered
	params.Push("request", transaction)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/transaction_show", params.Push(
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
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/transaction_show", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("transaction delivered successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
}

func (q *transactionController) prepare(c fiber.Ctx) (
	context.Context,
	*session.Middleware,
	*helper.FlashMessage,
	*models.Admin,
	string,
) {
	ctx := c.Context()
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	return ctx, sess, flash, currentAdmin, csrfToken
}

func (q *transactionController) prepareTransaction(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentAdmin *models.Admin,
	csrfToken string,
	withParams bool,
	transactionStatuses ...models.TransactionStatus,
) (
	*models.Transaction,
	models.Map,
	error,
) {
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, models.ErrTransactionNotFound
	}
	request := models.NewTransactionFilter().
		WithTransactionIDs(transactionID)
	if len(transactionStatuses) > 0 {
		request.WithTransactionStatuses(transactionStatuses...)
	}
	transaction, err := q.transactionService.FindTransaction(ctx, request)
	if err != nil && !errors.Is(err, models.ErrTransactionNotFound) {
		return nil, nil, err
	}
	if transaction == nil {
		return nil, nil, models.ErrTransactionNotFound
	}
	var params models.Map
	if withParams {
		params = models.NewMap().Push(
			"baseURL", c.BaseURL(),
			"flash", flash,
			"path", c.OriginalURL(),
			"csrfToken", csrfToken,
			"currentAdmin", currentAdmin,
			"pageTitle", helper.Sprintf("Detail Data Transaction|%s", transaction.ReferenceNumber),
			"transaction", transaction,
			"request", transaction,
		)
	}
	return transaction, params, nil
}
