package router

import (
	"context"

	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/migration"
	authUseCase "github.com/roysitumorang/laukpauk/modules/auth/usecase"
	regionQuery "github.com/roysitumorang/laukpauk/modules/region/query"
	regionUseCase "github.com/roysitumorang/laukpauk/modules/region/usecase"
	userQuery "github.com/roysitumorang/laukpauk/modules/user/query"
	userUseCase "github.com/roysitumorang/laukpauk/modules/user/usecase"
	"go.uber.org/zap"
)

type (
	Service struct {
		Migration     *migration.Migration
		AuthUseCase   authUseCase.AuthUseCase
		RegionUseCase regionUseCase.RegionUseCase
		UserUseCase   userUseCase.UserUseCase
	}
)

func MakeHandler() *Service {
	ctxt := "Router-MakeHandler"
	dbRead := config.GetDbReadOnly()
	dbWrite := config.GetDbWriteOnly()
	ctx := context.Background()
	tx, err := dbWrite.Begin(ctx)
	if err != nil {
		helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrBegin")
		return nil
	}
	migration := migration.NewMigration(tx)
	regionQuery := regionQuery.NewRegionQuery(dbRead, dbWrite)
	userQuery := userQuery.NewUserQuery(dbRead, dbWrite)
	authUseCase := authUseCase.NewAuthUseCase(userQuery)
	regionUseCase := regionUseCase.NewRegionUseCase(regionQuery)
	userUseCase := userUseCase.NewUserUseCase(userQuery)
	return &Service{
		Migration:     migration,
		AuthUseCase:   authUseCase,
		RegionUseCase: regionUseCase,
		UserUseCase:   userUseCase,
	}
}
