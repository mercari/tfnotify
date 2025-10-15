package apperr

import (
	"errors"
	"testing"
)

func TestHandleError(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		err      error
		exitCode int
	}{
		{
			name:     "case 0",
			err:      NewExitError(1, errors.New("error")),
			exitCode: 1,
		},
		{
			name:     "case 1",
			err:      NewExitError(0, errors.New("error")),
			exitCode: 1,
		},
		{
			name:     "case 2",
			err:      errors.New("error"),
			exitCode: 1,
		},
		{
			name:     "case 3",
			err:      NewExitError(0, nil),
			exitCode: 0,
		},
		{
			name:     "case 4",
			err:      NewExitError(1, nil),
			exitCode: 1,
		},
		{
			name:     "case 5",
			err:      nil,
			exitCode: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// TODO: test stderr
			exitCode := HandleExit(testCase.err)
			if exitCode != testCase.exitCode {
				t.Errorf("got %d but want %d", exitCode, testCase.exitCode)
			}
		})
	}
}
