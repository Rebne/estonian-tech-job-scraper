package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
)

const stebbyURL string = "https://stebby.bamboohr.com/careers/list"

type stebbyScraper struct {
	url       string
	retriever fetcher.HTMLRetriever
}

type stebbyResponse struct {
	Meta struct {
		TotalCount int `json:"totalCount"`
	} `json:"meta"`
	Result []stebbyJob `json:"result"`
}

type stebbyJob struct {
	JobOpeningName string `json:"jobOpeningName"`
	Location       struct {
		City string `json:"city"`
	} `json:"location"`
}

func NewStebbyScraper(retriever fetcher.HTMLRetriever) *stebbyScraper {
	return &stebbyScraper{url: stebbyURL, retriever: retriever}
}

func (ss *stebbyScraper) Name() string {
	return "stebby"
}

func (ss *stebbyScraper) GetJobs(ctx context.Context) ([]domain.Job, error) {
	var payload stebbyResponse
	if err := fetchJSON(ctx, ss.retriever, ss.url, &payload); err != nil {
		return nil, fmt.Errorf("failed to retrieve Stebby jobs: %w", err)
	}

	if payload.Meta.TotalCount == 0 {
		return []domain.Job{}, nil
	}

	result := make([]domain.Job, 0)
	for _, job := range payload.Result {
		title := strings.TrimSpace(job.JobOpeningName)
		if title == "" {
			continue
		}

		result = append(result, domain.
			NewJobBuilder().
			WithTitle(title).
			WithPage(ss.Name()).
			WithLocation(strings.TrimSpace(job.Location.City)).
			WithURL("https://stebby.bamboohr.com/careers").
			WithHashFrom(domain.HashFieldTitle, domain.HashFieldPage).
			Build(),
		)
	}

	return result, nil
}
