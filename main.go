//	@title			Laukpauk API
//	@version		0.1.0
//	@description	This is documentation of Laukpauk API.

//	@contact.name	Roy Situmorang
//	@contact.email	roy.situmorang@gmail.com

//	@host
//	@BasePath	/v1

//	@accept		json
//	@produce	json

//	@schemes	http https

//	@securitydefinitions.apikey	apiKey
//	@in							header
//	@name						x-api-key
//	@description				Laukpauk API call requires X-Api-Key request header

// @Security	ApiKeyAuth
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/router"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctxt := "Main"
	ctx := context.Background()
	helper.InitLogger()
	cmd := &cli.Command{
		Name: config.AppName,
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "print version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Printf("Version: %s\nCommit: %s\nBuild: %s\n", config.Version, config.Commit, config.Build)
					return nil
				},
			},
			{
				Name:  "run",
				Usage: "run app",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					err := godotenv.Load(".env")
					if err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrLoad")
						return err
					}
					if err = helper.InitHelper(); err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrInitHelper")
						return err
					}
					service, err := router.MakeHandler(ctx)
					if err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrMakeHandler")
						return err
					}
					helper.InitDbWrite(service.DbWrite)
					if err = service.Migration.Migrate(ctx); err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrMigrate")
						return err
					}
					var g errgroup.Group
					g.Go(func() error {
						return service.HTTPServerMain(ctx)
					})
					g.Go(func() error {
						return service.Consume(ctx)
					})
					g.Go(func() error {
						c := cron.New(cron.WithChain(
							cron.Recover(cron.DefaultLogger),
						))
						// run every minute
						entryID, err := c.AddFunc("* * * * *", func() {
							tx, err := helper.BeginTx(ctx)
							if err != nil {
								helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
								return
							}
							defer func() {
								errRollback := tx.Rollback(ctx)
								if errors.Is(errRollback, pgx.ErrTxClosed) {
									errRollback = nil
								}
								if errRollback != nil {
									helper.Log(ctx, slog.LevelError, errRollback.Error(), ctxt, "ErrRollback")
								}
							}()
							if err = tx.Commit(ctx); err != nil {
								helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
							}
						})
						if err != nil {
							helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrAddFunc")
							return err
						}
						helper.Log(ctx, slog.LevelInfo, fmt.Sprintf("cron: entry added with ID %d", entryID), ctxt, "")
						c.Start()
						helper.Log(ctx, slog.LevelInfo, "cron: scheduled tasks running!...", ctxt, "")
						return nil
					})
					err = g.Wait()
					if err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrWait")
					}
					return err
				},
			},
			{
				Name:  "migration",
				Usage: "new/run migration",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					now := time.Now()
					err := godotenv.Load(".env")
					if err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrLoad")
						return err
					}
					if err = helper.InitHelper(); err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrInitHelper")
						return err
					}
					service, err := router.MakeHandler(ctx)
					if err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrMakeHandler")
						return err
					}
					helper.InitDbWrite(service.DbWrite)
					var activity string
					switch cmd.Args().First() {
					case "new":
						if err = service.Migration.CreateMigrationFile(ctx); err != nil {
							helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCreateMigrationFile")
							return err
						}
						activity = "creating"
					case "run":
						if err = service.Migration.Migrate(ctx); err != nil {
							helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrMigrate")
							return err
						}
						activity = "running"
					default:
						return errors.New("requires an option: new | run")
					}
					duration := time.Since(now)
					helper.Log(ctx, slog.LevelInfo, fmt.Sprintf("%s migration successfully in %s", activity, duration.String()), ctxt, "")
					return nil
				},
			},
		},
	}
	if err := cmd.Run(ctx, os.Args); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrRun")
	}
}
