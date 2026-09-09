package controllers

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"os"
	"strconv"
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

type (
	authController struct {
		adminService           services.AdminService
		sellerService          services.SellerService
		buyerService           services.BuyerService
		companyService         services.CompanyService
		jsonWebTokenService    services.JsonWebTokenService
		oneTimePasswordService services.OneTimePasswordService
		textMessageService     services.TextMessageService
		kafkaService           *externals.KafkaClient
	}
)

func NewAuthController(
	adminService services.AdminService,
	sellerService services.SellerService,
	buyerService services.BuyerService,
	companyService services.CompanyService,
	jsonWebTokenService services.JsonWebTokenService,
	oneTimePasswordService services.OneTimePasswordService,
	textMessageService services.TextMessageService,
	kafkaService *externals.KafkaClient,
) *authController {
	return &authController{
		adminService:           adminService,
		sellerService:          sellerService,
		buyerService:           buyerService,
		companyService:         companyService,
		jsonWebTokenService:    jsonWebTokenService,
		oneTimePasswordService: oneTimePasswordService,
		textMessageService:     textMessageService,
		kafkaService:           kafkaService,
	}
}

func (q *authController) Mount(r fiber.Router) {
	adminSessionAuth := middlewares.AdminSessionAuthMiddleware(q.adminService)
	sellerSessionAuth := middlewares.SellerSessionAuthMiddleware(q.sellerService)
	buyerKeyAuth := middlewares.BuyerKeyAuthMiddleware(q.buyerService, q.jsonWebTokenService)
	companyVerify := middlewares.CompanyVerifyMiddleware(q.companyService)
	r.Group("/admin").
		Get("/login", q.AdminNewSession).
		Post("/login", q.AdminCreateSession).
		Get("/me", adminSessionAuth, q.AdminProfile).
		Get("/logout", q.AdminLogout)
	r.Group("/seller").
		Get("/otp/new", q.SellerNewOTP).
		Post("/otp", q.SellerCreateOTP).
		Post("/login", q.SellerCreateSession).
		Get("/me", sellerSessionAuth, q.SellerProfile).
		Get("/logout", q.SellerLogout)
	r.Group("/:slug/v1/buyer", companyVerify).
		Post("/otp", q.V1BuyerCreateOTP).
		Post("/login", q.V1BuyerLogin).
		Get("/me", buyerKeyAuth, q.V1BuyerProfile)
}

func (q *authController) AdminNewSession(c fiber.Ctx) error {
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	if sess.Get(models.AdminID) != nil {
		return flash.Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	request := new(models.LoginRequest)
	return c.Render("admin/auth_login", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", request,
	))
}

func (q *authController) AdminCreateSession(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-AdminCreateSession"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	if sess.Get(models.AdminID) != nil {
		return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/admin")
	}
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", new(models.LoginRequest),
	)
	request, statusCode, err := sanitizers.ValidateLogin(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateLogin")
		c.Response().SetStatusCode(statusCode)
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	admin, err := q.adminService.FindAdmin(
		ctx,
		models.NewAdminFilter().
			WithLogin(request.Login).
			WithStatuses(models.AdminStatusActive),
	)
	if err != nil {
		if !errors.Is(err, models.ErrAdminNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
		}
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if admin == nil ||
		admin.EncryptedPassword == nil {
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	matched, err := helper.VerifyPassword(request.Password, *admin.EncryptedPassword)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if !matched {
		return c.Render("admin/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	if err = sess.RegenerateWithContext(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRegenerate")
	}
	sess.Set(models.AdminID, admin.ID)
	next, ok := sess.Get("next").(string)
	if !ok {
		next = "/admin"
	} else {
		sess.Delete("next")
	}
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, next)
}

func (q *authController) AdminProfile(c fiber.Ctx) error {
	sess := session.FromContext(c)
	currentAdmin, _ := sess.Get(models.CurrentAdmin).(*models.Admin)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentAdmin", currentAdmin,
	)
	return c.Render("admin/auth_me", params)
}

func (q *authController) AdminLogout(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-AdminLogout"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	if csrfHandler := csrf.HandlerFromContext(c); csrfHandler != nil {
		if err := csrfHandler.DeleteToken(c); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteToken")
		}
	}
	if err := sess.ResetWithContext(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrReset")
	}
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/admin/auth/login")
}

func (q *authController) SellerNewOTP(c fiber.Ctx) error {
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
	return c.Render("seller/auth_otp", models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"request", request,
	))
}

