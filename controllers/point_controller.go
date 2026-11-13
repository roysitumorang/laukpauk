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

type pointController struct {
	sellerService       services.SellerService
	buyerService        services.BuyerService
	pointService        services.PointService
	notificationService services.NotificationService
}

func NewPointController(
	sellerService services.SellerService,
	buyerService services.BuyerService,
	pointService services.PointService,
	notificationService services.NotificationService,
) *pointController {
	return &pointController{
		sellerService:       sellerService,
		buyerService:        buyerService,
		pointService:        pointService,
		notificationService: notificationService,
	}
}

func (q *pointController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
}

func (q *pointController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
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
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/point_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.pointService.FindPoints(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoints")
		return c.Render("/seller/point_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("/seller/point_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *pointController) SellerNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-SellerNew"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Point{}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("/seller/point_new", params.Push("buyers", buyers), "seller/layouts/authenticated")
}

func (q *pointController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Point{}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
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
		return c.Render("/seller/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if _, ok := mapBuyerIDs[request.BuyerID]; !ok {
		return c.Render("/seller/point_new", params.Push(
			"flash", flash.Danger("buyer_id: not found"),
		), "seller/layouts/authenticated")
	}
	createdBy := currentSeller.ID
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
		return c.Render("/seller/point_new", params.Push(
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
	if _, err = q.pointService.CreatePoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePoint")
		return c.Render("/seller/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/seller/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point created successfully").
		Redirect(c, sess.Session, "/seller/points")
}

func (q *pointController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	if point == nil {
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("/seller/point_edit", params.Push("buyers", buyers), "seller/layouts/authenticated")
}

func (q *pointController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	if point == nil {
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
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
		return c.Render("/seller/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	request, statusCode, err := sanitizers.ValidatePoint(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePoint")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/seller/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	request.ID = point.ID
	request.RecordStatus = point.RecordStatus
	request.PointType = point.PointType
	request.BuyerID = point.BuyerID
	request.TransactionID = point.TransactionID
	request.UpdatedBy = &currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/seller/point_edit", params.Push(
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
	if _, err = q.pointService.UpdatePoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
		return c.Render("/seller/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/seller/point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point updated successfully").
		Redirect(c, sess.Session, "/seller/points")
}

func (q *pointController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	pointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	point, err := q.pointService.FindPoint(
		ctx,
		models.NewPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithPointIDs(pointID),
	)
	if err != nil && !errors.Is(err, models.ErrPointNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	if point == nil {
		return flash.Danger(models.ErrPointNotFound.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	point.RecordStatus = models.RecordStatusDeleted
	point.DeletedBy = &currentSeller.ID
	point.DeletedAt = new(time.Now())
	point.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.pointService.UpdatePoint(ctx, tx, point); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/points")
	}
	return flash.Clear(c, sess.Session).
		Success("point deleted successfully").
		Redirect(c, sess.Session, "/seller/points")
}
