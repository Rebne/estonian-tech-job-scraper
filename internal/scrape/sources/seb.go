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

const sebURL string = "https://jobs.eu.lever.co/seb?location=Tallinn&commitment=Data%20%26%20IT"

type sebScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
}

func NewSebScraper(retriever fetcher.HTMLRetriever) *sebScraper {
	return &sebScraper{
		url:       sebURL,
		retriever: retriever,
	}
}

func (ss *sebScraper) Name() string {
	return "seb"
}

func (ss *sebScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := ss.retriever.Fetch(ctx, ss.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve SEB html: %w", err)
	}

	jobs, err := ss.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SEB jobs: %w", err)
	}

	return jobs, nil
}

func (ss *sebScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find("div.posting")
	if jobs.Length() == 0 {
		return nil, errors.New("seb document missing jobs")
	}

	result := make([]domain.Job, 0)
	jobs.Each(func(_ int, job *goquery.Selection) {
		locationParts := make([]string, 0)
		job.Find("div span").Each(func(_ int, span *goquery.Selection) {
			part := strings.TrimSpace(span.Text())
			if part != "" {
				locationParts = append(locationParts, part)
			}
		})

		location := strings.Join(locationParts, " ")
		title := strings.TrimSpace(job.Find("h5").First().Text())
		if split := strings.SplitN(title, " | ", 2); len(split) > 0 {
			title = strings.TrimSpace(split[0])
		}

		if title == "" {
			return
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ss.Name()).
			WithLocation(location).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			WithURL(sebURL).
			Build(),
		)
	})

	return result, nil
}
