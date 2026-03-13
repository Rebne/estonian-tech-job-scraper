package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Rebne/scrapy_project_v2/internal/sqlc/jobs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	repo Repository
}

func NewApp() App {
	config := NewConfig()
	return App{
		repo: NewRepository(config),
	}
}

type Repository struct {
	jobs *jobs.Queries
}

type Config struct {
	database_url string
}

func NewConfig() Config {
	var result Config
	if result.database_url = os.Getenv("DATABASE_URL"); result.database_url == "" {
		log.Fatal("DATABASE_URL is unset")
	}
	return result
}

func NewPostgres(url string) *pgxpool.Pool {

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("database unreachable:", err)
	}

	log.Println("PostgreSQL connected")

	return pool
}

func NewRepository(config Config) Repository {
	pool := NewPostgres(config.database_url)
	return Repository{
		jobs: jobs.New(pool),
	}
}

func main() {
	app := NewApp()
	ctx := context.Background()
	jobs, err := app.repo.jobs.ListAllJobs(ctx)
	if err != nil {
		message := fmt.Sprint("Could not list all jobs: %w", err)
		log.Fatal(message)
	}
	fmt.Println(jobs)
}
