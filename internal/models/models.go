package models

import "crypto/sha256"

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

type JobBuilder struct {
	job *job
}

func (jb *JobBuilder) Build() *job {
	return jb.job
}

func (jb *JobBuilder) WithURL(url string) *JobBuilder {
	jb.job.url = url
	return jb
}

func (jb *JobBuilder) WithCategory(category string) *JobBuilder {
	jb.job.category = category
	return jb
}

func (jb *JobBuilder) WithEmploymentType(employmentType string) *JobBuilder {
	jb.job.employmentType = employmentType
	return jb
}

func (jb *JobBuilder) WithDescription(description string) *JobBuilder {
	jb.job.description = description
	return jb
}

func (jb *JobBuilder) WithLocation(location string) *JobBuilder {
	jb.job.location = location
	return jb
}

func (jb *JobBuilder) WithCompany(company string) *JobBuilder {
	jb.job.company = company
	return jb
}

func (jb *JobBuilder) WithTile(title string) *JobBuilder {
	jb.job.title = title
	return jb
}

func (jb *JobBuilder) WithPage(page string) *JobBuilder {
	jb.job.page = page
	return jb
}

func (jb *JobBuilder) WithUniqueIdentifier(identifier string) *JobBuilder {
	sum := sha256.Sum256([]byte(identifier))
	jb.job.hash = sum[:]
	return jb
}
