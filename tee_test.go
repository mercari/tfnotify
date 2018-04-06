package main

import (
	"bytes"
	"io"
	"testing"
)

func TestTee(t *testing.T) {
	testCases := []struct {
		stdin  io.Reader
		stdout string
		body   string
	}{
		{
			// Regular
			stdin:  bytes.NewBufferString("Plan: 1 to add\n"),
			stdout: "Plan: 1 to add\n",
			body:   "Plan: 1 to add\n",
		},
		{
			// ANSI color codes are included
			stdin:  bytes.NewBufferString("\033[mPlan: 1 to add\033[m\n"),
			stdout: "\033[mPlan: 1 to add\033[m\n",
			body:   "Plan: 1 to add\n",
		},
	}

	for _, testCase := range testCases {
		stdout := new(bytes.Buffer)
		body := tee(testCase.stdin, stdout)
		if body != testCase.body {
			t.Errorf("got %q but want %q", body, testCase.body)
		}
		if stdout.String() != testCase.stdout {
			t.Errorf("got %q but want %q", stdout.String(), testCase.stdout)
		}
	}
}
