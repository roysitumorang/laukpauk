package middlewares

import (
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/basicauth"
	"github.com/roysitumorang/laukpauk/helper"
)

func BasicAuthMiddleware() fiber.Handler {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{
			os.Getenv("BASIC_AUTH_USERNAME"): os.Getenv("BASIC_AUTH_PASSWORD"),
		},
		Unauthorized: func(c fiber.Ctx) error {
			return helper.NewResponse(fiber.StatusUnauthorized).SetMessage("Unauthorized").WriteResponse(c)
		},
	})
}
