package controllers

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/url"
	"os"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/externals"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
	"github.com/roysitumorang/laukpauk/sanitizers"
	"github.com/roysitumorang/laukpauk/services"
)

type oneTimePasswordController struct {
	sellerService          services.SellerService
	buyerService           services.BuyerService
	companyService         services.CompanyService
	oneTimePasswordService services.OneTimePasswordService
	textMessageService     services.TextMessageService
	kafkaClient            *externals.KafkaClient
}

func NewOneTimePasswordController(
	sellerService services.SellerService,
	buyerService services.BuyerService,
	companyService services.CompanyService,
	oneTimePasswordService services.OneTimePasswordService,
	textMessageService services.TextMessageService,
	kafkaClient *externals.KafkaClient,
) *oneTimePasswordController {
	return &oneTimePasswordController{
		sellerService:          sellerService,
		buyerService:           buyerService,
		companyService:         companyService,
		oneTimePasswordService: oneTimePasswordService,
		textMessageService:     textMessageService,
		kafkaClient:            kafkaClient,
	}
}

func (q *oneTimePasswordController) Mount(r fiber.Router) {
	r.Group("/seller").
		Get("/new", q.SellerNew).
		Post("", q.SellerCreate)
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	r.Group("/v1/buyer", companyVerify).
		Post("", q.V1BuyerCreate)
}

func (q *oneTimePasswordController) SellerNew(c fiber.Ctx) error {
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
	request := &models.OtpRequest{
		Login: c.Query("login"),
	}
	return c.Render("seller/one_time_password_new", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", request,
	))
}

func (q *oneTimePasswordController) SellerCreate(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "OneTimePasswordController-SellerCreate"
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
		return c.Render("seller/one_time_password_new", params.Push(
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
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if seller == nil || seller.EncryptedPassword == nil {
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(models.ErrLoginNotFound.Error()),
		))
	}
	matched, err := helper.VerifyPassword(request.Password, *seller.EncryptedPassword)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	if !matched {
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(models.ErrLoginNotFound.Error()),
		))
	}
	now := time.Now().In(time.Local)
	otpDuration, err := time.ParseDuration(os.Getenv("OTP_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	otpToken := helper.RandomNumber(6)
	otpHash, err := helper.HashPassword(otpToken)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	otpMessage, err := helper.Encrypt(helper.Sprintf("%s is your OTP token, valid for %d minutes", otpToken, int64(math.Round(otpDuration.Minutes()))))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrEncrypt")
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/one_time_password_new", params.Push(
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
	oneTimePassword, err := q.oneTimePasswordService.CreateOneTimePassword(ctx, tx, &models.OneTimePassword{
		SellerID:      seller.ID,
		BuyerID:       uuid.Nil(),
		Hash:          otpHash,
		AttemptsCount: 0,
		Valid:         true,
		CreatedAt:     now,
		ExpiredAt:     now.Add(otpDuration).Add(time.Nanosecond),
	})
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateOneTimePassword")
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	textMessage, err := q.textMessageService.CreateTextMessage(ctx, tx, &models.TextMessage{
		AdminID:     uuid.Nil(),
		SellerID:    seller.ID,
		BuyerID:     uuid.Nil(),
		MobilePhone: seller.MobilePhone,
		MessageType: models.TextMessageTypeOtp,
		Body:        otpMessage,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTextMessage")
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/one_time_password_new", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if err = q.kafkaClient.Publish(
		ctx,
		models.TopicMessage{
			Topic: models.TopicTextMessage,
			Event: models.TopicTextMessageEventSend,
			ID:    textMessage.ID,
		},
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
	}
	sess.Set(models.CurrentOtpTrackingID, oneTimePassword.TrackingID)
	return flash.Success("please check your phone for OTP token SMS").
		Redirect(c, sess.Session, helper.Sprintf("/seller/auth/sessions/new?login=%s", url.QueryEscape(request.Login)))
}

func (q *oneTimePasswordController) V1BuyerCreate(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "OneTimePasswordController-V1BuyerCreate"
	company, _ := c.Locals(models.CurrentCompany).(*models.Company)
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
			WithCompanyIDs(company.ID).
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
	otpMessage, err := helper.Encrypt(helper.Sprintf("%s is your OTP token, valid for %d minutes", otpToken, int64(math.Round(otpDuration.Minutes()))))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrEncrypt")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
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
	response, err := q.oneTimePasswordService.CreateOneTimePassword(ctx, tx, &models.OneTimePassword{
		SellerID:      uuid.Nil(),
		BuyerID:       buyer.ID,
		Hash:          otpHash,
		AttemptsCount: 0,
		Valid:         true,
		CreatedAt:     now,
		ExpiredAt:     now.Add(otpDuration).Add(time.Nanosecond),
	})
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateOneTimePassword")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	textMessage, err := q.textMessageService.CreateTextMessage(ctx, tx, &models.TextMessage{
		AdminID:     uuid.Nil(),
		SellerID:    uuid.Nil(),
		BuyerID:     buyer.ID,
		MobilePhone: buyer.MobilePhone,
		MessageType: models.TextMessageTypeOtp,
		Body:        otpMessage,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTextMessage")
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
	go func() {
		if err = q.kafkaClient.Publish(
			ctx,
			models.TopicMessage{
				Topic: models.TopicTextMessage,
				Event: models.TopicTextMessageEventSend,
				ID:    textMessage.ID,
			},
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
		}
	}()
	return helper.NewResponse(fiber.StatusCreated).
		SetData(response).
		WriteResponse(c)
}
