package app

import (
	"context"

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
	return nil
}
