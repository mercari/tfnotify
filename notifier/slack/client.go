package slack

import (
	"errors"
	"os"
	"strings"

	"github.com/mercari/tfnotify/terraform"
	"github.com/lestrrat-go/slack"
)

// EnvToken is Slack API Token
const EnvToken = "SLACK_TOKEN"

// Client is a API client for Slack
type Client struct {
	*slack.Client

	Config Config

	common service

	Notify *NotifyService

	API API
}

// Config is a configuration for GitHub client
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
	client := slack.New(token)
	c := &Client{
		Config: cfg,
		Client: client,
	}
	c.common.client = c
	c.Notify = (*NotifyService)(&c.common)
	c.API = &Slack{
		Client:  client,
		Channel: cfg.Channel,
		Botname: cfg.Botname,
	}
	return c, nil
}
