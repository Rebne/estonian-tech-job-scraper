package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources/shared"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const confidoURL string = "https://www.cvkeskus.ee/arstikeskus-confido-ou-toopakkumised-196423"

type confidoScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewConfidoScraper(retriever fetcher.HTMLRetriever) *confidoScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.TitleIncludeFilter{}).
		Add(jobfilter.TitleExcludeFilter{})

	return &confidoScraper{
		url:       confidoURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (cs *confidoScraper) Name() string {
	return "confido"
}

func (cs *confidoScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := cs.retriever.Fetch(ctx, cs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Confido html: %w", err)
	}

	jobs, err := cs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Confido jobs: %w", err)
	}

	return shared.FilterJobs(jobs, cs.filters), nil
}

func (cs *confidoScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	container := doc.Find("#company-jobs").First()
	if container.Length() == 0 {
		return nil, errors.New("confido document missing job container")
	}

	location := strings.TrimSpace(container.Find("span.location").First().Text())
	titles := container.Find("article h2")
	if titles.Length() == 0 {
		return nil, shared.ErrNoJobsFound
	}

	result := make([]domain.Job, 0)
	titles.Each(func(_ int, titleNode *goquery.Selection) {
		title := strings.TrimSpace(titleNode.Text())
		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(cs.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	})

	return result, nil
}
