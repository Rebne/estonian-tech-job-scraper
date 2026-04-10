package notifier

import "fmt"

type Notifier interface {
	Notify(messages []string) error
}

type stdOutNotifier struct{}

func NewStdOutNotifier() *stdOutNotifier {
	return &stdOutNotifier{}
}

func (son stdOutNotifier) Notify(messages []string) error {
	for _, message := range messages {
		fmt.Println(message)
	}
	return nil
}
