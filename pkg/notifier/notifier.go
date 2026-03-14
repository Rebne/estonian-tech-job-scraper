package notifier

type Notifier interface {
	Notify(messages []string) error
}
