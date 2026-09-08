package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
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
	buyerController struct {
		adminService services.AdminService
		buyerService services.BuyerService
	}
)

func NewBuyerController(
	userService services.AdminService,
	buyerService services.BuyerService,
) *buyerController {
	return &buyerController{
		adminService: userService,
		buyerService: buyerService,
	}
}

func (q *buyerController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *buyerController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.BuyerFilter{}
	rows := []*models.Buyer{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Buyer Laukpauk",
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
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	request, err = sanitizers.FindBuyers(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("admin/buyer_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.buyerService.FindBuyers(
		ctx,
		request.WithStatuses(models.BuyerStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
		return c.Render("admin/buyer_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/buyer_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *buyerController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Buyer{}
	buyers := []*models.Buyer{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Buyer",
		"request", request,
		"buyer", request,
		"buyers", buyers,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/buyer_new", params.Push("buyers", buyers), "admin/layouts/authenticated")
}

func (q *buyerController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Buyer{}
	buyers := []*models.Buyer{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Buyer",
		"request", request,
		"buyer", request,
		"buyers", buyers,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBuyer(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBuyer")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithUsername(request.Username),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if buyer != nil {
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueBuyerUsernameViolation.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if buyer, err = q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithMobilePhone(request.MobilePhone),
	); err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if buyer != nil {
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueBuyerMobilePhoneViolation.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if request.Email != nil {
		if buyer, err = q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithEmail(*request.Email),
		); err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return c.Render("admin/buyer_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
		if buyer != nil {
			return c.Render("admin/buyer_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueBuyerEmailViolation.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
	}
	now := time.Now().In(time.Local)
	request.ReferralPoints = 0
	request.RewardPoints = 0
	request.Wallet = 0
	request.TransactionVolume = 0
	request.CreatedAt = now
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("BUYER_DEFAULT_PHOTO")
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
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
	if _, err = q.buyerService.CreateBuyer(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBuyer")
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/buyer_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/buyer_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("buyer created successfully").
		Redirect(c, sess.Session, "/admin/buyers")
}

func (q *buyerController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	buyerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBuyerNotFound.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithBuyerIDs(buyerID),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	if buyer == nil {
		return flash.Danger(models.ErrBuyerNotFound.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	buyers := []*models.Buyer{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Buyer|%s", buyer.Name),
		"request", buyer,
		"buyer", buyer,
		"buyers", buyers,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/buyer_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	return c.Render("admin/buyer_edit", params, "admin/layouts/authenticated")
}

func (q *buyerController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	buyerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBuyerNotFound.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithBuyerIDs(buyerID),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	if buyer == nil {
		return flash.Danger(models.ErrBuyerNotFound.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	buyers := []*models.Buyer{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Buyer|%s", buyer.Name),
		"request", buyer,
		"buyer", buyer,
		"buyers", buyers,
	)
	if buyers, _, err = q.buyerService.FindBuyers(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.BuyerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/buyer_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("buyers", buyers)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBuyer(ctx, c)
	if photo == nil {
		request.Photo = buyer.Photo
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBuyer")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/buyer_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if buyer.Username != request.Username {
		otherBuyer, err := q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithUsername(request.Username).
				WithExcludedBuyerIDs(buyer.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return c.Render("admin/buyer_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherBuyer != nil {
			return c.Render("admin/buyer_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueBuyerUsernameViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if buyer.MobilePhone != request.MobilePhone {
		otherBuyer, err := q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithMobilePhone(request.MobilePhone).
				WithExcludedBuyerIDs(buyer.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return c.Render("admin/buyer_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherBuyer != nil {
			return c.Render("admin/admin_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueBuyerMobilePhoneViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if request.Email != nil &&
		(buyer.Email == nil ||
			*buyer.Email != *request.Email) {
		otherBuyer, err := q.buyerService.FindBuyer(
			ctx,
			models.NewBuyerFilter().
				WithEmail(*request.Email).
				WithExcludedBuyerIDs(buyer.ID),
		)
		if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
			return c.Render("admin/buyer_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherBuyer != nil {
			return c.Render("admin/buyer_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueBuyerEmailViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	now := time.Now().In(time.Local)
	buyer.Name = request.Name
	buyer.Username = request.Username
	buyer.Photo = request.Photo
	buyer.Address = request.Address
	buyer.ReferralPoints = 0
	buyer.RewardPoints = 0
	buyer.Wallet = 0
	buyer.TransactionVolume = 0
	buyer.ReferrerID = request.ReferrerID
	buyer.UpdatedAt = now
	buyer.DeactivatedBy = nil
	buyer.DeactivatedAt = nil
	buyer.DeactivationReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/buyer_edit", params.Push(
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
	if _, err = q.buyerService.UpdateBuyer(ctx, tx, buyer); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
		return c.Render("admin/buyer_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/buyer_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/buyer_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("buyer updated successfully").
		Redirect(c, sess.Session, "/admin/buyers")
}

func (q *buyerController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BuyerController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	buyerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrBuyerNotFound.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithStatuses(models.BuyerStatusActive).
			WithBuyerIDs(buyerID),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	if buyer == nil {
		return flash.Danger(models.ErrBuyerNotFound.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	buyer.Status = models.BuyerStatusDeactivated
	buyer.DeactivatedBy = &currentAdmin.ID
	buyer.DeactivatedAt = new(time.Now())
	buyer.DeactivationReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.buyerService.UpdateBuyer(ctx, tx, buyer); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/buyers")
	}
	return flash.Clear(c, sess.Session).
		Success("buyer deactivated successfully").
		Redirect(c, sess.Session, "/admin/buyers")
}
