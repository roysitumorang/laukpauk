package usecase

import (
	"context"
	"time"
	"unsafe"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/keys"
	authModel "github.com/roysitumorang/laukpauk/modules/auth/model"
	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
	userQuery "github.com/roysitumorang/laukpauk/modules/user/query"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type (
	authUseCaseImplementation struct {
		userQuery userQuery.UserQuery
	}
)

func NewAuthUseCase(
	userQuery userQuery.UserQuery,
) AuthUseCase {
	return &authUseCaseImplementation{
		userQuery: userQuery,
	}
}

func (q *authUseCaseImplementation) Login(ctx context.Context, roleID int64, request authModel.LoginRequest) (response authModel.LoginResponse, err error) {
	ctxt := "AuthUseCase-Login"
	users, err := q.userQuery.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs:      []int64{roleID},
			Status:       []int{userModel.StatusActive},
			MobilePhones: []string{request.MobilePhone},
		},
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindUsers")
		return
	}
	if len(users) == 0 {
		err = authModel.ErrLoginFailed
		return
	}
	user := users[0]
	encryptedPassword := unsafe.Slice(unsafe.StringData(user.Password), len(user.Password))
	password := unsafe.Slice(unsafe.StringData(request.Password), len(request.Password))
	if err = bcrypt.CompareHashAndPassword(encryptedPassword, password); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCompareHashAndPassword")
		err = authModel.ErrLoginFailed
		return
	}
	expiryTime := time.Now().Add(time.Hour * 72).Unix()
	claims := jwt.MapClaims{
		"id":  user.ID,
		"exp": expiryTime,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	privateKey, err := keys.InitPrivateKey()
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrInitPrivateKey")
		return
	}
	if response.IDToken, err = token.SignedString(privateKey); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSignedString")
		return
	}
	response.ExpiresIn = expiryTime
	response.Profile = user
	return
}
