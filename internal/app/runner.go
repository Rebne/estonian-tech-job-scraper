package app

import (
	"context"

	"github.com/Rebne/scrapy_project_v2/internal/repository/sqlc/jobs"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
)

type Runner struct {
	scrapers []scrape.Scraper
	filters  jobfilter.JobFilterChain
	repo     jobs.Queries
	notifier notifier.Notifier
}

func (r *Runner) Run(ctx context.Context) error {
	return nil
}
