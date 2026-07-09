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
	rewardPointController struct {
		adminService       services.AdminService
		buyerService       services.BuyerService
		rewardPointService services.RewardPointService
	}
)

func NewRewardPointController(
	adminService services.AdminService,
	buyerService services.BuyerService,
	rewardPointService services.RewardPointService,
) *rewardPointController {
	return new(rewardPointController{
		adminService:       adminService,
		buyerService:       buyerService,
		rewardPointService: rewardPointService,
	})
}

func (q *rewardPointController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *rewardPointController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "RewardPointController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.RewardPointFilter)
	rows := []*models.RewardPoint{}
	pagination := new(models.Pagination)
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
	request, err = sanitizers.FindRewardPoints(ctx, c)
	request.WithRecordStatuses(models.RecordStatusActive)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRewardPoints")
		return c.Render("/admin/reward_point_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.rewardPointService.FindRewardPoints(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRewardPoints")
		return c.Render("/admin/reward_point_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/reward_point_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *rewardPointController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "RewardPointController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.RewardPoint{
		RewardPointType: models.RewardPointTypeByAdmin,
	})
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Point User",
		"request", request,
		"buyers", buyers,
		"rewardPointTypeByAdmin", models.RewardPointTypeByAdmin,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter(),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/reward_point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/reward_point_new", params.Push("buyers", buyers), "admin/layouts/authenticated")
}

func (q *rewardPointController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.RewardPoint{
		RewardPointType: models.RewardPointTypeByAdmin,
	})
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Tambah Point User",
		"request", request,
		"buyers", buyers,
		"rewardPointTypeByAdmin", models.RewardPointTypeByAdmin,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter(),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/reward_point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	mapBuyerIDs := map[uuid.UUID]struct{}{}
	for _, buyer := range buyers {
		mapBuyerIDs[buyer.ID] = struct{}{}
	}
	request, statusCode, err := sanitizers.ValidateRewardPoint(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRewardPoint")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/reward_point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if _, ok := mapBuyerIDs[request.BuyerID]; !ok {
		return c.Render("/admin/reward_point_new", params.Push(
			"flash", flash.Danger("buyer_id: not found"),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.RewardPointType = models.RewardPointTypeByAdmin
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedBy = &createdBy
	request.UpdatedAt = now
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/reward_point_new", params.Push(
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
	if _, err = q.rewardPointService.CreateRewardPoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateRewardPoint")
		return c.Render("/admin/reward_point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/reward_point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("reward point created successfully").
		Redirect(c, sess.Session, "/admin/reward_points")
}

func (q *rewardPointController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "RewardPointController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	rewardPointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrRewardPointNotFound.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	rewardPoint, err := q.rewardPointService.FindRewardPoint(
		ctx,
		models.NewRewardPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRewardPointIDs(rewardPointID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrRewardPointNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRewardPoint")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	if rewardPoint == nil {
		return flash.Danger(models.ErrRewardPointNotFound.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Poin User",
		"request", rewardPoint,
		"point", rewardPoint,
		"buyers", buyers,
		"rewardPointTypeByAdmin", models.RewardPointTypeByAdmin,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(rewardPoint.BuyerID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/reward_point_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("/admin/reward_point_edit", params.Push("buyers", buyers), "admin/layouts/authenticated")
}

func (q *rewardPointController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	rewardPointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrRewardPointNotFound.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	rewardPoint, err := q.rewardPointService.FindRewardPoint(
		ctx,
		models.NewRewardPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRewardPointIDs(rewardPointID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrRewardPointNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRewardPoint")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	if rewardPoint == nil {
		return flash.Danger(models.ErrRewardPointNotFound.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	var buyers []*models.Buyer
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Edit Data Poin User",
		"request", rewardPoint,
		"point", rewardPoint,
		"buyers", buyers,
		"rewardPointTypeByAdmin", models.RewardPointTypeByAdmin,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(rewardPoint.BuyerID),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("/admin/reward_point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	request, statusCode, err := sanitizers.ValidateRewardPoint(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateRewardPoint")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/admin/reward_point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request.ID = rewardPoint.ID
	request.RecordStatus = rewardPoint.RecordStatus
	request.RewardPointType = rewardPoint.RewardPointType
	request.BuyerID = rewardPoint.BuyerID
	request.UpdatedBy = &currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if request.RewardPointType == models.RewardPointTypeByTransaction {
		request.Description = rewardPoint.Description
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/admin/reward_point_edit", params.Push(
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
	if _, err = q.rewardPointService.UpdateRewardPoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRewardPoint")
		return c.Render("/admin/reward_point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/admin/reward_point_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("reward point updated successfully").
		Redirect(c, sess.Session, "/admin/reward_points")
}

func (q *rewardPointController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "RewardPointController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	rewardPointID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrRewardPointNotFound.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	rewardPoint, err := q.rewardPointService.FindRewardPoint(
		ctx,
		models.NewRewardPointFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithRewardPointIDs(rewardPointID),
	)
	if err != nil {
		if !errors.Is(err, models.ErrRewardPointNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindRewardPoint")
		}
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	if rewardPoint == nil {
		return flash.Danger(models.ErrRewardPointNotFound.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	request := rewardPoint
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.rewardPointService.UpdateRewardPoint(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateRewardPoint")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/reward_points")
	}
	return flash.Clear(c, sess.Session).
		Success("reward point deleted successfully").
		Redirect(c, sess.Session, "/admin/reward_points")
}
