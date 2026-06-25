package controllers

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"os"
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
	bazaarController struct {
		userService   services.UserService
		bazaarService services.BazaarService
	}
)

func NewBazaarController(
	userService services.UserService,
	bazaarService services.BazaarService,
) *bazaarController {
	return new(bazaarController{
		userService:   userService,
		bazaarService: bazaarService,
	})
}

func (q *bazaarController) Mount(r fiber.Router) {
	adminSessionAuth := middleware.AdminSessionAuth(q.userService)
	r.Group("/admin", adminSessionAuth).
		Get("", q.AdminIndex).
		Get("/new", q.AdminNew).
		Post("", q.AdminCreate).
		Get("/:id/edit", q.AdminEdit).
		Post("/:id", q.AdminUpdate).
		Post("/:id/delete", q.AdminDelete)
}

func (q *bazaarController) AdminIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BazaarController-AdminIndex"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.BazaarFilter)
	rows := []*models.Bazaar{}
	pagination := new(models.Pagination)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Data Pasar",
		"limits", models.Limits,
		"request", request,
		"rows", rows,
		"pagination", pagination,
	)
	if request, err = sanitizers.FindBazaars(ctx, c); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBazaars")
		return c.Render("admin/bazaar_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	request = request.WithRecordStatus(models.RecordStatusActive)
	if rows, pagination, err = q.bazaarService.FindBazaars(
		c,
		request,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBazaars")
		return c.Render("admin/bazaar_index", params.Push(
			"flash", flash.Danger(err.Error()),
		), "admin/layouts/authenticated")
	}
	return c.Render("admin/bazaar_index", params.Push(
		"request", request,
		"rows", rows,
		"pagination", pagination,
	), "admin/layouts/authenticated")
}

func (q *bazaarController) AdminNew(c fiber.Ctx) (err error) {
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Bazaar)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Pasar",
		"request", request,
	)
	return c.Render("admin/bazaar_new", params, "admin/layouts/authenticated")
}

func (q *bazaarController) AdminCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BazaarController-AdminCreate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.Bazaar)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", "Add New Pasar",
		"request", request,
	)
	request, photo, statusCode, err := sanitizers.ValidateBazaar(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBazaar")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/bazaar_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	now := time.Now()
	createdBy := currentAdmin.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = &createdBy
	request.CreatedAt = now
	request.UpdatedBy = &createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("BAZAAR_DEFAULT_PHOTO")
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/bazaar_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.bazaarService.CreateBazaar(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBazaar")
		return c.Render("admin/bazaar_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/bazaar_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/bazaar_new", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("bazaar created successfully").Redirect(c, sess.Session, "/admin/bazaars")
}

func (q *bazaarController) AdminEdit(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BazaarController-AdminEdit"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	bazaarID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("bazaar not found").Redirect(c, sess.Session, "/admin/bazaars")
	}
	bazaars, _, err := q.bazaarService.FindBazaars(
		ctx,
		models.NewBazaarFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBazaarIDs(bazaarID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBazaars")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/bazaars")
	}
	if len(bazaars) == 0 {
		return flash.Danger("bazaar not found").Redirect(c, sess.Session, "/admin/bazaars")
	}
	request := bazaars[0]
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", fmt.Sprintf("Edit Data Pasar|%s", request.Name),
		"request", request,
		"bazaarID", request.ID,
		"bazaar", request,
	)
	return c.Render("admin/bazaar_edit", params, "admin/layouts/authenticated")
}

func (q *bazaarController) AdminUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BazaarController-AdminUpdate"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	csrfToken := csrf.TokenFromContext(c)
	bazaarID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("bazaar not found").Redirect(c, sess.Session, "/admin/bazaars")
	}
	bazaars, _, err := q.bazaarService.FindBazaars(
		ctx,
		models.NewBazaarFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBazaarIDs(bazaarID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBazaars")
		return flash.Danger("bazaar not found").Redirect(c, sess.Session, "/admin/bazaars")
	}
	if len(bazaars) == 0 {
		return flash.Danger("bazaar not found").Redirect(c, sess.Session, "/admin/bazaars")
	}
	bazaar := bazaars[0]
	request := new(models.Bazaar)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
		"pageTitle", fmt.Sprintf("Edit Data Pasar|%s", bazaar.Name),
		"request", request,
		"bazaarID", bazaar.ID,
		"bazaar", bazaar,
	)
	request, photo, statusCode, err := sanitizers.ValidateBazaar(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBazaar")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/bazaar_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	request.ID = bazaar.ID
	request.RecordStatus = bazaar.RecordStatus
	request.UpdatedBy = &currentAdmin.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	if photo == nil {
		request.Photo = bazaar.Photo
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("admin/bazaar_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if _, err = q.bazaarService.UpdateBazaar(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBazaar")
		return c.Render("admin/bazaar_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = q.saveFile(c, photo, request.Photo); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
		}
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
		return c.Render("admin/bazaar_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("admin/bazaar_edit", params.Push(
			"flash", flash.Danger(err.Error()),
			"request", request,
		), "admin/layouts/authenticated")
	}
	return flash.Success("bazaar updated successfully").Redirect(c, sess.Session, "/admin/bazaars")
}

func (q *bazaarController) AdminDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BazaarController-AdminDelete"
	flash := helper.NewFlashMessage()
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	bazaarID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return flash.Danger("bazaar not found").Redirect(c, sess.Session, "/admin/bazaars")
	}
	bazaars, _, err := q.bazaarService.FindBazaars(
		ctx,
		models.NewBazaarFilter().
			WithRecordStatus(models.RecordStatusActive).
			WithBazaarIDs(bazaarID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBazaars")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/bazaars")
	}
	if len(bazaars) == 0 {
		return flash.Danger("bazaar not found").Redirect(c, sess.Session, "/admin/bazaars")
	}
	request := bazaars[0]
	request.RecordStatus = models.RecordStatusDeleted
	request.DeletedBy = &currentAdmin.ID
	request.DeletedAt = new(time.Now())
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/bazaars")
	}
	if _, err = q.bazaarService.UpdateBazaar(ctx, tx, request); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBazaar")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/bazaars")
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return flash.Danger(err.Error()).Redirect(c, sess.Session, "/admin/bazaars")
	}
	return flash.Success("bazaar deleted successfully").Redirect(c, sess.Session, "/admin/bazaars")
}

func (q *bazaarController) saveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	if err := c.SaveFile(file, fmt.Sprintf(models.BazaarPhotoPath, filename)); err != nil {
		return err
	}
	return c.SaveFile(file, fmt.Sprintf(models.BazaarPhotoCompressPath, filename))
}
