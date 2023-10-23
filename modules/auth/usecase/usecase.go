package usecase

import (
	"context"

	"github.com/roysitumorang/laukpauk/modules/auth/model"
)

type (
	AuthUseCase interface {
		Login(ctx context.Context, roleIDs []int64, request model.LoginRequest) (response model.LoginResponse, err error)
	}
)
