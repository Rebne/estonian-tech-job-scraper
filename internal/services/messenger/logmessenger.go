package messenger

import (
	"fmt"

	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
)

type LogMessenger struct {
	notifier notifier.Notifier
}

func NewLogMessenger(notifier notifier.Notifier) *LogMessenger {
	return &LogMessenger{
		notifier: notifier,
	}
}

func (lm *LogMessenger) Send(logJson string) error {
	err := lm.notifier.Notify([]string{logJson})
	if err != nil {
		return fmt.Errorf("log messenger failed to notify: %w", err)
	}
	return nil
}
