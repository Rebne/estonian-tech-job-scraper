package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
)

const taltechURL string = "https://career.taltech.ee/pakkumised/?sectors=infotehnoloogia"

type taltechScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
}

func NewTaltechScraper(retriever fetcher.HTMLRetriever) *taltechScraper {
	return &taltechScraper{
		url:       taltechURL,
		retriever: retriever,
	}
}

func (ts *taltechScraper) Name() string {
	return "taltech"
}

func (ts *taltechScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := ts.retriever.Fetch(ctx, ts.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve TalTech html: %w", err)
	}

	jobs, err := ts.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TalTech jobs: %w", err)
	}

	return jobs, nil
}

func (ts *taltechScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	rows := doc.Find("tr")
	if rows.Length() == 0 {
		return nil, errors.New("taltech document missing rows")
	}

	rows = rows.Slice(1, goquery.ToEnd)
	if rows.Length() == 0 {
		return nil, errors.New("taltech document missing job rows")
	}

	result := make([]domain.Job, 0)
	rows.Each(func(_ int, row *goquery.Selection) {
		title := strings.TrimSpace(row.Find(".title").First().Text())
		company := strings.TrimSpace(row.Find(".name").First().Text())
		category := strings.TrimSpace(row.Find(".meta").First().Text())
		url, _ := row.Find(".title").First().Attr("href")

		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ts.Name()).
			WithCompany(company).
			WithCategory(category).
			WithURL(strings.TrimSpace(url)).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	})

	return result, nil
}
