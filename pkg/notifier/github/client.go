package github

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// EnvBaseURL is GitHub base URL. This can be set to a domain endpoint to use with GitHub Enterprise.
const EnvBaseURL = "GITHUB_BASE_URL"

// Client is a API client for GitHub
type Client struct {
	*github.Client

	Debug    bool
	Config   *Config
	common   service
	Comment  *CommentService
	Commits  *CommitsService
	Notify   *NotifyService
	User     *UserService
	v4Client *githubv4.Client
	API      API
}

// Config is a configuration for GitHub client
type Config struct {
	BaseURL         string
	GraphQLEndpoint string
	Owner           string
	Repo            string
	PR              PullRequest
	CI              string
	Parser          terraform.Parser
	// Template is used for all Terraform command output
	Template           *terraform.Template
	ParseErrorTemplate *terraform.Template
	// ResultLabels is a set of labels to apply depending on the plan result
	ResultLabels     ResultLabels
	Vars             map[string]string
	EmbeddedVarNames []string
	Templates        map[string]string
	UseRawOutput     bool
	Patch            bool
	SkipNoChanges    bool
	IgnoreWarning    bool
	Masks            []*config.Mask
}

// PullRequest represents GitHub Pull Request metadata
type PullRequest struct {
	Revision string
	Number   int
}

type service struct {
	client *Client
}

func getToken() (string, error) {
	if token := os.Getenv("TFCMT_GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	return "", errors.New("github token is missing")
}

// NewClient returns Client initialized with Config
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	client := github.NewClient(tc)

	baseURL := cfg.BaseURL
	baseURL = strings.TrimPrefix(baseURL, "$")
	if baseURL == EnvBaseURL {
		baseURL = os.Getenv(EnvBaseURL)
	}
	if baseURL != "" {
		var err error
		client, err = github.NewClient(tc).WithEnterpriseURLs(baseURL, baseURL)
		if err != nil {
			return &Client{}, errors.New("failed to create a new github api client")
		}
	}

	c := &Client{
		Config: cfg,
		Client: client,
	}
	if cfg.GraphQLEndpoint == "" {
		c.v4Client = githubv4.NewClient(tc)
	} else {
		c.v4Client = githubv4.NewEnterpriseClient(cfg.GraphQLEndpoint, tc)
	}

	c.common.client = c
	c.Comment = (*CommentService)(&c.common)
	c.Commits = (*CommitsService)(&c.common)
	c.Notify = (*NotifyService)(&c.common)
	c.User = (*UserService)(&c.common)

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
	AddOrUpdateLabel      string
	DestroyLabel          string
	NoChangesLabel        string
	PlanErrorLabel        string
	AddOrUpdateLabelColor string
	DestroyLabelColor     string
	NoChangesLabelColor   string
	PlanErrorLabelColor   string
}

// HasAnyLabelDefined returns true if any of the internal labels are set
func (r *ResultLabels) HasAnyLabelDefined() bool {
	return r.AddOrUpdateLabel != "" || r.DestroyLabel != "" || r.NoChangesLabel != "" || r.PlanErrorLabel != ""
}

// IsResultLabel returns true if a label matches any of the internal labels
func (r *ResultLabels) IsResultLabel(label string) bool {
	switch label {
	case "":
		return false
	case r.AddOrUpdateLabel, r.DestroyLabel, r.NoChangesLabel, r.PlanErrorLabel:
		return true
	default:
		return false
	}
}
