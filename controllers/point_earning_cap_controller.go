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
	pointEarningCapController struct {
		adminService           services.AdminService
		pointEarningCapService services.PointEarningCapService
	}
)

func NewPointEarningCapController(
	adminService services.AdminService,
	pointEarningCapService services.PointEarningCapService,
) *pointEarningCapController {
	return &pointEarningCapController{
		adminService:           adminService,
		pointEarningCapService: pointEarningCapService,
	}
}

func (q *pointEarningCapController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *pointEarningCapController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
		"currentAdmin", currentAdmin,
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
		return c.Render("/admin/point_earning_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.pointEarningCapService.FindPointEarningCaps(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCaps")
		return c.Render("/admin/point_earning_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/point_earning_cap_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *pointEarningCapController) AdminNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Tarif Akumulasi Poin",
		"request", request,
		"pointEarningCap", request,
	)
	return c.Render("/admin/point_earning_cap_new", params, "admin/layouts/authenticated")
}

func (q *pointEarningCapController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
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
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Tarif Akumulasi Poin",
		"request", request,
		"pointEarningCap", request,
	)
	request, statusCode, err := sanitizers.ValidatePointEarningCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointEarningCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_earning_cap_new", params.Push(
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
	if request.MinimumShoppingCost > 0 && request.MaximumShoppingCost > 0 {
		pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
			ctx,
			models.NewPointEarningCapFilter().
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost),
		)
		if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
			return c.Render("admin/point_earning_cap_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if pointEarningCap != nil {
			if pointEarningCap.RecordStatus == models.RecordStatusActive {
				return c.Render("admin/point_earning_cap_new", params.Push(
					"flash", flash.Danger(models.ErrUniquePointEarningCapShoppingCostViolation.Error()),
				), "admin/layouts/authenticated")
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
		return c.Render("/admin/point_earning_cap_new", params.Push(
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
	if _, err = q.pointEarningCapService.SavePointEarningCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSavePointEarningCap")
		return c.Render("/admin/point_earning_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_earning_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point earning cap created successfully").
		Redirect(c, sess.Session, "/admin/point_earning_caps")
}

func (q *pointEarningCapController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	if pointEarningCap == nil {
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Tarif Akumulasi Poin|%d", pointEarningCap.MinimumShoppingCost),
		"request", pointEarningCap,
		"pointEarningCap", pointEarningCap,
	)
	return c.Render("/admin/point_earning_cap_edit", params, "admin/layouts/authenticated")
}

func (q *pointEarningCapController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	if pointEarningCap == nil {
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Tarif Akumulasi Poin|%d", pointEarningCap.MinimumShoppingCost),
		"request", pointEarningCap,
		"pointEarningCap", pointEarningCap,
	)
	request, statusCode, err := sanitizers.ValidatePointEarningCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointEarningCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_earning_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if (pointEarningCap.MinimumShoppingCost != request.MinimumShoppingCost) &&
		(pointEarningCap.MaximumShoppingCost != request.MaximumShoppingCost) {
		otherPointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
			ctx,
			models.NewPointEarningCapFilter().
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost).
				WithExcludedPointEarningCapIDs(pointEarningCap.ID),
		)
		if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
			return c.Render("admin/point_earning_cap_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherPointEarningCap != nil {
			return c.Render("admin/point_earning_cap_edit", params.Push(
				"flash", flash.Danger(models.ErrUniquePointEarningCapShoppingCostViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = pointEarningCap.ID
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/point_earning_cap_edit", params.Push(
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
	if _, err = q.pointEarningCapService.UpdatePointEarningCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointEarningCap")
		return c.Render("/admin/point_earning_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_earning_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point earning cap updated successfully").
		Redirect(c, sess.Session, "/admin/point_earning_caps")
}

func (q *pointEarningCapController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointEarningCapController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	pointEarningCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	pointEarningCap, err := q.pointEarningCapService.FindPointEarningCap(
		ctx,
		models.NewPointEarningCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointEarningCapIDs(pointEarningCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointEarningCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointEarningCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	if pointEarningCap == nil {
		return flash.Danger(models.ErrPointEarningCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	pointEarningCap.RecordStatus = models.RecordStatusDeleted
	pointEarningCap.DeletedBy = &currentAdmin.ID
	pointEarningCap.DeletedAt = new(time.Now())
	pointEarningCap.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_earning_caps")
	}
	return flash.Clear(c, sess.Session).
		Success("point earning cap deleted successfully").
		Redirect(c, sess.Session, "/admin/point_earning_caps")
}
