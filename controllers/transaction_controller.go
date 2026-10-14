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
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type transactionController struct {
	companyService      services.CompanyService
	sellerService       services.SellerService
	buyerService        services.BuyerService
	transactionService  services.TransactionService
	jsonWebTokenService services.JsonWebTokenService
	kafkaClient         *externals.KafkaClient
}

func NewTransactionController(
	companyService services.CompanyService,
	sellerService services.SellerService,
	buyerService services.BuyerService,
	transactionService services.TransactionService,
	jsonWebTokenService services.JsonWebTokenService,
	kafkaClient *externals.KafkaClient,
) *transactionController {
	return &transactionController{
		companyService:      companyService,
		sellerService:       sellerService,
		buyerService:        buyerService,
		transactionService:  transactionService,
		jsonWebTokenService: jsonWebTokenService,
		kafkaClient:         kafkaClient,
	}
}

func (q *transactionController) Mount(r fiber.Router) {
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", companyVerify).
		Get("", buyerKeyAuth, q.V1BuyerIndex).
		Post("", buyerKeyAuth, q.V1BuyerCreate).
		Get("/:id", buyerKeyAuth, q.V1BuyerShow).
		Post("/pay", q.V1Pay).
		Post("/:id/withdraw", buyerKeyAuth, q.V1BuyerWithdraw)
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/:id", q.SellerShow).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id/withdraw", q.SellerWithdraw).
		Post("/:id/process", q.SellerProcess).
		Post("/:id/ship", q.SellerShip).
		Post("/:id/complete", q.SellerComplete)

}

