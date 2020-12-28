package notifier

import "context"

// Notifier is a notification interface
type Notifier interface {
	Notify(ctx context.Context, body string) (exit int, err error)
}
