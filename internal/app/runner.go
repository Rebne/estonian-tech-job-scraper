package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	internalerrors "github.com/Rebne/scrapy_project_v2/internal/errors"
	"github.com/Rebne/scrapy_project_v2/internal/repository/sqlc/jobs"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources"

	apisources "github.com/Rebne/scrapy_project_v2/internal/scrape/sources/api"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobformatter"
	"github.com/Rebne/scrapy_project_v2/internal/services/messenger"
	"github.com/Rebne/scrapy_project_v2/pkg/logger"
	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runner interface {
	Run(context.Context)
	Close() error
}

type runner struct {
	scrapers     []scrape.Scraper
	db           *pgxpool.Pool
	repo         *jobs.Queries
	jobMessenger *messenger.JobMessenger
	logMessenger *messenger.LogMessenger
	retrievers   []fetcher.HTMLRetriever
	scrapeFunc   scrapeFunc
	options      runnerOptions
}

type runnerOptions struct {
	devMode bool
}

type scrapeFunc func(context.Context, []scrape.Scraper) []domain.Job

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
		runner.jobMessenger = messenger.NewJobMessenger(
			notifier.NewTelegramNotifier(config.TelegramBotToken, config.TelegramChatID),
			jobformatter.NewTelegramHTMLFormatter(),
		)
		runner.logMessenger = messenger.NewLogMessenger(
			notifier.NewTelegramNotifier(
				config.TelegramBotToken, config.TelegramChatID, notifier.WithThreadID(config.TelegramLogThreadID),
			),
		)
	} else {
		runner.jobMessenger = messenger.NewJobMessenger(
			notifier.NewStdOutNotifier(),
			jobformatter.NewTelegramHTMLFormatter(),
		)
	}

	if config.Async {
		runner.scrapeFunc = scrapeAsync
	} else {
		runner.scrapeFunc = scrapeSync
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

	playwrightRetriever, err := fetcher.NewPlaywrightFetcher(fetcher.PlaywrightFetcherOptions{})
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

func (r *runner) Run(ctx context.Context) {
	bufLog := logger.NewBufferedLogger(slog.LevelInfo)

	ctx = logger.ContextWithLogger(context.Background(), bufLog.Logger)

	bufLog.Info("running scraper")
	scrapedJobs := r.scrapeFunc(ctx, r.scrapers)
	// in devmode notify all jobs, no persistence
	if r.options.devMode {
		err := r.jobMessenger.Send(scrapedJobs)
		if err != nil {
			panic(err)
		}
	}

	err := r.persistAndNotify(ctx, scrapedJobs)
	if err != nil {
		bufLog.Error("runner failed", "err", err)
	}
	bufLog.Info("scraper finished")
	if r.logMessenger != nil {
		err := r.logMessenger.Send(bufLog.Read())
		if err != nil {
			bufLog.Error("log messenger failed to save", "err", err)
		}
	}
}

func scrapeSync(ctx context.Context, scrapers []scrape.Scraper) []domain.Job {
	slog := logger.Logger(ctx)
	result := make([]domain.Job, 0)
	for _, scraper := range scrapers {
		jobs, err := scraper.GetJobs(ctx)
		if err != nil {
			if errors.Is(err, internalerrors.ErrNoJobsFound) {
				slog.Warn("no jobs found", "source", scraper.Name())
				continue
			}
			slog.Error(fmt.Sprintf("scraping source %s failed", scraper.Name()), "err", err)
			continue
		}
		result = append(result, jobs...)
	}

	return result
}

func scrapeAsync(ctx context.Context, scrapers []scrape.Scraper) []domain.Job {
	result := make([]domain.Job, 0)
	var mu sync.Mutex
	slog := logger.Logger(ctx)

	var wg sync.WaitGroup
	for _, scraper := range scrapers {
		scraper := scraper
		wg.Go(func() {
			scrapedJobs, err := scraper.GetJobs(ctx)
			if err != nil {
				if errors.Is(err, internalerrors.ErrNoJobsFound) {
					slog.Warn("no jobs found", "source", scraper.Name())
					return
				}
				slog.Error(fmt.Sprintf("scraping source %s failed", scraper.Name()), "err", err)
				return
			}

			mu.Lock()
			result = append(result, scrapedJobs...)
			mu.Unlock()
		})
	}
	wg.Wait()

	return result
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

	filteredJobs := []domain.Job{}
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
		filteredJobs = append(filteredJobs, scrapedJob)
	}

	// Delete job if it is no longer scraped from the web
	for _, job := range existingJobs {
		if relevant := existingKeys[string(job.JobHash)]; !relevant {
			r.repo.DeleteJob(ctx, job.JobHash)
		}
	}

	if len(filteredJobs) == 0 {
		return nil
	}

	if err := r.jobMessenger.Send(filteredJobs); err != nil {
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
