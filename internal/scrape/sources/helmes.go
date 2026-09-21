package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	internalerrors "github.com/Rebne/scrapy_project_v2/internal/errors"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources/shared"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const helmesURL string = "https://www.helmes.com/career/"

type helmesScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewHelmesScraper(retriever fetcher.HTMLRetriever) *helmesScraper {
	return &helmesScraper{
		url:       helmesURL,
		retriever: retriever,
		filters: jobfilter.NewJobFilterChain().
			Add(jobfilter.LocationEstoniaFilter{}),
	}
}

func (hs *helmesScraper) Name() string {
	return "helmes"
}

func (hs *helmesScraper) GetJobs(ctx context.Context) (scrape.ScrapeResult, error) {
	html, err := hs.retriever.Fetch(ctx, hs.url)
	if err != nil {
		return scrape.ScrapeResult{}, fmt.Errorf("failed to retrieve Helmes html: %w", err)
	}

	jobs, err := hs.parseJobs(html)
	if err != nil {
		return scrape.ScrapeResult{}, fmt.Errorf("failed to parse Helmes jobs: %w", err)
	}

	return scrape.ScrapeResult{Source: hs.Name(), Jobs: shared.FilterJobs(jobs, hs.filters), Status: scrape.ScrapeStatusSuccess}, nil
}

func (hs *helmesScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	listing := doc.Find("section.helmes-jobs-td")
	if listing.Length() == 0 {
		return nil, errors.New("helmes document missing jobs listing")
	}

	jobRows := listing.Find("a.hj-row")
	if jobRows.Length() == 0 {
		return nil, errors.New("helmes jobs listing missing job rows")
	}

	result := make([]domain.Job, 0)
	jobRows.Each(func(_ int, jobRow *goquery.Selection) {
		title := strings.TrimSpace(jobRow.Find(".hj-role").First().Text())
		if title == "" {
			return
		}
		location := strings.TrimSpace(jobRow.Find(".hj-loc").First().Text())
		url := strings.TrimSpace(jobRow.AttrOr("href", ""))
		if url == "" {
			url = helmesURL
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(hs.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			WithURL(url).
			Build(),
		)
	})

	if len(result) == 0 {
		return nil, internalerrors.ErrNoJobsFound
	}

	return result, nil
}
