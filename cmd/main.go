package main

import (
	"context"
	"flag"
	"log"

	"github.com/Rebne/scrapy_project_v2/internal/app"
)

func main() {
	async := flag.Bool("async", false, "run scrapers concurrently")
	flag.Parse()

	config, err := app.BuildConfig(*async)
	if err != nil {
		log.Fatal("building config failed: ", err)
	}

	runner, err := app.NewRunner(config)
	if err != nil {
		log.Fatal("initializing runner failed: ", err)
	}
	defer runner.Close()

	err = runner.Run(context.Background())
	if err != nil {
		log.Printf("errors: %v", err)
	}
}
