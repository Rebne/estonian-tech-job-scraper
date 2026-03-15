package messageformatter

import (
	"github.com/Rebne/scrapy_project_v2/internal/domain"
)

type JobFormatter interface {
	FormatJobs(jobs []domain.Job) ([]string, error)
}

