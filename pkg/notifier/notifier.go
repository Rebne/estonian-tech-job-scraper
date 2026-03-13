package notifier

type INotifier interface {
	Notify() error
}

type Notifier struct {
}

func (n Notifier) Notify() error {
	return nil
}
