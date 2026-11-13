package controllers

import (
	"context"
	"errors"
	"log/slog"
	"strings"
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

type transactionController struct {
	buyerService        services.BuyerService
	transactionService  services.TransactionService
	jsonWebTokenService services.JsonWebTokenService
	kafkaClient         *externals.KafkaClient
	userService         services.UserService
}

func NewTransactionController(
	buyerService services.BuyerService,
	transactionService services.TransactionService,
	jsonWebTokenService services.JsonWebTokenService,
	kafkaClient *externals.KafkaClient,
	userService services.UserService,
) *transactionController {
	return &transactionController{
		buyerService:        buyerService,
		transactionService:  transactionService,
		jsonWebTokenService: jsonWebTokenService,
		kafkaClient:         kafkaClient,
		userService:         userService,
	}
}

func (q *transactionController) Mount(r fiber.Router) {
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer").
		Get("", buyerKeyAuth, q.V1BuyerIndex).
		Post("", buyerKeyAuth, q.V1BuyerCreate).
		Get("/:id", buyerKeyAuth, q.V1BuyerShow).
		Post("/pay", q.V1Pay).
		Post("/:id/withdraw", buyerKeyAuth, q.V1BuyerWithdraw)
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Get("/:id", q.V1SellerShow).
		Post("/:id/withdraw", q.V1SellerWithdraw).
		Post("/:id/process", q.V1SellerProcess).
		Post("/:id/ship", q.V1SellerShip).
		Post("/:id/complete", q.V1SellerComplete)

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
		ctx,
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
	request.BuyerID = currentBuyer.ID
	response, notificationIDs, err := q.transactionService.CreateTransaction(ctx, currentBuyer, request)
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

func (q *transactionController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-V1SellerIndex"
	ctx := c.Context()
	request, err := sanitizers.FindTransactions(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.transactionService.FindTransactions(
		ctx,
		request,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
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

func (q *transactionController) V1SellerShow(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-V1SellerShow"
	ctx := c.Context()
	transaction, err := q.prepareTransaction(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(transaction).
		WriteResponse(c)
}

func (q *transactionController) V1SellerWithdraw(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-V1SellerWithdraw"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, err := q.prepareTransaction(ctx, c, models.TransactionStatusWithdrawalPending)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	request.Status = models.TransactionStatusWithdrawn
	request.WithdrawnBy = &currentUser.ID
	request.WithdrawnAt = new(time.Now().In(time.Local))
	saved, err := q.transactionService.WithdrawTransaction(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrWithdrawTransaction")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("transaction withdrawn successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *transactionController) V1SellerProcess(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-V1SellerProcess"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	transaction, err := q.prepareTransaction(ctx, c, models.TransactionStatusPaid)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if transaction == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidateTransaction(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransaction")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	n := len(transaction.Items)
	mapTransactionItemIDs := make(map[uuid.UUID]int, n)
	for i, transactionItem := range transaction.Items {
		mapTransactionItemIDs[transactionItem.ID] = i
	}
	for i, requestItem := range request.Items {
		j, ok := mapTransactionItemIDs[requestItem.ID]
		if !ok {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(helper.Sprintf("items[%d].id: not found", i)).
				WriteResponse(c)
		}
		transactionItem := transaction.Items[j]
		requestItem.RevisionNote = strings.TrimSpace(requestItem.RevisionNote)
		if requestItem.Quantity > transactionItem.Quantity {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(helper.Sprintf("items[%d].quantity: cannot exceed %d", i, transactionItem.Quantity)).
				WriteResponse(c)
		} else if requestItem.Quantity < transactionItem.Quantity && requestItem.RevisionNote == "" {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(helper.Sprintf("items[%d].revison_note: is required", i)).
				WriteResponse(c)
		}
		transactionItem.Quantity = max(requestItem.Quantity, 0)
		transactionItem.OutOfStock = transactionItem.Quantity == 0
		transactionItem.RevisionNote = requestItem.RevisionNote
		transaction.Items[j] = transactionItem
	}
	transaction.Revise()
	now := time.Now()
	transaction.Status = models.TransactionStatusProcessed
	transaction.ProcessedBy = &currentUser.ID
	transaction.ProcessedAt = &now
	transaction.UpdatedAt = now
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
	saved, err := q.transactionService.UpdateTransaction(ctx, tx, transaction)
	if err != nil {
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
		SetMessage("transaction processed successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *transactionController) V1SellerShip(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-V1SellerShip"
	ctx := c.Context()
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	transaction, err := q.prepareTransaction(ctx, c, models.TransactionStatusProcessed)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if transaction == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	request, _, err := sanitizers.ValidateTransactionShipment(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateTransactionShipment")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	transaction.Status = models.TransactionStatusShipped
	transaction.ShippedBy = &currentUser.ID
	transaction.ShippedAt = &now
	transaction.DeliveryActualCost = request.DeliveryActualCost
	transaction.UpdatedAt = now
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
	saved, err := q.transactionService.UpdateTransaction(ctx, tx, transaction)
	if err != nil {
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
		SetMessage("transaction shipped successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *transactionController) V1SellerComplete(c fiber.Ctx) (err error) {
	ctxt := "TransactionController-V1SellerComplete"
	ctx := c.Context()
	transaction, err := q.prepareTransaction(ctx, c, models.TransactionStatusShipped)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPrepareTransaction")
		return helper.NewResponse(notFoundStatusCode(err)).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if transaction == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrTransactionNotFound.Error()).
			WriteResponse(c)
	}
	saved, err := q.transactionService.CompleteTransaction(ctx, transaction)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCompleteTransaction")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("transaction completed successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *transactionController) prepareTransaction(
	ctx context.Context,
	c fiber.Ctx,
	transactionStatuses ...models.TransactionStatus,
) (
	*models.Transaction,
	error,
) {
	transactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, models.ErrTransactionNotFound
	}
	request := models.NewTransactionFilter().
		WithTransactionIDs(transactionID)
	if len(transactionStatuses) > 0 {
		request.WithTransactionStatuses(transactionStatuses...)
	}
	transaction, err := q.transactionService.FindTransaction(ctx, request)
	if err != nil && !errors.Is(err, models.ErrTransactionNotFound) {
		return nil, err
	}
	if transaction == nil {
		return nil, models.ErrTransactionNotFound
	}
	return transaction, nil
}
