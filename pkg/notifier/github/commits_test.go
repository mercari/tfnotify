package github

import (
	"testing"
)

func TestPRNumber(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "xxx")
	testCases := []struct {
		prNumber int
		ok       bool
		revision string
	}{
		{
			prNumber: 1,
			ok:       true,
			revision: "xxx",
		},
	}

	for _, testCase := range testCases {
		cfg := newFakeConfig()
		client, err := NewClient(t.Context(), &cfg)
		if err != nil {
			t.Fatal(err)
		}
		api := newFakeAPI()
		client.API = &api
		prNumber, err := client.Commits.PRNumber(t.Context(), testCase.revision)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if prNumber != testCase.prNumber {
			t.Errorf("got %d but want %d", prNumber, testCase.prNumber)
		}
	}
}
