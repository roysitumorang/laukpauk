package router

import (
	"github.com/roysitumorang/laukpauk/config"
	authUseCase "github.com/roysitumorang/laukpauk/modules/auth/usecase"
	regionQuery "github.com/roysitumorang/laukpauk/modules/region/query"
	regionUseCase "github.com/roysitumorang/laukpauk/modules/region/usecase"
	userQuery "github.com/roysitumorang/laukpauk/modules/user/query"
	userUseCase "github.com/roysitumorang/laukpauk/modules/user/usecase"
)

type (
	Service struct {
		AuthUseCase   authUseCase.AuthUseCase
		RegionUseCase regionUseCase.RegionUseCase
		UserUseCase   userUseCase.UserUseCase
	}
)

func MakeHandler() *Service {
	dbRead := config.GetDbReadOnly()
	dbWrite := config.GetDbWriteOnly()
	regionQuery := regionQuery.NewRegionQuery(dbRead, dbWrite)
	userQuery := userQuery.NewUserQuery(dbRead, dbWrite)
	authUseCase := authUseCase.NewAuthUseCase(userQuery)
	regionUseCase := regionUseCase.NewRegionUseCase(regionQuery)
	userUseCase := userUseCase.NewUserUseCase(userQuery)
	return &Service{
		AuthUseCase:   authUseCase,
		RegionUseCase: regionUseCase,
		UserUseCase:   userUseCase,
	}
}
