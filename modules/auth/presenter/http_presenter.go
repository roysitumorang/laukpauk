package presenter

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/laukpauk/helper"
	middlewareJWT "github.com/roysitumorang/laukpauk/middleware/jwt"
	"github.com/roysitumorang/laukpauk/modules/auth/sanitizer"
	authUseCase "github.com/roysitumorang/laukpauk/modules/auth/usecase"
	roleModel "github.com/roysitumorang/laukpauk/modules/role/model"
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
	bearerVerifier := middlewareJWT.NewJWT()
	admin := r.Group("/admin")
	admin.Post("/login", q.AdminLogin)
	admin.Use(bearerVerifier).
		Get("/profile", q.AdminGetProfile).
		Put("/password/change", q.AdminChangePassword)
	buyer := r.Group("/buyer")
	buyer.Post("/register", q.BuyerRegister).
		Put("/:token/activate", q.BuyerActivate).
		Post("/login", q.BuyerLogin)
	buyer.Use(bearerVerifier).
		Get("/profile", q.BuyerGetProfile).
		Put("/password/change", q.BuyerChangePassword)
	seller := r.Group("/seller")
	seller.Post("/register", q.SellerRegister).
		Put("/:token/activate", q.SellerActivate).
		Post("/login", q.SellerLogin)
	seller.Use(bearerVerifier).
		Get("/profile", q.SellerGetProfile).
		Put("/password/change", q.SellerChangePassword)
}

func (q *authHTTPHandler) AdminLogin(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-AdminLogin"
	request, statusCode, err := sanitizer.Login(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrLogin")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	response, err := q.authUseCase.Login(ctx, []int64{roleModel.RoleSuperAdmin, roleModel.RoleAdmin}, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrLogin")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) AdminGetProfile(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-AdminGetProfile"
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok || userID < 1 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	users, err := q.userUseCase.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs: []int64{roleModel.RoleSuperAdmin, roleModel.RoleAdmin},
			Status:  []int{userModel.StatusActive},
			UserIDs: []int64{int64(userID)},
		},
	)
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

func (q *authHTTPHandler) AdminChangePassword(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-AdminChangePassword"
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok || userID < 1 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	users, err := q.userUseCase.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs: []int64{roleModel.RoleSuperAdmin, roleModel.RoleAdmin},
			Status:  []int{userModel.StatusActive},
			UserIDs: []int64{int64(userID)},
		},
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindUsers")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	if len(users) == 0 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	currentUser := users[0]
	request, statusCode, err := sanitizer.ChangePassword(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrChangePassword")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	if err = q.authUseCase.ChangePassword(ctx, currentUser.ID, currentUser.Password, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrChangePassword")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusNoContent, "", nil).WriteResponse(c)
}

func (q *authHTTPHandler) BuyerRegister(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-BuyerRegister"
	request, statusCode, err := sanitizer.Register(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRegister")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	request.RoleID = roleModel.RoleBuyer
	response, err := q.authUseCase.Register(ctx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRegister")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) BuyerActivate(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-BuyerActivate"
	activationToken := c.Params("token")
	response, err := q.authUseCase.Activate(ctx, roleModel.RoleBuyer, activationToken)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrActivate")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) BuyerLogin(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-BuyerLogin"
	request, statusCode, err := sanitizer.Login(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrLogin")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	response, err := q.authUseCase.Login(ctx, []int64{roleModel.RoleBuyer}, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrLogin")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) BuyerGetProfile(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-BuyerGetProfile"
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok || userID < 1 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	users, err := q.userUseCase.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs: []int64{roleModel.RoleBuyer},
			Status:  []int{userModel.StatusActive},
			UserIDs: []int64{int64(userID)},
		},
	)
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

func (q *authHTTPHandler) BuyerChangePassword(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-BuyerChangePassword"
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok || userID < 1 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	users, err := q.userUseCase.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs: []int64{roleModel.RoleBuyer},
			Status:  []int{userModel.StatusActive},
			UserIDs: []int64{int64(userID)},
		},
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindUsers")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	if len(users) == 0 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	currentUser := users[0]
	request, statusCode, err := sanitizer.ChangePassword(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrChangePassword")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	if err = q.authUseCase.ChangePassword(ctx, currentUser.ID, currentUser.Password, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrChangePassword")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusNoContent, "", nil).WriteResponse(c)
}

func (q *authHTTPHandler) SellerRegister(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-SellerRegister"
	request, statusCode, err := sanitizer.Register(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRegister")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	request.RoleID = roleModel.RoleSeller
	response, err := q.authUseCase.Register(ctx, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRegister")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) SellerActivate(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-SellerActivate"
	activationToken := c.Params("token")
	response, err := q.authUseCase.Activate(ctx, roleModel.RoleSeller, activationToken)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrActivate")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) SellerLogin(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-SellerLogin"
	request, statusCode, err := sanitizer.Login(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrLogin")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	response, err := q.authUseCase.Login(ctx, []int64{roleModel.RoleSeller}, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrLogin")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}

func (q *authHTTPHandler) SellerGetProfile(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-SellerGetProfile"
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok || userID < 1 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	users, err := q.userUseCase.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs: []int64{roleModel.RoleSeller},
			Status:  []int{userModel.StatusActive},
			UserIDs: []int64{int64(userID)},
		},
	)
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

func (q *authHTTPHandler) SellerChangePassword(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "AuthPresenter-SellerChangePassword"
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok || userID < 1 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	users, err := q.userUseCase.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs: []int64{roleModel.RoleSeller},
			Status:  []int{userModel.StatusActive},
			UserIDs: []int64{int64(userID)},
		},
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindUsers")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	if len(users) == 0 {
		return helper.NewResponse(fiber.StatusUnauthorized, "unauthorized", nil).WriteResponse(c)
	}
	currentUser := users[0]
	request, statusCode, err := sanitizer.ChangePassword(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrChangePassword")
		return helper.NewResponse(statusCode, err.Error(), nil).WriteResponse(c)
	}
	if err = q.authUseCase.ChangePassword(ctx, currentUser.ID, currentUser.Password, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrChangePassword")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusNoContent, "", nil).WriteResponse(c)
}
