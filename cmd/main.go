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

	app.NewRunner(config).Run(context.Background())
}
