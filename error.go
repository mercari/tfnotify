package main

import (
	"fmt"
	"os"

	"github.com/mercari/tfnotify/notifier"
)

// Exit codes are int values for the exit code that shell interpreter can interpret
const (
	ExitCodeOK    int = 0
	ExitCodeError int = iota
)

// ErrorFormatter is the interface for format
type ErrorFormatter interface {
	Format(s fmt.State, verb rune)
}

// ExitCoder is the wrapper interface for urfave/cli
type ExitCoder interface {
	error
	ExitCode() int
}

// ExitError is the wrapper struct for urfave/cli
type ExitError struct {
	exitCode int
	err      error
}

// NewExitError makes a new ExitError
func NewExitError(exitCode int, err error) *ExitError {
	return &ExitError{
		exitCode: exitCode,
		err:      err,
	}
}

// Error returns the string message, fulfilling the interface required by `error`
func (ee *ExitError) Error() string {
	if ee.err == nil {
		return ""
	}
	return fmt.Sprintf("%v", ee.err)
}

// ExitCode returns the exit code, fulfilling the interface required by `ExitCoder`
func (ee *ExitError) ExitCode() int {
	return ee.exitCode
}

// HandleExit returns int value that shell interpreter can interpret as the exit code
// If err has error message, it will be displayed to stderr
// This function is heavily inspired by urfave/cli.HandleExitCoder
func HandleExit(err error) int {
	if err == nil {
		return ExitCodeOK
	}

	// Ignore nop
	if err == notifier.ErrNop {
		return ExitCodeOK
	}

	if exitErr, ok := err.(ExitCoder); ok {
		if err.Error() != "" {
			if _, ok := exitErr.(ErrorFormatter); ok {
				fmt.Fprintf(os.Stderr, "%+v\n", err)
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		return exitErr.ExitCode()
	}

	if _, ok := err.(error); ok {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return ExitCodeError
	}

	return ExitCodeOK
}
