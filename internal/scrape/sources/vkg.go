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

const vkgURL string = "https://www.cvkeskus.ee/viru-keemia-grupp-as-toopakkumised-3950"

type vkgScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewVkgScraper(retriever fetcher.HTMLRetriever) *vkgScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.TitleIncludeFilter{}).
		Add(jobfilter.TitleExcludeFilter{})

	return &vkgScraper{
		url:       vkgURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (vs *vkgScraper) Name() string {
	return "vkg"
}

func (vs *vkgScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := vs.retriever.Fetch(ctx, vs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve VKG html: %w", err)
	}

	jobs, err := vs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VKG jobs: %w", err)
	}

	return shared.FilterJobs(jobs, vs.filters), nil
}

func (vs *vkgScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	container := doc.Find("#company-jobs").First()
	if container.Length() == 0 {
		return nil, errors.New("vkg document missing job container")
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
			WithPage(vs.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			WithURL(vkgURL).
			Build(),
		)
	})

	return result, nil
}
