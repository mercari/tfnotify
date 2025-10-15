package github

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	testCases := []struct {
		name     string
		config   Config
		envToken string
		expect   string
	}{
		{
			name:     "specify via env but not to be set env (part 1)",
			config:   Config{},
			envToken: "",
			expect:   "github token is missing",
		},
		{
			name:     "specify via env (part 1)",
			config:   Config{},
			envToken: "abcdefg",
			expect:   "",
		},
		{
			name:     "specify via env but not to be set env (part 2)",
			config:   Config{},
			envToken: "",
			expect:   "github token is missing",
		},
		{
			name:     "specify via env (part 2)",
			config:   Config{},
			envToken: "abcdefg",
			expect:   "",
		},
		{
			name:     "no specification (part 1)",
			config:   Config{},
			envToken: "",
			expect:   "github token is missing",
		},
		{
			name:     "no specification (part 2)",
			config:   Config{},
			envToken: "abcdefg",
			expect:   "github token is missing",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", testCase.envToken)
			cfg := testCase.config
			_, err := NewClient(t.Context(), &cfg)
			if err == nil {
				return
			}
			if err.Error() != testCase.expect {
				t.Errorf("got %q but want %q", err.Error(), testCase.expect)
			}
		})
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "xxx")
	t.Setenv(EnvBaseURL, "")

	testCases := []struct {
		name       string
		config     Config
		envBaseURL string
		expect     string
	}{
		{
			name: "specify directly",
			config: Config{
				BaseURL: "https://git.example.com/api/v3/",
			},
			envBaseURL: "",
			expect:     "https://git.example.com/api/v3/",
		},
		{
			name: "specify via env but not to be set env (part 1)",
			config: Config{
				BaseURL: "GITHUB_BASE_URL",
			},
			envBaseURL: "",
			expect:     "https://api.github.com/",
		},
		{
			name: "specify via env (part 1)",
			config: Config{
				BaseURL: "GITHUB_BASE_URL",
			},
			envBaseURL: "https://git.example.com/api/v3/",
			expect:     "https://git.example.com/api/v3/",
		},
		{
			name: "specify via env but not to be set env (part 2)",
			config: Config{
				BaseURL: "$GITHUB_BASE_URL",
			},
			envBaseURL: "",
			expect:     "https://api.github.com/",
		},
		{
			name: "specify via env (part 2)",
			config: Config{
				BaseURL: "$GITHUB_BASE_URL",
			},
			envBaseURL: "https://git.example.com/api/v3/",
			expect:     "https://git.example.com/api/v3/",
		},
		{
			name:       "no specification (part 1)",
			config:     Config{},
			envBaseURL: "",
			expect:     "https://api.github.com/",
		},
		{
			name:       "no specification (part 2)",
			config:     Config{},
			envBaseURL: "https://git.example.com/api/v3/",
			expect:     "https://api.github.com/",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(EnvBaseURL, testCase.envBaseURL)
			cfg := testCase.config
			c, err := NewClient(t.Context(), &cfg)
			if err != nil {
				t.Fatal(err)
			}
			url := c.BaseURL.String()
			if url != testCase.expect {
				t.Errorf("got %q but want %q", url, testCase.expect)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
			rl:   ResultLabels{},
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
	t.Parallel()
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
			rl:    ResultLabels{},
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
