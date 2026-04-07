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
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources/shared"
)

const helmesURL string = "https://www.helmes.com/career/"

type helmesScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
}

func NewHelmesScraper(retriever fetcher.HTMLRetriever) *helmesScraper {
	return &helmesScraper{
		url:       helmesURL,
		retriever: retriever,
	}
}

func (hs *helmesScraper) Name() string {
	return "helmes"
}

func (hs *helmesScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := hs.retriever.Fetch(ctx, hs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Helmes html: %w", err)
	}

	jobs, err := hs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Helmes jobs: %w", err)
	}

	return jobs, nil
}

func (hs *helmesScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	containers := doc.Find(".job-offers__city")
	if containers.Length() == 0 {
		return nil, errors.New("helmes document missing job city containers")
	}

	locationRegex := regexp.MustCompile(`(?i)Estonia|Tartu`)
	var targetContainer *goquery.Selection
	location := ""

	containers.EachWithBreak(func(_ int, container *goquery.Selection) bool {
		header := container.Find("h4").First()
		candidateLocation := strings.TrimSpace(header.Text())
		if candidateLocation == "" || !locationRegex.MatchString(candidateLocation) {
			return true
		}

		targetContainer = container
		location = candidateLocation
		return false
	})

	if targetContainer == nil {
		return nil, errors.New("helmes estonian job container not found")
	}

	jobLinks := targetContainer.Find("li a")
	if jobLinks.Length() == 0 {
		return nil, shared.ErrNoJobsFound
	}

	result := make([]domain.Job, 0)
	jobLinks.Each(func(_ int, jobLink *goquery.Selection) {
		title := strings.TrimSpace(jobLink.Text())
		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(hs.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	})

	if len(result) == 0 {
		return nil, shared.ErrNoJobsFound
	}

	return result, nil
}
