package jobformatter

import "github.com/Rebne/scrapy_project_v2/internal/domain"

type JobFormatter interface {
	FormatJob(job domain.Job) (string, error)
	MustFormatJob(job domain.Job) string
}
