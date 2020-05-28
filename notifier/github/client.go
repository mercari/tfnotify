package github

import (
	"errors"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/mercari/tfnotify/terraform"
	"golang.org/x/oauth2"
)

// EnvToken is GitHub API Token
const EnvToken = "GITHUB_TOKEN"

// EnvBaseURL is GitHub base URL. This can be set to a domain endpoint to use with GitHub Enterprise.
const EnvBaseURL = "GITHUB_BASE_URL"

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
	Token        string
	BaseURL      string
	Owner        string
	Repo         string
	PR           PullRequest
	CI           string
	Parser       terraform.Parser
	UseRawOutput bool
	WarnDestroy  bool
	// Template is used for all Terraform command output
	Template terraform.Template
	// DestroyWarningTemplate is used only for additional warning
	// the plan result contains destroy operation
	DestroyWarningTemplate terraform.Template
	// ResultLabels is a set of labels to apply depending on the plan result
	ResultLabels ResultLabels
}

// PullRequest represents GitHub Pull Request metadata
type PullRequest struct {
	Revision              string
	Title                 string
	Message               string
	Number                int
	DestroyWarningTitle   string
	DestroyWarningMessage string
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

	baseURL := cfg.BaseURL
	baseURL = strings.TrimPrefix(baseURL, "$")
	if baseURL == EnvBaseURL {
		baseURL = os.Getenv(EnvBaseURL)
	}
	if baseURL != "" {
		var err error
		client, err = github.NewEnterpriseClient(baseURL, baseURL, tc)
		if err != nil {
			return &Client{}, errors.New("failed to create a new github api client")
		}
	}

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

// ResultLabels represents the labels to add to the PR depending on the plan result
type ResultLabels struct {
	ChangesLabel   string
	DestroyLabel   string
	ErrorLabel     string
	NoChangesLabel string
}

// HasAnyLabelDefined returns true if any of the internal labels are set
func (r *ResultLabels) HasAnyLabelDefined() bool {
	return r.ChangesLabel != "" || r.DestroyLabel != "" || r.ErrorLabel != "" || r.NoChangesLabel != ""
}

// ToStringSlice returns all the defined labels as a string slice
func (r *ResultLabels) ToStringSlice() []string {
	var result []string
	if r.ChangesLabel != "" {
		result = append(result, r.ChangesLabel)
	}
	if r.DestroyLabel != "" {
		result = append(result, r.DestroyLabel)
	}
	if r.ErrorLabel != "" {
		result = append(result, r.ErrorLabel)
	}
	if r.NoChangesLabel != "" {
		result = append(result, r.NoChangesLabel)
	}
	return result
}

func (r *ResultLabels) IsResultLabel(label string) bool {
	switch label {
	case "":
		return false
	case r.ChangesLabel, r.DestroyLabel, r.ErrorLabel, r.NoChangesLabel:
		return true
	default:
		return false
	}
}
