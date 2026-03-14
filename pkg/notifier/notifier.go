package notifier

type INotifier interface {
	Notify() error
}
