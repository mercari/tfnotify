package slack

import (
	"fmt"
	"strings"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/sirupsen/logrus"
	slackgo "github.com/slack-go/slack"
)

// escapeSlackText escapes special characters for Slack mrkdwn format
func escapeSlackText(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}

// slackMaxTextLength is the maximum length of a Slack message text field.
// Slack rejects messages longer than 40,000 characters.
const slackMaxTextLength = 40000

// truncateForSlack truncates text so it fits in a single Slack message
func truncateForSlack(text string) string {
	if len(text) <= slackMaxTextLength {
		return text
	}
	const note = "\n... (output truncated by tfnotify)"
	return text[:slackMaxTextLength-len(note)] + note
}

// buildParentMessage builds the parent (summary) message of a thread
func buildParentMessage(title, message, status string) string {
	parent := ""
	if title != "" {
		parent = fmt.Sprintf("*%s*\n\n", title)
	}
	if message != "" {
		parent += fmt.Sprintf("%s\n\n", message)
	}
	return parent + status
}

// buildThreadMessage wraps the result details in a code block for a thread reply
func buildThreadMessage(details string) string {
	if details == "" {
		details = "(no output captured)"
	}
	// Leave room for the code fences and the truncation note so the
	// closing fence is never cut off.
	const overhead = 50
	if len(details) > slackMaxTextLength-overhead {
		details = details[:slackMaxTextLength-overhead] + "\n... (output truncated by tfnotify)"
	}
	return fmt.Sprintf("```\n%s\n```", details)
}

// Client is a Slack client
type Client struct {
	*slackgo.Client
	Config *Config
}

// Config is a Slack configuration
type Config struct {
	Token              string
	ChannelID          string
	BotName            string
	CI                 config.CI
	Parser             terraform.Parser
	Template           *terraform.Template
	ParseErrorTemplate *terraform.Template
	Terraform          config.Terraform
	Vars               map[string]string
	Templates          map[string]string
	UseRawOutput       bool
	// Default templates for Slack messages
	Title        string
	Message      string
	ApplyTitle   string
	ApplyMessage string
	PlanTitle    string
	PlanMessage  string
	// Notification control
	NotifyOnPlanError  bool
	NotifyOnApplyError bool
	UseThreads         bool
}

// NewClient creates a new Slack client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("slack token is required")
	}
	if cfg.ChannelID == "" {
		return nil, fmt.Errorf("slack channel ID is required")
	}

	client := slackgo.New(cfg.Token)

	// Optionally set bot name if provided
	if cfg.BotName != "" {
		// Bot name can be set in message options when posting
		logrus.WithField("bot_name", cfg.BotName).Debug("Slack bot name configured")
	}

	return &Client{
		Client: client,
		Config: cfg,
	}, nil
}

type service struct {
	client *Client
}

// NotifyService handles Slack notifications
type NotifyService service

// Notify returns NotifyService
func (c *Client) Notify() *NotifyService {
	return &NotifyService{
		client: c,
	}
}
