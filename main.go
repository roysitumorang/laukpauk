package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/roysitumorang/laukpauk/router"
	"golang.org/x/sync/errgroup"
)

func main() {
	var g errgroup.Group
	g.Go(func() error {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatalln("can't load .env file")
		}
		service := router.MakeHandler()
		return service.HTTPServerMain()
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
