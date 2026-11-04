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

type companyController struct {
	adminService        services.AdminService
	sellerService       services.SellerService
	buyerService        services.BuyerService
	companyService      services.CompanyService
	jsonWebTokenService services.JsonWebTokenService
	notificationService services.NotificationService
}

func NewCompanyController(
	adminService services.AdminService,
	sellerService services.SellerService,
	buyerService services.BuyerService,
	companyService services.CompanyService,
	jsonWebTokenService services.JsonWebTokenService,
	notificationService services.NotificationService,
) *companyController {
	return &companyController{
		adminService:        adminService,
		sellerService:       sellerService,
		buyerService:        buyerService,
		companyService:      companyService,
		jsonWebTokenService: jsonWebTokenService,
		notificationService: notificationService,
	}
}

func (q *companyController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService, q.notificationService)
	r.Group("/seller", sellerSessionAuth).
		Get("/mine", q.SellerEdit).
		Post("/mine", q.SellerUpdate)
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	r.Group("/v1/buyer", companyVerify).
		Get("/current", q.V1BuyerShow)
}

func (q *companyController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CompanyController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.CompanyFilter{}
	rows := []*models.Company{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Company",
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
				Field:     models.CompanySortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.CompanySortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	request, err = sanitizers.FindCompanies(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompanies")
		return c.Render("admin/company_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.companyService.FindCompanies(
		ctx,
		request.WithExcludedCompanyIDs(uuid.Nil()),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompanies")
		return c.Render("admin/company_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/company_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *companyController) AdminNew(c fiber.Ctx) (err error) {
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Company{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Company",
		"request", request,
		"company", request,
	)
	return c.Render("admin/company_new", params, "admin/layouts/authenticated")
}

func (q *companyController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CompanyController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Company)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Company",
		"request", request,
		"company", request,
	)
	request, statusCode, err := sanitizers.ValidateCompany(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateCompany")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/company_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.SubscriptionStatus = models.SubscriptionStatusTrial
	request.ReferralPointsSlots = []int64{}
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = uuid.Nil()
	request.UpdatedAt = now
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/company_new", params.Push(
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
	company, err := q.companyService.CreateCompany(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateCompany")
		return c.Render("admin/company_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	seller, err := q.sellerService.CreateSeller(ctx, tx, &models.Seller{
		CompanyID:   company.ID,
		Status:      models.SellerStatusActive,
		SellerLevel: models.SellerLevelOwner,
		Name:        company.Name,
		MobilePhone: company.WhatsappPhoneNumber,
		Username:    helper.RandomString(10),
		Photo:       "seller-default.png",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSeller")
		return c.Render("admin/company_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	company.OwnerID = seller.ID
	company.UpdatedBy = seller.ID
	if _, err = q.companyService.UpdateCompany(ctx, tx, company); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCompany")
		return c.Render("admin/company_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/company_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("company created successfully").
		Redirect(c, sess.Session, "/admin/companies")
}

func (q *companyController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CompanyController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	companyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	company, err := q.companyService.FindCompany(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(companyID),
	)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompany")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	if company == nil {
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Company|%s", company.Name),
		"request", company,
		"company", company,
	)
	return c.Render("admin/company_edit", params, "admin/layouts/authenticated")
}

func (q *companyController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CompanyController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	companyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	company, err := q.companyService.FindCompany(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(companyID),
	)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompany")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	if company == nil {
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Company|%s", company.Name),
		"request", company,
		"company", company,
	)
	request, statusCode, err := sanitizers.ValidateCompany(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateCompany")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/company_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	company.Name = request.Name
	company.Subdomain = request.Subdomain
	company.Billed = request.Billed
	company.WhatsappPhoneNumber = request.WhatsappPhoneNumber
	company.UpdatedAt = time.Now()
	company.DeactivatedBy = nil
	company.DeactivatedAt = nil
	company.DeactivationReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/company_edit", params.Push(
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
	if _, err = q.companyService.UpdateCompany(ctx, tx, company); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCompany")
		return c.Render("admin/company_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/company_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("company updated successfully").
		Redirect(c, sess.Session, "/admin/companies")
}

func (q *companyController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CompanyController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := c.Locals(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	companyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	company, err := q.companyService.FindCompany(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(companyID),
	)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompany")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	if company == nil {
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	company.DeactivatedBy = &currentAdmin.ID
	company.DeactivatedAt = new(time.Now())
	company.DeactivationReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.companyService.UpdateCompany(ctx, tx, company); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCompany")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/companies")
	}
	return flash.Clear(c, sess.Session).
		Success("company deactivated successfully").
		Redirect(c, sess.Session, "/admin/companies")
}

func (q *companyController) SellerEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CompanyController-SellerEdit"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	company, err := q.companyService.FindCompany(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(currentSeller.CompanyID),
	)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompany")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/companies/mine")
	}
	if company == nil {
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/seller/companies/mine")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Data Company|%s", company.Name),
		"request", company,
		"company", company,
	)
	return c.Render("seller/company_edit", params, "seller/layouts/authenticated")
}

func (q *companyController) SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "CompanyController-SellerUpdate"
	sess := session.FromContext(c)
	currentSeller, _ := c.Locals(models.CurrentSeller).(*models.Seller)
	notifications, _ := c.Locals(models.CurrentNotifications).([]*models.Notification)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	company, err := q.companyService.FindCompany(
		ctx,
		models.NewCompanyFilter().
			WithCompanyIDs(currentSeller.CompanyID),
	)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompany")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/seller/companies/mine")
	}
	if company == nil {
		return flash.Danger(models.ErrCompanyNotFound.Error()).Redirect(c, sess.Session, "/seller/companies/mine")
	}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
		"notifications", notifications,
		"pageTitle", helper.Sprintf("Edit Data Company|%s", company.Name),
		"request", company,
		"company", company,
	)
	request, statusCode, err := sanitizers.ValidateCompany(ctx, c, true)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateCompany")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/company_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	request.ID = company.ID
	request.Billed = company.Billed
	request.BillingInterval = company.BillingInterval
	request.SubscriptionStatus = company.SubscriptionStatus
	request.SubscriptionStartDate = company.SubscriptionStartDate
	request.SubscriptionEndDate = company.SubscriptionEndDate
	request.UpdatedBy = currentSeller.ID
	request.UpdatedAt = time.Now()
	request.DeactivatedBy = nil
	request.DeactivatedAt = nil
	request.DeactivationReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/company_edit", params.Push(
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
	if _, err = q.companyService.UpdateCompany(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCompany")
		return c.Render("seller/company_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/company_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "seller/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("company updated successfully").
		Redirect(c, sess.Session, "/seller/companies/mine")
}

func (q *companyController) V1BuyerShow(c fiber.Ctx) error {
	currentCompany, _ := c.Locals(models.CurrentCompany).(*models.Company)
	return helper.NewResponse(fiber.StatusOK).
		SetData(currentCompany).
		WriteResponse(c)
}
