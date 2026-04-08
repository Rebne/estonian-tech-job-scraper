package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/repository/sqlc/jobs"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources"

	apisources "github.com/Rebne/scrapy_project_v2/internal/scrape/sources/api"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobformatter"
	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runner interface {
	Run(context.Context) error
	Close() error
}

type runner struct {
	scrapers   []scrape.Scraper
	db         *pgxpool.Pool
	repo       *jobs.Queries
	formatter  jobformatter.JobFormatter
	notifier   notifier.Notifier
	retrievers []fetcher.HTMLRetriever
	scrapeFunc scrapeFunc
	options    runnerOptions
}

type runnerOptions struct {
	devMode bool
}

type scrapeFunc func(context.Context, []scrape.Scraper) ([]domain.Job, error)

func NewRunner(config Config) (Runner, error) {
	var runner runner

	if config.Mode.IsTest() || config.Mode.IsProd() {
		db, err := newDB(config.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		runner.db = db
		runner.repo = jobs.New(db)
	}

	if config.Mode.IsProd() {
		runner.notifier = notifier.NewTelegramNotifier(config.TelegramBotToken, config.TelegramChatID)
	} else {
		runner.notifier = notifier.NewStdOutNotifier()
	}

	if config.Async {
		runner.scrapeFunc = scrapeAsync
	} else {
		runner.scrapeFunc = scrapeSync
	}

	runner.formatter = jobformatter.NewTelegramHTMLFormatter()

	if config.ProxyURL == "" {
		log.Printf("warning: %s is not set, continuing without proxy", ProxyURL)
	}

	httpRetriever, err := fetcher.NewHTTPFetcher(fetcher.HTTPFetcherOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http fetcher: %w", err)
	}

	chromeRetriever, err := fetcher.NewChromeFetcher(fetcher.ChromeFetcherOptions{
		ExecutablePath: config.ChromeExecutablePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chrome fetcher: %w", err)
	}

	playwrightRetriever, err := fetcher.NewPlaywrightFetcher(fetcher.PlaywrightFetcherOptions{
		ProxyURL: config.ProxyURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize playwright fetcher: %w", err)
	}

	runner.retrievers = []fetcher.HTMLRetriever{httpRetriever, chromeRetriever, playwrightRetriever}
	runner.addScraper(sources.NewCgiScraper(chromeRetriever))
	runner.addScraper(sources.NewCodeborneScraper(httpRetriever))
	runner.addScraper(sources.NewConfidoScraper(httpRetriever))
	runner.addScraper(sources.NewFoxwayScraper(httpRetriever))
	runner.addScraper(sources.NewGotoAndPlayScraper(httpRetriever))
	runner.addScraper(sources.NewHelmesScraper(httpRetriever))
	runner.addScraper(sources.NewPipedriveScraper(httpRetriever))
	runner.addScraper(sources.NewPlaytechScraper(httpRetriever))
	runner.addScraper(sources.NewProekspertScraper(httpRetriever))
	runner.addScraper(sources.NewSebScraper(httpRetriever))
	runner.addScraper(sources.NewSwedbankScraper(httpRetriever))
	runner.addScraper(sources.NewTaltechScraper(httpRetriever))
	runner.addScraper(sources.NewVkgScraper(httpRetriever))
	runner.addScraper(sources.NewWiseScraper(httpRetriever))
	runner.addScraper(sources.NewBoltScraper(playwrightRetriever))
	runner.addScraper(apisources.NewCveeScraper(httpRetriever))
	runner.addScraper(apisources.NewGliaScraper(httpRetriever))
	runner.addScraper(apisources.NewMicrosoftScraper(httpRetriever))
	runner.addScraper(apisources.NewStebbyScraper(httpRetriever))

	runner.options.devMode = config.Mode.IsDev()

	return &runner, nil
}

func (r *runner) Close() error {
	closeErrors := make([]error, 0)
	for _, retriever := range r.retrievers {
		if retriever == nil {
			continue
		}

		if err := retriever.Close(); err != nil {
			closeErrors = append(closeErrors, err)
		}
	}

	return errors.Join(closeErrors...)
}

func (r *runner) addScraper(scraper scrape.Scraper) {
	r.scrapers = append(r.scrapers, scraper)
}

func (r *runner) Run(ctx context.Context) error {
	scrapedJobs, scrapeErr := r.scrapeFunc(ctx, r.scrapers)
	var runErr error
	// in devmode notify all jobs, no persistence
	if r.options.devMode {
		var messages []string
		for _, job := range scrapedJobs {
			messages = append(messages, r.formatter.MustFormatJob(job))
		}
		runErr = r.notifier.Notify(messages)
	} else {
		runErr = r.persistAndNotify(ctx, scrapedJobs)
	}

	return errors.Join(scrapeErr, runErr)
}

func scrapeSync(ctx context.Context, scrapers []scrape.Scraper) ([]domain.Job, error) {
	result := make([]domain.Job, 0)
	scrapeErrors := make([]error, 0)
	for _, scraper := range scrapers {
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

func scrapeAsync(ctx context.Context, scrapers []scrape.Scraper) ([]domain.Job, error) {
	result := make([]domain.Job, 0)
	var mu sync.Mutex
	scrapeErrors := make([]error, 0)

	var wg sync.WaitGroup
	for _, scraper := range scrapers {
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

func (r *runner) persistAndNotify(ctx context.Context, scrapedJobs []domain.Job) error {
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
