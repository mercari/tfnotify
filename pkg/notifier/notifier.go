package notifier

import (
	"context"
)

// Notifier is a notification interface
type Notifier interface {
	Apply(ctx context.Context, param *ParamExec) error
	Plan(ctx context.Context, param *ParamExec) error
}

type ParamExec struct {
	Stdout         string
	Stderr         string
	CombinedOutput string
	CIName         string
	ExitCode       int
}
