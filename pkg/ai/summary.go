package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

// Template file paths relative to project root
const (
	TemplatePlanSuccess  = "./templates/plan-success.md"
	TemplatePlanFailure  = "./templates/plan-failure.md"
	TemplateApplySuccess = "./templates/apply-success.md"
	TemplateApplyFailure = "./templates/apply-failure.md"
	TemplateDefault      = "./templates/default.md"
)

// SummaryConfig holds configuration for AI summary generation
type SummaryConfig struct {
	Enabled       bool
	Provider      string // "openai", "anthropic", "google", "litellm", "devin"
	APIKey        string
	Model         string
	Template      string
	TemplateFile  string
	MaxTokens     int
	OperationType string   // "plan" or "apply"
	IsSuccess     bool     // true for success, false for failure
	SessionID     string   // Optional: existing Devin session ID to use
	PlaybookIDs   []string // Optional: Devin playbook IDs to guide behavior (for Devin provider only)
}

// PlanData contains the terraform plan data to be summarized
type PlanData struct {
	Result                 string
	CreatedResources       []string
	UpdatedResources       []string
	DeletedResources       []string
	ReplacedResources      []string
	MovedResources         []string
	ImportedResources      []string
	HasDestroy             bool
	HasError               bool
	Warning                string
	ChangeOutsideTerraform string
	ErrorMessages          []string
	ExitCode               int
	CombinedOutput         string
	PRNumber               int
	RepoOwner              string
	RepoName               string
}

// Summarizer generates AI summaries of Terraform plans
type Summarizer struct {
	config SummaryConfig
	logger *logrus.Entry
}

// NewSummarizer creates a new AI summarizer
func NewSummarizer(config SummaryConfig) *Summarizer {
	return &Summarizer{
		config: config,
		logger: logrus.WithField("component", "ai-summarizer"),
	}
}

// GenerateSummary creates an AI-powered summary of the plan
func (s *Summarizer) GenerateSummary(ctx context.Context, data interface{}) (string, error) {
	s.logger.Info("GenerateSummary called")

	if !s.config.Enabled {
		s.logger.Warn("AI summary disabled in config")
		return "", nil
	}

	s.logger.WithField("provider", s.config.Provider).Info("AI summary enabled")

	// Convert interface{} to PlanData
	planDataMap, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid data type, expected map[string]interface{}")
	}

	planData := &PlanData{
		Result:                 getString(planDataMap, "Result"),
		CreatedResources:       getStringSlice(planDataMap, "CreatedResources"),
		UpdatedResources:       getStringSlice(planDataMap, "UpdatedResources"),
		DeletedResources:       getStringSlice(planDataMap, "DeletedResources"),
		ReplacedResources:      getStringSlice(planDataMap, "ReplacedResources"),
		ImportedResources:      getStringSlice(planDataMap, "ImportedResources"),
		HasDestroy:             getBool(planDataMap, "HasDestroy"),
		HasError:               getBool(planDataMap, "HasError"),
		Warning:                getString(planDataMap, "Warning"),
		ChangeOutsideTerraform: getString(planDataMap, "ChangeOutsideTerraform"),
		ErrorMessages:          getStringSlice(planDataMap, "ErrorMessages"),
		ExitCode:               getInt(planDataMap, "ExitCode"),
		CombinedOutput:         getString(planDataMap, "CombinedOutput"),
		PRNumber:               getInt(planDataMap, "PRNumber"),
		RepoOwner:              getString(planDataMap, "RepoOwner"),
		RepoName:               getString(planDataMap, "RepoName"),
	}

	// Extract OperationType and IsSuccess for template selection
	if operationType := getString(planDataMap, "OperationType"); operationType != "" {
		s.config.OperationType = operationType
	}
	if isSuccess, ok := planDataMap["IsSuccess"].(bool); ok {
		s.config.IsSuccess = isSuccess
	}

	s.logger.WithFields(logrus.Fields{
		"created":        len(planData.CreatedResources),
		"updated":        len(planData.UpdatedResources),
		"deleted":        len(planData.DeletedResources),
		"replaced":       len(planData.ReplacedResources),
		"has_error":      planData.HasError,
		"error_count":    len(planData.ErrorMessages),
		"exit_code":      planData.ExitCode,
		"operation_type": s.config.OperationType,
		"is_success":     s.config.IsSuccess,
	}).Info("plan data parsed")

	// Load template
	tmplContent, err := s.loadTemplate()
	if err != nil {
		return "", fmt.Errorf("load template: %w", err)
	}
	s.logger.WithField("template_length", len(tmplContent)).Debug("template loaded")

	// Generate prompt from template
	prompt, err := s.renderPrompt(tmplContent, planData)
	if err != nil {
		return "", fmt.Errorf("render prompt: %w", err)
	}

	s.logger.WithField("prompt_length", len(prompt)).Info("AI prompt generated, calling provider...")

	// Call AI provider
	summary, err := s.callAIProvider(ctx, s.config.SessionID, prompt)
	if err != nil {
		return "", fmt.Errorf("call AI provider: %w", err)
	}

	s.logger.WithField("summary_length", len(summary)).Info("AI summary received")
	return summary, nil
}

