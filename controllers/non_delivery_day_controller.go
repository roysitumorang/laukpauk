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
	nonDeliveryDayController struct {
		adminService          services.AdminService
		nonDeliveryDayService services.NonDeliveryDayService
		deliveryPlanService   services.DeliveryPlanService
	}
)

func NewNonDeliveryDayController(
	adminService services.AdminService,
	nonDeliveryDayService services.NonDeliveryDayService,
	deliveryPlanService services.DeliveryPlanService,
) *nonDeliveryDayController {
	return new(nonDeliveryDayController{
		adminService:          adminService,
		nonDeliveryDayService: nonDeliveryDayService,
		deliveryPlanService:   deliveryPlanService,
	})
}

func (q *nonDeliveryDayController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *nonDeliveryDayController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NonDeliveryDayController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.NonDeliveryDayFilter)
	rows := []*models.NonDeliveryDay{}
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
				Field:     models.NonDeliveryDaySortingFieldDate,
				Direction: models.SortingDirectionAsc,
				Label:     "Date A-Z",
			},
			{
				Field:     models.NonDeliveryDaySortingFieldDate,
				Direction: models.SortingDirectionDesc,
				Label:     "Date Z-A",
			},
		},
	)
	request, err = sanitizers.FindNonDeliveryDays(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNonDeliveryDays")
		return c.Render("admin/non_delivery_day_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.WithRecordStatuses(models.RecordStatusActive)
	if rows, pagination, err = q.nonDeliveryDayService.FindNonDeliveryDays(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNonDeliveryDays")
		return c.Render("admin/non_delivery_day_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/non_delivery_day_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *nonDeliveryDayController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NonDeliveryDayController-AdminNew"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.NonDeliveryDay{
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
		return c.Render("admin/non_delivery_day_new", params, "admin/layouts/authenticated")
	}
	params.Push("deliveryPlans", deliveryPlans)
	return c.Render("admin/non_delivery_day_new", params, "admin/layouts/authenticated")
}

func (q *nonDeliveryDayController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NonDeliveryDayController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.NonDeliveryDay{
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
		return c.Render("admin/non_delivery_day_new", params, "admin/layouts/authenticated")
	}
	params.Push("deliveryPlans", deliveryPlans)
	mapDeliveryPlanIDs := map[uuid.UUID]struct{}{}
	for _, deliveryPlan := range deliveryPlans {
		mapDeliveryPlanIDs[deliveryPlan.ID] = struct{}{}
	}
	request, statusCode, err := sanitizers.ValidateNonDeliveryDay(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateNonDeliveryDay")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/non_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if _, ok := mapDeliveryPlanIDs[request.DeliveryPlanID]; !ok {
		return c.Render("admin/non_delivery_day_new", params.Push(
			"flash", flash.Danger("non_delivery_day_id: not found"),
		), "admin/layouts/authenticated")
	}
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = currentAdmin.ID
	request.CreatedAt = time.Now()
	params.Push("request", request)
	nonDeliveryDay, err := q.nonDeliveryDayService.FindNonDeliveryDay(
		ctx,
		models.NewNonDeliveryDayFilter().
			WithDates(request.Date).
			WithDeliveryPlanIDs(request.DeliveryPlanID),
	)
	if err != nil && !errors.Is(err, models.ErrNonDeliveryDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNonDeliveryDay")
		return c.Render("admin/non_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if nonDeliveryDay != nil {
		if nonDeliveryDay.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/non_delivery_day_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = nonDeliveryDay.ID
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/non_delivery_day_new", params.Push(
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
		if _, err = q.nonDeliveryDayService.CreateNonDeliveryDay(ctx, tx, request); err != nil {
			if !errors.Is(err, models.ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation) {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateNonDeliveryDay")
			}
			return c.Render("admin/non_delivery_day_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	} else if _, err = q.nonDeliveryDayService.UpdateNonDeliveryDay(ctx, tx, request); err != nil {
		if !errors.Is(err, models.ErrUniqueNonDeliveryDayDateDeliveryPlanIdViolation) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateNonDeliveryDay")
		}
		return c.Render("admin/non_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/non_delivery_day_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("non delivery day created successfully").
		Redirect(c, sess.Session, "/admin/non_delivery_days")
}

func (q *nonDeliveryDayController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "NonDeliveryDayController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	nonDeliveryDayID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrNonDeliveryDayNotFound.Error()).Redirect(c, sess.Session, "/admin/non_delivery_days")
	}
	nonDeliveryDay, err := q.nonDeliveryDayService.FindNonDeliveryDay(
		ctx,
		models.NewNonDeliveryDayFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithNonDeliveryDayIDs(nonDeliveryDayID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrNonDeliveryDayNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindNonDeliveryDay")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/non_delivery_days")
	}
	if nonDeliveryDay == nil {
		return flash.Danger(models.ErrNonDeliveryDayNotFound.Error()).Redirect(c, sess.Session, "/admin/non_delivery_days")
	}
	request := nonDeliveryDay
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/non_delivery_days")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.nonDeliveryDayService.UpdateNonDeliveryDay(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateNonDeliveryDay")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/non_delivery_days")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/non_delivery_days")
	}
	return flash.Clear(c, sess.Session).
		Success("non delivery day deleted successfully").
		Redirect(c, sess.Session, "/admin/non_delivery_days")
}
