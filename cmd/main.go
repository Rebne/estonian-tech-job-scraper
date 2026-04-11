package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	"github.com/Rebne/scrapy_project_v2/internal/app"
	"github.com/Rebne/scrapy_project_v2/pkg/logger"
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

	bufLogger := logger.NewBufferedLogger(slog.LevelError)

	ctx := logger.ContextWithLogger(context.Background(), bufLogger.Logger)

	err = runner.Run(ctx)
	if err != nil {
		bufLogger.Error("runner failed", "err", err)
	}
}
