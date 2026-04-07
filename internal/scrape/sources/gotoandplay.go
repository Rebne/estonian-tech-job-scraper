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

const gotoAndPlayURL string = "https://play.ee/jobs/"

type gotoAndPlayScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
}

func NewGotoAndPlayScraper(retriever fetcher.HTMLRetriever) *gotoAndPlayScraper {
	return &gotoAndPlayScraper{
		url:       gotoAndPlayURL,
		retriever: retriever,
	}
}

func (gs *gotoAndPlayScraper) Name() string {
	return "gotoandplay"
}

func (gs *gotoAndPlayScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := gs.retriever.Fetch(ctx, gs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve GotoAndPlay html: %w", err)
	}

	jobs, err := gs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GotoAndPlay jobs: %w", err)
	}

	return jobs, nil
}

func (gs *gotoAndPlayScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find("div.card__content")
	if jobs.Length() == 0 {
		return nil, errors.New("gotoandplay document missing jobs")
	}

	result := make([]domain.Job, 0)
	jobs.Each(func(_ int, job *goquery.Selection) {
		title := strings.TrimSpace(job.Find("h2").First().Text())
		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(gs.Name()).
			WithLocation("Tartu, Estonia").
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			WithURL(gotoAndPlayURL).
			Build(),
		)
	})

	return result, nil
}
