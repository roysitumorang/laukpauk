package middlewares

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/keys"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

func BuyerKeyAuthMiddleware(
	buyerService services.BuyerService,
	jsonWebTokenService services.JsonWebTokenService,
) fiber.Handler {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "header: %s", fiber.HeaderAuthorization)
	return keyauth.New(keyauth.Config{
		SuccessHandler: func(c fiber.Ctx) error {
			return c.Next()
		},
		ErrorHandler: func(c fiber.Ctx, err error) error {
			if err == nil {
				err = keyauth.ErrMissingOrMalformedAPIKey
			}
			return helper.NewResponse(fiber.StatusUnauthorized).SetMessage(err.Error()).WriteResponse(c)
		},
		Validator: func(c fiber.Ctx, token string) (bool, error) {
			claims, err := bearerVerify(token)
			if err != nil {
				return false, err
			}
			buyerID, err := uuid.Parse(claims.ID)
			if err != nil {
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}
			ctx := c.Context()
			currentBuyer, err := buyerService.FindBuyer(
				ctx,
				models.NewBuyerFilter().
					WithBuyerIDs(buyerID),
			)
			if err != nil || currentBuyer == nil {
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}
			jsonWebTokenID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}
			currentJsonWebToken, err := jsonWebTokenService.FindJsonWebToken(
				ctx,
				models.NewJsonWebTokenFilter().
					WithJsonWebTokenIDs(jsonWebTokenID),
			)
			if err != nil || currentJsonWebToken == nil {
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}
			c.Locals(models.CurrentBuyer, currentBuyer)
			c.Locals(models.CurrentJsonWebToken, currentJsonWebToken)
			return true, nil
		},
	})
}

func bearerVerify(tokenString string) (*jwt.RegisteredClaims, error) {
	var claimsStruct jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(_ *jwt.Token) (any, error) {
			return keys.InitPublicKey()
		},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid JWT")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errors.New("invalid JWT")
	}
	if claims.Issuer != helper.GetJwtIssuer() {
		return nil, errors.New("iss is invalid")
	}
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("JWT is expired")
	}
	return claims, nil
}
