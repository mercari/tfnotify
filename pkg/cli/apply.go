package cli

import (
	"context"
	"os"

	"github.com/mercari/tfnotify/v1/pkg/ai"
	"github.com/mercari/tfnotify/v1/pkg/controller"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func cmdApply(ctx context.Context, cmd *cli.Command) error {
	logLevel := cmd.String("log-level")
	setLogLevel(logLevel)

	cfg, err := newConfig(cmd)
	if err != nil {
		return err
	}

	if logLevel == "" {
		logLevel = cfg.Log.Level
		setLogLevel(logLevel)
	}

	if err := parseOpts(cmd, &cfg, os.Environ()); err != nil {
		return err
	}

	// Configure AI summary if enabled via flags
	if cmd.Bool("summary") {
		cfg.AISummary.Enabled = true
		if provider := cmd.String("summary-provider"); provider != "" {
			cfg.AISummary.Provider = provider
		}
		if model := cmd.String("summary-model"); model != "" {
			cfg.AISummary.Model = model
		}
		if template := cmd.String("summary-template"); template != "" {
			cfg.AISummary.TemplateFile = template
		}
	}

	// Get session ID if provided
	sessionID := cmd.String("session-id")

	t := &controller.Controller{
		Config:             cfg,
		Parser:             terraform.NewApplyParser(),
		Template:           terraform.NewApplyTemplate(cfg.Terraform.Apply.Template),
		ParseErrorTemplate: terraform.NewApplyParseErrorTemplate(cfg.Terraform.Apply.WhenParseError.Template),
	}

	// Add AI summarizer if enabled
	if cfg.AISummary.Enabled {
		apiKey := os.Getenv("LITELLM_API_KEY")
		switch cfg.AISummary.Provider {
		case "anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "devin":
			apiKey = os.Getenv("DEVIN_API_KEY")
		}

		summarizer := ai.NewSummarizer(ai.SummaryConfig{
			Enabled:       cfg.AISummary.Enabled,
			Provider:      cfg.AISummary.Provider,
			APIKey:        apiKey,
			Model:         cfg.AISummary.Model,
			Template:      cfg.AISummary.Template,
			TemplateFile:  cfg.AISummary.TemplateFile,
			MaxTokens:     cfg.AISummary.MaxTokens,
			SessionID:     sessionID,
			OperationType: "apply",
		})
		t.AISummarizer = summarizer
		logrus.WithFields(logrus.Fields{
			"provider": cfg.AISummary.Provider,
			"model":    cfg.AISummary.Model,
		}).Info("AI summary enabled for apply")
	}

	args := cmd.Args()

	return t.Apply(ctx, controller.Command{
		Cmd:  args.First(),
		Args: args.Tail(),
	})
}
