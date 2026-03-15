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
	"github.com/Rebne/scrapy_project_v2/internal/services/messageformatter"
	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
)

type Runner struct {
	scrapers  []scrape.Scraper
	repo      *jobs.Queries
	formatter messageformatter.JobFormatter
	notifier  notifier.Notifier
}

func NewRunner(config Config, db jobs.DBTX) *Runner {
	return &Runner{
		scrapers: []scrape.Scraper{
			sources.NewCgiScraper(),
		},
		repo:      jobs.New(db),
		formatter: messageformatter.NewTelegramHTMLFormatter(),
		notifier:  notifier.NewTelegramNotifier(config.TelegramBotToken, config.TelegramChatID),
	}
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
		wg.Add(1)
		go func() {
			defer wg.Done()

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
		}()
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

	existingKeys := make(map[string]struct{}, len(existingJobs))
	for _, existingJob := range existingJobs {
		existingKeys[string(existingJob.JobHash)] = struct{}{}
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

		existingKeys[key] = struct{}{}
		messages = append(messages, r.formatter.MustFormatJob(scrapedJob))
	}

	if len(messages) == 0 {
		return nil
	}

	if err := r.notifier.Notify(messages); err != nil {
		return fmt.Errorf("sending notifications failed: %w", err)
	}

	return nil
}
