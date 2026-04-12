package messenger

import (
	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobformatter"
	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
)

type JobMessenger struct {
	formatter jobformatter.JobFormatter
	notifier  notifier.Notifier
}

func NewJobMessenger(n notifier.Notifier, f jobformatter.JobFormatter) *JobMessenger {
	return &JobMessenger{
		notifier:  n,
		formatter: f,
	}
}

func (jm *JobMessenger) Send(jobs []domain.Job) error {
	var messages []string
	for _, job := range jobs {
		messages = append(messages, jm.formatter.MustFormatJob(job))
	}
	err := jm.notifier.Notify(messages)
	if err != nil {
		return err
	}
	return nil
}
