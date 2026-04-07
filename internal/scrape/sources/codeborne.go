package sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources/shared"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const codeborneURL string = "https://codeborne.com/en/jobs/"

type codeborneScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewCodeborneScraper(retriever fetcher.HTMLRetriever) *codeborneScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.LocationEstoniaFilter{}).
		Add(jobfilter.TitleExcludeFilter{})
	return &codeborneScraper{
		url:       codeborneURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (cs *codeborneScraper) Name() string {
	return "codeborne"
}

func (cs *codeborneScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := cs.retriever.Fetch(ctx, cs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Codeborne html: %w", err)
	}

	jobs, err := cs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Codeborne jobs: %w", err)
	}

	return shared.FilterJobs(jobs, cs.filters), nil
}

func (cs *codeborneScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	result := make([]domain.Job, 0)
	jobs := doc.Find("div.job-card")

	jobs.Each(func(i int, job *goquery.Selection) {
		title := strings.TrimSpace(job.Find("div.font-bold").First().Text())
		location := strings.TrimSpace(job.Find("div div.text-lg").First().Text())
		level := strings.TrimSpace(job.Find("p.text-lg").First().Text())

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(cs.Name()).
			WithLocation(location).
			WithEmploymentType(level).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			WithURL(codeborneURL).
			Build(),
		)
	})

	return result, nil
}
