package api

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const cveeURL string = "https://cv.ee/api/v1/vacancy-search-service/search?limit=250&categories[]=INFORMATION_TECHNOLOGY"

var mediatedOfferRegex = regexp.MustCompile(`(?i)vahendatud pakkumised`)

type cveeScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters jobfilter.JobFilterChain
}

type cveeSearchResponse struct {
	Vacancies []cveeVacancy `json:"vacancies"`
}

type cveeVacancy struct {
	ID           int64  `json:"id"`
	Position     string `json:"positionTitle"`
	EmployerName string `json:"employerName"`
}

func NewCveeScraper(retriever fetcher.HTMLRetriever) *cveeScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.TitleExcludeFilter{}).
		Add(jobfilter.TitleIncludeFilter{})
	return &cveeScraper{
		url: cveeURL,
		retriever: retriever,
		filters: filterChain,
	}
}

func (cs *cveeScraper) Name() string {
	return "cvee"
}

func (cs *cveeScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	var payload cveeSearchResponse
	if err := fetchJSON(ctx, cs.retriever, cs.url, &payload); err != nil {
		return nil, fmt.Errorf("failed to retrieve CVEE jobs: %w", err)
	}

	result := make([]domain.Job, 0)
	for _, vacancy := range payload.Vacancies {
		title := strings.TrimSpace(vacancy.Position)
		employer := strings.TrimSpace(vacancy.EmployerName)
		if title == "" || vacancy.ID == 0 {
			continue
		}

		if mediatedOfferRegex.MatchString(employer) {
			continue
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(cs.Name()).
			WithCompany(employer).
			WithURL(fmt.Sprintf("https://cv.ee/et/vacancy/%d", vacancy.ID)).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	}

	return filterJobs(result, cs.filters), nil
}
