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

const wiseURL string = "https://wise.jobs/jobs?options=33,300&page=1&size=50"

type wiseScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewWiseScraper(retriever fetcher.HTMLRetriever) *wiseScraper {
	filterChain := jobfilter.NewJobFilterChain().Add(jobfilter.TitleExcludeFilter{})

	return &wiseScraper{
		url:       wiseURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (ws *wiseScraper) Name() string {
	return "wise"
}

func (ws *wiseScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := ws.retriever.Fetch(ctx, ws.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Wise html: %w", err)
	}

	jobs, err := ws.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Wise jobs: %w", err)
	}

	return filterJobs(jobs, ws.filters), nil
}

func (ws *wiseScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find(".attrax-vacancy-tile")
	if jobs.Length() == 0 {
		return nil, errors.New("wise document missing jobs")
	}

	result := make([]domain.Job, 0)
	jobs.Each(func(_ int, job *goquery.Selection) {
		title := strings.TrimSpace(job.Find(".attrax-vacancy-tile__title").First().Text())
		href, _ := job.Find(".attrax-vacancy-tile__learn-more").First().Attr("href")
		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ws.Name()).
			WithLocation("Tallinn").
			WithURL("https://wise.jobs"+strings.TrimSpace(href)).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	})

	return result, nil
}
