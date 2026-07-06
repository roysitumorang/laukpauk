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
	deliveryPlanController struct {
		userService         services.UserService
		deliveryPlanService services.DeliveryPlanService
	}
)

func NewDeliveryPlanController(
	userService services.UserService,
	deliveryPlanService services.DeliveryPlanService,
) *deliveryPlanController {
	return new(deliveryPlanController{
		userService:         userService,
		deliveryPlanService: deliveryPlanService,
	})
}

func (q *deliveryPlanController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *deliveryPlanController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
		"currentAdmin", currentAdmin,
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
		return c.Render("admin/delivery_plan_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.WithRecordStatuses(models.RecordStatusActive)
	if rows, pagination, err = q.deliveryPlanService.FindDeliveryPlans(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return c.Render("admin/delivery_plan_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/delivery_plan_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *deliveryPlanController) AdminNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.DeliveryPlan)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Jadwal Delivery",
		"request", request,
	)
	return c.Render("admin/delivery_plan_new", params, "admin/layouts/authenticated")
}

func (q *deliveryPlanController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.DeliveryPlan)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Jadwal Delivery",
		"request", request,
	)
	request, statusCode, err := sanitizers.ValidateDeliveryPlan(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryPlan")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithName(request.Name),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return c.Render("admin/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if deliveryPlan != nil {
		return c.Render("admin/delivery_plan_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueDeliveryPlanNameViolation.Error()),
		), "admin/layouts/authenticated")
	}
	deliveryPlan, err = q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithTransactionWindowMin(new(request.TransactionWindowMin)).
			WithTransactionWindowMax(new(request.TransactionWindowMax)),
	)
	if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		return c.Render("admin/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if deliveryPlan != nil {
		return c.Render("admin/delivery_plan_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueDeliveryTransactionWindowFrameViolation.Error()),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/delivery_plan_new", params.Push(
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
	if _, err = q.deliveryPlanService.CreateDeliveryPlan(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateDeliveryPlan")
		return c.Render("admin/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("delivery plan created successfully").
		Redirect(c, sess.Session, "/admin/delivery_plans")
}

func (q *deliveryPlanController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrDeliveryPlanNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	if deliveryPlan == nil {
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Jadwal Delivery|%s", deliveryPlan.Name),
		"request", deliveryPlan,
		"deliveryPlan", deliveryPlan,
	)
	return c.Render("admin/delivery_plan_edit", params, "admin/layouts/authenticated")
}

func (q *deliveryPlanController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrDeliveryPlanNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	if deliveryPlan == nil {
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	request := deliveryPlan
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Jadwal Delivery|%s", deliveryPlan.Name),
		"request", request,
		"deliveryPlan", deliveryPlan,
	)
	request, statusCode, err := sanitizers.ValidateDeliveryPlan(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateDeliveryPlan")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/delivery_plan_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if deliveryPlan.Name != request.Name {
		otherDeliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
			ctx,
			models.NewDeliveryPlanFilter().
				WithName(request.Name).
				WithExcludedDeliveryPlanIDs(deliveryPlan.ID),
		)
		if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
			return c.Render("admin/delivery_plan_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherDeliveryPlan != nil {
			return c.Render("admin/delivery_plan_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueDeliveryPlanNameViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if deliveryPlan.TransactionWindowMin != request.TransactionWindowMin ||
		deliveryPlan.TransactionWindowMax != request.TransactionWindowMax {
		otherDeliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
			ctx,
			models.NewDeliveryPlanFilter().
				WithTransactionWindowMin(new(request.TransactionWindowMin)).
				WithTransactionWindowMax(new(request.TransactionWindowMax)).
				WithExcludedDeliveryPlanIDs(deliveryPlan.ID),
		)
		if err != nil && !errors.Is(err, models.ErrDeliveryPlanNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
			return c.Render("admin/delivery_plan_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherDeliveryPlan != nil {
			return c.Render("admin/delivery_plan_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueDeliveryTransactionWindowFrameViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = deliveryPlan.ID
	request.RecordStatus = deliveryPlan.RecordStatus
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/delivery_plan_edit", params.Push(
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
	if _, err = q.deliveryPlanService.UpdateDeliveryPlan(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
		return c.Render("admin/delivery_plan_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/delivery_plan_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("delivery plan updated successfully").
		Redirect(c, sess.Session, "/admin/delivery_plans")
}

func (q *deliveryPlanController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "DeliveryPlanController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	deliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	deliveryPlan, err := q.deliveryPlanService.FindDeliveryPlan(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryPlanIDs(deliveryPlanID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrDeliveryPlanNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlan")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	if deliveryPlan == nil {
		return flash.Danger(models.ErrDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	request := deliveryPlan
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
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
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryPlan")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/delivery_plans")
	}
	return flash.Clear(c, sess.Session).
		Success("delivery plan deleted successfully").
		Redirect(c, sess.Session, "/admin/delivery_plans")
}
