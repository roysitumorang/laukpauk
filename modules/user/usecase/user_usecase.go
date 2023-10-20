package usecase

import (
	"context"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/user/model"
	userQuery "github.com/roysitumorang/laukpauk/modules/user/query"
	"go.uber.org/zap"
)

type (
	userUseCaseImplementation struct {
		userQuery userQuery.UserQuery
	}
)

func NewUserUseCase(
	userQuery userQuery.UserQuery,
) UserUseCase {
	return &userUseCaseImplementation{
		userQuery: userQuery,
	}
}

func (q *userUseCaseImplementation) FindUsers(ctx context.Context, filter model.UserFilter) (response []model.User, err error) {
	ctxt := "UserUseCase-FindUsers"
	if response, err = q.userQuery.FindUsers(ctx, filter); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindUsers")
	}
	return
}
