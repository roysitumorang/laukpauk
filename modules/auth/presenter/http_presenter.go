package presenter

import (
	"context"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/keys"
	"github.com/roysitumorang/laukpauk/modules/auth/sanitizer"
	authUseCase "github.com/roysitumorang/laukpauk/modules/auth/usecase"
	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
	userUseCase "github.com/roysitumorang/laukpauk/modules/user/usecase"
	"go.uber.org/zap"
)

type (
	authHTTPHandler struct {
		authUseCase authUseCase.AuthUseCase
		userUseCase userUseCase.UserUseCase
	}
)

func NewAuthHTTPHandler(
	authUseCase authUseCase.AuthUseCase,
	userUseCase userUseCase.UserUseCase,
) *authHTTPHandler {
	return &authHTTPHandler{
		authUseCase: authUseCase,
		userUseCase: userUseCase,
	}
}

func (q *authHTTPHandler) Mount(r fiber.Router) {
	privateKey, _ := keys.InitPrivateKey()
	r.Post("/signin", q.SignIn)
	r.Use(
		jwtware.New(jwtware.Config{
			SigningKey: jwtware.SigningKey{
				JWTAlg: jwtware.RS256,
				Key:    privateKey.Public(),
			},
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				code := fiber.StatusUnauthorized
				return ctx.Status(code).JSON(map[string]interface{}{
					"code":    code,
					"message": err.Error(),
				})
			},
		}),
	).
		Get("/me", q.AboutMe)
}

func (q *authHTTPHandler) SignIn(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-SignIn"
	request, statusCode, err := sanitizer.SignIn(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSignIn")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	response, err := q.authUseCase.SignIn(ctx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSignIn")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) AboutMe(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-AboutMe"
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok || userID < 1 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	users, err := q.userUseCase.FindUsers(ctx, userModel.UserFilter{UserIDs: []int64{int64(userID)}})
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindUsers")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	if len(users) == 0 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	response := users[0]
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}
