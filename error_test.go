package main

import (
	"errors"
	"testing"
)

func TestHandleError(t *testing.T) {
	testCases := []struct {
		err      error
		exitCode int
	}{
		{
			err:      NewExitError(1, errors.New("error")),
			exitCode: 1,
		},
		{
			err:      NewExitError(0, errors.New("error")),
			exitCode: 0,
		},
		{
			err:      errors.New("error"),
			exitCode: 1,
		},
		{
			err:      NewExitError(0, nil),
			exitCode: 0,
		},
		{
			err:      NewExitError(1, nil),
			exitCode: 1,
		},
		{
			err:      nil,
			exitCode: 0,
		},
	}

	for _, testCase := range testCases {
		// TODO: test stderr
		exitCode := HandleExit(testCase.err)
		if exitCode != testCase.exitCode {
			t.Errorf("got %q but want %q", exitCode, testCase.exitCode)
		}
	}
}
