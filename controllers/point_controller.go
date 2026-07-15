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
	pointController struct {
		adminService services.AdminService
		buyerService services.BuyerService
		pointService services.PointService
	}
)

func NewPointController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	pointService services.PointService,
) *pointController {
	return &pointController{
		adminService: adminService,
		buyerService: buyerService,
		pointService: pointService,
	}
}

func (q *pointController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *pointController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.PointFilter{}
	rows := []*models.Point{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Histori Point User",
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
	request, err = sanitizers.FindPoints(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoints")
		return c.Render("/admin/point_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.WithRecordStatuses(models.RecordStatusActive)
	if rows, pagination, err = q.pointService.FindPoints(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoints")
		return c.Render("/admin/point_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/point_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *pointController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminNew"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Point)
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Point User",
		"request", request,
		"point", request,
		"buyers", buyers,
		"pointTypes", models.PointTypes,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter(),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/point_new", params.Push("buyers", buyers), "admin/layouts/authenticated")
}

func (q *pointController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Point)
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Point User",
		"request", request,
		"point", request,
		"buyers", buyers,
		"pointTypes", models.PointTypes,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter(),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	mapBuyerIDs := map[uuid.UUID]struct{}{}
	for _, buyer := range buyers {
		mapBuyerIDs[buyer.ID] = struct{}{}
	}
	request, statusCode, err := sanitizers.ValidatePoint(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePoint")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if _, ok := mapBuyerIDs[request.BuyerID]; !ok {
		return c.Render("/admin/point_new", params.Push(
			"flash", flash.Danger("buyer_id: not found"),
		), "admin/layouts/authenticated")
	}
	createdBy := currentAdmin.ID
	now := time.Now()
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedBy = &createdBy
	request.UpdatedAt = now
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/point_new", params.Push(
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
	if _, err = q.pointService.CreatePoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePoint")
		return c.Render("/admin/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point created successfully").
		Redirect(c, sess.Session, "/admin/points")
}

func (q *pointController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	if point == nil {
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Poin User",
		"request", point,
		"point", point,
		"buyers", buyers,
		"pointTypes", models.PointTypes,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(point.BuyerID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/point_edit", params.Push("buyers", buyers), "admin/layouts/authenticated")
}

func (q *pointController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	if point == nil {
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Poin User",
		"request", point,
		"point", point,
		"buyers", buyers,
		"pointTypes", models.PointTypes,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(point.BuyerID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	request, statusCode, err := sanitizers.ValidatePoint(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePoint")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = point.ID
	request.RecordStatus = point.RecordStatus
	request.PointType = point.PointType
	request.BuyerID = point.BuyerID
	request.TransactionID = point.TransactionID
	request.UpdatedBy = &currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/point_edit", params.Push(
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
	if _, err = q.pointService.UpdatePoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
		return c.Render("/admin/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point updated successfully").
		Redirect(c, sess.Session, "/admin/points")
}

func (q *pointController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrBlogNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	if point == nil {
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	request := point
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.pointService.UpdatePoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/points")
	}
	return flash.Clear(c, sess.Session).
		Success("point deleted successfully").
		Redirect(c, sess.Session, "/admin/points")
}
