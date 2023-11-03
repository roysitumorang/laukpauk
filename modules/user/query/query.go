package query

import (
	"context"

	authModel "github.com/roysitumorang/laukpauk/modules/auth/model"
	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
)

type (
	UserQuery interface {
		FindUsers(ctx context.Context, filter userModel.UserFilter) (response []userModel.User, err error)
		ChangePassword(ctx context.Context, userID int64, encryptedPassword string) (err error)
		Register(ctx context.Context, request authModel.RegisterRequest) (response *authModel.RegisterResponse, err error)
		Activate(ctx context.Context, roleID int64, activationToken string) (response int64, err error)
	}
)
