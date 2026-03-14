package domain

import (
	"crypto/sha256"
	"strings"
)

type Job interface {
	Hash() []byte
	Page() string
	Title() string
	Company() string
	Location() string
	Description() string
	EmploymentType() string
	Category() string
	URL() string
}

type HashField string

const (
	HashFieldPage           HashField = "page"
	HashFieldTitle          HashField = "title"
	HashFieldCompany        HashField = "company"
	HashFieldLocation       HashField = "location"
	HashFieldDescription    HashField = "description"
	HashFieldEmploymentType HashField = "employment_type"
	HashFieldCategory       HashField = "category"
	HashFieldURL            HashField = "url"
)

type job struct {
	hash           []byte
	page           string
	title          string
	company        string
	location       string
	description    string
	employmentType string
	category       string
	url            string
}

func (j *job) Hash() []byte {
	return j.hash
}

func (j *job) Page() string {
	return j.page
}

func (j *job) Title() string {
	return j.title
}

func (j *job) Company() string {
	return j.company
}

func (j *job) Location() string {
	return j.location
}

func (j *job) Description() string {
	return j.description
}

func (j *job) EmploymentType() string {
	return j.employmentType
}

func (j *job) Category() string {
	return j.category
}

func (j *job) URL() string {
	return j.url
}

type jobBuilder struct {
	job *job
}

func NewJobBuilder() *jobBuilder {
	return &jobBuilder{job: &job{}}
}

func (jb *jobBuilder) Build() Job {
	return jb.job
}

func (jb *jobBuilder) WithURL(url string) *jobBuilder {
	jb.job.url = url
	return jb
}

func (jb *jobBuilder) WithCategory(category string) *jobBuilder {
	jb.job.category = category
	return jb
}

func (jb *jobBuilder) WithEmploymentType(employmentType string) *jobBuilder {
	jb.job.employmentType = employmentType
	return jb
}

func (jb *jobBuilder) WithDescription(description string) *jobBuilder {
	jb.job.description = description
	return jb
}

func (jb *jobBuilder) WithLocation(location string) *jobBuilder {
	jb.job.location = location
	return jb
}

func (jb *jobBuilder) WithCompany(company string) *jobBuilder {
	jb.job.company = company
	return jb
}

func (jb *jobBuilder) WithTitle(title string) *jobBuilder {
	jb.job.title = title
	return jb
}

func (jb *jobBuilder) WithPage(page string) *jobBuilder {
	jb.job.page = page
	return jb
}

func (jb *jobBuilder) WithUniqueIdentifier(identifier string) *jobBuilder {
	sum := sha256.Sum256([]byte(identifier))
	jb.job.hash = sum[:]
	return jb
}

func (jb *jobBuilder) WithHashFrom(fields ...HashField) *jobBuilder {

	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, normalizeHashValue(jb.valueForField(field)))
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	jb.job.hash = sum[:]
	return jb
}

func (jb *jobBuilder) valueForField(field HashField) string {
	switch field {
	case HashFieldPage:
		return jb.job.page
	case HashFieldTitle:
		return jb.job.title
	case HashFieldCompany:
		return jb.job.company
	case HashFieldLocation:
		return jb.job.location
	case HashFieldDescription:
		return jb.job.description
	case HashFieldEmploymentType:
		return jb.job.employmentType
	case HashFieldCategory:
		return jb.job.category
	case HashFieldURL:
		return jb.job.url
	default:
		return ""
	}
}

func normalizeHashValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
