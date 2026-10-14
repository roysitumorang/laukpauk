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

type pointEarningCapController struct {
	sellerService          services.SellerService
	companyService         services.CompanyService
	pointEarningCapService services.PointEarningCapService
	buyerService           services.BuyerService
	jsonWebTokenService    services.JsonWebTokenService
	notificationService    services.NotificationService
}

func NewPointEarningCapController(
	sellerService services.SellerService,
	companyService services.CompanyService,
	pointEarningCapService services.PointEarningCapService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
	notificationService services.NotificationService,
) *pointEarningCapController {
	return &pointEarningCapController{
		sellerService:          sellerService,
		companyService:         companyService,
		pointEarningCapService: pointEarningCapService,
		buyerService:           buyerService,
		jsonWebTokenService:    jsonWebTokenService,
		notificationService:    notificationService,
	}
}

func (q *pointEarningCapController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", companyVerify, buyerKeyAuth).
		Get("", q.V1BuyerIndex)
}

func (q *pointEarningCapController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.PointEarningCapFilter{}
	rows := []*models.PointEarningCap{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Tarif Akumulasi Poin",
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
	request, err = sanitizers.FindPointEarningCaps(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return c.Render("/seller/point_earning_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.pointEarningCapService.FindPointEarningCaps(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return c.Render("/seller/point_earning_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("/seller/point_earning_cap_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *pointEarningCapController) SellerNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.PointEarningCap{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Tambah Tarif Akumulasi Poin",
		"request", request,
		"pointEarningCap", request,
	)
	return c.Render("/seller/point_earning_cap_new", params, "seller/layouts/authenticated")
}

func (q *pointEarningCapController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.PointEarningCap{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", "Tambah Tarif Akumulasi Poin",
		"request", request,
		"pointEarningCap", request,
	)
	request, statusCode, err := sanitizers.ValidatePointEarningCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointEarningCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/seller/point_earning_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	createdBy := currentSeller.ID
	now := time.Now()
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	params.Push("request", request)
	if request.MinimumShoppingCost > 0 && request.MaximumShoppingCost > 0 {
		pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
			ctx,
			models.NewPointEarningCapFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost),
		)
		if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
			return c.Render("seller/point_earning_cap_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if pointEarningCap != nil {
			if pointEarningCap.RecordStatus == models.RecordStatusActive {
				return c.Render("seller/point_earning_cap_new", params.Push(
					"flash", flash.Danger(models.ErrPointEarningCapShoppingCostOverlapping.Error()),
				), "seller/layouts/authenticated")
			}
			request.ID = pointEarningCap.ID
			request.DeletedBy = nil
			request.DeletedAt = nil
			request.DeletionReason = nil
		}
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/seller/point_earning_cap_new", params.Push(
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
	if _, err = q.pointEarningCapService.SavePointEarningCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSavePointEarningCap")
		return c.Render("/seller/point_earning_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/seller/point_earning_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point earning cap created successfully").
		Redirect(c, sess.Session, "/seller/point_earning_caps")
}

func (q *pointEarningCapController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	if pointEarningCap == nil {
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Tarif Akumulasi Poin|%d", pointEarningCap.MinimumShoppingCost),
		"request", pointEarningCap,
		"pointEarningCap", pointEarningCap,
	)
	return c.Render("/seller/point_earning_cap_edit", params, "seller/layouts/authenticated")
}

func (q *pointEarningCapController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	if pointEarningCap == nil {
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Tarif Akumulasi Poin|%d", pointEarningCap.MinimumShoppingCost),
		"request", pointEarningCap,
		"pointEarningCap", pointEarningCap,
	)
	request, statusCode, err := sanitizers.ValidatePointEarningCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointEarningCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/seller/point_earning_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if (pointEarningCap.MinimumShoppingCost != request.MinimumShoppingCost) &&
		(pointEarningCap.MaximumShoppingCost != request.MaximumShoppingCost) {
		otherPointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
			ctx,
			models.NewPointEarningCapFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost).
				WithExcludedPointEarningCapIDs(pointEarningCap.ID),
		)
		if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
			return c.Render("seller/point_earning_cap_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherPointEarningCap != nil {
			return c.Render("seller/point_earning_cap_edit", params.Push(
				"flash", flash.Danger(models.ErrPointEarningCapShoppingCostOverlapping.Error()),
			), "seller/layouts/authenticated")
		}
	}
	request.ID = pointEarningCap.ID
	request.CompanyID = currentSeller.CompanyID
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/seller/point_earning_cap_edit", params.Push(
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
	if _, err = q.pointEarningCapService.UpdatePointEarningCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
		return c.Render("/seller/point_earning_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/seller/point_earning_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point earning cap updated successfully").
		Redirect(c, sess.Session, "/seller/point_earning_caps")
}

func (q *pointEarningCapController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	if pointEarningCap == nil {
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	pointEarningCap.RecordStatus = models.RecordStatusDeleted
	pointEarningCap.DeletedBy = &currentSeller.ID
	pointEarningCap.DeletedAt = new(time.Now())
	pointEarningCap.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.pointEarningCapService.UpdatePointEarningCap(ctx, tx, pointEarningCap); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_earning_caps")
	}
	return flash.Clear(c, sess.Session).
		Success("point earning cap deleted successfully").
		Redirect(c, sess.Session, "/seller/point_earning_caps")
}

func (q *pointEarningCapController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-V1BuyerIndex"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	response, _, err := q.pointEarningCapService.FindPointEarningCaps(
		ctx,
		models.NewPointEarningCapFilter().
			WithCompanyIDs(currentBuyer.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
