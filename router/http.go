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
	authPresenter "github.com/roysitumorang/laukpauk/modules/auth/presenter"
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
			response := map[string]interface{}{
				"code":    code,
				"message": err.Error(),
			}
			err = ctx.Status(code).JSON(response)
			if err != nil {
				response["code"] = fiber.StatusInternalServerError
				response["message"] = "Internal Server Error"
				return ctx.Status(fiber.StatusInternalServerError).JSON(response)
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
				"/api/v1/provinces":      "/api/v1/region/provinces",
				"/api/v1/provinces/*":    "/api/v1/region/provinces/$1",
				"/api/v1/cities/*":       "/api/v1/region/cities/$1",
				"/api/v1/subdistricts/*": "/api/v1/region/subdistricts/$1",
			},
			StatusCode: fiber.StatusMovedPermanently,
		}),
	)
	api := r.Group("/api")
	v1 := api.Group("/v1")
	authPresenter.NewAuthHTTPHandler(q.AuthUseCase, q.UserUseCase).Mount(v1.Group("/auth"))
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