package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

const microsoftURL string = "https://apply.careers.microsoft.com/api/pcsx/search?domain=microsoft.com&query=&location=Estonia,%20Harjumaa,%20Tallinn&start=0&sort_by=distance&filter_distance=160&filter_include_remote=1&filter_seniority=Intern&filter_seniority=Entry"

type microsoftScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
	filters jobfilter.JobFilterChain
}

type microsoftSearchResponse struct {
	Data struct {
		Positions []microsoftPosition `json:"positions"`
	} `json:"data"`
}

type microsoftPosition struct {
	ID                 int64    `json:"id"`
	Title              string   `json:"title"`
	Locations          []string `json:"locations"`
	PositionURL        string   `json:"positionUrl"`
	WorkLocationOption string   `json:"workLocationOption"`
}

func NewMicrosoftScraper(retriever fetcher.HTMLRetriever) *microsoftScraper {
	filterChain := jobfilter.NewJobFilterChain().
		Add(jobfilter.TitleExcludeFilter{})

	return &microsoftScraper{
		url: microsoftURL,
		retriever: retriever,
		filters: filterChain,
	}
}

func (ms *microsoftScraper) Name() string {
	return "microsoft"
}

func (ms *microsoftScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	var payload microsoftSearchResponse
	if err := fetchJSON(ctx, ms.retriever, ms.url, &payload); err != nil {
		return nil, fmt.Errorf("failed to retrieve Microsoft jobs: %w", err)
	}

	result := make([]domain.Job, 0)
	for _, position := range payload.Data.Positions {
		title := strings.TrimSpace(position.Title)
		if title == "" || position.ID == 0 {
			continue
		}

		location := ""
		if len(position.Locations) > 0 {
			location = strings.TrimSpace(position.Locations[0])
		}

		jobURL := ms.url
		if relative := strings.TrimSpace(position.PositionURL); relative != "" {
			jobURL = "https://apply.careers.microsoft.com" + relative
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ms.Name()).
			WithLocation(location).
			WithEmploymentType(strings.TrimSpace(position.WorkLocationOption)).
			WithURL(jobURL).
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	}

	return result, nil
}
