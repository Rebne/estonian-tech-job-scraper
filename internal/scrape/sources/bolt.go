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

const boltURL string = "https://bolt.eu/en/careers/positions/?location%5B0%5D=Estonia-J%C3%B5hvi&location%5B1%5D=Estonia-Tallinn&location%5B2%5D=Estonia-Tartu&location%5B3%5D=Estonia&teams%5B0%5D=Engineering"

type boltScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
}

func NewBoltScraper(retriever fetcher.HTMLRetriever) *boltScraper {
	return &boltScraper{
		url:       boltURL,
		retriever: retriever,
	}
}

func (bs *boltScraper) Name() string {
	return "bolt"
}

func (bs *boltScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := bs.retriever.Fetch(ctx, bs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Bolt html: %w", err)
	}

	jobs, err := bs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Bolt jobs: %w", err)
	}

	return jobs, nil
}

func (bs *boltScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find(`div[class*="Accordion_AccordionItem"]`)
	if jobs.Length() == 0 {
		return nil, errors.New("bolt document missing jobs")
	}

	result := make([]domain.Job, 0)
	jobs.Each(func(_ int, job *goquery.Selection) {
		title := strings.TrimSpace(job.Find(".rt-Text.rt-r-size-2.rt-r-weight-medium").First().Text())
		location := strings.TrimSpace(job.Find(".rt-Text.rt-r-size-2.rt-r-weight-regular.rt-truncate").First().Text())
		if title == "" || location == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(bs.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			WithURL(boltURL).
			Build(),
		)
	})

	return result, nil
}
