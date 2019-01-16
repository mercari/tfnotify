package typetalk

import (
	"errors"
	"os"
	"strconv"

	"github.com/mercari/tfnotify/terraform"
	typetalk "github.com/nulab/go-typetalk/typetalk/v1"
)

// EnvToken is Typetalk API Token
const EnvToken = "TYPETALK_TOKEN"

// EnvTopicID is Typetalk topic ID
const EnvTopicID = "TYPETALK_TOPIC_ID"

// Client represents Typetalk API client.
type Client struct {
	*typetalk.Client
	Config Config
	common service
	Notify *NotifyService
	API    API
}

// Config is a configuration for Typetalk Client
type Config struct {
	Token    string
	Title    string
	TopicID  string
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
	token := os.ExpandEnv(cfg.Token)
	if token == EnvToken {
		token = os.Getenv(EnvToken)
	}
	if token == "" {
		return &Client{}, errors.New("Typetalk token is missing")
	}

	topicIDString := os.ExpandEnv(cfg.TopicID)
	if topicIDString == EnvTopicID {
		topicIDString = os.Getenv(EnvTopicID)
	}
	if topicIDString == "" {
		return &Client{}, errors.New("Typetalk topic ID is missing")
	}

	topicID, err := strconv.Atoi(topicIDString)
	if err != nil {
		return &Client{}, errors.New("Typetalk topic ID is not numeric value")
	}

	client := typetalk.NewClient(nil)
	client.SetTypetalkToken(token)
	c := &Client{
		Config: cfg,
		Client: client,
	}
	c.common.client = c
	c.Notify = (*NotifyService)(&c.common)
	c.API = &Typetalk{
		Client:  client,
		TopicID: topicID,
	}

	return c, nil
}
