package controllers

import (
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	settingController struct {
		adminService   services.AdminService
		settingService services.SettingService
	}
)

func NewSettingController(
	adminService services.AdminService,
	settingService services.SettingService,
) *settingController {
	return new(settingController{
		adminService:   adminService,
		settingService: settingService,
	})
}

func (q *settingController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.adminService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/switch", q.AdminSwitch)
}

func (q *settingController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SettingController-AdminIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.SettingFilter)
	rows := []*models.Setting{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Pengaturan",
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
				Field:     models.SettingSortingFieldKey,
				Direction: models.SortingDirectionAsc,
				Label:     "Key A-Z",
			},
			{
				Field:     models.SettingSortingFieldKey,
				Direction: models.SortingDirectionDesc,
				Label:     "Key Z-A",
			},
		},
	)
	if request, err = sanitizers.FindSettings(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSettings")
		return c.Render("admin/setting_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if rows, pagination, err = q.settingService.FindSettings(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSettings")
		return c.Render("admin/setting_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/setting_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *settingController) AdminNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Setting)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Pengaturan",
		"request", request,
	)
	return c.Render("admin/setting_new", params, "admin/layouts/authenticated")
}

func (q *settingController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SettingController-AdminCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Setting)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Pengaturan",
		"request", request,
	)
	request, statusCode, err := sanitizers.ValidateSetting(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSetting")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/setting_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	setting, err := q.settingService.FindSetting(
		ctx,
		models.NewSettingFilter().
			WithKeys(request.Key),
	)
	if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
		return c.Render("admin/setting_new", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.Active = true
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	if setting != nil {
		return c.Render("admin/setting_new", params.Push(
			"flash", flash.Danger(models.ErrUniqueSettingKeyViolation.Error()),
		), "admin/layouts/authenticated")
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/setting_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.settingService.CreateSetting(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateSetting")
		return c.Render("admin/setting_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/setting_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("setting created successfully").Redirect(c, sess.Session, "/admin/settings")
}

func (q *settingController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SettingController-AdminEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	settingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrSettingNotFound.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	setting, err := q.settingService.FindSetting(
		ctx,
		models.NewSettingFilter().
			WithSettingIDs(settingID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSettings")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	if setting == nil {
		return flash.Danger(models.ErrSettingNotFound.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	request := setting
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Pengaturan|%s", request.Key),
		"request", request,
		"setting", setting,
	)
	return c.Render("admin/setting_edit", params, "admin/layouts/authenticated")
}

func (q *settingController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SettingController-AdminUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	settingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrSettingNotFound.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	setting, err := q.settingService.FindSetting(
		ctx,
		models.NewSettingFilter().
			WithSettingIDs(settingID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSettings")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	if setting == nil {
		return flash.Danger(models.ErrSettingNotFound.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	request := new(models.Setting)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", helper.Sprintf("Edit Data Pengaturan|%s", setting.Key),
		"request", request,
		"setting", setting,
	)
	request, statusCode, err := sanitizers.ValidateSetting(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateSetting")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/setting_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if setting.Key != request.Key {
		otherSetting, err := q.settingService.FindSetting(
			ctx,
			models.NewSettingFilter().
				WithKeys(request.Key).
				WithExcludedSettingIDs(setting.ID),
		)
		if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
			return c.Render("admin/setting_edit", params.Push(
				"flash", flash.Danger(err.Error()),
			), "admin/layouts/authenticated")
		}
		if otherSetting != nil {
			return c.Render("admin/setting_edit", params.Push(
				"flash", flash.Danger(models.ErrUniqueSettingKeyViolation.Error()),
			), "admin/layouts/authenticated")
		}
	}
	request.ID = setting.ID
	request.Active = setting.Active
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	params.Push("request", request)
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/setting_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if _, err = q.settingService.UpdateSetting(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSetting")
		return c.Render("admin/setting_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/setting_edit", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return flash.Success("setting updated successfully").Redirect(c, sess.Session, "/admin/settings")
}

func (q *settingController) AdminSwitch(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "SettingController-AdminSwitch"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	settingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger(models.ErrSettingNotFound.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	setting, err := q.settingService.FindSetting(
		ctx,
		models.NewSettingFilter().
			WithSettingIDs(settingID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSettings")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	if setting == nil {
		return flash.Danger(models.ErrSettingNotFound.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	request := setting
	request.Active = !setting.Active
	request.UpdatedBy = currentAdmin.ID
	request.UpdatedAt = time.Now()
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	if _, err = q.settingService.UpdateSetting(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateSetting")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/settings")
	}
	return flash.Success("setting switched successfully").Redirect(c, sess.Session, "/admin/settings")
}
