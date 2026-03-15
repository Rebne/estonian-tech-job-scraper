package messageformatter

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
)

type JobFormatter interface {
	FormatJobs(jobs []domain.Job) ([]string, error)
	FormatJob(job domain.Job) (string, error)
	MustFormatJob(job domain.Job) string
}

type TelegramHTMLFormatter struct {
	tmpl *template.Template
}

type telegramTemplateData struct {
	Company        string
	Title          string
	Location       string
	Description    string
	EmploymentType string
	Category       string
	URL            string
}

const telegramJobTemplate = `New Job Listing:
{{if .Company}}🏢 Company: {{.Company}}
{{end}}{{if .Title}}📋 Title: {{.Title}}
{{end}}{{if .Location}}📍 Location: {{.Location}}
{{end}}{{if .Description}}📝 Description: {{.Description}}
{{end}}{{if .EmploymentType}}💼 Employment Type: {{.EmploymentType}}
{{end}}{{if .Category}}🔍 Category: {{.Category}}
{{end}}{{if .URL}}🆔 <a href="{{.URL}}">Job link</a>{{end}}`

func NewTelegramHTMLFormatter() *TelegramHTMLFormatter {
	return &TelegramHTMLFormatter{tmpl: template.Must(template.New("telegram-job").Parse(telegramJobTemplate))}
}

func (f *TelegramHTMLFormatter) FormatJobs(jobs []domain.Job) ([]string, error) {
	messages := make([]string, 0, len(jobs))
	for _, job := range jobs {
		formatted, err := f.FormatJob(job)
		if err != nil {
			return nil, fmt.Errorf("failed to format jobs: %w", err)
		}
		messages = append(messages, formatted)
	}

	return messages, nil
}

func (f *TelegramHTMLFormatter) FormatJob(job domain.Job) (string, error) {
	return f.formatJob(job)
}

func (f *TelegramHTMLFormatter) MustFormatJob(job domain.Job) string {
	result, err := f.formatJob(job)
	if err != nil {
		panic(err)
	}
	return result
}

func (f *TelegramHTMLFormatter) formatJob(job domain.Job) (string, error) {
	data := telegramTemplateData{
		Title:          strings.TrimSpace(job.Title()),
		Location:       strings.TrimSpace(job.Location()),
		Description:    strings.TrimSpace(job.Description()),
		EmploymentType: strings.TrimSpace(job.EmploymentType()),
		Category:       strings.TrimSpace(job.Category()),
		URL:            strings.TrimSpace(job.URL()),
	}
	if company := strings.TrimSpace(job.Company()); company != "" {
		data.Company = strings.ToUpper(company)
	} else if page := strings.TrimSpace(job.Page()); page != "" {
		data.Company = strings.ToUpper(page)
	}

	var b bytes.Buffer
	if err := f.tmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("failed to render telegram template: %w", err)
	}

	return b.String(), nil
}
