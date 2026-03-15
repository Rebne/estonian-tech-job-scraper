package main

import (
	"context"
	"log"

	"github.com/Rebne/scrapy_project_v2/internal/app"
)

func main() {
	config, err := app.BuildConfig()
	if err != nil {
		log.Fatal("building config failed: ", err)
	}

	runner, err := app.NewRunner(config)
	if err != nil {
		log.Fatal("initializing runner failed: ", err)
	}

	runner.Run(context.Background())
}
