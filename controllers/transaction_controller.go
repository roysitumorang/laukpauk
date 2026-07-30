package controllers

import (
	"context"
	"errors"
	"log/slog"
	"strings"
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
	transactionController struct {
		adminService        services.AdminService
		buyerService        services.BuyerService
		transactionService  services.TransactionService
		jsonWebTokenService services.JsonWebTokenService
	}
)

func NewTransactionController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	transactionService services.TransactionService,
	jsonWebTokenService services.JsonWebTokenService,
) *transactionController {
	return &transactionController{
		adminService:        adminService,
		buyerService:        buyerService,
		transactionService:  transactionService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *transactionController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Get("", buyerKeyAuth, q.V1BuyerIndex).
		Post("", buyerKeyAuth, q.V1BuyerCreate).
		Get("/:id", buyerKeyAuth, q.V1BuyerShow).
		Post("/pay", q.V1Pay).
		Post("/:id/withdraw", buyerKeyAuth, q.V1BuyerWithdraw)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/:id", q.AdminShow).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id/withdraw", q.AdminWithdraw).
		Post("/:id/process", q.AdminProcess).
		Post("/:id/ship", q.AdminShip).
		Post("/:id/complete", q.AdminComplete)

}

func (q *transactionController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerIndex"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, err := sanitizers.FindTransactions(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.transactionService.FindTransactions(
		c,
		request.WithBuyerIDs(currentBuyer.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusBadRequest).
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

func (q *transactionController) V1BuyerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerCreate"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateTransaction(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		c.Response().SetStatusCode(statusCode)
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response, err := q.transactionService.CreateTransaction(ctx, currentBuyer, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("transaction created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *transactionController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerShow"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.transactionService.FindTransaction(
		c,
		models.NewTransactionFilter().
			WithBuyerIDs(currentBuyer.ID).
			WithTransactionIDs(transactionID),
	)
	if err != nil && !errors.Is(err, models.ErrTransactionNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *transactionController) V1Pay(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1Pay"
	request, statusCode, err := sanitizers.ValidateTransactionPayment(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransactionPayment")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	transactionID, err := uuid.Parse(request.PaymentCapture.Value.Data.ReferenceID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	transaction, err := q.transactionService.FindTransaction(
		ctx,
		models.NewTransactionFilter().
			WithTransactionIDs(transactionID).
			WithTransactionStatuses(models.TransactionStatusPending),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if transaction == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	if transaction.PaymentStatus = &request.PaymentCapture.Value.Data.Status; *transaction.PaymentStatus == "SUCCEEDED" {
		transaction.Status = models.TransactionStatusPaid
		transaction.PaidAt = &request.PaymentCapture.Value.Data.Created
	} else {
		transaction.PaymentFailureCode = &request.PaymentFailure.Value.Data.FailureCode
	}
	transaction.PaymentCapturedAt = &request.PaymentCapture.Value.Data.Created
	transaction.PaymentID = &request.PaymentCapture.Value.Data.PaymentID
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
	return helper.NewResponse(fiber.StatusNoContent).
		WriteResponse(c)
}

func (q *transactionController) V1BuyerWithdraw(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerWithdraw"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	transaction, err := q.transactionService.FindTransaction(
		c,
		models.NewTransactionFilter().
			WithBuyerIDs(currentBuyer.ID).
			WithTransactionIDs(transactionID).
			WithTransactionStatuses(models.TransactionStatusPaid),
	)
	if err != nil && !errors.Is(err, models.ErrTransactionNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if transaction == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateTransactionWithdrawal(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransactionWithdrawal")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	transaction.WithdrawalReason = &request.Reason
	response, err := q.transactionService.RequestWithdrawal(ctx, transaction)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRequestWithdrawal")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
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

func (q *transactionController) AdminWithdraw(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminWithdraw"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request, _, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
		models.TransactionStatusWithdrawalPending,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if request == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	request.Status = models.TransactionStatusWithdrawn
	request.WithdrawnBy = &currentAdmin.ID
	request.WithdrawnAt = new(time.Now().In(time.Local))
	if _, err = q.transactionService.WithdrawTransaction(ctx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrWithdrawTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", request.ID))
	}
	return flash.Success("transaction withdrawn successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", request.ID))
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
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
	n := len(transaction.Items)
	mapTransactionItemIDs := make(map[uuid.UUID]int, n)
	for i, transactionItem := range transaction.Items {
		mapTransactionItemIDs[transactionItem.ID] = i
	}
	for i, requestItem := range request.Items {
		j, ok := mapTransactionItemIDs[requestItem.ID]
		if !ok {
			return c.Render("admin/transaction_edit", params.Push(
				"flash", flash.Danger(helper.Sprintf("items[%d].id: not found", i)),
			), "admin/layouts/authenticated")
		}
		transactionItem := transaction.Items[j]
		requestItem.RevisionNote = strings.TrimSpace(requestItem.RevisionNote)
		if requestItem.Quantity > transactionItem.Quantity {
			return c.Render("admin/transaction_edit", params.Push(
				"flash", flash.Danger(helper.Sprintf("items[%d].quantity: cannot exceed %d", i, transactionItem.Quantity)),
			), "admin/layouts/authenticated")
		} else if requestItem.Quantity < transactionItem.Quantity && requestItem.RevisionNote == "" {
			return c.Render("admin/transaction_edit", params.Push(
				"flash", flash.Danger(helper.Sprintf("items[%d].revison_note: is required", i)),
			), "admin/layouts/authenticated")
		}
		transactionItem.Quantity = max(requestItem.Quantity, 0)
		transactionItem.OutOfStock = transactionItem.Quantity == 0
		transactionItem.RevisionNote = requestItem.RevisionNote
		transaction.Items[j] = transactionItem
	}
	transaction.Revise()
	now := time.Now()
	transaction.Status = models.TransactionStatusProcessed
	transaction.ProcessedBy = &currentAdmin.ID
	transaction.ProcessedAt = &now
	transaction.UpdatedAt = now
	params.Push("request", transaction)
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
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
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
	transaction, _, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
		models.TransactionStatusProcessed,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	request, _, err := sanitizers.ValidateTransactionShipment(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransactionShipment")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
	}
	now := time.Now()
	transaction.Status = models.TransactionStatusShipped
	transaction.ShippedBy = &currentAdmin.ID
	transaction.ShippedAt = &now
	transaction.DeliveryActualCost = request.DeliveryActualCost
	transaction.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
	}
	return flash.Success("transaction shipped successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
}

func (q *transactionController) AdminComplete(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-AdminComplete"
	ctx, sess, flash, currentAdmin, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, _, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentAdmin,
		csrfToken,
		false,
		models.TransactionStatusShipped,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/admin/transactions")
	}
	if _, err = q.transactionService.CompleteTransaction(ctx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCompleteTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
	}
	return flash.Success("transaction completed successfully").Redirect(c, sess.Session, helper.Sprintf("/admin/transactions/%s", transaction.ID))
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
