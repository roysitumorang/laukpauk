package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/router"
	"github.com/roysitumorang/laukpauk/services/messagingconsumer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctxt := "Main"
	ctx := context.Background()
	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "print version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Version: %s\nCommit: %s\nBuild: %s\n", config.Version, config.Commit, config.Build)
		},
	}
	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "run app",
		Run: func(_ *cobra.Command, _ []string) {
			if err := godotenv.Load(".env"); err != nil {
				helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrLoad")
			}
			var g errgroup.Group
			service := router.MakeHandler()
			g.Go(func() error {
				return service.HTTPServerMain()
			})
			g.Go(func() error {
				messagingConsumer := messagingconsumer.GetMessagingConsumerService()
				messagingConsumer.Consume(service)
				return nil
			})
			if err := g.Wait(); err != nil {
				helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrWait")
			}
		},
	}
	cmdMigration := &cobra.Command{
		Use:   "migration",
		Short: "new/run migration",
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) == 0 {
				err = errors.New("requires at least 1 arg (new|run")
				return
			}
			if args[0] != "new" && args[0] != "run" {
				err = fmt.Errorf("invalid first flag specified: %s", args[0])
			}
			return
		},
		Run: func(_ *cobra.Command, args []string) {
			if err := godotenv.Load(".env"); err != nil {
				helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrLoad")
			}
			service := router.MakeHandler()
			now := time.Now()
			switch args[0] {
			case "new":
				if err := service.Migration.CreateMigrationFile(ctx); err != nil {
					helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrCreateMigrationFile")
				}
				duration := time.Since(now)
				helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("creating migration successfully in %s", duration.String()), ctxt, "")
			case "run":
				service.Migration.Migrate(ctx)
				duration := time.Since(now)
				helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("running migration successfully in %s", duration.String()), ctxt, "")
			}
		},
	}
	rootCmd := &cobra.Command{Use: config.AppName}
	rootCmd.AddCommand(
		cmdVersion,
		cmdRun,
		cmdMigration,
	)
	rootCmd.SuggestionsMinimumDistance = 1
	_ = rootCmd.Execute()
}
