package slack

import (
	"errors"
	"os"
	"strings"

	"github.com/lestrrat-go/slack"
	"github.com/mercari/tfnotify/terraform"
)

// EnvToken is Slack API Token
const EnvToken = "SLACK_TOKEN"

// EnvChannelID is Slack channel ID
const EnvChannelID = "SLACK_CHANNEL_ID"

// EnvBotName is Slack bot name
const EnvBotName = "SLACK_BOT_NAME"

// Client is a API client for Slack
type Client struct {
	*slack.Client

	Config Config

	common service

	Notify *NotifyService

	API API
}

// Config is a configuration for Slack client
type Config struct {
	Token    string
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
	token := cfg.Token
	token = strings.TrimPrefix(token, "$")
	if token == EnvToken {
		token = os.Getenv(EnvToken)
	}
	if token == "" {
		return &Client{}, errors.New("slack token is missing")
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

	client := slack.New(token)
	c := &Client{
		Config: cfg,
		Client: client,
	}
	c.common.client = c
	c.Notify = (*NotifyService)(&c.common)
	c.API = &Slack{
		Client:  client,
		Channel: channel,
		Botname: botname,
	}
	return c, nil
}
