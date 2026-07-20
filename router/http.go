package router

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/coregx/coregex"
	"github.com/dustin/go-humanize"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/contrib/v3/monitor"
	fibersentry "github.com/gofiber/contrib/v3/sentry"
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
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
	"github.com/roysitumorang/laukpauk/middleware"
	"github.com/roysitumorang/laukpauk/models"
	slogfiber "github.com/samber/slog-fiber"
)

const (
	DefaultPort uint16 = 8080
)

func (q *Service) HTTPServerMain(ctx context.Context, logger *slog.Logger) error {
	ctxt := "Router-HTTPServerMain"
	debug := helper.GetEnv() == "development"
	// Create a new engine
	engine := jet.New("./views", ".jet")
	engine.Debug(debug)
	if debug {
		engine.ShouldReload = true
		engine.Verbose = true
	}
	engine.AddFunc("Comma", func(v int64) string {
		return humanize.Comma(v)
	})
	engine.AddFunc("HasPrefix", func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	})
	sessionMiddleware, sessionStore := session.NewWithStore(session.Config{
		Storage: q.Storage,
	})
	sessionStore.RegisterType(&helper.FlashMessage{})
	sessionStore.RegisterType(uuid.UUID{})
	sessionStore.RegisterType(&models.Admin{})
	app := fiber.New(fiber.Config{
		RegexHandler: coregex.MustCompile,
		JSONEncoder: func(v any) ([]byte, error) {
			return json.Marshal(v)
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
		SingleUseToken: true,
	})
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
			Rules: map[string]string{
				"/admin/auth/*":                  "/auth/admin/$1",
				"/admin":                         "/home/admin",
				"/admin/articles":                "/articles/admin",
				"/admin/articles/*":              "/articles/admin/$1",
				"/admin/blogs":                   "/blogs/admin",
				"/admin/blogs/*":                 "/blogs/admin/$1",
				"/admin/blocks/*":                "/blocks/admin/$1",
				"/admin/blocks/*/*":              "/blocks/admin/$1/$2",
				"/admin/blocks/*/*/*":            "/blocks/admin/$1/$2/$3",
				"/admin/product_categories":      "/product_categories/admin",
				"/admin/product_categories/*":    "/product_categories/admin/$1",
				"/admin/product_categories/*/*":  "/product_categories/admin/$1/$2",
				"/admin/products":                "/products/admin",
				"/admin/products/*":              "/products/admin/$1",
				"/admin/products/*/*":            "/products/admin/$1/$2",
				"/admin/settings":                "/settings/admin",
				"/admin/settings/*":              "/settings/admin/$1",
				"/admin/settings/*/*":            "/settings/admin/$1/$2",
				"/admin/delivery_plans":          "/delivery_plans/admin",
				"/admin/delivery_plans/*":        "/delivery_plans/admin/$1",
				"/admin/delivery_plans/*/*":      "/delivery_plans/admin/$1/$2",
				"/admin/non_delivery_days":       "/non_delivery_days/admin",
				"/admin/non_delivery_days/*":     "/non_delivery_days/admin/$1",
				"/admin/non_delivery_days/*/*":   "/non_delivery_days/admin/$1/$2",
				"/admin/points":                  "/points/admin",
				"/admin/points/*":                "/points/admin/$1",
				"/admin/points/*/*":              "/points/admin/$1/$2",
				"/admin/vouchers":                "/vouchers/admin",
				"/admin/vouchers/*":              "/vouchers/admin/$1",
				"/admin/vouchers/*/*":            "/vouchers/admin/$1/$2",
				"/admin/point_earning_caps":      "/point_earning_caps/admin",
				"/admin/point_earning_caps/*":    "/point_earning_caps/admin/$1",
				"/admin/point_earning_caps/*/*":  "/point_earning_caps/admin/$1/$2",
				"/admin/point_spending_caps":     "/point_spending_caps/admin",
				"/admin/point_spending_caps/*":   "/point_spending_caps/admin/$1",
				"/admin/point_spending_caps/*/*": "/point_spending_caps/admin/$1/$2",
				"/admin/transactions":            "/transactions/admin",
				"/admin/transactions/*":          "/transactions/admin/$1",
				"/admin/transactions/*/*":        "/transactions/admin/$1/$2",
				"/v1/buyer/auth/*":               "/auth/v1/buyer/$1",
				"/v1/buyer/fcm_registrations":    "/fcm_registrations/v1/buyer",
				"/v1/buyer/product_categories":   "/product_categories/v1/buyer",
				"/v1/buyer/products":             "/products/v1/buyer",
				"/v1/buyer/products/*":           "/products/v1/buyer/$1",
				"/v1/buyer/cart":                 "/cart/v1/buyer",
				"/v1/buyer/cart/*":               "/cart/v1/buyer/$1",
				"/v1/buyer/cart/*/*":             "/cart/v1/buyer/$1/$2",
				"/v1/buyer/vouchers/*":           "/vouchers/v1/buyer/$1",
				"/v1/buyer/delivery_plans":       "/delivery_plans/v1/buyer",
				"/v1/buyer/transactions":         "/transactions/v1/buyer",
				"/v1/buyer/transactions/*":       "/transactions/v1/buyer/$1",
				"/v1/buyer/transactions/*/*":     "/transactions/v1/buyer/$1/$2",
				"/v1/buyer/delivery_routes":      "/delivery_routes/v1/buyer",
				"/":                              "/blocks",
				"/blogs":                         "/blocks/blogs",
				"/read/*":                        "/blocks/blogs/$1",
				"/contact-us":                    "/blocks/contact-us",
				"/faq":                           "/blocks/faq",
				"/privacy-policy":                "/blocks/privacy-policy",
			},
		}),
		cors.New(),
		pprof.New(),
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
	basicAuth := middleware.BasicAuth()
	if debug {
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
	).Mount(app.Group("/auth"))
	controllers.NewHomeController(q.AdminService).Mount(app.Group("/home"))
	controllers.NewArticleController(q.AdminService, q.ArticleService).Mount(app.Group("/articles"))
	controllers.NewBlogController(q.AdminService, q.BlogService).Mount(app.Group("/blogs"))
	controllers.NewBlockController(
		q.AdminService,
		q.BlockService,
		q.BlogService,
		q.MessageService,
	).Mount(app.Group("/blocks"))
	controllers.NewProductCategoryController(q.AdminService, q.ProductCategoryService).Mount(app.Group("/product_categories"))
	controllers.NewProductController(
		q.AdminService,
		q.ProductCategoryService,
		q.ProductService,
		q.ProductVariantService,
	).Mount(app.Group("/products"))
	controllers.NewDeliveryPlanController(
		q.AdminService,
		q.DeliveryPlanService,
		q.BuyerService,
		q.JsonWebTokenService,
	).Mount(app.Group("/delivery_plans"))
	controllers.NewNonDeliveryDayController(
		q.AdminService,
		q.NonDeliveryDayService,
		q.DeliveryPlanService,
	).Mount(app.Group("/non_delivery_days"))
	controllers.NewPointController(
		q.AdminService,
		q.BuyerService,
		q.PointService,
	).Mount(app.Group("/points"))
	controllers.NewVoucherController(
		q.AdminService,
		q.VoucherService,
		q.BuyerService,
		q.JsonWebTokenService,
	).Mount(app.Group("/vouchers"))
	controllers.NewPointEarningCapController(
		q.AdminService,
		q.PointEarningCapService,
	).Mount(app.Group("/point_earning_caps"))
	controllers.NewPointSpendingCapController(
		q.AdminService,
		q.PointSpendingCapService,
	).Mount(app.Group("/point_spending_caps"))
	controllers.NewTransactionController(
		q.AdminService,
		q.BuyerService,
		q.TransactionService,
		q.TransactionCounterService,
		q.CartService,
		q.CartItemService,
		q.DeliveryRouteService,
		q.SettingService,
		q.VoucherService,
		q.PointSpendingCapService,
		q.PointService,
		q.ProductVariantService,
		q.JsonWebTokenService,
	).Mount(app.Group("/transactions"))
	controllers.NewFcmRegistrationController(
		q.BuyerService,
		q.FcmRegistrationService,
		q.JsonWebTokenService,
	).Mount(app.Group("/fcm_registrations"))
	controllers.NewCartController(
		q.BuyerService,
		q.ProductVariantService,
		q.CartService,
		q.CartItemService,
		q.JsonWebTokenService,
	).Mount(app.Group("/cart"))
	controllers.NewDeliveryRouteController(
		q.AdminService,
		q.BuyerService,
		q.DeliveryRouteService,
		q.SettingService,
		q.JsonWebTokenService,
		q.RoutesApiConsumer,
	).Mount(app.Group("/delivery_routes"))
	app.Use(func(c fiber.Ctx) error {
		if strings.HasPrefix(c.OriginalURL(), "/v1/") {
			return helper.NewResponse(fiber.StatusNotFound).
				SetMessage(http.StatusText(fiber.StatusNotFound)).
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
	err := app.Listen(listenerPort)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrListen")
	}
	return err
}
