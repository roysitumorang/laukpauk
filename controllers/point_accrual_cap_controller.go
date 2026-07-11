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
	pointAccrualCapController struct {
		adminService           services.AdminService
		pointAccrualCapService services.PointAccrualCapService
	}
)

func NewPointAccrualCapController(
	adminService services.AdminService,
	pointAccrualCapService services.PointAccrualCapService,
) *pointAccrualCapController {
	return new(pointAccrualCapController{
		adminService:           adminService,
		pointAccrualCapService: pointAccrualCapService,
	})
}

func (q *pointAccrualCapController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *pointAccrualCapController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointAccrualCapController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.PointAccrualCapFilter)
	rows := []*models.PointAccrualCap{}
	pagination := new(models.Pagination)
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
			{
				Field:     models.PointAccrualCapSortingFieldMinimumPurchase,
				Direction: models.SortingDirectionDesc,
				Label:     "Max-Min",
			},
			{
				Field:     models.PointAccrualCapSortingFieldMinimumPurchase,
				Direction: models.SortingDirectionAsc,
				Label:     "Min-Max",
			},
		},
	)
	request, err = sanitizers.FindPointAccrualCaps(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCaps")
		return c.Render("/admin/point_accrual_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.pointAccrualCapService.FindPointAccrualCaps(
		c,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCaps")
		return c.Render("/admin/point_accrual_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/point_accrual_cap_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *pointAccrualCapController) AdminNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.PointAccrualCap)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Tarif Akumulasi Poin",
		"request", request,
	)
	return c.Render("/admin/point_accrual_cap_new", params, "admin/layouts/authenticated")
}

func (q *pointAccrualCapController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointAccrualCapController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.PointAccrualCap)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Tarif Akumulasi Poin",
		"request", request,
	)
	request, statusCode, err := sanitizers.ValidatePointAccrualCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointAccrualCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_accrual_cap_new", params.Push(
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
	if request.MinimumPurchase != nil {
		pointAccrualCap, err := q.pointAccrualCapService.FindPointAccrualCap(
			ctx,
			models.NewPointAccrualCapFilter().
				WithMinimumPurchases(*request.MinimumPurchase),
		)
		if err != nil && !errors.Is(err, models.ErrPointAccrualCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCap")
			return c.Render("admin/point_accrual_cap_new", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if pointAccrualCap != nil {
			if pointAccrualCap.RecordStatus == models.RecordStatusActive {
				return c.Render("admin/point_accrual_cap_new", params.Push(
					"flash", flash.Danger(models.ErrUniquePointAccrualCapMinimumPurchaseViolation.Error()),
				), "admin/layouts/authenticated")
			}
			request.ID = pointAccrualCap.ID
			request.DeletedMinimumPurchase = nil
			request.DeletedBy = nil
			request.DeletedAt = nil
			request.DeletionReason = nil
		}
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/point_accrual_cap_new", params.Push(
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
		if _, err = q.pointAccrualCapService.CreatePointAccrualCap(ctx, tx, request); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePointAccrualCap")
		}
	} else if _, err = q.pointAccrualCapService.UpdatePointAccrualCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointAccrualCap")
	}
	if err != nil {
		return c.Render("/admin/point_accrual_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_accrual_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point accrual cap created successfully").
		Redirect(c, sess.Session, "/admin/point_accrual_caps")
}

func (q *pointAccrualCapController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointAccrualCapController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	pointAccrualCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointAccrualCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	pointAccrualCap, err := q.pointAccrualCapService.FindPointAccrualCap(
		ctx,
		models.NewPointAccrualCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointAccrualCapIDs(pointAccrualCapID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrPointAccrualCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCap")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	if pointAccrualCap == nil {
		return flash.Danger(models.ErrPointAccrualCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Tarif Akumulasi Poin|%d", pointAccrualCap.MinimumPurchase),
		"request", pointAccrualCap,
		"pointAccrualCap", pointAccrualCap,
	)
	return c.Render("/admin/point_accrual_cap_edit", params, "admin/layouts/authenticated")
}

func (q *pointAccrualCapController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointAccrualCapController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	pointAccrualCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointAccrualCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	pointAccrualCap, err := q.pointAccrualCapService.FindPointAccrualCap(
		ctx,
		models.NewPointAccrualCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointAccrualCapIDs(pointAccrualCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointAccrualCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	if pointAccrualCap == nil {
		return flash.Danger(models.ErrPointAccrualCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Tarif Akumulasi Poin|%d", pointAccrualCap.MinimumPurchase),
		"request", pointAccrualCap,
		"point", pointAccrualCap,
	)
	request, statusCode, err := sanitizers.ValidatePointAccrualCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointAccrualCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_accrual_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if pointAccrualCap.MinimumPurchase != nil &&
		request.MinimumPurchase != nil &&
		*pointAccrualCap.MinimumPurchase != *request.MinimumPurchase {
		otherPointAccrualCap, err := q.pointAccrualCapService.FindPointAccrualCap(
			ctx,
			models.NewPointAccrualCapFilter().
				WithMinimumPurchases(*request.MinimumPurchase).
				WithExcludedPointAccrualCapIDs(pointAccrualCap.ID),
		)
		if err != nil && !errors.Is(err, models.ErrPointAccrualCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCap")
			return c.Render("admin/point_accrual_cap_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherPointAccrualCap != nil {
			return c.Render("admin/point_accrual_cap_edit", params.Push(
				"flash", flash.Danger(models.ErrUniquePointAccrualCapMinimumPurchaseViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = pointAccrualCap.ID
	request.DeletedMinimumPurchase = nil
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/point_accrual_cap_edit", params.Push(
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
	if _, err = q.pointAccrualCapService.UpdatePointAccrualCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointAccrualCap")
		return c.Render("/admin/point_accrual_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_accrual_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point accrual cap updated successfully").
		Redirect(c, sess.Session, "/admin/point_accrual_caps")
}

func (q *pointAccrualCapController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointAccrualCapController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	pointAccrualCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointAccrualCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	pointAccrualCap, err := q.pointAccrualCapService.FindPointAccrualCap(
		ctx,
		models.NewPointAccrualCapFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointAccrualCapIDs(pointAccrualCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointAccrualCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointAccrualCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	if pointAccrualCap == nil {
		return flash.Danger(models.ErrPointAccrualCapNotFound.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	request := pointAccrualCap
	if request.MinimumPurchase != nil {
		request.DeletedMinimumPurchase = new(*pointAccrualCap.MinimumPurchase)
	}
	request.MinimumPurchase = nil
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.pointAccrualCapService.UpdatePointAccrualCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointAccrualCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/point_accrual_caps")
	}
	return flash.Clear(c, sess.Session).
		Success("point accrual cap deleted successfully").
		Redirect(c, sess.Session, "/admin/point_accrual_caps")
}
