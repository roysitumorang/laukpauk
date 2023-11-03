package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/keys"
	authModel "github.com/roysitumorang/laukpauk/modules/auth/model"
	regionQuery "github.com/roysitumorang/laukpauk/modules/region/query"
	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
	userQuery "github.com/roysitumorang/laukpauk/modules/user/query"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type (
	authUseCaseImplementation struct {
		userQuery   userQuery.UserQuery
		regionQuery regionQuery.RegionQuery
	}
)

func NewAuthUseCase(
	userQuery userQuery.UserQuery,
	regionQuery regionQuery.RegionQuery,
) AuthUseCase {
	return &authUseCaseImplementation{
		userQuery:   userQuery,
		regionQuery: regionQuery,
	}
}

func (q *authUseCaseImplementation) Login(ctx context.Context, roleIDs []int64, request authModel.LoginRequest) (response authModel.LoginResponse, err error) {
	ctxt := "AuthUseCase-Login"
	users, err := q.userQuery.FindUsers(
		ctx,
		userModel.UserFilter{
			RoleIDs:      roleIDs,
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
	encryptedPassword := helper.String2ByteSlice(user.Password)
	password := helper.String2ByteSlice(request.Password)
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

func (q *authUseCaseImplementation) ChangePassword(ctx context.Context, userID int64, encryptedOldPassword string, request authModel.ChangePassword) (err error) {
	ctxt := "AuthUseCase-ChangePassword"
	encryptedOldPasswordByte := helper.String2ByteSlice(encryptedOldPassword)
	oldPassword := helper.String2ByteSlice(request.OldPassword)
	newPassword := helper.String2ByteSlice(request.NewPassword)
	if err = bcrypt.CompareHashAndPassword(encryptedOldPasswordByte, oldPassword); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCompareHashAndOldPassword")
		return errors.New("invalid old password")
	}
	if err = bcrypt.CompareHashAndPassword(encryptedOldPasswordByte, newPassword); err == nil {
		return errors.New("reusing old password is prohibited")
	}
	encryptedNewPassword, err := bcrypt.GenerateFromPassword(newPassword, bcrypt.DefaultCost)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGenerateFromPassword")
		return
	}
	if err = q.userQuery.ChangePassword(ctx, userID, helper.ByteSlice2String(encryptedNewPassword)); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrChangePassword")
	}
	return
}

func (q *authUseCaseImplementation) Register(ctx context.Context, request authModel.RegisterRequest) (response *authModel.RegisterResponse, err error) {
	ctxt := "AuthUseCase-Register"
	encryptedNewPassword, err := bcrypt.GenerateFromPassword(helper.String2ByteSlice(request.Password), bcrypt.DefaultCost)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGenerateFromPassword")
		return
	}
	request.Password = helper.ByteSlice2String(encryptedNewPassword)
	village, err := q.regionQuery.FindVillageByID(ctx, request.VillageID)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVillageByID")
		return
	}
	if village == nil || village.ID == 0 {
		err = errors.New("village not found")
		return
	}
	request.SubdistrictID = village.SubdistrictID
	if response, err = q.userQuery.Register(ctx, request); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRegister")
	}
	return
}

func (q *authUseCaseImplementation) Activate(ctx context.Context, roleID int64, activationToken string) (response authModel.LoginResponse, err error) {
	ctxt := "AuthUseCase-Login"
	userID, err := q.userQuery.Activate(ctx, roleID, activationToken)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrActivate")
		return
	}
	if userID == 0 {
		err = authModel.ErrActivationFailed
		return
	}
	users, err := q.userQuery.FindUsers(
		ctx,
		userModel.UserFilter{
			UserIDs: []int64{userID},
		},
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindUsers")
		return
	}
	if len(users) == 0 {
		err = authModel.ErrActivationFailed
		return
	}
	user := users[0]
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
