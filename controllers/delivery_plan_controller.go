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
	deliveryPlanController struct {
		sellerService       services.SellerService
		companyService      services.CompanyService
		deliveryPlanService services.DeliveryPlanService
		buyerService        services.BuyerService
		jsonWebTokenService services.JsonWebTokenService
	}
)

func NewDeliveryPlanController(
	sellerService services.SellerService,
	companyService services.CompanyService,
	deliveryPlanService services.DeliveryPlanService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
) *deliveryPlanController {
	return &deliveryPlanController{
		sellerService:       sellerService,
		companyService:      companyService,
		deliveryPlanService: deliveryPlanService,
		buyerService:        buyerService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *deliveryPlanController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	r.Group("/:slug/v1/buyer", companyVerify, buyerKeyAuth).
		Get("/now", q.V1BuyerShow)
}

func (q *deliveryPlanController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.DeliveryPlanFilter)
	rows := []*models.DeliveryPlan{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Jadwal Delivery",
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
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.SortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	request, err = sanitizers.FindDeliveryPlans(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return c.Render("seller/delivery_plan_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.deliveryPlanService.FindDeliveryPlans(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return c.Render("seller/delivery_plan_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/delivery_plan_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *deliveryPlanController) SellerNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.DeliveryPlan{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Add New Jadwal Delivery",
		"request", request,
		"deliveryPlan", request,
		"deliveryWeekdays", models.DeliveryWeekdays,
	)
	return c.Render("seller/delivery_plan_new", params, "seller/layouts/authenticated")
}

func (q *deliveryPlanController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.DeliveryPlan{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Add New Jadwal Delivery",
		"request", request,
		"deliveryPlan", request,
		"deliveryWeekdays", models.DeliveryWeekdays,
	)
	request, statusCode, err := sanitizers.ValidateDeliveryPlan(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryPlan")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithDeliveryWeekdays(request.Weekday).
			WithTransactionWindow(request.TransactionWindowMin, request.TransactionWindowMax),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return c.Render("seller/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if deliveryPlan != nil {
		return c.Render("seller/delivery_plan_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueDeliveryPlanTransactionWindowViolation.Error()),
		), "seller/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentSeller.ID
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/delivery_plan_new", params.Push(
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
	if _, err = q.deliveryPlanService.CreateDeliveryPlan(ctx, tx, request); err != nil {
		if !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateDeliveryPlan")
		}
		return c.Render("seller/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("delivery plan created successfully").
		Redirect(c, sess.Session, "/seller/delivery_plans")
}

func (q *deliveryPlanController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	if deliveryPlan == nil {
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Edit Data Jadwal Delivery",
		"request", deliveryPlan,
		"deliveryPlan", deliveryPlan,
		"deliveryWeekdays", models.DeliveryWeekdays,
	)
	return c.Render("seller/delivery_plan_edit", params, "seller/layouts/authenticated")
}

func (q *deliveryPlanController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	if deliveryPlan == nil {
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Edit Data Jadwal Delivery",
		"request", deliveryPlan,
		"deliveryPlan", deliveryPlan,
		"deliveryWeekdays", models.DeliveryWeekdays,
	)
	request, statusCode, err := sanitizers.ValidateDeliveryPlan(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryPlan")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/delivery_plan_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if deliveryPlan.Weekday != request.Weekday ||
		deliveryPlan.TransactionWindowMin != request.TransactionWindowMin ||
		deliveryPlan.TransactionWindowMax != request.TransactionWindowMax {
		otherDeliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
			ctx,
			models.NewDeliveryPlanFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithDeliveryWeekdays(request.Weekday).
				WithTransactionWindow(request.TransactionWindowMin, request.TransactionWindowMax).
				WithExcludedDeliveryPlanIDs(deliveryPlan.ID),
		)
		if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
			return c.Render("seller/delivery_plan_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherDeliveryPlan != nil {
			return c.Render("seller/delivery_plan_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueDeliveryPlanTransactionWindowViolation.Error()),
			), "seller/layouts/authenticated")
		}
	}
	request.ID = deliveryPlan.ID
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = deliveryPlan.RecordStatus
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/delivery_plan_edit", params.Push(
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
	if _, err = q.deliveryPlanService.UpdateDeliveryPlan(ctx, tx, request); err != nil {
		if !errors.Is(err, models.ErrUniqueDeliveryPlanTransactionWindowViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
		}
		return c.Render("seller/delivery_plan_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/delivery_plan_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("delivery plan updated successfully").
		Redirect(c, sess.Session, "/seller/delivery_plans")
}

func (q *deliveryPlanController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	if deliveryPlan == nil {
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	deliveryPlan.RecordStatus = models.RecordStatusDeleted
	deliveryPlan.DeletedBy = &currentSeller.ID
	deliveryPlan.DeletedAt = new(time.Now())
	deliveryPlan.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.deliveryPlanService.UpdateDeliveryPlan(ctx, tx, deliveryPlan); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/delivery_plans")
	}
	return flash.Clear(c, sess.Session).
		Success("delivery plan deleted successfully").
		Redirect(c, sess.Session, "/seller/delivery_plans")
}

func (q *deliveryPlanController) V1BuyerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-V1BuyerShow"
	now := time.Now().In(time.Local)
	response, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithTransactionDatetime(now),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if response == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrDeliveryPlanNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
