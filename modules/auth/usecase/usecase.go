package usecase

import (
	"context"

	"github.com/roysitumorang/laukpauk/modules/auth/model"
)

type (
	AuthUseCase interface {
		SignIn(ctx context.Context, request model.SignInRequest) (response model.SignInResponse, err error)
	}
)
