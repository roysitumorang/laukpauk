package router

import (
	"context"

	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/migration"
	authUseCase "github.com/roysitumorang/laukpauk/modules/auth/usecase"
	bannerQuery "github.com/roysitumorang/laukpauk/modules/banner/query"
	bannerUseCase "github.com/roysitumorang/laukpauk/modules/banner/usecase"
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
		BannerUseCase bannerUseCase.BannerUseCase
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
	bannerQuery := bannerQuery.NewBannerQuery(dbRead, dbWrite)
	regionQuery := regionQuery.NewRegionQuery(dbRead, dbWrite)
	userQuery := userQuery.NewUserQuery(dbRead, dbWrite)
	authUseCase := authUseCase.NewAuthUseCase(userQuery)
	bannerUseCase := bannerUseCase.BannerUseCase(bannerQuery)
	regionUseCase := regionUseCase.NewRegionUseCase(regionQuery)
	userUseCase := userUseCase.NewUserUseCase(userQuery)
	return &Service{
		Migration:     migration,
		AuthUseCase:   authUseCase,
		BannerUseCase: bannerUseCase,
		RegionUseCase: regionUseCase,
		UserUseCase:   userUseCase,
	}
}
