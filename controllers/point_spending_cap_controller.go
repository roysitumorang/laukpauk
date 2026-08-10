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

type pointSpendingCapController struct {
	adminService            services.AdminService
	pointSpendingCapService services.PointSpendingCapService
}

func NewPointSpendingCapController(
	adminService services.AdminService,
	pointSpendingCapService services.PointSpendingCapService,
) *pointSpendingCapController {
	return &pointSpendingCapController{
		adminService:            adminService,
		pointSpendingCapService: pointSpendingCapService,
	}
}

func (q *pointSpendingCapController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *pointSpendingCapController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.PointSpendingCapFilter{}
	rows := []*models.PointSpendingCap{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tarif Penggunaan Poin",
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
	request, err = sanitizers.FindPointSpendingCaps(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
		return c.Render("/admin/point_spending_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.pointSpendingCapService.FindPointSpendingCaps(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
		return c.Render("/admin/point_spending_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/point_spending_cap_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *pointSpendingCapController) AdminNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.PointSpendingCap{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Tarif Penggunaan Poin",
		"request", request,
		"pointSpendingCap", request,
	)
	return c.Render("/admin/point_spending_cap_new", params, "admin/layouts/authenticated")
}

func (q *pointSpendingCapController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.PointSpendingCap{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Tarif Penggunaan Poin",
		"request", request,
		"pointSpendingCap", request,
	)
	request, statusCode, err := sanitizers.ValidatePointSpendingCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointSpendingCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	createdBy := currentAdmin.ID
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	params.Push("request", request)
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return c.Render("admin/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if pointSpendingCap != nil {
		if pointSpendingCap.RecordStatus == models.RecordStatusActive {
			return c.Render("admin/point_spending_cap_new", params.Push(
				"flash", flash.Danger(models.ErrPointSpendingCapShoppingCostOverlapping.Error()),
			), "admin/layouts/authenticated")
		}
		request.ID = pointSpendingCap.ID
		request.DeletedBy = nil
		request.DeletedAt = nil
		request.DeletionReason = nil
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/point_spending_cap_new", params.Push(
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
	if _, err = q.pointSpendingCapService.SavePointSpendingCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSavePointSpendingCap")
		return c.Render("/admin/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point spending cap created successfully").
		Redirect(c, sess.Session, "/admin/point_spending_caps")
}

func (q *pointSpendingCapController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrPointSpendingCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	if pointSpendingCap == nil {
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Tarif Penggunaan Poin",
		"request", pointSpendingCap,
		"pointSpendingCap", pointSpendingCap,
	)
	return c.Render("/admin/point_spending_cap_edit", params, "admin/layouts/authenticated")
}

func (q *pointSpendingCapController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	if pointSpendingCap == nil {
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Tarif Penggunaan Poin",
		"request", pointSpendingCap,
		"pointSpendingCap", pointSpendingCap,
	)
	request, statusCode, err := sanitizers.ValidatePointSpendingCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointSpendingCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_spending_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if pointSpendingCap.MinimumShoppingCost != request.MinimumShoppingCost ||
		pointSpendingCap.MaximumShoppingCost != request.MaximumShoppingCost {
		otherPointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
			ctx,
			models.NewPointSpendingCapFilter().
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost).
				WithExcludedPointSpendingCapIDs(pointSpendingCap.ID),
		)
		if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
			return c.Render("admin/point_spending_cap_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherPointSpendingCap != nil {
			return c.Render("admin/point_spending_cap_edit", params.Push(
				"flash", flash.Danger(models.ErrPointSpendingCapShoppingCostOverlapping.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = pointSpendingCap.ID
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/point_spending_cap_edit", params.Push(
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
	if _, err = q.pointSpendingCapService.UpdatePointSpendingCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointSpendingCap")
		return c.Render("/admin/point_spending_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_spending_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point spending cap updated successfully").
		Redirect(c, sess.Session, "/admin/point_spending_caps")
}

func (q *pointSpendingCapController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	request, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	if request == nil {
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.pointSpendingCapService.UpdatePointSpendingCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointSpendingCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_spending_caps")
	}
	return flash.Clear(c, sess.Session).
		Success("point spending cap deleted successfully").
		Redirect(c, sess.Session, "/admin/point_spending_caps")
}
