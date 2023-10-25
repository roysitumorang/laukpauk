package query

import (
	"context"

	"github.com/roysitumorang/laukpauk/modules/user/model"
)

type (
	UserQuery interface {
		FindUsers(ctx context.Context, filter model.UserFilter) (response []model.User, err error)
		ChangePassword(ctx context.Context, userID int64, encryptedPassword string) (err error)
	}
)
