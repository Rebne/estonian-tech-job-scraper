package jobformatter

import "github.com/Rebne/scrapy_project_v2/internal/domain"

type JobFormatter interface {
	FormatJobs(jobs []domain.Job) ([]string, error)
	FormatJob(job domain.Job) (string, error)
	MustFormatJob(job domain.Job) string
}
