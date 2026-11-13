package router

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/coregx/coregex"
	"github.com/dustin/go-humanize"
	"github.com/getsentry/sentry-go"
	"github.com/goccy/go-json/v2"
	"github.com/gofiber/contrib/v3/monitor"
	fibersentry "github.com/gofiber/contrib/v3/sentry"
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/rewrite"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/fiber/v3/middleware/skip"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/jet/v3"
	"github.com/gofiber/utils/v2"
	"github.com/joho/godotenv"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/controllers"
	_ "github.com/roysitumorang/laukpauk/docs"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/middlewares"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
	slogfiber "github.com/samber/slog-fiber"
)

const DefaultPort uint16 = 8080

func (q *Service) HTTPServerMain(ctx context.Context, logger *slog.Logger) error {
	ctxt := "Router-HTTPServerMain"
	debug := helper.GetEnv() == "development"
	// Create a new engine
	engine := jet.New("./views", ".jet")
	engine.Debug(debug)
	engine.AddFunc("Comma", func(v int64) string {
		return humanize.Comma(v)
	})
	engine.AddFunc("HasPrefix", func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	})
	engine.AddFunc("QueryEscape", func(s string) string {
		return url.QueryEscape(s)
	})
	sessionMiddleware, sessionStore := session.NewWithStore(session.Config{
		Storage:         q.Storage,
		IdleTimeout:     30 * time.Minute,
		AbsoluteTimeout: 24 * time.Hour,
	})
	sessionStore.RegisterType(&helper.FlashMessage{})
	sessionStore.RegisterType(uuid.UUID{})
	sessionStore.RegisterType(&models.Admin{})
	app := fiber.New(fiber.Config{
		StrictRouting: true,
		RegexHandler:  coregex.MustCompile,
		JSONEncoder: func(v any) ([]byte, error) {
			buffer := pools.BufferPool.Get()
			defer func() {
				buffer.Reset()
				pools.BufferPool.Put(buffer)
			}()
			if err := json.MarshalWrite(buffer, v); err != nil {
				return nil, err
			}
			return buffer.Bytes(), nil
		},
		JSONDecoder: func(data []byte, v any) error {
			return json.Unmarshal(data, v)
		},
		Views: engine,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			statusCode := fiber.StatusInternalServerError
			if e, ok := errors.AsType[*fiber.Error](err); ok {
				statusCode = e.Code
			}
			return helper.NewResponse(statusCode).SetMessage(err.Error()).WriteResponse(ctx)
		},
	})
	csrfMiddleware := csrf.New(csrf.Config{
		CookieName:        "__Host-csrf_",
		CookieSecure:      true,
		CookieHTTPOnly:    true, // false for SPAs
		CookieSameSite:    "Lax",
		CookieSessionOnly: true,
		Extractor:         extractors.FromForm("_csrf"),
		Session:           sessionStore,
		// Redaction is enabled by default. Set DisableValueRedaction when you must expose tokens or storage keys in diagnostics.
		// DisableValueRedaction: true,
		SingleUseToken: false,
	})
	basicAuth := middlewares.BasicAuthMiddleware()
	middlewares := []any{
		recover.New(recover.Config{
			EnableStackTrace: true,
		}),
		slogfiber.NewWithConfig(
			logger,
			slogfiber.Config{
				WithUserAgent: true,
				WithRequestID: true,
				WithClientIP:  true,
			},
		),
		requestid.New(requestid.Config{
			Next:      nil,
			Header:    fiber.HeaderXRequestID,
			Generator: utils.UUIDv4,
		}),
		compress.New(),
		rewrite.New(rewrite.Config{
			RuleList: []rewrite.Rule{
				{From: "/admin/auth/sessions", To: "/auth/admin/sessions"},
				{From: "/admin/auth/sessions/*", To: "/auth/admin/sessions/$1"},
				{From: "/admin/auth/*", To: "/auth/admin/$1"},
				{From: "/admin", To: "/home/admin"},
				{From: "/admin/admins", To: "/admins/admin"},
				{From: "/admin/admins/*", To: "/admins/admin/$1"},
				{From: "/admin/admins/*/*", To: "/admins/admin/$1/$2"},
				{From: "/admin/notification_targets", To: "/notification_targets/admin"},
				{From: "/admin/text_messages", To: "/text_messages/admin"},
				{From: "/admin/text_messages/*", To: "/text_messages/admin/$1"},
				{From: "/admin/notifications", To: "/notifications/admin"},
				{From: "/admin/notifications/*", To: "/notifications/admin/$1"},
				{From: "/v1/seller/product_categories", To: "/product_categories/v1/seller"},
				{From: "/v1/seller/product_categories/*", To: "/product_categories/v1/seller/$1"},
				{From: "/v1/seller/product_categories/*/*", To: "/product_categories/v1/seller/$1/$2"},
				{From: "/v1/seller/products", To: "/products/v1/seller"},
				{From: "/v1/seller/products/*", To: "/products/v1/seller/$1"},
				{From: "/v1/seller/products/*/*", To: "/products/v1/seller/$1/$2"},
				{From: "/v1/seller/vouchers", To: "/vouchers/v1/seller"},
				{From: "/v1/seller/vouchers/*", To: "/vouchers/v1/seller/$1"},
				{From: "/v1/seller/vouchers/*/*", To: "/vouchers/v1/seller/$1/$2"},
				{From: "/v1/seller/delivery_plans", To: "/delivery_plans/v1/seller"},
				{From: "/v1/seller/delivery_plans/*", To: "/delivery_plans/v1/seller/$1"},
				{From: "/v1/seller/delivery_plans/*/*", To: "/delivery_plans/v1/seller/$1/$2"},
				{From: "/v1/seller/point_earning_caps", To: "/point_earning_caps/v1/seller"},
				{From: "/v1/seller/point_earning_caps/*", To: "/point_earning_caps/v1/seller/$1"},
				{From: "/v1/seller/point_earning_caps/*/*", To: "/point_earning_caps/v1/seller/$1/$2"},
				{From: "/v1/seller/point_spending_caps", To: "/point_spending_caps/v1/seller"},
				{From: "/v1/seller/point_spending_caps/*", To: "/point_spending_caps/v1/seller/$1"},
				{From: "/v1/seller/point_spending_caps/*/*", To: "/point_spending_caps/v1/seller/$1/$2"},
				{From: "/v1/seller/banners", To: "/banners/v1/seller"},
				{From: "/v1/seller/banners/*", To: "/banners/v1/seller/$1"},
				{From: "/v1/seller/banners/*/*", To: "/banners/v1/seller/$1/$2"},
				{From: "/v1/seller/sellers", To: "/sellers/v1/seller"},
				{From: "/v1/seller/sellers/*", To: "/sellers/v1/seller/$1"},
				{From: "/v1/seller/sellers/*/*", To: "/sellers/v1/seller/$1/$2"},
				{From: "/v1/seller/buyers", To: "/buyers/v1/seller"},
				{From: "/v1/seller/buyers/*", To: "/buyers/v1/seller/$1"},
				{From: "/v1/seller/buyers/*/*", To: "/buyers/v1/seller/$1/$2"},
				{From: "/v1/seller/recipes", To: "/recipes/v1/seller"},
				{From: "/v1/seller/recipes/*", To: "/recipes/v1/seller/$1"},
				{From: "/v1/seller/recipes/*/*", To: "/recipes/v1/seller/$1/$2"},
				{From: "/v1/seller/points", To: "/points/v1/seller"},
				{From: "/v1/seller/points/*", To: "/points/v1/seller/$1"},
				{From: "/v1/seller/points/*/*", To: "/points/v1/seller/$1/$2"},
				{From: "/v1/seller/notifications", To: "/notifications/v1/seller"},
				{From: "/v1/seller/notifications/*", To: "/notifications/v1/seller/$1"},
				{From: "/v1/seller/notifications/*/*", To: "/notifications/v1/seller/$1/$2"},
				{From: "/v1/seller/notification_targets", To: "/notification_targets/v1/seller"},
				{From: "/v1/seller/transactions", To: "/transactions/v1/seller"},
				{From: "/v1/seller/transactions/*", To: "/transactions/v1/seller/$1"},
				{From: "/v1/seller/transactions/*/*", To: "/transactions/v1/seller/$1/$2"},
				{From: "/v1/seller/users", To: "/users/v1/seller"},
				{From: "/v1/seller/users/*", To: "/users/v1/seller/$1"},
				{From: "/v1/seller/users/*/*", To: "/users/v1/seller/$1/$2"},
				{From: "/v1/seller/setting", To: "/setting/v1/seller"},
				{From: "/v1/buyer/users", To: "/users/v1/buyer"},
				{From: "/v1/buyer/users/*", To: "/users/v1/buyer/$1"},
				{From: "/v1/buyer/users/*/*", To: "/users/v1/buyer/$1/$2"},
				{From: "/v1/buyer/setting", To: "/setting/v1/buyer"},
				{From: "/v1/buyer/buyers", To: "/buyers/v1/buyer"},
				{From: "/v1/buyer/auth/*", To: "/auth/v1/buyer/$1"},
				{From: "/v1/buyer/notification_targets", To: "/notification_targets/v1/buyer"},
				{From: "/v1/buyer/product_categories", To: "/product_categories/v1/buyer"},
				{From: "/v1/buyer/products", To: "/products/v1/buyer"},
				{From: "/v1/buyer/products/*", To: "/products/v1/buyer/$1"},
				{From: "/v1/buyer/cart", To: "/cart/v1/buyer"},
				{From: "/v1/buyer/cart/*", To: "/cart/v1/buyer/$1"},
				{From: "/v1/buyer/cart/*/*", To: "/cart/v1/buyer/$1/$2"},
				{From: "/v1/buyer/vouchers", To: "/vouchers/v1/buyer"},
				{From: "/v1/buyer/delivery_plans/*", To: "/delivery_plans/v1/buyer/$1"},
				{From: "/v1/buyer/transactions", To: "/transactions/v1/buyer"},
				{From: "/v1/buyer/transactions/*", To: "/transactions/v1/buyer/$1"},
				{From: "/v1/buyer/transactions/*/*", To: "/transactions/v1/buyer/$1/$2"},
				{From: "/v1/buyer/banners", To: "/banners/v1/buyer"},
				{From: "/v1/buyer/banners/*", To: "/banners/v1/buyer/$1"},
				{From: "/v1/buyer/banners/*/*", To: "/banners/v1/buyer/$1/$2"},
				{From: "/v1/buyer/delivery_routes", To: "/delivery_routes/v1/buyer"},
				{From: "/v1/buyer/point_earning_caps", To: "/point_earning_caps/v1/buyer"},
				{From: "/v1/buyer/point_spending_caps", To: "/point_spending_caps/v1/buyer"},
				{From: "/v1/buyer/payment_channels", To: "/payment_channels/v1/buyer"},
				{From: "/", To: "/home"},
				{From: "/contact-us", To: "/home/contact-us"},
				{From: "/faq", To: "/home/faq"},
				{From: "/privacy-policy", To: "/home/privacy-policy"},
				{From: "/firebase-messaging-sw.js", To: "/assets/js/firebase-messaging-sw.js"},
			},
		}),
		cors.New(),
		pprof.New(),
		helmet.New(),
		skip.New(sessionMiddleware, func(c fiber.Ctx) bool {
			return strings.HasPrefix(c.OriginalURL(), "/v1/")
		}),
		skip.New(csrfMiddleware, func(c fiber.Ctx) bool {
			return strings.HasPrefix(c.OriginalURL(), "/v1/")
		}),
	}
	if sentryEnabled := os.Getenv("SENTRY_ENABLED") == "1"; sentryEnabled {
		_ = sentry.Init(sentry.ClientOptions{
			Dsn: os.Getenv("SENTRY_DSN"),
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				return event
			},
			Debug:            true,
			AttachStacktrace: true,
			EnableTracing:    true,
		})
		middlewares = append(
			middlewares,
			fibersentry.New(fibersentry.Config{
				Repanic:         true,
				WaitForDelivery: true,
			}),
		)
	}
	app.Use(middlewares...)
	if debug {
		app.Get("/debug/pprof/*", adaptor.HTTPHandler(http.DefaultServeMux))
		app.Get("/swagger/*", swaggo.HandlerDefault)
	}
	app.Get("/ping", func(c fiber.Ctx) error {
		return helper.NewResponse(fiber.StatusOK).
			SetData(map[string]any{
				"version": config.Version,
				"commit":  config.Commit,
				"build":   config.Build,
				"upsince": config.Now.Format(time.RFC3339),
				"uptime":  time.Since(config.Now).String(),
			}).WriteResponse(c)
	}).
		Get("/metrics", basicAuth, monitor.New(monitor.Config{
			APIOnly: true,
		})).
		Get("/env", basicAuth, func(c fiber.Ctx) error {
			envMap, err := godotenv.Read(".env")
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRead")
				return helper.NewResponse(fiber.StatusBadRequest).SetMessage(err.Error()).WriteResponse(c)
			}
			envMap["GO_VERSION"] = runtime.Version()
			return helper.NewResponse(fiber.StatusOK).SetData(envMap).WriteResponse(c)
		}).
		Get("/assets*", static.New("", static.Config{
			FS: os.DirFS("assets"),
		})).
		Get("/images*", static.New("", static.Config{
			FS: os.DirFS("images"),
		}))
	controllers.NewAuthController(
		q.AdminService,
		q.BuyerService,
		q.JsonWebTokenService,
		q.TextMessageService,
		q.KafkaClient,
	).Mount(app.Group("/auth"))
	controllers.NewHomeController(
		q.AdminService,
		q.MessageService,
	).Mount(app.Group("/home"))
	controllers.NewProductCategoryController(
		q.ProductCategoryService,
		q.ProductService,
		q.ProductVariantService,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/product_categories"))
	controllers.NewProductController(
		q.ProductCategoryService,
		q.ProductService,
		q.ProductVariantService,
		q.RecipeItemService,
		q.Storage,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/products"))
	controllers.NewDeliveryPlanController(
		q.DeliveryPlanService,
		q.BuyerService,
		q.JsonWebTokenService,
		q.UserService,
	).Mount(app.Group("/delivery_plans"))
	controllers.NewPointController(
		q.BuyerService,
		q.PointService,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/points"))
	controllers.NewVoucherController(
		q.VoucherService,
		q.BuyerService,
		q.JsonWebTokenService,
		q.UserService,
	).Mount(app.Group("/vouchers"))
	controllers.NewPointEarningCapController(
		q.PointEarningCapService,
		q.BuyerService,
		q.JsonWebTokenService,
		q.UserService,
	).Mount(app.Group("/point_earning_caps"))
	controllers.NewPointSpendingCapController(
		q.PointSpendingCapService,
		q.BuyerService,
		q.JsonWebTokenService,
		q.UserService,
	).Mount(app.Group("/point_spending_caps"))
	controllers.NewTransactionController(
		q.BuyerService,
		q.TransactionService,
		q.JsonWebTokenService,
		q.KafkaClient,
		q.UserService,
	).Mount(app.Group("/transactions"))
	controllers.NewNotificationTargetController(
		q.AdminService,
		q.BuyerService,
		q.NotificationTargetService,
		q.JsonWebTokenService,
		q.UserService,
	).Mount(app.Group("/notification_targets"))
	controllers.NewCartController(
		q.BuyerService,
		q.ProductVariantService,
		q.CartService,
		q.JsonWebTokenService,
	).Mount(app.Group("/cart"))
	controllers.NewDeliveryRouteController(
		q.BuyerService,
		q.DeliveryRouteService,
		q.SettingService,
		q.JsonWebTokenService,
		q.RoutesApiClient,
	).Mount(app.Group("/delivery_routes"))
	controllers.NewSettingController(
		q.UserService,
		q.SettingService,
		q.JsonWebTokenService,
	).Mount(app.Group("/setting"))
	controllers.NewPaymentChannelController(
		q.BuyerService,
		q.PaymentChannelService,
		q.JsonWebTokenService,
	).Mount(app.Group("/payment_channels"))
	controllers.NewBannerController(
		q.BannerService,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/banners"))
	controllers.NewUserController(
		q.UserService,
		q.JsonWebTokenService,
		q.TextMessageService,
		q.NotificationService,
		q.KafkaClient,
	).Mount(app.Group("/users"))
	controllers.NewBuyerController(
		q.BuyerService,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/buyers"))
	controllers.NewAdminController(
		q.AdminService,
	).Mount(app.Group("/admins"))
	controllers.NewTextMessageController(
		q.AdminService,
		q.BuyerService,
		q.TextMessageService,
	).Mount(app.Group("/text_messages"))
	controllers.NewNotificationController(
		q.AdminService,
		q.BuyerService,
		q.NotificationService,
		q.KafkaClient,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/notifications"))
	controllers.NewRecipeController(
		q.BuyerService,
		q.RecipeService,
		q.RecipeGuideService,
		q.RecipeItemService,
		q.ProductService,
		q.ProductVariantService,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/recipes"))
	controllers.NewSellerController(
		q.SellerService,
		q.UserService,
		q.JsonWebTokenService,
	).Mount(app.Group("/sellers"))
	app.Use(func(c fiber.Ctx) error {
		if strings.HasPrefix(c.OriginalURL(), "/v1/") {
			return helper.NewResponse(fiber.StatusNotFound).
				SetMessage(utils.StatusMessage(fiber.StatusNotFound)).
				WriteResponse(c)
		}
		c.Response().SetStatusCode(fiber.StatusNotFound)
		return c.Render("404", fiber.Map{
			"baseURL": c.BaseURL(),
		})
	})
	port := DefaultPort
	if envPort, ok := os.LookupEnv("PORT"); ok && envPort != "" {
		if portInt, _ := strconv.Atoi(envPort); portInt >= 0 && portInt <= math.MaxUint16 {
			port = uint16(portInt)
		}
	}
	listenerPort := helper.Sprintf(":%d", port)
	err := app.Listen(listenerPort, fiber.ListenConfig{
		EnablePrefork: !debug,
	})
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrListen")
	}
	return err
}
