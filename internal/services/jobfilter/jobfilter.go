package jobfilter

import (
	"github.com/Rebne/scrapy_project_v2/internal/models"
)
type JobFilter interface {
	Ok(job models.Job) bool
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
