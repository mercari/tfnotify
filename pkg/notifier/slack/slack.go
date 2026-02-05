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
