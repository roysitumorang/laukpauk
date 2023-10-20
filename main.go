package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/router"
	"github.com/roysitumorang/laukpauk/services/messagingconsumer"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func main() {
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
				log.Fatalln("can't load .env file")
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
				log.Fatal(err)
			}
		},
	}
	rootCmd := &cobra.Command{Use: config.AppName}
	rootCmd.AddCommand(
		cmdVersion,
		cmdRun,
	)
	rootCmd.SuggestionsMinimumDistance = 1
	_ = rootCmd.Execute()
}
