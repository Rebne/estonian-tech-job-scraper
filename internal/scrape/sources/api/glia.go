package api

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/sources/shared"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const gliaURL string = "https://boards-api.greenhouse.io/v1/boards/glia/departments/"

var gliaRelevantDepartments = []string{
	"Engineering",
	"Product",
	"Automation",
	"Client Engineering",
	"Data Analytics",
	"IT Team",
	"Platform",
}

type gliaScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters   jobfilter.JobFilterChain
}

type gliaDepartmentsResponse struct {
	Departments []gliaDepartment `json:"departments"`
}

type gliaDepartment struct {
	Name string    `json:"name"`
	Jobs []gliaJob `json:"jobs"`
}

type gliaJob struct {
	ID          int64 `json:"id"`
	Title       string
	AbsoluteURL string `json:"absolute_url"`
	Location    struct {
		Name string `json:"name"`
	} `json:"location"`
}

func NewGliaScraper(retriever fetcher.HTMLRetriever) *gliaScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.TitleExcludeFilter{}).
		Add(jobfilter.TitleIncludeFilter{}).
		Add(jobfilter.LocationEstoniaFilter{})

	return &gliaScraper{
		url:       gliaURL,
		retriever: retriever,
		filters:   filterChain,
	}
}

func (gs *gliaScraper) Name() string {
	return "glia"
}

func (gs *gliaScraper) GetJobs(ctx context.Context) (scrape.ScrapeResult, error) {
	var payload gliaDepartmentsResponse
	if err := shared.FetchJSON(ctx, gs.retriever, gs.url, &payload); err != nil {
		return scrape.ScrapeResult{}, fmt.Errorf("failed to retrieve Glia jobs: %w", err)
	}

	result := make([]domain.Job, 0)
	for _, department := range payload.Departments {
		if !slices.Contains(gliaRelevantDepartments, strings.TrimSpace(department.Name)) {
			continue
		}

		for _, job := range department.Jobs {
			title := strings.TrimSpace(job.Title)
			location := strings.TrimSpace(job.Location.Name)
			if title == "" || job.ID == 0 {
				continue
			}

			jobDomain := domain.
				NewJobBuilder().
				WithTitle(title).
				WithPage(gs.Name()).
				WithLocation(location).
				WithURL(strings.TrimSpace(job.AbsoluteURL)).
				WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
				Build()

			result = append(result, jobDomain)
		}
	}

	return scrape.ScrapeResult{Source: gs.Name(), Jobs: shared.FilterJobs(result, gs.filters), Status: scrape.ScrapeStatusSuccess}, nil
}
