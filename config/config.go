package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
	"gopkg.in/yaml.v3"
)

// Config is for tfnotify config structure
type Config struct {
	CI                 CI                `json:"-" yaml:"-"`
	Terraform          Terraform         `json:"terraform,omitempty"`
	Vars               map[string]string `json:"-" yaml:"-"`
	EmbeddedVarNames   []string          `json:"embedded_var_names,omitempty" yaml:"embedded_var_names"`
	Templates          map[string]string `json:"templates,omitempty"`
	Log                Log               `json:"log,omitempty"`
	GHEBaseURL         string            `json:"ghe_base_url,omitempty" yaml:"ghe_base_url"`
	GHEGraphQLEndpoint string            `json:"ghe_graphql_endpoint,omitempty" yaml:"ghe_graphql_endpoint"`
	PlanPatch          bool              `json:"plan_patch,omitempty" yaml:"plan_patch"`
	RepoOwner          string            `json:"repo_owner,omitempty" yaml:"repo_owner"`
	RepoName           string            `json:"repo_name,omitempty" yaml:"repo_name"`
	Output             string            `json:"-" yaml:"-"`
	Masks              []*Mask           `json:"-" yaml:"-"`

	// Legacy fields for backward compatibility
	Notifier Notifier `yaml:"notifier"`

	path string
}

type Mask struct {
	Type   string
	Value  string
	Regexp *regexp.Regexp
}

type CI struct {
	Name     string
	Owner    string
	Repo     string
	SHA      string
	Link     string
	PRNumber int
}

type Log struct {
	Level string `json:"level,omitempty"`
}

// Notifier is a notification notifier
type Notifier struct {
	Github   GithubNotifier   `yaml:"github"`
	Gitlab   GitlabNotifier   `yaml:"gitlab"`
	Slack    SlackNotifier    `yaml:"slack"`
	Typetalk TypetalkNotifier `yaml:"typetalk"`
}

// GithubNotifier is a notifier for GitHub
type GithubNotifier struct {
	Token      string     `yaml:"token"`
	BaseURL    string     `yaml:"base_url"`
	Repository Repository `yaml:"repository"`
}

// GitlabNotifier is a notifier for GitLab
type GitlabNotifier struct {
	Token      string     `yaml:"token"`
	BaseURL    string     `yaml:"base_url"`
	Repository Repository `yaml:"repository"`
}

// Repository represents a GitHub repository
type Repository struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

// SlackNotifier is a notifier for Slack
type SlackNotifier struct {
	Token   string `yaml:"token"`
	Channel string `yaml:"channel"`
	Bot     string `yaml:"bot"`
}

// TypetalkNotifier is a notifier for Typetalk
type TypetalkNotifier struct {
	Token   string `yaml:"token"`
	TopicID string `yaml:"topic_id"`
}

// Terraform represents terraform configurations
type Terraform struct {
	Plan         Plan  `json:"plan,omitempty"`
	Apply        Apply `json:"apply,omitempty"`
	UseRawOutput bool  `json:"use_raw_output,omitempty" yaml:"use_raw_output"`

	// Legacy fields for backward compatibility
	Default  Default  `yaml:"default"`
	Fmt      Fmt      `yaml:"fmt"`
	Validate Validate `yaml:"validate"`
}

// Default is a default setting for terraform commands
type Default struct {
	Template string `yaml:"template"`
}

// Fmt is a terraform fmt config
type Fmt struct {
	Template string `yaml:"template"`
}

// Validate is a terraform validate config
type Validate struct {
	Template string `yaml:"template"`
}

// Plan is a terraform plan config
type Plan struct {
	Template            string              `json:"template,omitempty"`
	WhenAddOrUpdateOnly WhenAddOrUpdateOnly `json:"when_add_or_update_only,omitempty" yaml:"when_add_or_update_only"`
	WhenDestroy         WhenDestroy         `json:"when_destroy,omitempty" yaml:"when_destroy"`
	WhenNoChanges       WhenNoChanges       `json:"when_no_changes,omitempty" yaml:"when_no_changes"`
	WhenPlanError       WhenPlanError       `json:"when_plan_error,omitempty" yaml:"when_plan_error"`
	WhenParseError      WhenParseError      `json:"when_parse_error,omitempty" yaml:"when_parse_error"`
	DisableLabel        bool                `json:"disable_label,omitempty" yaml:"disable_label"`
	IgnoreWarning       bool                `json:"ignore_warning,omitempty" yaml:"ignore_warning"`
}

// WhenAddOrUpdateOnly is a configuration to notify the plan result contains new or updated in place resources
type WhenAddOrUpdateOnly struct {
	Label        string `json:"label,omitempty"`
	Color        string `json:"label_color,omitempty" yaml:"label_color"`
	DisableLabel bool   `json:"disable_label,omitempty" yaml:"disable_label"`
}

