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
	sellerService           services.SellerService
	companyService          services.CompanyService
	pointSpendingCapService services.PointSpendingCapService
	buyerService            services.BuyerService
	jsonWebTokenService     services.JsonWebTokenService
}

func NewPointSpendingCapController(
	sellerService services.SellerService,
	companyService services.CompanyService,
	pointSpendingCapService services.PointSpendingCapService,
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
) *pointSpendingCapController {
	return &pointSpendingCapController{
		sellerService:           sellerService,
		companyService:          companyService,
		pointSpendingCapService: pointSpendingCapService,
		buyerService:            buyerService,
		jsonWebTokenService:     jsonWebTokenService,
	}
}

func (q *pointSpendingCapController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	r.Group("/v1/buyer", companyVerify, buyerKeyAuth).
		Get("", q.V1BuyerIndex)
}

func (q *pointSpendingCapController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
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
		"currentSeller", currentSeller,
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
		return c.Render("/seller/point_spending_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.pointSpendingCapService.FindPointSpendingCaps(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
		return c.Render("/seller/point_spending_cap_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("/seller/point_spending_cap_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *pointSpendingCapController) SellerNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
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
		"currentSeller", currentSeller,
		"pageTitle", "Tambah Tarif Penggunaan Poin",
		"request", request,
		"pointSpendingCap", request,
	)
	return c.Render("/seller/point_spending_cap_new", params, "seller/layouts/authenticated")
}

func (q *pointSpendingCapController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
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
		"currentSeller", currentSeller,
		"pageTitle", "Tambah Tarif Penggunaan Poin",
		"request", request,
		"pointSpendingCap", request,
	)
	request, statusCode, err := sanitizers.ValidatePointSpendingCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointSpendingCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/seller/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	createdBy := currentSeller.ID
	now := time.Now()
	request.CompanyID = currentSeller.CompanyID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	params.Push("request", request)
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return c.Render("seller/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if pointSpendingCap != nil {
		if pointSpendingCap.RecordStatus == models.RecordStatusActive {
			return c.Render("seller/point_spending_cap_new", params.Push(
				"flash", flash.Danger(models.ErrPointSpendingCapShoppingCostOverlapping.Error()),
			), "seller/layouts/authenticated")
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
		return c.Render("/seller/point_spending_cap_new", params.Push(
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
	if _, err = q.pointSpendingCapService.SavePointSpendingCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSavePointSpendingCap")
		return c.Render("/seller/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/seller/point_spending_cap_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point spending cap created successfully").
		Redirect(c, sess.Session, "/seller/point_spending_caps")
}

func (q *pointSpendingCapController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	if pointSpendingCap == nil {
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Edit Tarif Penggunaan Poin",
		"request", pointSpendingCap,
		"pointSpendingCap", pointSpendingCap,
	)
	return c.Render("/seller/point_spending_cap_edit", params, "seller/layouts/authenticated")
}

func (q *pointSpendingCapController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	pointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	if pointSpendingCap == nil {
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Edit Tarif Penggunaan Poin",
		"request", pointSpendingCap,
		"pointSpendingCap", pointSpendingCap,
	)
	request, statusCode, err := sanitizers.ValidatePointSpendingCap(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidatePointSpendingCap")
		c.Response().SetStatusCode(statusCode)
		return c.Render("/seller/point_spending_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if pointSpendingCap.MinimumShoppingCost != request.MinimumShoppingCost ||
		pointSpendingCap.MaximumShoppingCost != request.MaximumShoppingCost {
		otherPointSpendingCap, err := q.pointSpendingCapService.FindPointSpendingCap(
			ctx,
			models.NewPointSpendingCapFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithShoppingCost(request.MinimumShoppingCost, request.MaximumShoppingCost).
				WithExcludedPointSpendingCapIDs(pointSpendingCap.ID),
		)
		if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
			return c.Render("seller/point_spending_cap_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherPointSpendingCap != nil {
			return c.Render("seller/point_spending_cap_edit", params.Push(
				"flash", flash.Danger(models.ErrPointSpendingCapShoppingCostOverlapping.Error()),
			), "seller/layouts/authenticated")
		}
	}
	request.ID = pointSpendingCap.ID
	request.CompanyID = currentSeller.CompanyID
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("/seller/point_spending_cap_edit", params.Push(
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
	if _, err = q.pointSpendingCapService.UpdatePointSpendingCap(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePointSpendingCap")
		return c.Render("/seller/point_spending_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("/seller/point_spending_cap_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("point spending cap updated successfully").
		Redirect(c, sess.Session, "/seller/point_spending_caps")
}

func (q *pointSpendingCapController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	pointSpendingCapID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	request, err := q.pointSpendingCapService.FindPointSpendingCap(
		ctx,
		models.NewPointSpendingCapFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithRecordStatuses(models.RecordStatusActive).
			WithPointSpendingCapIDs(pointSpendingCapID),
	)
	if err != nil && !errors.Is(err, models.ErrPointSpendingCapNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCap")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	if request == nil {
		return flash.Danger(models.ErrPointSpendingCapNotFound.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentSeller.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
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
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/point_spending_caps")
	}
	return flash.Clear(c, sess.Session).
		Success("point spending cap deleted successfully").
		Redirect(c, sess.Session, "/seller/point_spending_caps")
}

func (q *pointSpendingCapController) V1BuyerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "PointSpendingCapController-V1BuyerIndex"
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	response, _, err := q.pointSpendingCapService.FindPointSpendingCaps(
		ctx,
		models.NewPointSpendingCapFilter().
			WithCompanyIDs(currentBuyer.CompanyID).
			WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPointSpendingCaps")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(response).
		WriteResponse(c)
}
