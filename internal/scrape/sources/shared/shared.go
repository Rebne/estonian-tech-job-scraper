package shared

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/scrape/fetcher"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
)

var ErrNoJobsFound = errors.New("no jobs found")

func FilterJobs(jobs []domain.Job, filterChain jobfilter.JobFilterChain) []domain.Job {
	var result []domain.Job
	for _, job := range jobs {
		if filterChain.Match(job) {
			result = append(result, job)
		}
	}
	return result
}

func FetchJSON(ctx context.Context, retriever fetcher.HTMLRetriever, url string, target any) error {
	body, err := retriever.Fetch(ctx, url)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(body), target)
}

func FallbackOnEmptyString(target, fallback string) string {
	if target == "" {
		return fallback
	}
	return target
}
