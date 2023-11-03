package jwt

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/keys"
)

func NewJWT() func(*fiber.Ctx) error {
	privateKey, _ := keys.InitPrivateKey()
	return jwtware.New(
		jwtware.Config{
			SigningKey: jwtware.SigningKey{
				JWTAlg: jwtware.RS256,
				Key:    privateKey.Public(),
			},
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				return helper.NewResponse(fiber.StatusUnauthorized, err.Error(), nil).WriteResponse(ctx)
			},
		},
	)
}
