package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
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

type voucherController struct {
	voucherService      services.VoucherService
	buyerService        services.BuyerService
	jsonWebTokenService services.JsonWebTokenService
	userService         services.UserService
}

func NewVoucherController(
	voucherService services.VoucherService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
	userService services.UserService,
) *voucherController {
	return &voucherController{
		voucherService:      voucherService,
		buyerService:        buyerService,
		jsonWebTokenService: jsonWebTokenService,
		userService:         userService,
	}
}

func (q *voucherController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/delete", q.V1SellerDelete)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("", q.V1BuyerIndex)
}

func (q *voucherController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-V1SellerIndex"
	pagination := new(models.Pagination)
	request, err := sanitizers.FindVouchers(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.voucherService.FindVouchers(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
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

func (q *voucherController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	now := time.Now().In(time.Local)
	t := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	).Format(time.DateOnly)
	request := &models.Voucher{
		RawEffectiveStartTime: "00:00:00",
		RawEffectiveEndTime:   "23:59:59",
		RawEffectiveStartDate: t,
		RawEffectiveEndDate:   t,
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateVoucher(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateVoucher")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if request.Code != nil {
		voucher, err := q.voucherService.FindVoucher(
			ctx,
			models.NewVoucherFilter().
				WithCodes(*request.Code),
		)
		if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if voucher != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueVoucherCodeViolation.Error()).
				WriteResponse(c)
		}
	}
	createdBy := currentUser.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("VOUCHER_DEFAULT_PHOTO")
	}
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
	saved, err := q.voucherService.CreateVoucher(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateVoucher")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
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
		SetMessage("voucher created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *voucherController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-V1SellerShow"
	voucherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	voucher, err := q.voucherService.FindVoucher(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithVoucherIDs(voucherID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if voucher == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(voucher).
		WriteResponse(c)
}

func (q *voucherController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	voucherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	voucher, err := q.voucherService.FindVoucher(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithVoucherIDs(voucherID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if voucher == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateVoucher(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateVoucher")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if voucher.Code != nil && request.Code != nil && *voucher.Code != *request.Code {
		otherVoucher, err := q.voucherService.FindVoucher(
			ctx,
			models.NewVoucherFilter().
				WithCodes(*request.Code).
				WithExcludedVoucherIDs(voucher.ID),
		)
		if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otherVoucher != nil {
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(models.ErrUniqueVoucherCodeViolation.Error()).
				WriteResponse(c)
		}
	}
	request.ID = voucher.ID
	request.RecordStatus = voucher.RecordStatus
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	request.BuyerIDs = voucher.BuyerIDs
	if photo == nil {
		request.Photo = voucher.Photo
	}
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
	saved, err := q.voucherService.UpdateVoucher(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateVoucher")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
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
		SetMessage("voucher updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *voucherController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	voucherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	voucher, err := q.voucherService.FindVoucher(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithVoucherIDs(voucherID),
	)
	if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if voucher == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	request := voucher
	if request.Code != nil {
		request.DeletedCode = new(*request.Code)
	}
	request.Code = nil
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentUser.ID
	request.DeletedAt = new(time.Now())
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
	saved, err := q.voucherService.UpdateVoucher(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateVoucher")
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
		SetMessage("voucher deleted successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *voucherController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-V1BuyerIndex"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	response, _, err := q.voucherService.FindVouchers(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBuyerIDs(currentBuyer.ID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	for i := range response {
		response[i].Photo = response[i].PhotoPath(c.BaseURL())
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