func (q *authController) SellerCreateOTP(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-SellerCreateOTP"
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
		return c.Render("seller/auth_otp", params.Push(
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
		return c.Render("seller/auth_otp", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if seller == nil {
		return c.Render("seller/auth_otp", params.Push(
			"flash", flash.Danger(models.ErrLoginNotFound.Error()),
		))
	}
	now := time.Now().In(time.Local)
	otpDuration, err := time.ParseDuration(os.Getenv("OTP_DURATION"))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseDuration")
		return c.Render("seller/auth_otp", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	otpToken := helper.RandomNumber(6)
	otpHash, err := helper.HashPassword(otpToken)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrHashPassword")
		return c.Render("seller/auth_otp", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	otpMessage, err := helper.Encrypt(helper.Sprintf("%s is your OTP token, valid for %d minutes", otpToken, int64(math.Round(otpDuration.Minutes()))))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrEncrypt")
		return c.Render("seller/auth_otp", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/auth_otp", params.Push(
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
		return c.Render("seller/auth_otp", params.Push(
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
		return c.Render("seller/auth_otp", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/auth_otp", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	go func() {
		if err = q.kafkaService.Publish(
			ctx,
			models.TopicRequest{
				Topic: models.TopicTextMessage,
				Messages: []models.TopicMessage{
					{
						Topic: models.TopicTextMessage,
						Event: models.TopicTextMessageEventSend,
						ID:    textMessage.ID,
					},
				},
			},
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
		}
	}()
	return c.Render("seller/auth_login", params.Push(
		"flash", flash.Success("please check your phone for OTP token SMS"),
		"request", &models.OtpLoginRequest{
			Login:      request.Login,
			TrackingID: oneTimePassword.TrackingID,
		},
	))
}

func (q *authController) SellerCreateSession(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-SellerCreateSession"
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
		"request", &models.OtpLoginRequest{},
	)
	request, statusCode, err := sanitizers.ValidateOtpLogin(ctx, c)
	params.Push("request", request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateOtpLogin")
		c.Response().SetStatusCode(statusCode)
		return c.Render("seller/auth_login", params.Push(
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
		return c.Render("seller/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if seller == nil {
		return c.Render("seller/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	oneTimePassword, err := q.oneTimePasswordService.FindOneTimePassword(
		ctx,
		models.NewOneTimePasswordFilter().
			WithSellerIDs(seller.ID).
			WithTrackingIDs(request.TrackingID),
	)
	if err != nil {
		if errors.Is(err, models.ErrOneTimePasswordNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindOneTimePassword")
		}
		return c.Render("seller/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	now := time.Now().In(time.Local)
	if oneTimePassword == nil ||
		!oneTimePassword.Valid ||
		!oneTimePassword.ExpiredAt.After(now) {
		return c.Render("seller/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	matched, err := helper.VerifyPassword(request.OtpToken, oneTimePassword.Hash)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return c.Render("seller/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return c.Render("seller/auth_login", params.Push(
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
	if !matched {
		otpMaxAttempts, _ := strconv.Atoi(os.Getenv("OTP_MAX_ATTEMPTS"))
		if oneTimePassword.AttemptsCount < otpMaxAttempts-1 {
			oneTimePassword.AttemptsCount++
		} else {
			oneTimePassword.Valid = false
		}
		if _, err = q.oneTimePasswordService.UpdateOneTimePassword(
			ctx,
			tx,
			oneTimePassword,
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateOneTimePassword")
			return c.Render("seller/auth_login", params.Push(
				"flash", flash.Danger(err.Error()),
			))
		}
	}
	oneTimePassword.Valid = false
	if _, err = q.oneTimePasswordService.UpdateOneTimePassword(
		ctx,
		tx,
		oneTimePassword,
	); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateOneTimePassword")
		return c.Render("seller/auth_login", params.Push(
			"flash", flash.Danger(err.Error()),
		))
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return c.Render("seller/auth_login", params.Push(
			"flash", flash.Danger(models.ErrLoginFailed.Error()),
		))
	}
	if err = sess.RegenerateWithContext(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRegenerate")
	}
	sess.Set(models.SellerID, seller.ID)
	next, ok := sess.Get("next").(string)
	if !ok {
		next = "/seller"
	} else {
		sess.Delete("next")
	}
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, next)
}

func (q *authController) SellerProfile(c fiber.Ctx) error {
	sess := session.FromContext(c)
	currentSeller, _ := sess.Get(models.CurrentSeller).(*models.Seller)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	defer flash.Clear(c, sess.Session)
	csrfToken := csrf.TokenFromContext(c)
	params := models.NewMap().Push(
		"baseURL", c.BaseURL(),
		"flash", flash,
		"path", c.OriginalURL(),
		"csrfToken", csrfToken,
		"currentSeller", currentSeller,
	)
	return c.Render("seller/auth_me", params)
}

func (q *authController) SellerLogout(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-SellerLogout"
	sess := session.FromContext(c)
	flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
	if !ok {
		flash = helper.NewFlashMessage()
	}
	if csrfHandler := csrf.HandlerFromContext(c); csrfHandler != nil {
		if err := csrfHandler.DeleteToken(c); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDeleteToken")
		}
	}
	if err := sess.ResetWithContext(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrReset")
	}
	return flash.Clear(c, sess.Session).Redirect(c, sess.Session, "/seller/auth/otp/new")
}

func (q *authController) V1BuyerCreateOTP(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-V1BuyerCreateOTP"
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
		if err = q.kafkaService.Publish(
			ctx,
			models.TopicRequest{
				Topic: models.TopicTextMessage,
				Messages: []models.TopicMessage{
					{
						Topic: models.TopicTextMessage,
						Event: models.TopicTextMessageEventSend,
						ID:    textMessage.ID,
					},
				},
			},
		); err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
		}
	}()
	return helper.NewResponse(fiber.StatusCreated).
		SetData(response).
		WriteResponse(c)
}

func (q *authController) V1BuyerLogin(c fiber.Ctx) error {
	ctx := c.Context()
	ctxt := "AuthController-V1BuyerLogin"
	company, _ := c.Locals(models.CurrentCompany).(*models.Company)
	loginRequest, statusCode, err := sanitizers.ValidateOtpLogin(ctx, c)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidateOtpLogin")
		return helper.NewResponse(statusCode).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	buyer, err := q.buyerService.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithCompanyIDs(company.ID).
			WithLogin(loginRequest.Login).
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
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	oneTimePassword, err := q.oneTimePasswordService.FindOneTimePassword(
		ctx,
		models.NewOneTimePasswordFilter().
			WithBuyerIDs(buyer.ID).
			WithTrackingIDs(loginRequest.TrackingID),
	)
	if err != nil {
		if errors.Is(err, models.ErrOneTimePasswordNotFound) {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindOneTimePassword")
		}
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	now := time.Now().In(time.Local)
	if oneTimePassword == nil ||
		!oneTimePassword.Valid ||
		!oneTimePassword.ExpiredAt.After(now) {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	matched, err := helper.VerifyPassword(loginRequest.OtpToken, oneTimePassword.Hash)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrVerifyPassword")
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	if !matched {
		return helper.NewResponse(fiber.StatusBadRequest).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	expiredAt := now.Add(helper.GetAccessTokenAge()).Add(time.Nanosecond)
	jsonWebTokenRequest := &models.JsonWebToken{
		BuyerID:   buyer.ID,
		CreatedAt: now,
		ExpiredAt: expiredAt,
	}
	tx, err := repositories.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrLoginFailed.Error()).
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
	jsonWebToken, err := q.jsonWebTokenService.CreateJsonWebToken(ctx, tx, jsonWebTokenRequest)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateJsonWebToken")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	tokenString, err := helper.GenerateAccessToken(buyer.ID.String(), jsonWebToken.ID.String(), buyer.Username, jsonWebToken.CreatedAt, jsonWebToken.ExpiredAt)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrGenerateAccessToken")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(models.ErrLoginFailed.Error()).
			WriteResponse(c)
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return helper.NewResponse(fiber.StatusUnprocessableEntity).
			SetMessage(err.Error()).
			WriteResponse(c)
	}
	response := models.BuyerLoginResponse{
		IDToken:   tokenString,
		ExpiredAt: jsonWebToken.ExpiredAt,
		Buyer:     buyer,
	}
	return helper.NewResponse(fiber.StatusCreated).
		SetData(response).
		WriteResponse(c)
}

func (q *authController) V1BuyerProfile(c fiber.Ctx) error {
	currentBuyer, _ := c.Locals(models.CurrentBuyer).(*models.Buyer)
	return helper.NewResponse(fiber.StatusOK).
		SetData(currentBuyer).
		WriteResponse(c)
}
