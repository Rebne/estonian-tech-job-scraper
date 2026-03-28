package sources

import (
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobfilter"
	"errors"
)

var ErrNoJobsFound = errors.New("no jobs found")

func filterJobs(jobs []domain.Job, filterChain jobfilter.JobFilterChain) []domain.Job {
	var result []domain.Job
	for _, job := range jobs {
		if filterChain.Match(job) {
			result = append(result, job)
		}
	}
	return result
}
