package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/roysitumorang/laukpauk/router"
	"github.com/roysitumorang/laukpauk/services/messagingconsumer"
	"golang.org/x/sync/errgroup"
)

func main() {
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
}
