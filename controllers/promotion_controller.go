package controllers

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type promotionController struct {
	promotionService    services.PromotionService
	jsonWebTokenService services.JsonWebTokenService
	userService         services.UserService
	transactionService  services.TransactionService
}

func NewPromotionController(
	promotionService services.PromotionService,
	jsonWebTokenService services.JsonWebTokenService,
	userService services.UserService,
	transactionService services.TransactionService,
) *promotionController {
	return &promotionController{
		promotionService:    promotionService,
		jsonWebTokenService: jsonWebTokenService,
		userService:         userService,
		transactionService:  transactionService,
	}
}

func (q *promotionController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Put("/:id", q.V1SellerUpdate).
		Delete("/:id", q.V1SellerDelete)
	buyerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleBuyer, q.userService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex).
		Get("/:id", q.V1BuyerShow)
}

func (q *promotionController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PromotionController-V1SellerIndex"
	pagination := new(models.Pagination)
	request, err := sanitizers.FindPromotions(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotions")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.promotionService.FindPromotions(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotions")
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

func (q *promotionController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PromotionController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	now := time.Now().In(time.Local)
	request, statusCode, err := sanitizers.ValidatePromotion(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePromotion")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request.Code != nil {
		promotion, err := q.promotionService.FindPromotion(
			ctx,
			models.NewPromotionFilter().
				WithCodes(*request.Code),
		)
		if errors.Is(err, models.ErrPromotionNotFound) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotion")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if promotion != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniquePromotionCodeViolation.Error()).
				WriteResponse(c)
		}
	}
	createdBy := currentUser.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
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
	response, err := q.promotionService.CreatePromotion(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePromotion")
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
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("promotion created successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *promotionController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PromotionController-V1SellerShow"
	promotionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	response, err := q.promotionService.FindPromotion(
		ctx,
		models.NewPromotionFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPromotionIDs(promotionID),
	)
	if errors.Is(err, models.ErrPromotionNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotion")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *promotionController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PromotionController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	promotionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	promotion, err := q.promotionService.FindPromotion(
		ctx,
		models.NewPromotionFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPromotionIDs(promotionID),
	)
	if errors.Is(err, models.ErrPromotionNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotion")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request, statusCode, err := sanitizers.ValidatePromotion(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePromotion")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if promotion.Code != nil && request.Code != nil && *promotion.Code != *request.Code {
		otherPromotion, err := q.promotionService.FindPromotion(
			ctx,
			models.NewPromotionFilter().
				WithCodes(*request.Code).
				WithExcludedPromotionIDs(promotion.ID),
		)
		if errors.Is(err, models.ErrPromotionNotFound) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotion")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherPromotion != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniquePromotionCodeViolation.Error()).
				WriteResponse(c)
		}
	}
	request.ID = promotion.ID
	request.RecordStatus = promotion.RecordStatus
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
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
	response, err := q.promotionService.UpdatePromotion(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePromotion")
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
		SetMessage("promotion updated successfully").
		SetData(response).
		WriteResponse(c)
}

func (q *promotionController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PromotionController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	promotionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	promotion, err := q.promotionService.FindPromotion(
		ctx,
		models.NewPromotionFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPromotionIDs(promotionID),
	)
	if errors.Is(err, models.ErrPromotionNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotion")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if promotion.Code != nil {
		promotion.DeletedCode = new(*promotion.Code)
	}
	promotion.Code = nil
	promotion.RecordStatus = models.RecordStatusDeleted
	promotion.DeletedBy = &currentUser.ID
	promotion.DeletedAt = new(time.Now())
	promotion.DeletionReason = nil
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
	if _, err = q.promotionService.UpdatePromotion(ctx, tx, promotion); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePromotion")
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

func (q *promotionController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PromotionController-V1BuyerIndex"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, err := sanitizers.FindPromotions(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	promotions, _, err := q.promotionService.FindPromotions(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive).
			WithTransactionDatetime(now),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotions")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	n := len(promotions)
	var response []*models.Promotion
	if n == 0 {
		return helper.NewResponse(fiber.StatusOK).
			SetData(response).
			WriteResponse(c)
	}
	promotionIDs := make([]uuid.UUID, n)
	for i := range promotions {
		promotionIDs[i] = promotions[i].ID
	}
	mapPromotionsUsages, err := q.transactionService.FindPromotionsUsages(
		ctx,
		currentUser.ID,
		promotionIDs...,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotionsUsages")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	for _, promotion := range promotions {
		if promotion.Usages = mapPromotionsUsages[promotion.ID]; promotion.Usages < promotion.QuotaPerBuyer {
			response = append(response, promotion)
		}
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}

func (q *promotionController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PromotionController-V1BuyerShow"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	promotionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	response, err := q.promotionService.FindPromotion(
		ctx,
		models.NewPromotionFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithTransactionDatetime(now).
			WithPromotionIDs(promotionID),
	)
	if errors.Is(err, models.ErrPromotionNotFound) {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotions")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	mapPromotionsUsages, err := q.transactionService.FindPromotionsUsages(
		ctx,
		currentUser.ID,
		response.ID,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPromotionsUsages")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response.Usages = mapPromotionsUsages[response.ID]; response.Usages >= response.QuotaPerBuyer {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrPromotionNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
