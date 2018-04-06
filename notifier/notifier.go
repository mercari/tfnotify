package notifier

// Notifier is a notification interface
type Notifier interface {
	Notify(body string) (exit int, err error)
}
