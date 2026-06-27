package scrape

import (
	"context"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
)

type Scraper interface {
	Name() string
	GetJobs(context.Context) ([]domain.Job, error)
type ScrapeStatus string

const (
	ScrapeStatusSuccess ScrapeStatus = "Success"
	ScrapeStatusFailed  ScrapeStatus = "Failed"
)

type ScrapeResult struct {
	Source string
	Jobs   []domain.Job
	Status ScrapeStatus
}
