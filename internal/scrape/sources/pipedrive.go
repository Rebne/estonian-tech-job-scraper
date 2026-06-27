package sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/errors"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources/shared"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const pipedriveURL string = "https://www.pipedrive.com/jobs/open-positions?location=estonia"

type pipedriveScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewPipedriveScraper(retriever fetcher.HTMLRetriever) *pipedriveScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.TitleExcludeFilter{})

	return &pipedriveScraper{
		url:       pipedriveURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (ps *pipedriveScraper) Name() string {
	return "pipedrive"
}

func (ps *pipedriveScraper) GetJobs(ctx context.Context) (scrape.ScrapeResult, error) {
	html, err := ps.retriever.Fetch(ctx, ps.url)
	if err != nil {
		return scrape.ScrapeResult{}, fmt.Errorf("failed to retrieve Pipedrive html: %w", err)
	}

	jobs, err := ps.parseJobs(html)
	if err != nil {
		return scrape.ScrapeResult{}, fmt.Errorf("failed to parse Pipedrive jobs: %w", err)
	}

	return scrape.ScrapeResult{Source: ps.Name(), Jobs: shared.FilterJobs(jobs, ps.filters), Status: scrape.ScrapeStatusSuccess}, nil
}

func (ps *pipedriveScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find("div[role='list'] > div[role='listitem']")
	if jobs.Length() == 0 {
		return nil, errors.ErrNoJobsFound
	}

	result := make([]domain.Job, 0, jobs.Length())
	jobs.Each(func(_ int, job *goquery.Selection) {
		title := strings.TrimSpace(job.Find("h3").First().Text())
		location := strings.TrimSpace(job.Find("div p").First().Text())
		href, _ := job.Find("a").First().Attr("href")
		href = strings.TrimSpace(href)

		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ps.Name()).
			WithLocation(location).
			WithURL(shared.FallbackOnEmptyString(href, pipedriveURL)).
			WithHashFrom(domain.HashFieldPage, domain.HashFieldTitle, domain.HashFieldURL).
			Build(),
		)
	})

	if len(result) == 0 {
		return nil, errors.ErrNoJobsFound
	}

	return result, nil
}
