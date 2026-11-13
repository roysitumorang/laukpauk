package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type bannerController struct {
	bannerService       services.BannerService
	userService         services.UserService
	jsonWebTokenService services.JsonWebTokenService
}

func NewBannerController(
	bannerService services.BannerService,
	userService services.UserService,
	jsonWebTokenService services.JsonWebTokenService,
) *bannerController {
	return &bannerController{
		bannerService:       bannerService,
		userService:         userService,
		jsonWebTokenService: jsonWebTokenService,
	}
}

func (q *bannerController) Mount(r fiber.Router) {
	sellerKeyAuth := middlewares.KeyAuthMiddleware(models.RoleSeller, q.userService, q.jsonWebTokenService)
	r.Group("/v1/seller", sellerKeyAuth).
		Get("", q.V1SellerIndex).
		Post("", q.V1SellerCreate).
		Get("/:id", q.V1SellerShow).
		Post("/:id", q.V1SellerUpdate).
		Post("/:id/switch", q.V1SellerSwitchVisibility).
		Post("/:id/delete", q.V1SellerDelete)
}

func (q *bannerController) V1SellerIndex(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-V1SellerIndex"
	request, err := sanitizers.FindBanners(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanners")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	rows, pagination, err := q.bannerService.FindBanners(
		ctx,
		request.WithRecordStatuses(models.RecordStatusActive),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanners")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	pagination.Rows = make([]any, len(rows))
	for i := range rows {
		pagination.Rows[i] = rows[i]
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(pagination).
		WriteResponse(c)
}

func (q *bannerController) V1SellerCreate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-V1SellerCreate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBanner(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBanner")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	now := time.Now()
	createdBy := currentUser.ID
	request.RecordStatus = models.RecordStatusActive
	request.CreatedBy = createdBy
	request.CreatedAt = now
	request.UpdatedBy = createdBy
	request.UpdatedAt = now
	if photo == nil {
		request.Photo = os.Getenv("ARTICLE_DEFAULT_PHOTO")
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	saved, err := q.bannerService.CreateBanner(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetMessage("banner created successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *bannerController) V1SellerShow(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-V1SellerShow"
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if banner == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetData(banner).
		WriteResponse(c)
}

func (q *bannerController) V1SellerUpdate(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-V1SellerUpdate"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if banner == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	request, photo, mimetype, statusCode, err := sanitizers.ValidateBanner(ctx, c)
	if photo == nil {
		request.Photo = banner.Photo
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateBanner")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	request.ID = banner.ID
	request.RecordStatus = banner.RecordStatus
	request.Visible = banner.Visible
	request.UpdatedBy = currentUser.ID
	request.UpdatedAt = time.Now()
	request.DeletedBy = nil
	request.DeletedAt = nil
	request.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	saved, err := q.bannerService.UpdateBanner(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if photo != nil {
		if err = helper.SaveFile(c, photo, mimetype, request.Photo); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrSaveFile")
			return helper.NewResponse(fiber.StatusUnprocessableEntity).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("banner updated successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *bannerController) V1SellerSwitchVisibility(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-V1SellerSwitchVisibility"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if banner == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	banner.Visible = !banner.Visible
	banner.UpdatedBy = currentUser.ID
	banner.UpdatedAt = time.Now()
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	saved, err := q.bannerService.UpdateBanner(ctx, tx, banner)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("banner visibility switched successfully").
		SetData(saved).
		WriteResponse(c)
}

func (q *bannerController) V1SellerDelete(c fiber.Ctx) (err error) {
	ctx := c.Context()
	ctxt := "BannerController-V1SellerDelete"
	currentUser, _ := c.Locals(models.CurrentUser).(*models.User)
	bannerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParse")
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	banner, err := q.bannerService.FindBanner(
		ctx,
		models.NewBannerFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithBannerIDs(bannerID),
	)
	if err != nil && !errors.Is(err, models.ErrBannerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if banner == nil {
		return helper.NewResponse(fiber.StatusNotFound).
			SetMessage(models.ErrBannerNotFound.Error()).
			WriteResponse(c)
	}
	banner.RecordStatus = models.RecordStatusDeleted
	banner.DeletedBy = &currentUser.ID
	banner.DeletedAt = new(time.Now())
	banner.DeletionReason = nil
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	saved, err := q.bannerService.UpdateBanner(ctx, tx, banner)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBanner")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK).
		SetMessage("banner deleted successfully").
		SetData(saved).
		WriteResponse(c)
}
