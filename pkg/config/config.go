package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"

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
	AISummary          AISummary         `json:"ai_summary,omitempty" yaml:"ai_summary"`
}

type Mask struct {
	Type   string
	Value  string
	Regexp *regexp.Regexp
}

// AISummary configuration for AI-powered summary generation
type AISummary struct {
	Enabled      bool     `json:"enabled,omitempty" yaml:"enabled"`
	Provider     string   `json:"provider,omitempty" yaml:"provider"`
	Model        string   `json:"model,omitempty" yaml:"model"`
	Template     string   `json:"template,omitempty" yaml:"template"`
	TemplateFile string   `json:"template_file,omitempty" yaml:"template_file"`
	MaxTokens    int      `json:"max_tokens,omitempty" yaml:"max_tokens"`
	PlaybookIDs  []string `json:"playbook_ids,omitempty" yaml:"playbook_ids"`
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
	// Format string
}

// Terraform represents terraform configurations
type Terraform struct {
	Plan         Plan  `json:"plan,omitempty"`
	Apply        Apply `json:"apply,omitempty"`
	UseRawOutput bool  `json:"use_raw_output,omitempty" yaml:"use_raw_output"`
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
	Template       string         `json:"template,omitempty"`
	WhenParseError WhenParseError `json:"when_parse_error,omitempty" yaml:"when_parse_error"`
}

// LoadFile binds the config file to Config structure
func (c *Config) LoadFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%s: no config file", path)
	}
	raw, _ := os.ReadFile(path)
	return yaml.Unmarshal(raw, c)
}

// Validate validates config file
func (c *Config) Validate() error {
	if c.Output != "" {
		return nil
	}
	if c.CI.Owner == "" {
		return errors.New("repository owner is missing")
	}

	if c.CI.Repo == "" {
		return errors.New("repository name is missing")
	}

	if c.CI.SHA == "" && c.CI.PRNumber <= 0 {
		return errors.New("pull request number or SHA (revision) is needed")
	}
	return nil
}

// Find returns config path
func (c *Config) Find(file string) (string, error) {
	if file != "" {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
		return "", errors.New("config for tfnotify is not found at all")
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get a current directory path: %w", err)
	}
	if p := findconfig.Find(wd, findconfig.Exist, "tfnotify.yaml", "tfnotify.yml", ".tfnotify.yaml", ".tfnotify.yml"); p != "" {
		return p, nil
	}
	return "", nil
}
