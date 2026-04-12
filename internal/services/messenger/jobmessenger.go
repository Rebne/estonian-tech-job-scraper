package messenger

import (
	"fmt"

	"github.com/Rebne/scrapy_project_v2/internal/domain"
	"github.com/Rebne/scrapy_project_v2/internal/services/jobformatter"
	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
)

type JobMessenger struct {
	formatter jobformatter.JobFormatter
	notifier  notifier.Notifier
}

func NewJobMessenger(notifier notifier.Notifier, formatter jobformatter.JobFormatter) *JobMessenger {
	return &JobMessenger{
		notifier:  notifier,
		formatter: formatter,
	}
}

func (jm *JobMessenger) Send(jobs []domain.Job) error {
	var messages []string
	for _, job := range jobs {
		messages = append(messages, jm.formatter.MustFormatJob(job))
	}
	err := jm.notifier.Notify(messages)
	if err != nil {
		return fmt.Errorf("job messenger failed to notify: %w", err)
	}
	return nil
}
