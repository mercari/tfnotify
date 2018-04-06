package github

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	githubToken := os.Getenv(EnvToken)
	defer func() {
		os.Setenv(EnvToken, githubToken)
	}()
	os.Setenv(EnvToken, "")

	testCases := []struct {
		config   Config
		envToken string
		expect   string
	}{
		{
			// specify directly
			config:   Config{Token: "abcdefg"},
			envToken: "",
			expect:   "",
		},
		{
			// specify via env but not to be set env (part 1)
			config:   Config{Token: "GITHUB_TOKEN"},
			envToken: "",
			expect:   "github token is missing",
		},
		{
			// specify via env (part 1)
			config:   Config{Token: "GITHUB_TOKEN"},
			envToken: "abcdefg",
			expect:   "",
		},
		{
			// specify via env but not to be set env (part 2)
			config:   Config{Token: "$GITHUB_TOKEN"},
			envToken: "",
			expect:   "github token is missing",
		},
		{
			// specify via env (part 2)
			config:   Config{Token: "$GITHUB_TOKEN"},
			envToken: "abcdefg",
			expect:   "",
		},
		{
			// no specification (part 1)
			config:   Config{},
			envToken: "",
			expect:   "github token is missing",
		},
		{
			// no specification (part 2)
			config:   Config{},
			envToken: "abcdefg",
			expect:   "github token is missing",
		},
	}
	for _, testCase := range testCases {
		os.Setenv(EnvToken, testCase.envToken)
		_, err := NewClient(testCase.config)
		if err == nil {
			continue
		}
		if err.Error() != testCase.expect {
			t.Errorf("got %q but want %q", err.Error(), testCase.expect)
		}
	}
}

func TestIsNumber(t *testing.T) {
	testCases := []struct {
		pr   PullRequest
		isPR bool
	}{
		{
			pr: PullRequest{
				Number: 0,
			},
			isPR: false,
		},
		{
			pr: PullRequest{
				Number: 123,
			},
			isPR: true,
		},
	}
	for _, testCase := range testCases {
		if testCase.pr.IsNumber() != testCase.isPR {
			t.Errorf("got %v but want %v", testCase.pr.IsNumber(), testCase.isPR)
		}
	}
}
