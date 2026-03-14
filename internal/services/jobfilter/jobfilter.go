package jobfilter

import (
	"regexp"
	"strings"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
)

type JobFilter interface {
	Ok(job domain.Job) bool
}

type JobFilterChain interface {
	Add(filter JobFilter) JobFilterChain
	Match(job domain.Job) bool
}

type jobFilterChain struct {
	filters []JobFilter
}

func NewJobFilterChain() *jobFilterChain {
	return &jobFilterChain{filters: []JobFilter{}}
}

func (jfc *jobFilterChain) Add(filter JobFilter) JobFilterChain {
	jfc.filters = append(jfc.filters, filter)
	return jfc
}

func (jfc *jobFilterChain) Match(job domain.Job) bool {
	for _, filter := range jfc.filters {
		if !filter.Ok(job) {
			return false
		}
	}

	return true
}

type TitleIncludeFilter struct{}

func (TitleIncludeFilter) Ok(job domain.Job) bool {
	for _, key := range includeKeywords {
		if found := strings.Contains(job.Title(), key); found {
			return true
		}
	}
	return false
}

type TitleExcludeFilter struct{}

func (TitleExcludeFilter) Ok(job domain.Job) bool {
	for _, key := range excludeKeywords {
		if found := strings.Contains(job.Title(), key); found {
			return false
		}
	}
	return true
}

type LocationEstoniaFilter struct{}

func (LocationEstoniaFilter) Ok(job domain.Job) bool {
	re := regexp.MustCompile(`(?i)estonia|tallinn|tartu`)
	return re.MatchString(job.Location())
}

var excludeKeywords = []string{
	"staff",
	"lektor",
	"ekspert",
	"owner",
	"omanik",
	"ohvitser",
	"consultant",
	"konsultant",
	"arhitekt",
	"architect",
	"vanem",
	"senior",
	"lead",
	"juht",
	"manager",
}

var includeKeywords = []string{
	"arendaja",
	"developer",
	"full-stack",
	"full stack",
	"engineer",
	"junior",
	"algaja",
	"intern",
	"praktikant",
	"tester",
	"testija",
	"ui/ux",
}
