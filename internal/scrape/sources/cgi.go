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
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const URL string = "https://cgi.njoyn.com/corp/xweb/xweb.asp?CLID=21001&page=joblisting&CountryID=EE"

type cgiScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

func (cs *cgiScraper) Name() string {
	return "cgi"
}

func NewCgiScraper(retriever fetcher.HTMLRetriever) *cgiScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.LocationEstoniaFilter{}).
		Add(jobfilter.TitleIncludeFilter{}).
		Add(jobfilter.TitleExcludeFilter{})
	return &cgiScraper{
		url:       URL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (cs *cgiScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	html, err := cs.retriever.Fetch(ctx, cs.url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve CGI html: %w", err)
	}
	jobs, err := cs.parseJobs(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CGI jobs: %w", err)
	}
	return shared.FilterJobs(jobs, cs.filters), nil
}

func (cs *cgiScraper) parseJobs(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var result []domain.Job

	rows := doc.Find("tr")
	if rows.Length() == 0 {
		return nil, errors.New("cgi document missing table rows")
	}

	// Select all table rows except the header
	rows = rows.Slice(1, goquery.ToEnd)
	if rows.Length() == 0 {
		return nil, shared.ErrNoJobsFound
	}

	categoryRegex := regexp.MustCompile(`(?i)Software Development`)

	rows.Each(func(i int, row *goquery.Selection) {
		cols := row.Find("td")

		if cols.Length() < 5 {
			return
		}

		title := strings.TrimSpace(cols.Eq(1).Text())
		category := strings.TrimSpace(cols.Eq(2).Text())
		location := strings.TrimSpace(
			cols.Eq(3).Text() + " " + cols.Eq(4).Text(),
		)

		if categoryRegex.MatchString(category) {
			job := domain.
				NewJobBuilder().
				WithTitle(title).
				WithPage(cs.Name()).
				WithLocation(location).
				WithCategory(category).
				WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
				Build()
			result = append(result, job)
		}
	})

	return result, nil
}
