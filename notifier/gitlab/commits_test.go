package gitlab

import (
	"testing"
)

func TestCommitsList(t *testing.T) {
	testCases := []struct {
		revision string
		ok       bool
	}{
		{
			revision: "04e0917e448b662c2b16330fad50e97af16ff27a",
			ok:       true,
		},
		{
			revision: "",
			ok:       false,
		},
	}

	for _, testCase := range testCases {
		cfg := newFakeConfig()
		client, err := NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		api := newFakeAPI()
		client.API = &api
		_, err = client.Commits.List(testCase.revision)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}

func TestCommitsLastOne(t *testing.T) {
	testCases := []struct {
		commits  []string
		revision string
		lastRev  string
		ok       bool
	}{
		{
			// ok
			commits: []string{
				"04e0917e448b662c2b16330fad50e97af16ff27a",
				"04e0917e448b662c2b16330fad50e97af16ff27b",
				"04e0917e448b662c2b16330fad50e97af16ff27c",
			},
			revision: "04e0917e448b662c2b16330fad50e97af16ff27a",
			lastRev:  "04e0917e448b662c2b16330fad50e97af16ff27b",
			ok:       true,
		},
		{
			// no revision
			commits: []string{
				"04e0917e448b662c2b16330fad50e97af16ff27a",
				"04e0917e448b662c2b16330fad50e97af16ff27b",
				"04e0917e448b662c2b16330fad50e97af16ff27c",
			},
			revision: "",
			lastRev:  "",
			ok:       false,
		},
		{
			// no commits
			commits:  []string{},
			revision: "04e0917e448b662c2b16330fad50e97af16ff27a",
			lastRev:  "",
			ok:       false,
		},
	}

	for _, testCase := range testCases {
		cfg := newFakeConfig()
		client, err := NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		api := newFakeAPI()
		client.API = &api
		commit, err := client.Commits.lastOne(testCase.commits, testCase.revision)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if commit != testCase.lastRev {
			t.Errorf("got %q but want %q", commit, testCase.lastRev)
		}
	}
}
