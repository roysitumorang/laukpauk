package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
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
	voucherController struct {
		adminService        services.AdminService
		voucherService      services.VoucherService
		buyerService        services.BuyerService
		jsonWebTokenService services.JsonWebTokenService
	}
)

func NewVoucherController(
	adminService services.AdminService,
	voucherService services.VoucherService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
) *voucherController {
	return &voucherController{
		adminService:        adminService,
		voucherService:      voucherService,
		buyerService:        buyerService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *voucherController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	buyerKeyAuth := middleware.BuyerKeyAuth(q.buyerService, q.jsonWebTokenService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
	r.Group("/v1/buyer", buyerKeyAuth).
		Get("/:id", q.V1BuyerShow)
}

func (q *voucherController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-AdminIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.VoucherFilter)
	rows := []*models.Voucher{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Voucher",
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
			{
				Field:     models.VoucherSortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.VoucherSortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	if request, err = sanitizers.FindVouchers(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return c.Render("admin/voucher_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.WithRecordStatuses(models.RecordStatusActive)
	if rows, pagination, err = q.voucherService.FindVouchers(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return c.Render("admin/voucher_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/voucher_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *voucherController) AdminNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
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
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Voucher",
		"request", request,
		"voucher", request,
		"voucherTypes", models.VoucherTypes,
		"targetAudiences", models.VoucherTargetAudiences,
	)
	return c.Render("admin/voucher_new", params, "admin/layouts/authenticated")
}

func (q *voucherController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-AdminCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
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
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Voucher",
		"voucher", request,
		"voucherTypes", models.VoucherTypes,
		"targetAudiences", models.VoucherTargetAudiences,
	)
	request, photo, statusCode, err := sanitizers.ValidateVoucher(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateVoucher")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/voucher_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if request.Code != nil {
		voucher, err := q.voucherService.FindVoucher(
			ctx,
			models.NewVoucherFilter().
				WithCodes(*request.Code),
		)
		if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
			return c.Render("admin/voucher_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if voucher != nil {
			return c.Render("admin/voucher_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueVoucherCodeViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	createdBy := currentAdmin.ID
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
		return c.Render("admin/voucher_new", params.Push(
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
	if _, err = q.voucherService.CreateVoucher(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateVoucher")
		return c.Render("admin/voucher_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/voucher_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/voucher_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("voucher created successfully").Redirect(c, sess.Session, "/admin/vouchers")
}

func (q *voucherController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-AdminEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	voucherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrVoucherNotFound.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	voucher, err := q.voucherService.FindVoucher(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithVoucherIDs(voucherID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	if voucher == nil {
		return flash.Danger(models.ErrVoucherNotFound.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Voucher|%s", voucher.Name),
		"request", voucher,
		"voucher", voucher,
		"voucherTypes", models.VoucherTypes,
		"targetAudiences", models.VoucherTargetAudiences,
	)
	return c.Render("admin/voucher_edit", params, "admin/layouts/authenticated")
}

func (q *voucherController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-AdminUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	voucherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrVoucherNotFound.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	voucher, err := q.voucherService.FindVoucher(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithVoucherIDs(voucherID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	if voucher == nil {
		return flash.Danger(models.ErrVoucherNotFound.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Voucher|%s", voucher.Name),
		"request", voucher,
		"voucher", voucher,
		"voucherTypes", models.VoucherTypes,
		"targetAudiences", models.VoucherTargetAudiences,
	)
	request, photo, statusCode, err := sanitizers.ValidateVoucher(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateVoucher")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/voucher_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
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
			return c.Render("admin/voucher_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherVoucher != nil {
			return c.Render("admin/voucher_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueVoucherCodeViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = voucher.ID
	request.RecordStatus = voucher.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	request.BuyerIDs = voucher.BuyerIDs
	if photo == nil {
		request.Photo = voucher.Photo
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/voucher_edit", params.Push(
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
	if _, err = q.voucherService.UpdateVoucher(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateVoucher")
		return c.Render("admin/voucher_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = helper.SaveFile(c, photo, request.Photo); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/voucher_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/voucher_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("voucher updated successfully").Redirect(c, sess.Session, "/admin/vouchers")
}

func (q *voucherController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	voucherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrVoucherNotFound.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	voucher, err := q.voucherService.FindVoucher(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithVoucherIDs(voucherID),
	)
	if err != nil && !errors.Is(err, models.ErrVoucherNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	if voucher == nil {
		return flash.Danger(models.ErrVoucherNotFound.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	request := voucher
	if request.Code != nil {
		request.DeletedCode = new(*request.Code)
	}
	request.Code = nil
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.voucherService.UpdateVoucher(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateVoucher")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/vouchers")
	}
	return flash.Clear(c, sess.Session).
		Success("voucher deleted successfully").
		Redirect(c, sess.Session, "/admin/vouchers")
}

func (q *voucherController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "VoucherController-V1BuyerShow"
	response, err := q.voucherService.FindVoucher(
		ctx,
		models.NewVoucherFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithCodes(c.Params("id")),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrVoucherNotFound.Error()).
			WriteResponse(c)
	}
	response.Photo = response.PhotoPath(c.BaseURL())
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
