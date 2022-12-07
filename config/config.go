package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config is for tfnotify config structure
type Config struct {
	CI        string    `yaml:"ci"`
	Notifier  Notifier  `yaml:"notifier"`
	Terraform Terraform `yaml:"terraform"`

	path string
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
	Default      Default `yaml:"default"`
	Fmt          Fmt     `yaml:"fmt"`
	Plan         Plan    `yaml:"plan"`
	Apply        Apply   `yaml:"apply"`
	Drift        Drift   `yaml:"drift"`
	UseRawOutput bool    `yaml:"use_raw_output,omitempty"`
}

// Default is a default setting for terraform commands
type Default struct {
	Template string `yaml:"template"`
}

// Fmt is a terraform fmt config
type Fmt struct {
	Template string `yaml:"template"`
}

// Drift is a terraform drift config
type Drift struct {
	Template string `yaml:"template"`
}

// Plan is a terraform plan config
type Plan struct {
	Template            string              `yaml:"template"`
	WhenAddOrUpdateOnly WhenAddOrUpdateOnly `yaml:"when_add_or_update_only,omitempty"`
	WhenDestroy         WhenDestroy         `yaml:"when_destroy,omitempty"`
	WhenNoChanges       WhenNoChanges       `yaml:"when_no_changes,omitempty"`
	WhenPlanError       WhenPlanError       `yaml:"when_plan_error,omitempty"`
}

// WhenAddOrUpdateOnly is a configuration to notify the plan result contains new or updated in place resources
type WhenAddOrUpdateOnly struct {
	Label string `yaml:"label,omitempty"`
}

// WhenDestroy is a configuration to notify the plan result contains destroy operation
type WhenDestroy struct {
	Label    string `yaml:"label,omitempty"`
	Template string `yaml:"template,omitempty"`
}

// WhenNoChange is a configuration to add a label when the plan result contains no change
type WhenNoChanges struct {
	Label string `yaml:"label,omitempty"`
}

// WhenPlanError is a configuration to notify the plan result returns an error
type WhenPlanError struct {
	Label string `yaml:"label,omitempty"`
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
	raw, _ := ioutil.ReadFile(cfg.path)
	return yaml.Unmarshal(raw, cfg)
}

// Validation validates config file
func (cfg *Config) Validation() error {
	switch strings.ToLower(cfg.CI) {
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
		return fmt.Errorf("%s: not supported yet", cfg.CI)
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
			return fmt.Errorf("Typetalk topic id is missing")
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
	var files []string
	if file == "" {
		files = []string{
			"tfnotify.yaml",
			"tfnotify.yml",
			".tfnotify.yaml",
			".tfnotify.yml",
		}
	} else {
		files = []string{file}
	}
	for _, file := range files {
		_, err := os.Stat(file)
		if err == nil {
			return file, nil
		}
	}
	return "", errors.New("config for tfnotify is not found at all")
}
