package sources

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const swedbankURL string = "https://jobs.swedbank.com/jobs?department_id=57"

type swedbankScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewSwedbankScraper(retriever fetcher.HTMLRetriever) *swedbankScraper {
	filterChain := jobfilter.NewJobFilterChain().Add(jobfilter.TitleExcludeFilter{})

	return &swedbankScraper{
		url:       swedbankURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (ss *swedbankScraper) Name() string {
	return "swedbank"
}

func (ss *swedbankScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := ss.retriever.Fetch(ctx, ss.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Swedbank html: %w", err)
	}

	jobs, err := ss.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Swedbank jobs: %w", err)
	}

	return filterJobs(jobs, ss.filters), nil
}

func (ss *swedbankScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find("ul#jobs_list_container li")
	if jobs.Length() == 0 {
		return nil, errors.New("swedbank document missing jobs")
	}

	locationRegex := regexp.MustCompile(`(?i)Tallinn|Tartu|Multiple`)
	result := make([]domain.Job, 0)

	jobs.Each(func(_ int, job *goquery.Selection) {
		location := ""
		locationSpan := job.Find("span.text-base span").First()
		if locationSpan.Length() > 0 {
			if title, ok := locationSpan.Attr("title"); ok {
				location = strings.TrimSpace(title)
			}
		}

		if location == "" {
			locationSpans := job.Find("span.text-base span")
			if locationSpans.Length() > 2 {
				location = strings.TrimSpace(locationSpans.Eq(2).Text())
			}
		}

		title := ""
		titleSpan := job.Find("span.text-block-base-link").First()
		if titleSpan.Length() > 0 {
			if titleAttr, ok := titleSpan.Attr("title"); ok {
				title = strings.TrimSpace(titleAttr)
			}
		}

		if location == "" || title == "" || !locationRegex.MatchString(location) {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ss.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	})

	return result, nil
}
