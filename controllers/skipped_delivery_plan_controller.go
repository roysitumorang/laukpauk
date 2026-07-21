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
	skippedDeliveryPlanController struct {
		adminService               services.AdminService
		skippedDeliveryPlanService services.SkippedDeliveryPlanService
		deliveryPlanService        services.DeliveryPlanService
	}
)

func NewSkippedDeliveryPlanController(
	adminService services.AdminService,
	skippedDeliveryPlanService services.SkippedDeliveryPlanService,
	deliveryPlanService services.DeliveryPlanService,
) *skippedDeliveryPlanController {
	return &skippedDeliveryPlanController{
		adminService:               adminService,
		skippedDeliveryPlanService: skippedDeliveryPlanService,
		deliveryPlanService:        deliveryPlanService,
	}
}

func (q *skippedDeliveryPlanController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *skippedDeliveryPlanController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SkippedDeliveryPlanController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.SkippedDeliveryPlanFilter)
	rows := []*models.SkippedDeliveryPlan{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Jadwal Tutup",
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
				Field:     models.SkippedDeliveryPlanSortingFieldDate,
				Direction: models.SortingDirectionAsc,
				Label:     "Date A-Z",
			},
			{
				Field:     models.SkippedDeliveryPlanSortingFieldDate,
				Direction: models.SortingDirectionDesc,
				Label:     "Date Z-A",
			},
		},
	)
	request, err = sanitizers.FindSkippedDeliveryPlans(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSkippedDeliveryPlans")
		return c.Render("admin/skipped_delivery_plan_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.WithRecordStatuses(models.RecordStatusActive)
	if rows, pagination, err = q.skippedDeliveryPlanService.FindSkippedDeliveryPlans(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSkippedDeliveryPlans")
		return c.Render("admin/skipped_delivery_plan_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/skipped_delivery_plan_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *skippedDeliveryPlanController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SkippedDeliveryPlanController-AdminNew"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.SkippedDeliveryPlan{
		Date: time.Now(),
	})
	var deliveryPlans []*models.DeliveryPlan
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Jadwal Tutup",
		"request", request,
		"deliveryPlans", deliveryPlans,
	)
	if deliveryPlans, _, err = q.deliveryPlanService.FindDeliveryPlans(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return c.Render("admin/skipped_delivery_plan_new", params, "admin/layouts/authenticated")
	}
	params.Push("deliveryPlans", deliveryPlans)
	return c.Render("admin/skipped_delivery_plan_new", params, "admin/layouts/authenticated")
}

func (q *skippedDeliveryPlanController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SkippedDeliveryPlanController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.SkippedDeliveryPlan{
		Date: time.Now(),
	})
	var deliveryPlans []*models.DeliveryPlan
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Jadwal Tutup",
		"request", request,
		"deliveryPlans", deliveryPlans,
	)
	if deliveryPlans, _, err = q.deliveryPlanService.FindDeliveryPlans(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return c.Render("admin/skipped_delivery_plan_new", params, "admin/layouts/authenticated")
	}
	params.Push("deliveryPlans", deliveryPlans)
	mapDeliveryPlanIDs := map[uuid.UUID]struct{}{}
	for _, deliveryPlan := range deliveryPlans {
		mapDeliveryPlanIDs[deliveryPlan.ID] = struct{}{}
	}
	request, statusCode, err := sanitizers.ValidateSkippedDeliveryPlan(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSkippedDeliveryPlan")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/skipped_delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if _, ok := mapDeliveryPlanIDs[request.DeliveryPlanID]; !ok {
		return c.Render("admin/skipped_delivery_plan_new", params.Push(
			"flash", flash.Danger("delivery_plan_id: not found"),
		), "admin/layouts/authenticated")
	}
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = time.Now()
	params.Push("request", request)
	skippedDeliveryPlan, err := q.skippedDeliveryPlanService.FindSkippedDeliveryPlan(
		ctx,
		models.NewSkippedDeliveryPlanFilter().
			WithDates(request.Date).
			WithDeliveryPlanIDs(request.DeliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrSkippedDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSkippedDeliveryPlan")
		return c.Render("admin/skipped_delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if skippedDeliveryPlan != nil {
		if skippedDeliveryPlan.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/skipped_delivery_plan_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueSkippedDeliveryPlanDateDeliveryPlanIdViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = skippedDeliveryPlan.ID
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/skipped_delivery_plan_new", params.Push(
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
	if _, err = q.skippedDeliveryPlanService.SaveSkippedDeliveryPlan(ctx, tx, request); err != nil {
		if !errors.Is(err, models.ErrUniqueSkippedDeliveryPlanDateDeliveryPlanIdViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveSkippedDeliveryPlan")
		}
		return c.Render("admin/skipped_delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/skipped_delivery_plan_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("skipped delivery plan created successfully").
		Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
}

func (q *skippedDeliveryPlanController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SkippedDeliveryPlanController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	skippedDeliveryPlanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrSkippedDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
	}
	request, err := q.skippedDeliveryPlanService.FindSkippedDeliveryPlan(
		ctx,
		models.NewSkippedDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithSkippedDeliveryPlanIDs(skippedDeliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrSkippedDeliveryPlanNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSkippedDeliveryPlan")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
	}
	if request == nil {
		return flash.Danger(models.ErrSkippedDeliveryPlanNotFound.Error()).Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
	}
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.skippedDeliveryPlanService.UpdateSkippedDeliveryPlan(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSkippedDeliveryPlan")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
	}
	return flash.Clear(c, sess.Session).
		Success("skipped delivery plan deleted successfully").
		Redirect(c, sess.Session, "/admin/skipped_delivery_plans")
}
