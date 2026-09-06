package controllers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type (
	otpController struct {
		sellerService services.SellerService
		buyerService  services.BuyerService
	}
)

func NewOtpController(
	sellerService services.SellerService,
	buyerService services.BuyerService,
) *otpController {
	return &otpController{
		sellerService: sellerService,
		buyerService:  buyerService,
	}
}

func (q *otpController) Mount(r fiber.Router) {
	r.Group("/seller").
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate)
	r.Group("/v1/buyer").
		Post("", q.V1BuyerCreate)
}

func (q *otpController) SellerNew(c fiber.Ctx) error {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	if sess.Get(models.SellerID) != nil {
		return flash.Redirect(c, sess.Session, "/seller")
	}
	csrfToken := csrf.TokenFromContext(c)
	return c.Render("seller/otp_new", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", &models.OtpRequest{},
	))
}

func (q *otpController) SellerCreate(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "OtpController-SellerCreate"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	if sess.Get(models.SellerID) != nil {
		return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/seller")
	}
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", &models.OtpRequest{},
	)
	request, statusCode, err := sanitizers.ValidateOtp(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateOtp")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/otp_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	seller, err := q.sellerService.FindSeller(
		ctx,
		models.NewSellerFilter().
			WithLogin(request.Login).
			WithStatuses(models.SellerStatusActive),
	)
	if err != nil {
		if !errors.Is(err, models.ErrSellerNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSeller")
		}
		return c.Render("seller/otp_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if seller == nil {
		return c.Render("seller/otp_new", params.Push(
			"flash", flash.Danger(models.ErrLoginNotFound.Error()),
		))
	}
	now := time.Now().In(time.Local)
	if seller.OtpHash == "" ||
		!seller.OtpExpiredAt.After(now) {
		otpDuration, err := time.ParseDuration(os.Getenv("OTP_DURATION"))
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
			return c.Render("seller/otp_new", params.Push(
				"flash", flash.Danger(err.Error()),
			))
		}
		otpToken := helper.RandomNumber(6)
		otpHash, err := helper.HashPassword(otpToken)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
			return c.Render("seller/otp_new", params.Push(
				"flash", flash.Danger(err.Error()),
			))
		}
		if otpHash != nil {
			seller.OtpHash = *otpHash
		}
		seller.OtpExpiredAt = now.Add(otpDuration + time.Nanosecond)
		tx, err := repositories.BeginTx(ctx)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
			return c.Render("seller/otp_new", params.Push(
				"flash", flash.Danger(err.Error()),
			))
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
			return c.Render("seller/otp_new", params.Push(
				"flash", flash.Danger(err.Error()),
			))
		}
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return c.Render("seller/otp_new", params.Push(
				"flash", flash.Danger(err.Error()),
			))
		}
	}
	return c.Render("seller/auth_login", params.Push(
		"request", &models.LoginRequest{Login: request.Login},
	))
}

func (q *otpController) V1BuyerCreate(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "OtpController-V1BuyerCreate"
	otpRequest, statusCode, err := sanitizers.ValidateOtp(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateOtp")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithLogin(otpRequest.Login).
			WithStatuses(models.BuyerStatusActive),
	)
	if err != nil && !errors.Is(err, models.ErrBuyerNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if buyer == nil {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginNotFound.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	if buyer.OtpHash == "" ||
		!buyer.OtpExpiredAt.After(now) {
		otpDuration, err := time.ParseDuration(os.Getenv("OTP_DURATION"))
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		otpToken := helper.RandomNumber(6)
		otpHash, err := helper.HashPassword(otpToken)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if otpHash != nil {
			buyer.OtpHash = *otpHash
		}
		buyer.OtpExpiredAt = now.Add(otpDuration + time.Nanosecond)
		tx, err := repositories.BeginTx(ctx)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
			return helper.NewResponse(fiber.StatusBadRequest).
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
		if _, err = q.buyerService.UpdateBuyer(ctx, tx, buyer); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBuyer")
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if err = tx.Commit(ctx); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
	}
	return helper.NewResponse(fiber.StatusNoContent).WriteResponse(c)
}
