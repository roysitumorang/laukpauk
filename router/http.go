package router

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/redirect"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/roysitumorang/laukpauk/helper"
	authPresenter "github.com/roysitumorang/laukpauk/modules/auth/presenter"
	bannerPresenter "github.com/roysitumorang/laukpauk/modules/banner/presenter"
	regionPresenter "github.com/roysitumorang/laukpauk/modules/region/presenter"
	"go.uber.org/zap"
)

const (
	DefaultPort = 8080
)

func (q *Service) HTTPServerMain() error {
	logger, _ := zap.NewProduction()
	r := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			if err = helper.NewResponse(code, err.Error(), nil).WriteResponse(ctx); err != nil {
				return helper.NewResponse(fiber.StatusInternalServerError, "Internal Server Error", nil).WriteResponse(ctx)
			}
			return nil
		},
	})
	r.Use(
		recover.New(),
		fiberzap.New(fiberzap.Config{
			Logger: logger,
		}),
		requestid.New(),
		compress.New(),
		redirect.New(redirect.Config{
			Rules: map[string]string{
				"/api/v1/cities/*":                    "/api/v1/region/cities/$1",
				"/api/v1/provinces":                   "/api/v1/region/provinces",
				"/api/v1/provinces/*":                 "/api/v1/region/provinces/$1",
				"/api/v1/subdistricts/*":              "/api/v1/region/subdistricts/$1",
				"/api/v1/admin/auth/login":            "/api/v1/auth/admin/login",
				"/api/v1/admin/auth/password/change":  "/api/v1/auth/admin/password/change",
				"/api/v1/admin/auth/profile":          "/api/v1/auth/admin/profile",
				"/api/v1/buyer/auth/login":            "/api/v1/auth/buyer/login",
				"/api/v1/buyer/auth/password/change":  "/api/v1/auth/buyer/password/change",
				"/api/v1/buyer/auth/profile":          "/api/v1/auth/buyer/profile",
				"/api/v1/buyer/auth/register":         "/api/v1/auth/buyer/register",
				"/api/v1/buyer/auth/*/activate":       "/api/v1/auth/buyer/$1/activate",
				"/api/v1/seller/auth/login":           "/api/v1/auth/seller/login",
				"/api/v1/seller/auth/password/change": "/api/v1/auth/seller/password/change",
				"/api/v1/seller/auth/profile":         "/api/v1/auth/seller/profile",
				"/api/v1/seller/auth/register":        "/api/v1/auth/seller/register",
				"/api/v1/seller/auth/*/activate":      "/api/v1/auth/seller/$1/activate",
			},
			StatusCode: fiber.StatusPermanentRedirect,
		}),
	)
	api := r.Group("/api")
	v1 := api.Group("/v1")
	authPresenter.NewAuthHTTPHandler(q.AuthUseCase, q.UserUseCase).Mount(v1.Group("/auth"))
	bannerPresenter.NewBannerHTTPHandler(q.BannerUseCase).Mount(v1.Group("/banners"))
	regionPresenter.NewRegionHTTPHandler(q.RegionUseCase).Mount(v1.Group("/region"))
	var port uint16
	if envPort, ok := os.LookupEnv("PORT"); ok {
		portInt, err := strconv.Atoi(envPort)
		if err != nil {
			port = DefaultPort
		} else {
			port = uint16(portInt)
		}
	} else {
		port = DefaultPort
	}
	listenerPort := fmt.Sprintf(":%d", port)
	return r.Listen(listenerPort)
}
