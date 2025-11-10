package apperr

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
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
	return ee.err.Error()
}

// ExitCode returns the exit code, fulfilling the interface required by `ExitCoder`
func (ee *ExitError) ExitCode() int {
	return ee.exitCode
}

// HandleExit returns int value that shell interpreter can interpret as the exit code
// If err has error message, it will be displayed to stderr
// This function is heavily inspired by urfave/cli.HandleExitCoder
func HandleExit(err error) (int, string) {
	if err == nil {
		return ExitCodeOK, ""
	}

	logE := logrus.NewEntry(logrus.New())

	if exitErr, ok := err.(ExitCoder); ok { //nolint:errorlint
		errMsg := err.Error()
		if errMsg != "" {
			if _, ok := exitErr.(ErrorFormatter); ok {
				logrus.Errorf("%+v", err)
			} else {
				logerr.WithError(logE, err).Error("tfnotify failed")
			}
		}
		if code := exitErr.ExitCode(); code != 0 {
			return code, errMsg
		}
		if errMsg == "" {
			return ExitCodeOK, ""
		}
		return ExitCodeError, errMsg
	}

	logerr.WithError(logE, err).Error("tfnotify failed")
	return ExitCodeError, err.Error()
}
