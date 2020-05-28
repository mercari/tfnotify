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

func TestNewClientWithBaseURL(t *testing.T) {
	githubBaseURL := os.Getenv(EnvBaseURL)
	defer func() {
		os.Setenv(EnvBaseURL, githubBaseURL)
	}()
	os.Setenv(EnvBaseURL, "")

	testCases := []struct {
		config     Config
		envBaseURL string
		expect     string
	}{
		{
			// specify directly
			config: Config{
				Token:   "abcdefg",
				BaseURL: "https://git.example.com/api/v3/",
			},
			envBaseURL: "",
			expect:     "https://git.example.com/api/v3/",
		},
		{
			// specify via env but not to be set env (part 1)
			config: Config{
				Token:   "abcdefg",
				BaseURL: "GITHUB_BASE_URL",
			},
			envBaseURL: "",
			expect:     "https://api.github.com/",
		},
		{
			// specify via env (part 1)
			config: Config{
				Token:   "abcdefg",
				BaseURL: "GITHUB_BASE_URL",
			},
			envBaseURL: "https://git.example.com/api/v3/",
			expect:     "https://git.example.com/api/v3/",
		},
		{
			// specify via env but not to be set env (part 2)
			config: Config{
				Token:   "abcdefg",
				BaseURL: "$GITHUB_BASE_URL",
			},
			envBaseURL: "",
			expect:     "https://api.github.com/",
		},
		{
			// specify via env (part 2)
			config: Config{
				Token:   "abcdefg",
				BaseURL: "$GITHUB_BASE_URL",
			},
			envBaseURL: "https://git.example.com/api/v3/",
			expect:     "https://git.example.com/api/v3/",
		},
		{
			// no specification (part 1)
			config:     Config{Token: "abcdefg"},
			envBaseURL: "",
			expect:     "https://api.github.com/",
		},
		{
			// no specification (part 2)
			config:     Config{Token: "abcdefg"},
			envBaseURL: "https://git.example.com/api/v3/",
			expect:     "https://api.github.com/",
		},
	}
	for _, testCase := range testCases {
		os.Setenv(EnvBaseURL, testCase.envBaseURL)
		c, err := NewClient(testCase.config)
		if err != nil {
			continue
		}
		url := c.Client.BaseURL.String()
		if url != testCase.expect {
			t.Errorf("got %q but want %q", url, testCase.expect)
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

func TestHasAnyLabelDefined(t *testing.T) {
	testCases := []struct {
		rl   ResultLabels
		want bool
	}{
		{
			rl: ResultLabels{
				AddOrUpdateLabel: "add-or-update",
				DestroyLabel:     "destroy",
				NoChangesLabel:   "no-changes",
				PlanErrorLabel:   "error",
			},
			want: true,
		},
		{
			rl: ResultLabels{
				AddOrUpdateLabel: "add-or-update",
				DestroyLabel:     "destroy",
				NoChangesLabel:   "",
				PlanErrorLabel:   "error",
			},
			want: true,
		},
		{
			rl: ResultLabels{
				AddOrUpdateLabel: "",
				DestroyLabel:     "",
				NoChangesLabel:   "",
				PlanErrorLabel:   "",
			},
			want: false,
		},
		{
			rl: ResultLabels{},
			want: false,
		},
	}
	for _, testCase := range testCases {
		if testCase.rl.HasAnyLabelDefined() != testCase.want {
			t.Errorf("got %v but want %v", testCase.rl.HasAnyLabelDefined(), testCase.want)
		}
	}
}

func TestIsResultLabels(t *testing.T) {
	testCases := []struct {
		rl    ResultLabels
		label string
		want  bool
	}{
		{
			rl: ResultLabels{
				AddOrUpdateLabel: "add-or-update",
				DestroyLabel:     "destroy",
				NoChangesLabel:   "no-changes",
				PlanErrorLabel:   "error",
			},
			label: "add-or-update",
			want:  true,
		},
		{
			rl: ResultLabels{
				AddOrUpdateLabel: "add-or-update",
				DestroyLabel:     "destroy",
				NoChangesLabel:   "no-changes",
				PlanErrorLabel:   "error",
			},
			label: "my-label",
			want:  false,
		},
		{
			rl: ResultLabels{
				AddOrUpdateLabel: "add-or-update",
				DestroyLabel:     "destroy",
				NoChangesLabel:   "no-changes",
				PlanErrorLabel:   "error",
			},
			label: "",
			want:  false,
		},
		{
			rl: ResultLabels{
				AddOrUpdateLabel: "",
				DestroyLabel:     "",
				NoChangesLabel:   "no-changes",
				PlanErrorLabel:   "",
			},
			label: "",
			want:  false,
		},
		{
			rl: ResultLabels{},
			label: "",
			want:  false,
		},
	}
	for _, testCase := range testCases {
		if testCase.rl.IsResultLabel(testCase.label) != testCase.want {
			t.Errorf("got %v but want %v", testCase.rl.IsResultLabel(testCase.label), testCase.want)
		}
	}
}