// Helper functions for type conversion
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]string); ok {
			return slice
		}
	}
	return nil
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

// loadTemplate loads the template from file or uses default
func (s *Summarizer) loadTemplate() (string, error) {
	// 1. If explicit template file is provided, use it
	if s.config.TemplateFile != "" {
		content, err := os.ReadFile(s.config.TemplateFile)
		if err != nil {
			return "", fmt.Errorf("read template file: %w", err)
		}
		return string(content), nil
	}

	// 2. If template string is provided, use it
	if s.config.Template != "" {
		return s.config.Template, nil
	}

	// 3. Auto-select template based on operation context
	templatePath := s.selectTemplateByContext()
	return s.loadTemplateFromPath(templatePath)
}

// selectTemplateByContext automatically selects the appropriate template
func (s *Summarizer) selectTemplateByContext() string {
	switch s.config.OperationType {
	case "plan":
		if s.config.IsSuccess {
			return TemplatePlanSuccess
		}
		return TemplatePlanFailure
	case "apply":
		if s.config.IsSuccess {
			return TemplateApplySuccess
		}
		return TemplateApplyFailure
	default:
		return TemplateDefault
	}
}

// loadTemplateFromPath loads a template from the templates directory
func (s *Summarizer) loadTemplateFromPath(templatePath string) (string, error) {
	// Try to find the template file relative to the executable
	// This allows templates to work in different deployment scenarios

	// Try multiple locations in order:
	locationsToTry := []string{}

	// 1. Current working directory
	locationsToTry = append(locationsToTry, templatePath)

	// 2. Relative to executable location
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		locationsToTry = append(locationsToTry, filepath.Join(execDir, templatePath))
	}

	// 3. Strip leading "./" and try again (for cases where it's already in templates/)
	cleanPath := templatePath
	if len(cleanPath) > 2 && cleanPath[:2] == "./" {
		cleanPath = cleanPath[2:]
		locationsToTry = append(locationsToTry, cleanPath)

		// Also try relative to executable
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			locationsToTry = append(locationsToTry, filepath.Join(execDir, cleanPath))
		}
	}

	// 4. Search upward from current directory to find templates folder
	// This handles cases where binary is called from a subdirectory
	if cwd, err := os.Getwd(); err == nil {
		currentDir := cwd
		// Search up to 10 levels up (reasonable limit)
		for i := 0; i < 10; i++ {
			// Try with original path
			candidatePath := filepath.Join(currentDir, templatePath)
			locationsToTry = append(locationsToTry, candidatePath)

			// Try with clean path (without ./)
			if cleanPath != templatePath {
				candidatePath = filepath.Join(currentDir, cleanPath)
				locationsToTry = append(locationsToTry, candidatePath)
			}

			// Move up one directory
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir {
				// Reached root directory
				break
			}
			currentDir = parentDir
		}
	}

	// Try each location
	for _, path := range locationsToTry {
		content, err := os.ReadFile(path)
		if err == nil {
			s.logger.WithField("template", path).Debug("loaded template")
			return string(content), nil
		}
		s.logger.WithFields(logrus.Fields{
			"path":  path,
			"error": err.Error(),
		}).Debug("template not found at path")
	}

	return "", fmt.Errorf("template file not found: %s (tried %d locations)", templatePath, len(locationsToTry))
}

// renderPrompt renders the template with plan data
func (s *Summarizer) renderPrompt(tmplContent string, data *PlanData) (string, error) {
	tmpl, err := template.New("ai-prompt").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	prompt := buf.String()

	// Debug logging to verify error data is in prompt
	s.logger.WithFields(logrus.Fields{
		"prompt_length":          len(prompt),
		"has_error_messages":     len(data.ErrorMessages) > 0,
		"error_message_count":    len(data.ErrorMessages),
		"exit_code":              data.ExitCode,
		"has_combined_output":    len(data.CombinedOutput) > 0,
		"combined_output_length": len(data.CombinedOutput),
	}).Debug("prompt rendered with error data")

	// Log a preview of the prompt if errors are present
	if len(data.ErrorMessages) > 0 || data.ExitCode != 0 {
		preview := prompt
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		s.logger.WithField("prompt_preview", preview).Debug("prompt preview with errors")
	}

	return prompt, nil
}

// callAIProvider calls the configured AI provider
func (s *Summarizer) callAIProvider(ctx context.Context, session_id string, prompt string) (string, error) {
	switch s.config.Provider {
	case "openai":
		return s.callOpenAI(ctx, prompt)
	case "anthropic":
		return s.callAnthropic(ctx, prompt)
	case "litellm":
		return s.callLiteLLM(ctx, prompt)
	case "devin":
		return s.callDevin(ctx, session_id, prompt)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", s.config.Provider)
	}
}
