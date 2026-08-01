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
	adminController struct {
		adminService services.AdminService
	}
)

func NewAdminController(
	adminService services.AdminService,
) *adminController {
	return &adminController{
		adminService: adminService,
	}
}

func (q *adminController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *adminController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AdminController-AdminIndex"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.AdminFilter{}
	rows := []*models.Admin{}
	pagination := &models.Pagination{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Admin Laukpauk",
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
				Field:     models.AdminSortingFieldName,
				Direction: models.SortingDirectionAsc,
				Label:     "Name A-Z",
			},
			{
				Field:     models.AdminSortingFieldName,
				Direction: models.SortingDirectionDesc,
				Label:     "Name Z-A",
			},
		},
	)
	request, err = sanitizers.FindAdmins(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmins")
		return c.Render("admin/admin_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.adminService.FindAdmins(
		ctx,
		request.WithUserStatuses(models.UserStatusActive),
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmins")
		return c.Render("admin/admin_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/admin_index", params.Push(
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *adminController) AdminNew(c fiber.Ctx) (err error) {
	ctx := c.Context()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Admin{}
	admins := []*models.Admin{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Admin",
		"request", request,
		"admin", request,
		"admins", admins,
		"adminTypes", models.AdminTypes,
	)
	if admins, _, err = q.adminService.FindAdmins(
		ctx,
		models.NewAdminFilter().
			WithUserStatuses(models.UserStatusActive).
			WithSortings(models.Sorting{
				Field:     models.AdminSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/admin_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/admin_new", params.Push("admins", admins), "admin/layouts/authenticated")
}

func (q *adminController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AdminController-AdminCreate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	request := &models.Admin{}
	admins := []*models.Admin{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Admin",
		"request", request,
		"admin", request,
		"admins", admins,
		"adminTypes", models.AdminTypes,
	)
	if admins, _, err = q.adminService.FindAdmins(
		ctx,
		models.NewAdminFilter().
			WithUserStatuses(models.UserStatusActive).
			WithSortings(models.Sorting{
				Field:     models.AdminSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/admin_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("admins", admins)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateAdmin(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateAdmin")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/admin_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	admin, err := q.adminService.FindAdmin(
		ctx,
		models.NewAdminFilter().
			WithUsername(request.Username),
	)
	if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
		return c.Render("admin/admin_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if admin != nil {
		return c.Render("admin/admin_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueAdminUsernameViolation.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if request.MobilePhone != nil {
		admin, err := q.adminService.FindAdmin(
			ctx,
			models.NewAdminFilter().
				WithMobilePhone(*request.MobilePhone),
		)
		if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
			return c.Render("admin/admin_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
		if admin != nil {
			return c.Render("admin/admin_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueAdminMobilePhoneViolation.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
	}
	if request.Email != nil {
		admin, err := q.adminService.FindAdmin(
			ctx,
			models.NewAdminFilter().
				WithEmail(*request.Email),
		)
		if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
			return c.Render("admin/admin_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
		if admin != nil {
			return c.Render("admin/admin_new", params.Push(
				"flash", flash.Danger(models.ErrUniqueAdminEmailViolation.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
	}
	now := time.Now().In(time.Local)
	createdBy := currentAdmin.UserID
	if request.PasswordIncluded {
		encryptedPasword, err := helper.HashPassword(request.PlaintextPassword)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
			return c.Render("admin/admin_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
		request.EncryptedPassword = encryptedPasword
		request.PasswordChangedAt = &now
	}
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedBy = &createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("BUYER_DEFAULT_PHOTO")
	}
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/admin_new", params.Push(
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
	if _, err = q.adminService.CreateAdmin(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateAdmin")
		return c.Render("admin/admin_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/admin_new", params.Push(
				"flash", flash.Danger(err.Error()),
				"request", request.SetPhoto(""),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/admin_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request.SetPhoto(""),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("admin created successfully").
		Redirect(c, sess.Session, "/admin/admins")
}

func (q *adminController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AdminController-AdminEdit"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrAdminNotFound.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	admin, err := q.adminService.FindAdmin(
		ctx,
		models.NewAdminFilter().
			WithUserStatuses(models.UserStatusActive).
			WithAdminIDs(adminID),
	)
	if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	if admin == nil {
		return flash.Danger(models.ErrAdminNotFound.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	admins := []*models.Admin{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Admin|%s", admin.Name),
		"request", admin,
		"admin", admin,
		"admins", admins,
		"adminTypes", models.AdminTypes,
	)
	if admins, _, err = q.adminService.FindAdmins(
		ctx,
		models.NewAdminFilter().
			WithUserStatuses(models.UserStatusActive).
			WithSortings(models.Sorting{
				Field:     models.AdminSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/admin_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("admins", admins)
	return c.Render("admin/admin_edit", params, "admin/layouts/authenticated")
}

func (q *adminController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AdminController-AdminUpdate"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	csrfToken := csrf.TokenFromContext(c)
	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrAdminNotFound.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	admin, err := q.adminService.FindAdmin(
		ctx,
		models.NewAdminFilter().
			WithUserStatuses(models.UserStatusActive).
			WithAdminIDs(adminID),
	)
	if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	if admin == nil {
		return flash.Danger(models.ErrAdminNotFound.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	admins := []*models.Admin{}
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Admin|%s", admin.Name),
		"request", admin,
		"admin", admin,
		"admins", admins,
		"adminTypes", models.AdminTypes,
	)
	if admins, _, err = q.adminService.FindAdmins(
		ctx,
		models.NewAdminFilter().
			WithUserStatuses(models.UserStatusActive).
			WithSortings(models.Sorting{
				Field:     models.AdminSortingFieldName,
				Direction: models.SortingDirectionAsc,
			}),
	); err != nil {
		return c.Render("admin/admin_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	params.Push("admins", admins)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateAdmin(ctx, c)
	if photo == nil {
		request.Photo = admin.Photo
	}
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateAdmin")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/admin_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if admin.Username != request.Username {
		otherAdmin, err := q.adminService.FindAdmin(
			ctx,
			models.NewAdminFilter().
				WithUsername(request.Username).
				WithExcludedAdminIDs(admin.ID),
		)
		if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
			return c.Render("admin/admin_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherAdmin != nil {
			return c.Render("admin/admin_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueAdminUsernameViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	now := time.Now().In(time.Local)
	if request.PasswordIncluded {
		encryptedPasword, err := helper.HashPassword(request.PlaintextPassword)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
			return c.Render("admin/admin_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		admin.EncryptedPassword = encryptedPasword
		admin.PasswordChangedAt = &now
	}
	admin.Name = request.Name
	admin.Username = request.Username
	admin.Photo = request.Photo
	admin.UpdatedBy = &currentAdmin.UserID
	admin.UpdatedAt = now
	admin.DeactivatedBy = nil
	admin.DeactivatedAt = nil
	admin.DeactivationReason = nil
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/admin_edit", params.Push(
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
	if _, err = q.adminService.UpdateAdmin(ctx, tx, admin); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateAdmin")
		return c.Render("admin/admin_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return c.Render("admin/admin_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/admin_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Clear(c, sess.Session).
		Success("admin updated successfully").
		Redirect(c, sess.Session, "/admin/admins")
}

func (q *adminController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "AdminController-AdminDelete"
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrAdminNotFound.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	admin, err := q.adminService.FindAdmin(
		ctx,
		models.NewAdminFilter().
			WithUserStatuses(models.UserStatusActive).
			WithAdminIDs(adminID),
	)
	if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	if admin == nil {
		return flash.Danger(models.ErrAdminNotFound.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	admin.UserStatus = models.UserStatusDeactivated
	admin.DeactivatedBy = &currentAdmin.UserID
	admin.DeactivatedAt = new(time.Now())
	admin.DeactivationReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = q.adminService.UpdateAdmin(ctx, tx, admin); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateAdmin")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/admins")
	}
	return flash.Clear(c, sess.Session).
		Success("admin deactivated successfully").
		Redirect(c, sess.Session, "/admin/admins")
}
