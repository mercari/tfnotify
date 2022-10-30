package mattermost

import (
	"errors"
	"os"
	"strings"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/mercari/tfnotify/terraform"
)

// EnvWebhook is Mattermost webhook
const EnvWebhook = "MATTERMOST_WEBHOOK"

// EnvChannelID is Slack channel ID
const EnvChannelID = "SLACK_CHANNEL_ID"

// EnvBotName is Slack bot name
const EnvBotName = "SLACK_BOT_NAME"

// Client is a API client for Slack
type Client struct {
	*slack.Payload

	Config Config

	common service

	Notify *NotifyService

	API API
}

// Config is a configuration for Mattermost client
type Config struct {
	Webhook  string
	Channel  string
	Botname  string
	Title    string
	Message  string
	CI       string
	Parser   terraform.Parser
	Template terraform.Template
}

type service struct {
	client *Client
}

// NewClient returns Client initialized with Config
func NewClient(cfg Config) (*Client, error) {
	webhook := cfg.Webhook
	webhook = strings.TrimPrefix(webhook, "$")
	if webhook == EnvWebhook {
		webhook = os.Getenv(EnvWebhook)
	}
	if webhook == "" {
		return &Client{}, errors.New("mattermost webhook is missing")
	}

	channel := cfg.Channel
	channel = strings.TrimPrefix(channel, "$")
	if channel == EnvChannelID {
		channel = os.Getenv(EnvChannelID)
	}

	botname := cfg.Botname
	botname = strings.TrimPrefix(botname, "$")
	if botname == EnvBotName {
		botname = os.Getenv(EnvBotName)
	}

	c := &Client{
		Config: cfg,
	}
	c.common.client = c
	c.Notify = (*NotifyService)(&c.common)
	c.API = &Mattermost{
		Webhook: webhook,
		Channel: channel,
		Botname: botname,
	}
	return c, nil
}
