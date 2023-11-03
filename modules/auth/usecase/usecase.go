package usecase

import (
	"context"

	"github.com/roysitumorang/laukpauk/modules/auth/model"
)

type (
	AuthUseCase interface {
		Login(ctx context.Context, roleIDs []int64, request model.LoginRequest) (response model.LoginResponse, err error)
		ChangePassword(ctx context.Context, userID int64, encryptedPassword string, request model.ChangePassword) (err error)
		Register(ctx context.Context, request model.RegisterRequest) (response *model.RegisterResponse, err error)
		Activate(ctx context.Context, roleID int64, activationToken string) (response model.LoginResponse, err error)
	}
)
