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
)

const foxwayURL string = "https://jobs.foxway.com/jobs"

type foxwayScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
}

func NewFoxwayScraper(retriever fetcher.HTMLRetriever) *foxwayScraper {
	return &foxwayScraper{
		url:       foxwayURL,
		retriever: retriever,
	}
}

func (fs *foxwayScraper) Name() string {
	return "foxway"
}

func (fs *foxwayScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := fs.retriever.Fetch(ctx, fs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Foxway html: %w", err)
	}

	jobs, err := fs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Foxway jobs: %w", err)
	}

	return jobs, nil
}

func (fs *foxwayScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find("#jobs_list_container li")
	if jobs.Length() == 0 {
		return nil, errors.New("foxway document missing jobs")
	}

	locationRegex := regexp.MustCompile(`(?i)Estonia|Tartu`)
	result := make([]domain.Job, 0)

	jobs.Each(func(_ int, job *goquery.Selection) {
		locationParts := make([]string, 0)
		job.Find("a div div span").Each(func(_ int, span *goquery.Selection) {
			value := strings.TrimSpace(span.Text())
			if value != "" {
				locationParts = append(locationParts, value)
			}
		})

		location := strings.Join(locationParts, " ")
		if !locationRegex.MatchString(location) {
			return
		}

		title := strings.TrimSpace(job.Find("a div span.company-link-style").First().Text())
		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(fs.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	})

	return result, nil
}
