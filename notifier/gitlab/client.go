package gitlab

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/mercari/tfnotify/terraform"

	gitlab "github.com/xanzy/go-gitlab"
)

// EnvToken is GitLab API Token
const EnvToken = "GITLAB_TOKEN"

// EnvBaseURL is GitLab base URL. This can be set to a domain endpoint to use with Private GitLab.
const EnvBaseURL = "GITLAB_BASE_URL"

// Client ...
type Client struct {
	*gitlab.Client
	Debug  bool
	Config Config

	common service

	Comment *CommentService
	Commits *CommitsService
	Notify  *NotifyService

	API API
}

// Config is a configuration for GitHub client
type Config struct {
	Token     string
	BaseURL   string
	NameSpace string
	Project   string
	MR        MergeRequest
	CI        string
	Parser    terraform.Parser
	Template  terraform.Template
}

// MergeRequest represents GitLab Merge Request metadata
type MergeRequest struct {
	Revision string
	Title    string
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
		return &Client{}, errors.New("gitlab token is missing")
	}
	client := gitlab.NewClient(http.DefaultClient, token)

	baseURL := cfg.BaseURL
	baseURL = strings.TrimPrefix(baseURL, "$")
	if baseURL == EnvBaseURL {
		baseURL = os.Getenv(EnvBaseURL)
	}
	if baseURL != "" {
		client.SetBaseURL(baseURL)
	}

	c := &Client{
		Config: cfg,
		Client: client,
	}
	c.common.client = c
	c.Comment = (*CommentService)(&c.common)
	c.Commits = (*CommitsService)(&c.common)
	c.Notify = (*NotifyService)(&c.common)

	c.API = &GitLab{
		Client:    client,
		namespace: cfg.NameSpace,
		project:   cfg.Project,
	}

	return c, nil
}

// IsNumber returns true if MergeRequest is Merge Request build
func (mr *MergeRequest) IsNumber() bool {
	return mr.Number != 0
}
