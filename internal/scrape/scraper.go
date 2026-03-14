package scrape

import (
	"github.com/Rebne/scrapy_project_v2/internal/domain"
)

type Scraper interface {
	GetJobs() ([]domain.Job, error)
}
