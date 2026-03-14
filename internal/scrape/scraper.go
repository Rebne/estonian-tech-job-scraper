package scrape

import (
	"context"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
)

type Scraper interface {
	GetJobs(context.Context) ([]domain.Job, error)
}
