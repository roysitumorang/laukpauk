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
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/router"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctxt := "Main"
	ctx := context.Background()
	logger := helper.NewLogger()
	cmd := &cli.Command{
		Name: config.AppName,
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "print version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					_, _ = os.Stdout.WriteString(helper.Sprintf("Version: %s\nCommit: %s\nBuild: %s\n", config.Version, config.Commit, config.Build))
					return nil
				},
			}, {
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
					if err = service.Migration.Migrate(ctx); err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrMigrate")
						return err
					}
					var g errgroup.Group
					g.Go(func() error {
						return service.HTTPServerMain(ctx, logger)
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
							smsMessages, _, err := service.SmsMessageService.FindSmsMessages(
								ctx,
								models.NewSmsMessageFilter().
									WithDeliveryStatusBlank(),
							)
							if err != nil {
								helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSmsMessages")
								return
							}
							n := len(smsMessages)
							if n == 0 {
								return
							}
							messages := make([]models.TopicMessage, n)
							for i, smsMessage := range smsMessages {
								messages[i] = models.TopicMessage{
									Topic: models.TopicSms,
									Event: models.TopicSmsEventCheckReport,
									ID:    smsMessage.ID,
								}
							}
							if err = service.KafkaConsumer.Publish(
								ctx,
								models.TopicRequest{
									Topic:    models.TopicSms,
									Messages: messages,
								},
							); err != nil {
								helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrPublish")
								return
							}
						})
						if err != nil {
							helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrAddFunc")
							return err
						}
						helper.Log(ctx, slog.LevelInfo, helper.Sprintf("cron: entry added with ID %d", entryID), ctxt, "")
						if entryID, err = c.AddFunc("* * * * *", func() {
							if err = service.TransactionService.CancelOverdueTransactions(ctx); err != nil {
								helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCancelOverdueTransactions")
							}
						}); err != nil {
							helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrAddFunc")
							return err
						}
						helper.Log(ctx, slog.LevelInfo, helper.Sprintf("cron: entry added with ID %d", entryID), ctxt, "")
						c.Start()
						helper.Log(ctx, slog.LevelInfo, "cron: scheduled tasks running!...", ctxt, "")
						return nil
					})
					if err = g.Wait(); err != nil {
						helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrWait")
					}
					return err
				},
			}, {
				Name:  "migration",
				Usage: "migration new|run",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Len() == 0 ||
						(cmd.Args().First() != "new" &&
							cmd.Args().First() != "run") {
						return errors.New("requires a parameter: new | run")
					}
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
					}
					duration := time.Since(now)
					helper.Log(ctx, slog.LevelInfo, helper.Sprintf("%s migration successfully in %s", activity, duration.String()), ctxt, "")
					return nil
				},
			},
		},
	}
	if err := cmd.Run(ctx, os.Args); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrRun")
	}
}
