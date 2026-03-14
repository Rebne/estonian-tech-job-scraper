package notifier

type Notifier interface {
	Notify() error
}
