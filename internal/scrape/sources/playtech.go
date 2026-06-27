package sources

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/errors"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources/shared"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const playtechURL string = "https://www.playtechpeople.com/jobs-our/"

var playtechIgnoreDepartmentRegex = regexp.MustCompile(`(?i)sales|human resources|finance|live operations|administrative services|corporate support|distribution|g&a - other|general management/executive|legal|manufacturing|marketing|regulatory affairs|retail operations|sales & marketing - other|sales support|services - other|training/education|transportation|web/e commerce`)

type playtechScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func NewPlaytechScraper(retriever fetcher.HTMLRetriever) *playtechScraper {
	filterChain := jobfilter.NewJobFilterChain().Add(jobfilter.TitleExcludeFilter{})

	return &playtechScraper{
		url:       playtechURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (ps *playtechScraper) Name() string {
	return "playtech"
}

func (ps *playtechScraper) GetJobs(ctx context.Context) (scrape.ScrapeResult, error) {
	html, err := ps.retriever.Fetch(ctx, ps.url)
	if err != nil {
		return scrape.ScrapeResult{}, fmt.Errorf("failed to retrieve Playtech html: %w", err)
	}

	jobs, err := ps.parseJobs(html)
	if err != nil {
		return scrape.ScrapeResult{}, fmt.Errorf("failed to parse Playtech jobs: %w", err)
	}

	return scrape.ScrapeResult{Source: ps.Name(), Jobs: shared.FilterJobs(jobs, ps.filters), Status: scrape.ScrapeStatusSuccess}, nil
}

func (ps *playtechScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	jobs := doc.Find("a.job-item")
	if jobs.Length() == 0 {
		return nil, errors.ErrNoJobsFound
	}

	locationRegex := regexp.MustCompile(`(?i)Estonia`)
	seen := make(map[string]bool)
	result := make([]domain.Job, 0)

	jobs.Each(func(_ int, job *goquery.Selection) {
		title := strings.TrimSpace(job.Find("h6").First().Text())
		location := strings.TrimSpace(job.Find("p.location-link").First().Text())
		category := strings.TrimSpace(job.Find("p.cat-link").First().Text())
		if title == "" || !locationRegex.MatchString(location) {
			return
		}
		if playtechIgnoreDepartmentRegex.MatchString(category) {
			return
		}
		if seen[title] {
			return
		}

		seen[title] = true
		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ps.Name()).
			WithLocation(location).
			WithCategory(category).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			WithURL(playtechURL).
			Build(),
		)
	})

	return result, nil
}
