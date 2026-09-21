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

type sellerController struct {
	sellerService services.SellerService
}

func NewSellerController(
	sellerService services.SellerService,
) *sellerController {
	return &sellerController{
		sellerService: sellerService,
	}
}

func (q *sellerController) Mount(r fiber.Router) {
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, models.SellerLevelOwner)
	r.Group("/seller", sellerSessionAuth).
		Get("", q.SellerIndex).
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate).
		Get("/:id/edit", q.SellerEdit).
		Post("/:id", q.SellerUpdate).
		Post("/:id/delete", q.SellerDelete)
}

func (q *sellerController) SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-SellerIndex"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.SellerFilter{}
	rows := []*models.Seller{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Seller Laukpauk",
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
				Field:     models.SellerSortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.SellerSortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	request, err = sanitizers.FindSellers(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSellers")
		return c.Render("seller/seller_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if rows, pagination, err = q.sellerService.FindSellers(
		ctx,
		request.WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSellers")
		return c.Render("seller/seller_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/seller_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "seller/layouts/authenticated")
}

func (q *sellerController) SellerNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Seller{}
	sellers := []*models.Seller{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Add New Seller",
		"request", request,
		"seller", request,
		"sellers", sellers,
		"sellerLevels", models.SellerLevels,
	)
	if sellers, _, err = q.sellerService.FindSellers(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SellerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return c.Render("seller/seller_new", params.Push("sellers", sellers), "seller/layouts/authenticated")
}

func (q *sellerController) SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-SellerCreate"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Seller{}
	sellers := []*models.Seller{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", "Add New Seller",
		"request", request,
		"seller", request,
		"sellers", sellers,
		"sellerLevels", models.SellerLevels,
	)
	if sellers, _, err = q.sellerService.FindSellers(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SellerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	params.Push("sellers", sellers)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateSeller(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSeller")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithUsername(request.Username),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if seller != nil {
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueSellerUsernameViolation.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if seller, err = q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithMobilePhone(request.MobilePhone),
	); err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if seller != nil {
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueSellerMobilePhoneViolation.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if request.Email != nil {
		if seller, err = q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithEmail(*request.Email),
		); err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return c.Render("seller/seller_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "seller/layouts/authenticated")
		}
		if seller != nil {
			return c.Render("seller/seller_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueSellerEmailViolation.Error()),
				"request", request.SetPhoto(""),
			), "seller/layouts/authenticated")
		}
	}
	now := time.Now().In(time.Local)
	createdBy := currentSeller.ID
	request.CompanyID = currentSeller.CompanyID
	request.SellerLevel = models.SellerLevelStaff
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedBy = &createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("SELLER_DEFAULT_PHOTO")
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
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
	if _, err = q.sellerService.CreateSeller(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSeller")
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/seller_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/seller_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("seller created successfully").
		Redirect(c, sess.Session, "/seller/sellers")
}

func (q *sellerController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	sellerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrSellerNotFound.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive).
			WithSellerIDs(sellerID),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	if seller == nil {
		return flash.Danger(models.ErrSellerNotFound.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	sellers := []*models.Seller{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", helper.Sprintf("Edit Data Seller|%s", seller.Name),
		"request", seller,
		"seller", seller,
		"sellers", sellers,
		"sellerLevels", models.SellerLevels,
	)
	if sellers, _, err = q.sellerService.FindSellers(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SellerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("seller/seller_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	params.Push("sellers", sellers)
	return c.Render("seller/seller_edit", params, "seller/layouts/authenticated")
}

func (q *sellerController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	sellerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrSellerNotFound.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive).
			WithSellerIDs(sellerID),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	if seller == nil {
		return flash.Danger(models.ErrSellerNotFound.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	sellers := []*models.Seller{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"pageTitle", helper.Sprintf("Edit Data Seller|%s", seller.Name),
		"request", seller,
		"seller", seller,
		"sellers", sellers,
		"sellerLevels", models.SellerLevels,
	)
	if sellers, _, err = q.sellerService.FindSellers(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive).
			WithSortings(models.Sorting{
				Field:     models.SellerSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("seller/seller_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	params.Push("sellers", sellers)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateSeller(ctx, c)
	if photo == nil {
		request.Photo = seller.Photo
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSeller")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/seller_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if seller.Username != request.Username {
		otherSeller, err := q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithUsername(request.Username).
				WithExcludedSellerIDs(seller.ID),
		)
		if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return c.Render("seller/seller_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherSeller != nil {
			return c.Render("seller/seller_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueSellerUsernameViolation.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if seller.MobilePhone != request.MobilePhone {
		otherSeller, err := q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithMobilePhone(request.MobilePhone).
				WithExcludedSellerIDs(seller.ID),
		)
		if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return c.Render("seller/seller_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherSeller != nil {
			return c.Render("seller/seller_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueSellerMobilePhoneViolation.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if request.Email != nil && (seller.Email == nil ||
		*seller.Email != *request.Email) {
		otherSeller, err := q.sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithCompanyIDs(currentSeller.CompanyID).
				WithEmail(*request.Email).
				WithExcludedSellerIDs(seller.ID),
		)
		if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
			return c.Render("seller/seller_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
		if otherSeller != nil {
			return c.Render("seller/seller_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueSellerEmailViolation.Error()),
			), "seller/layouts/authenticated")
		}
	}
	now := time.Now().In(time.Local)
	seller.Name = request.Name
	seller.Username = request.Username
	seller.MobilePhone = request.MobilePhone
	seller.Email = request.Email
	seller.Photo = request.Photo
	seller.UpdatedBy = &currentSeller.ID
	seller.UpdatedAt = now
	seller.DeactivatedBy = nil
	seller.DeactivatedAt = nil
	seller.DeactivationReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/seller_edit", params.Push(
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
	if _, err = q.sellerService.UpdateSeller(ctx, tx, seller); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSeller")
		return c.Render("seller/seller_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("seller/seller_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "seller/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/seller_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("seller updated successfully").
		Redirect(c, sess.Session, "/seller/sellers")
}

func (q *sellerController) SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SellerController-SellerDelete"
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	sellerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrSellerNotFound.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	if sellerID.Compare(currentSeller.ID) == 0 {
		return flash.Danger(models.ErrSellerSelfDeletingProhibited.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithCompanyIDs(currentSeller.CompanyID).
			WithStatuses(models.SellerStatusActive).
			WithSellerIDs(sellerID),
	)
	if err != nil && !errors.Is(err, models.ErrSellerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	if seller == nil {
		return flash.Danger(models.ErrSellerNotFound.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	seller.Status = models.SellerStatusDeactivated
	seller.DeactivatedBy = &currentSeller.ID
	seller.DeactivatedAt = new(time.Now())
	seller.DeactivationReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.sellerService.UpdateSeller(ctx, tx, seller); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSeller")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/sellers")
	}
	return flash.Clear(c, sess.Session).
		Success("seller deactivated successfully").
		Redirect(c, sess.Session, "/seller/sellers")
}
