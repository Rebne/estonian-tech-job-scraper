package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const proekspertURL string = "https://proekspert.com/join-us/"

type proekspertScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewProekspertScraper(retriever fetcher.HTMLRetriever) *proekspertScraper {
	filterChain := jobfilter.NewJobFilterChain().Add(jobfilter.TitleExcludeFilter{})

	return &proekspertScraper{
		url:       proekspertURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (ps *proekspertScraper) Name() string {
	return "proekspert"
}

func (ps *proekspertScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := ps.retriever.Fetch(ctx, ps.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Proekspert html: %w", err)
	}

	jobs, err := ps.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Proekspert jobs: %w", err)
	}

	return filterJobs(jobs, ps.filters), nil
}

func (ps *proekspertScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find(".job-positions__items a")
	if jobs.Length() == 0 {
		return nil, errors.New("proekspert document missing jobs")
	}

	result := make([]domain.Job, 0)
	jobs.Each(func(_ int, job *goquery.Selection) {
		title := strings.TrimSpace(job.Text())
		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ps.Name()).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	})

	return result, nil
}
