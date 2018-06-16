package github

import (
	"errors"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/mercari/tfnotify/config"
	"github.com/mercari/tfnotify/terraform"
	"golang.org/x/oauth2"
)

// EnvToken is GitHub API Token
const EnvToken = "GITHUB_TOKEN"

// Client is a API client for GitHub
type Client struct {
	*github.Client
	Debug bool

	Config Config

	common service

	Comment *CommentService
	Commits *CommitsService
	Notify  *NotifyService

	API API
}

// Config is a configuration for GitHub client
type Config struct {
	Token    string
	Owner    string
	Repo     string
	PR       PullRequest
	CI       string
	Parser   terraform.Parser
	Template terraform.Template
	Filters  *config.Filters
}

// PullRequest represents GitHub Pull Request metadata
type PullRequest struct {
	Revision string
	Message  string
	Number   int
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
		return &Client{}, errors.New("github token is missing")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	c := &Client{
		Config: cfg,
		Client: client,
	}
	c.common.client = c
	c.Comment = (*CommentService)(&c.common)
	c.Commits = (*CommitsService)(&c.common)
	c.Notify = (*NotifyService)(&c.common)

	c.API = &GitHub{
		Client: client,
		owner:  cfg.Owner,
		repo:   cfg.Repo,
	}

	return c, nil
}

// IsNumber returns true if PullRequest is Pull Request build
func (pr *PullRequest) IsNumber() bool {
	return pr.Number != 0
}
