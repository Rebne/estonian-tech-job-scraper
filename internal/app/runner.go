package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/repository/sqlc/jobs"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobformatter"
	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runner struct {
	scrapers  []scrape.Scraper
	db        *pgxpool.Pool
	repo      *jobs.Queries
	formatter jobformatter.JobFormatter
	notifier  notifier.Notifier
}

func NewRunner(config Config) (*Runner, error) {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return &Runner{
		scrapers: []scrape.Scraper{
			sources.NewCgiScraper(),
		},
		db:        db,
		repo:      jobs.New(db),
		formatter: jobformatter.NewTelegramHTMLFormatter(),
		notifier:  notifier.NewTelegramNotifier(config.TelegramBotToken, config.TelegramChatID),
	}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	return r.RunSync(ctx)
}

func (r *Runner) RunSync(ctx context.Context) error {
	scrapedJobs, scrapeErr := r.scrapeSync(ctx)
	persistErr := r.persistAndNotify(ctx, scrapedJobs)

	return errors.Join(scrapeErr, persistErr)
}

func (r *Runner) RunAsync(ctx context.Context) error {
	scrapedJobs, scrapeErr := r.scrapeAsync(ctx)
	persistErr := r.persistAndNotify(ctx, scrapedJobs)

	return errors.Join(scrapeErr, persistErr)
}

func (r *Runner) scrapeSync(ctx context.Context) ([]domain.Job, error) {
	result := make([]domain.Job, 0)
	scrapeErrors := make([]error, 0)
	for _, scraper := range r.scrapers {
		jobs, err := scraper.GetJobs(ctx)
		if err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("scraper %q failed: %w", scraper.Name(), err))
			continue
		}
		result = append(result, jobs...)
	}

	if len(scrapeErrors) > 0 {
		return result, errors.Join(scrapeErrors...)
	}

	return result, nil
}

func (r *Runner) scrapeAsync(ctx context.Context) ([]domain.Job, error) {
	result := make([]domain.Job, 0)
	var mu sync.Mutex
	scrapeErrors := make([]error, 0)

	var wg sync.WaitGroup
	for _, scraper := range r.scrapers {
		scraper := scraper
		wg.Go(func() {
			scrapedJobs, err := scraper.GetJobs(ctx)
			if err != nil {
				mu.Lock()
				scrapeErrors = append(scrapeErrors, fmt.Errorf("scraper %q failed: %w", scraper.Name(), err))
				mu.Unlock()
				return
			}

			mu.Lock()
			result = append(result, scrapedJobs...)
			mu.Unlock()
		})
	}
	wg.Wait()

	if len(scrapeErrors) > 0 {
		return result, errors.Join(scrapeErrors...)
	}

	return result, nil
}

func (r *Runner) persistAndNotify(ctx context.Context, scrapedJobs []domain.Job) error {
	if len(scrapedJobs) == 0 {
		return nil
	}

	existingJobs, err := r.repo.GetAllJobs(ctx)
	if err != nil {
		return fmt.Errorf("loading existing jobs failed: %w", err)
	}

	existingKeys := make(map[string]bool, len(existingJobs))
	for _, existingJob := range existingJobs {
		existingKeys[string(existingJob.JobHash)] = true
	}

	messages := make([]string, 0)
	for _, scrapedJob := range scrapedJobs {
		key := string(scrapedJob.Hash())
		if _, exists := existingKeys[key]; exists {
			continue
		}

		if err := r.repo.InsertJob(ctx, jobs.InsertJobParams{
			JobHash: scrapedJob.Hash(),
			Page:    scrapedJob.Page(),
			Title:   scrapedJob.Title(),
		}); err != nil {
			return fmt.Errorf("inserting job failed: %w", err)
		}

		existingKeys[key] = true
		messages = append(messages, r.formatter.MustFormatJob(scrapedJob))
	}

	// Delete job if it is no longer scraped from the web
	for _, job := range existingJobs {
		if relevant := existingKeys[string(job.JobHash)]; !relevant {
			r.repo.DeleteJob(ctx, job.JobHash)
		}
	}

	if len(messages) == 0 {
		return nil
	}

	if err := r.notifier.Notify(messages); err != nil {
		return fmt.Errorf("sending notifications failed: %w", err)
	}

	return nil
}

func newDB(url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database unreachable: :%w", err)
	}
	return pool, nil
}