func (q *transactionController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "TransactionController-V1BuyerIndex"
	company, _ := c.Locals(models.CurrentCompany).(*models.Company)
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, err := sanitizers.FindTransactions(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.transactionService.FindTransactions(
		ctx,
		request.WithCompanyIDs(company.ID).
			WithBuyerIDs(currentBuyer.ID),
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
	company, _ := c.Locals(models.CurrentCompany).(*models.Company)
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	request, statusCode, err := sanitizers.ValidateTransaction(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		c.Response().SetStatusCode(statusCode)
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.CompanyID = company.ID
	request.BuyerID = currentBuyer.ID
	response, notificationIDs, err := q.transactionService.CreateTransaction(ctx, company, currentBuyer, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	messages := make([]models.TopicMessage, len(notificationIDs))
	for i := range notificationIDs {
		messages[i] = models.TopicMessage{
			Topic: models.TopicNotification,
			Event: models.TopicNotificationEventSend,
			ID:    notificationIDs[i],
		}
	}
	if err = q.kafkaClient.Publish(ctx, messages...); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
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

func (q *transactionController) SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-SellerIndex"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request := &models.TransactionFilter{}
	rows := []*models.Transaction{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
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
		return c.Render("seller/transaction_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.transactionService.FindTransactions(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return c.Render("seller/transaction_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/transaction_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *transactionController) SellerShow(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-SellerShow"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	_, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	return c.Render("seller/transaction_show", params, "seller/layouts/authenticated")
}

func (q *transactionController) SellerEdit(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-SellerEdit"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	return c.Render("seller/transaction_edit", params, "seller/layouts/authenticated")
}

func (q *transactionController) SellerWithdraw(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-SellerWithdraw"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	request, _, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		false,
		models.TransactionStatusWithdrawalPending,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	if request == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	request.Status = models.TransactionStatusWithdrawn
	request.WithdrawnBy = &currentSeller.ID
	request.WithdrawnAt = new(time.Now().In(time.Local))
	if _, err = q.transactionService.WithdrawTransaction(ctx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrWithdrawTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", request.ID))
	}
	return flash.Success("transaction withdrawn successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", request.ID))
}

func (q *transactionController) SellerProcess(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-SellerProcess"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, params, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		true,
		models.TransactionStatusPaid,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	params.Push("pageTitle", helper.Sprintf("Edit Data Transaction|%s", transaction.ReferenceNumber))
	request, statusCode, err := sanitizers.ValidateTransaction(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	n := len(transaction.Items)
	mapTransactionItemIDs := make(map[uuid.UUID]int, n)
	for i, transactionItem := range transaction.Items {
		mapTransactionItemIDs[transactionItem.ID] = i
	}
	for i, requestItem := range request.Items {
		j, ok := mapTransactionItemIDs[requestItem.ID]
		if !ok {
			return c.Render("seller/transaction_edit", params.Push(
				"flash", flash.Danger(helper.Sprintf("items[%d].id: not found", i)),
			), "seller/layouts/authenticated")
		}
		transactionItem := transaction.Items[j]
		requestItem.RevisionNote = strings.TrimSpace(requestItem.RevisionNote)
		if requestItem.Quantity > transactionItem.Quantity {
			return c.Render("seller/transaction_edit", params.Push(
				"flash", flash.Danger(helper.Sprintf("items[%d].quantity: cannot exceed %d", i, transactionItem.Quantity)),
			), "seller/layouts/authenticated")
		} else if requestItem.Quantity < transactionItem.Quantity && requestItem.RevisionNote == "" {
			return c.Render("seller/transaction_edit", params.Push(
				"flash", flash.Danger(helper.Sprintf("items[%d].revison_note: is required", i)),
			), "seller/layouts/authenticated")
		}
		transactionItem.Quantity = max(requestItem.Quantity, 0)
		transactionItem.OutOfStock = transactionItem.Quantity == 0
		transactionItem.RevisionNote = requestItem.RevisionNote
		transaction.Items[j] = transactionItem
	}
	transaction.Revise()
	now := time.Now()
	transaction.Status = models.TransactionStatusProcessed
	transaction.ProcessedBy = &currentSeller.ID
	transaction.ProcessedAt = &now
	transaction.UpdatedAt = now
	params.Push("request", transaction)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/transaction_edit", params.Push(
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
	if _, err = q.transactionService.UpdateTransaction(ctx, tx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransaction")
		return c.Render("seller/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/transaction_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Success("transaction processed successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
}

func (q *transactionController) SellerShip(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-SellerShip"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, _, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		false,
		models.TransactionStatusProcessed,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	request, _, err := sanitizers.ValidateTransactionShipment(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransactionShipment")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
	}
	now := time.Now()
	transaction.Status = models.TransactionStatusShipped
	transaction.ShippedBy = &currentSeller.ID
	transaction.ShippedAt = &now
	transaction.DeliveryActualCost = request.DeliveryActualCost
	transaction.UpdatedAt = now
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
	}
	return flash.Success("transaction shipped successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
}

func (q *transactionController) SellerComplete(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-SellerComplete"
	ctx, sess, flash, currentSeller, csrfToken := q.prepare(c)
	defer flash.Clear(c, sess.Session)
	transaction, _, err := q.prepareTransaction(
		ctx,
		c,
		flash,
		currentSeller,
		csrfToken,
		false,
		models.TransactionStatusShipped,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	if transaction == nil {
		return flash.Danger(models.ErrTransactionNotFound.Error()).Redirect(c, sess.Session, "/seller/transactions")
	}
	if _, err = q.transactionService.CompleteTransaction(ctx, transaction); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCompleteTransaction")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
	}
	return flash.Success("transaction completed successfully").Redirect(c, sess.Session, helper.Sprintf("/seller/transactions/%s", transaction.ID))
}

func (q *transactionController) prepare(c fiber.Ctx) (
	context.Context,
	*session.Middleware,
	*helper.FlashMessage,
	*models.Seller,
	string,
) {
	ctx := c.Context()
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	csrfToken := csrf.TokenFromContext(c)
	return ctx, sess, flash, currentSeller, csrfToken
}

func (q *transactionController) prepareTransaction(
	ctx context.Context,
	c fiber.Ctx,
	flash *helper.FlashMessage,
	currentSeller *models.Seller,
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
		WithCompanyIDs(currentSeller.CompanyID).
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
			"currentSeller", currentSeller,
			"pageTitle", helper.Sprintf("Detail Data Transaction|%s", transaction.ReferenceNumber),
			"transaction", transaction,
			"request", transaction,
		)
	}
	return transaction, params, nil
}