// WhenDestroy is a configuration to notify the plan result contains destroy operation
type WhenDestroy struct {
	Label        string `json:"label,omitempty"`
	Color        string `json:"label_color,omitempty" yaml:"label_color"`
	DisableLabel bool   `json:"disable_label,omitempty" yaml:"disable_label"`
	Template     string `json:"template,omitempty" yaml:"template,omitempty"`
}

// WhenNoChanges is a configuration to add a label when the plan result contains no change
type WhenNoChanges struct {
	Label          string `json:"label,omitempty"`
	Color          string `json:"label_color,omitempty" yaml:"label_color"`
	DisableLabel   bool   `json:"disable_label,omitempty" yaml:"disable_label"`
	DisableComment bool   `json:"disable_comment,omitempty" yaml:"disable_comment"`
}

// WhenPlanError is a configuration to notify the plan result returns an error
type WhenPlanError struct {
	Label        string `json:"label,omitempty"`
	Color        string `json:"label_color,omitempty" yaml:"label_color"`
	DisableLabel bool   `json:"disable_label,omitempty" yaml:"disable_label"`
}

// WhenParseError is a configuration to notify the plan result returns an error
type WhenParseError struct {
	Template string `json:"template,omitempty"`
}

// Apply is a terraform apply config
type Apply struct {
	Template string `yaml:"template"`
}

// LoadFile binds the config file to Config structure
func (cfg *Config) LoadFile(path string) error {
	cfg.path = path
	_, err := os.Stat(cfg.path)
	if err != nil {
		return fmt.Errorf("%s: no config file", cfg.path)
	}
	raw, err := os.ReadFile(cfg.path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	return yaml.Unmarshal(raw, cfg)
}

// Validation validates config file
func (cfg *Config) Validation() error {
	switch strings.ToLower(cfg.CI.Name) {
	case "":
		return errors.New("ci: need to be set")
	case "circleci", "circle-ci":
		// ok pattern
	case "gitlabci", "gitlab-ci":
		// ok pattern
	case "travis", "travisci", "travis-ci":
		// ok pattern
	case "codebuild":
		// ok pattern
	case "teamcity":
		// ok pattern
	case "drone":
		// ok pattern
	case "jenkins":
		// ok pattern
	case "github-actions":
		// ok pattern
	case "cloud-build", "cloudbuild":
		// ok pattern
	default:
		return fmt.Errorf("%s: not supported yet", cfg.CI.Name)
	}
	if cfg.isDefinedGithub() {
		if cfg.Notifier.Github.Repository.Owner == "" {
			return fmt.Errorf("repository owner is missing")
		}
		if cfg.Notifier.Github.Repository.Name == "" {
			return fmt.Errorf("repository name is missing")
		}
	}
	if cfg.isDefinedGitlab() {
		if cfg.Notifier.Gitlab.Repository.Owner == "" {
			return fmt.Errorf("repository owner is missing")
		}
		if cfg.Notifier.Gitlab.Repository.Name == "" {
			return fmt.Errorf("repository name is missing")
		}
	}
	if cfg.isDefinedSlack() {
		if cfg.Notifier.Slack.Channel == "" {
			return fmt.Errorf("slack channel id is missing")
		}
	}
	if cfg.isDefinedTypetalk() {
		if cfg.Notifier.Typetalk.TopicID == "" {
			return fmt.Errorf("typetalk topic id is missing")
		}
	}
	notifier := cfg.GetNotifierType()
	if notifier == "" {
		return fmt.Errorf("notifier is missing")
	}
	return nil
}

func (cfg *Config) isDefinedGithub() bool {
	// not empty
	return cfg.Notifier.Github != (GithubNotifier{})
}

func (cfg *Config) isDefinedGitlab() bool {
	// not empty
	return cfg.Notifier.Gitlab != (GitlabNotifier{})
}

func (cfg *Config) isDefinedSlack() bool {
	// not empty
	return cfg.Notifier.Slack != (SlackNotifier{})
}

func (cfg *Config) isDefinedTypetalk() bool {
	// not empty
	return cfg.Notifier.Typetalk != (TypetalkNotifier{})
}

// GetNotifierType return notifier type described in Config
func (cfg *Config) GetNotifierType() string {
	if cfg.isDefinedGithub() {
		return "github"
	}
	if cfg.isDefinedGitlab() {
		return "gitlab"
	}
	if cfg.isDefinedSlack() {
		return "slack"
	}
	if cfg.isDefinedTypetalk() {
		return "typetalk"
	}
	return ""
}

// Find returns config path
func (cfg *Config) Find(file string) (string, error) {
	if file != "" {
		return file, nil
	}

	// Use findconfig to search for config files
	configFiles := []string{
		"tfnotify.yaml",
		"tfnotify.yml",
		".tfnotify.yaml",
		".tfnotify.yml",
		"tfcmt.yaml",
		"tfcmt.yml",
		".tfcmt.yaml",
		".tfcmt.yml",
	}

	return findconfig.Find("", func(s string) bool {
		for _, file := range configFiles {
			if s == file {
				return true
			}
		}
		return false
	})
}
