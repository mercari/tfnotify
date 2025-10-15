package localfile

import (
	"context"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/mercari/tfnotify/v1/pkg/notifier/github"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
)

// Client is a fake API client for write to local file
type Client struct {
	Debug bool

	Config *Config

	common service

	Notify  *NotifyService
	Output  *OutputService
	labeler Labeler
}

// Config is a configuration for local file
type Config struct {
	OutputFile string
	Parser     terraform.Parser
	// Template is used for all Terraform command output
	Template           *terraform.Template
	ParseErrorTemplate *terraform.Template
	Vars               map[string]string
	EmbeddedVarNames   []string
	Templates          map[string]string
	CI                 string
	UseRawOutput       bool
	Masks              []*config.Mask

	// For labeling
	DisableLabel bool
}

type GitHubLabelConfig struct {
	BaseURL         string
	GraphQLEndpoint string
	Owner           string
	Repo            string
	PRNumber        int
	Revision        string
	Labels          github.ResultLabels
}

type service struct {
	client *Client
}

type Labeler interface {
	UpdateLabels(ctx context.Context, result terraform.ParseResult) []string
}

// NewClient returns Client initialized with Config
func NewClient(cfg *Config, labeler Labeler) (*Client, error) {
	c := &Client{
		Config:  cfg,
		labeler: labeler,
	}

	c.common.client = c

	c.Notify = (*NotifyService)(&c.common)

	return c, nil
}
