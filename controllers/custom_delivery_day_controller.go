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
	customDeliveryDayController struct {
		adminService             services.AdminService
		customDeliveryDayService services.CustomDeliveryDayService
		deliveryPlanService      services.DeliveryPlanService
	}
)

func NewCustomDeliveryDayController(
	adminService services.AdminService,
	customDeliveryDayService services.CustomDeliveryDayService,
	deliveryPlanService services.DeliveryPlanService,
) *customDeliveryDayController {
	return new(customDeliveryDayController{
		adminService:             adminService,
		customDeliveryDayService: customDeliveryDayService,
		deliveryPlanService:      deliveryPlanService,
	})
}

func (q *customDeliveryDayController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *customDeliveryDayController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CustomDeliveryDayController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.CustomDeliveryDayFilter)
	rows := []*models.CustomDeliveryDay{}
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
				Field:     models.CustomDeliveryDaySortingFieldDate,
				Direction: models.SortingDirectionAsc,
				Label:     "Date A-Z",
			},
			{
				Field:     models.CustomDeliveryDaySortingFieldDate,
				Direction: models.SortingDirectionDesc,
				Label:     "Date Z-A",
			},
		},
	)
	request, err = sanitizers.FindCustomDeliveryDays(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCustomDeliveryDays")
		return c.Render("admin/custom_delivery_day_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.WithRecordStatuses(models.RecordStatusActive)
	if rows, pagination, err = q.customDeliveryDayService.FindCustomDeliveryDays(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCustomDeliveryDays")
		return c.Render("admin/custom_delivery_day_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/custom_delivery_day_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *customDeliveryDayController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CustomDeliveryDayController-AdminNew"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.CustomDeliveryDay{
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
		return c.Render("admin/custom_delivery_day_new", params, "admin/layouts/authenticated")
	}
	params.Push("deliveryPlans", deliveryPlans)
	return c.Render("admin/custom_delivery_day_new", params, "admin/layouts/authenticated")
}

func (q *customDeliveryDayController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CustomDeliveryDayController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.CustomDeliveryDay{
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
		return c.Render("admin/custom_delivery_day_new", params, "admin/layouts/authenticated")
	}
	params.Push("deliveryPlans", deliveryPlans)
	mapDeliveryPlanIDs := map[uuid.UUID]struct{}{}
	for _, deliveryPlan := range deliveryPlans {
		mapDeliveryPlanIDs[deliveryPlan.ID] = struct{}{}
	}
	request, statusCode, err := sanitizers.ValidateCustomDeliveryDay(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateCustomDeliveryDay")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/custom_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if _, ok := mapDeliveryPlanIDs[request.DeliveryPlanID]; !ok {
		return c.Render("admin/custom_delivery_day_new", params.Push(
			"flash", flash.Danger("delivery_plan_id: not found"),
		), "admin/layouts/authenticated")
	}
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = time.Now()
	params.Push("request", request)
	customDeliveryDay, err := q.customDeliveryDayService.FindCustomDeliveryDay(
		ctx,
		models.NewCustomDeliveryDayFilter().
			WithDates(request.Date).
			WithDeliveryPlanIDs(request.DeliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrCustomDeliveryDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCustomDeliveryDay")
		return c.Render("admin/custom_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if customDeliveryDay != nil {
		if customDeliveryDay.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/custom_delivery_day_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueCustomDeliveryDayDateDeliveryPlanIdViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = customDeliveryDay.ID
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/custom_delivery_day_new", params.Push(
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
	if request.ID.Compare(uuid.Nil()) == 0 {
		if _, err = q.customDeliveryDayService.CreateCustomDeliveryDay(ctx, tx, request); err != nil {
			if !errors.Is(err, models.ErrUniqueCustomDeliveryDayDateDeliveryPlanIdViolation) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateCustomDeliveryDay")
			}
			return c.Render("admin/custom_delivery_day_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	} else if _, err = q.customDeliveryDayService.UpdateCustomDeliveryDay(ctx, tx, request); err != nil {
		if !errors.Is(err, models.ErrUniqueCustomDeliveryDayDateDeliveryPlanIdViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCustomDeliveryDay")
		}
		return c.Render("admin/custom_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/custom_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("custom delivery day created successfully").
		Redirect(c, sess.Session, "/admin/custom_delivery_days")
}

func (q *customDeliveryDayController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CustomDeliveryDayController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	customDeliveryDayID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrCustomDeliveryDayNotFound.Error()).Redirect(c, sess.Session, "/admin/custom_delivery_days")
	}
	request, err := q.customDeliveryDayService.FindCustomDeliveryDay(
		ctx,
		models.NewCustomDeliveryDayFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithCustomDeliveryDayIDs(customDeliveryDayID),
	)
	if err != nil && !errors.Is(err, models.ErrCustomDeliveryDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCustomDeliveryDay")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/custom_delivery_days")
	}
	if request == nil {
		return flash.Danger(models.ErrCustomDeliveryDayNotFound.Error()).Redirect(c, sess.Session, "/admin/custom_delivery_days")
	}
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/custom_delivery_days")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.customDeliveryDayService.UpdateCustomDeliveryDay(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCustomDeliveryDay")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/custom_delivery_days")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/custom_delivery_days")
	}
	return flash.Clear(c, sess.Session).
		Success("custom delivery day deleted successfully").
		Redirect(c, sess.Session, "/admin/custom_delivery_days")
}
