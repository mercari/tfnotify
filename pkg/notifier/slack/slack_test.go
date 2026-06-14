package slack

import (
	"context"
	"strings"
	"testing"

	"github.com/mercari/tfnotify/v1/pkg/notifier"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
)

// newTestClient builds a Slack client that would fail loudly if any API call
// were attempted (invalid token), so tests can assert that no message is sent.
func newTestClient(t *testing.T, cfg *Config) *Client {
	t.Helper()
	cfg.Token = "xoxb-test-invalid"
	cfg.ChannelID = "C0000000000"
	if cfg.Parser == nil {
		cfg.Parser = terraform.NewPlanParser()
	}
	if cfg.Template == nil {
		cfg.Template = terraform.NewPlanTemplate("")
	}
	if cfg.ParseErrorTemplate == nil {
		cfg.ParseErrorTemplate = terraform.NewPlanParseErrorTemplate("")
	}
	if cfg.Vars == nil {
		cfg.Vars = map[string]string{}
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	return client
}

const failedPlanOutput = `Error: Invalid configuration

Something went wrong`

// TestPlanDisabledNeverPosts guards against threads being sent when plan
// notifications are disabled (regression test).
func TestPlanDisabledNeverPosts(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, &Config{
		NotifyOnPlanError: false, // plan notifications disabled
		UseThreads:        true,  // threads enabled must not matter
	})

	err := client.Notify().Plan(context.Background(), &notifier.ParamExec{
		CombinedOutput: failedPlanOutput,
		ExitCode:       1,
	})
	// Any attempted Slack API call would fail with an auth error here,
	// so a nil error proves nothing was posted.
	if err != nil {
		t.Errorf("Plan() with notifications disabled posted to Slack: %v", err)
	}
}

// TestPlanSucceededNeverThreads ensures successful plans don't post threads
// even when notify_on_plan_error and threads are enabled.
func TestPlanSucceededNeverThreads(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, &Config{
		NotifyOnPlanError: true,
		UseThreads:        true,
	})

	err := client.Notify().Plan(context.Background(), &notifier.ParamExec{
		CombinedOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
		ExitCode:       0,
	})
	if err != nil {
		t.Errorf("Plan() with exit code 0 posted to Slack: %v", err)
	}
}

// TestApplyDisabledNeverPosts mirrors the plan test for the apply path.
func TestApplyDisabledNeverPosts(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, &Config{
		NotifyOnApplyError: false,
		UseThreads:         true,
		Parser:             terraform.NewApplyParser(),
		Template:           terraform.NewApplyTemplate(""),
		ParseErrorTemplate: terraform.NewApplyParseErrorTemplate(""),
	})

	err := client.Notify().Apply(context.Background(), &notifier.ParamExec{
		CombinedOutput: failedPlanOutput,
		ExitCode:       1,
	})
	if err != nil {
		t.Errorf("Apply() with notifications disabled posted to Slack: %v", err)
	}
}

func TestBuildParentMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		title   string
		message string
		status  string
		want    string
	}{
		{
			name:    "title and message",
			title:   "Plan Result",
			message: "service-a",
			status:  "failed",
			want:    "*Plan Result*\n\nservice-a\n\nfailed",
		},
		{
			name:   "title only",
			title:  "Plan Result",
			status: "failed",
			want:   "*Plan Result*\n\nfailed",
		},
		{
			name:   "status only",
			status: "failed",
			want:   "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildParentMessage(tt.title, tt.message, tt.status)
			if got != tt.want {
				t.Errorf("buildParentMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildThreadMessage(t *testing.T) {
	t.Parallel()

	got := buildThreadMessage("Error: something broke")
	want := "```\nError: something broke\n```"
	if got != want {
		t.Errorf("buildThreadMessage() = %q, want %q", got, want)
	}

	// Empty details should still produce a valid code block
	got = buildThreadMessage("")
	if !strings.Contains(got, "(no output captured)") {
		t.Errorf("buildThreadMessage(\"\") = %q, want placeholder text", got)
	}

	// Very long details must be truncated but keep the closing fence
	long := strings.Repeat("x", slackMaxTextLength+1000)
	got = buildThreadMessage(long)
	if len(got) > slackMaxTextLength {
		t.Errorf("buildThreadMessage() length = %d, exceeds Slack limit %d", len(got), slackMaxTextLength)
	}
	if !strings.HasSuffix(got, "\n```") {
		t.Errorf("buildThreadMessage() truncated message does not end with closing code fence: %q", got[len(got)-30:])
	}
	if !strings.Contains(got, "output truncated") {
		t.Error("buildThreadMessage() truncated message missing truncation note")
	}
}

func TestTruncateForSlack(t *testing.T) {
	t.Parallel()

	short := "hello"
	if got := truncateForSlack(short); got != short {
		t.Errorf("truncateForSlack(short) = %q, want unchanged", got)
	}

	long := strings.Repeat("y", slackMaxTextLength+5000)
	got := truncateForSlack(long)
	if len(got) > slackMaxTextLength {
		t.Errorf("truncateForSlack() length = %d, exceeds limit %d", len(got), slackMaxTextLength)
	}
	if !strings.Contains(got, "output truncated") {
		t.Error("truncateForSlack() missing truncation note")
	}
}
