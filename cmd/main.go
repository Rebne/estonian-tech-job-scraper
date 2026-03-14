package main

import (
	"context"
	"log"

	"github.com/Rebne/scrapy_project_v2/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := app.BuildConfig()
	if err != nil {
		log.Fatal("building config failed: ", err)
	}
	pool, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("database unreachable:", err)
	}

	log.Println("PostgreSQL connected")

	app.NewRunner(config, pool).Run(context.Background())
}
