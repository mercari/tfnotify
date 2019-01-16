package notifier

import (
	"errors"
)

var (
	ErrNop = errors.New("notification wasn't operated")
)

// Notifier is a notification interface
type Notifier interface {
	Notify(body string) (exit int, err error)
}
