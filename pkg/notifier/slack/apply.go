package slack

import (
	"context"
	"fmt"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/mercari/tfnotify/v1/pkg/mask"
	"github.com/mercari/tfnotify/v1/pkg/notifier"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/sirupsen/logrus"
	slackgo "github.com/slack-go/slack"
)

// Apply posts Slack message for terraform apply results
func (s *NotifyService) Apply(ctx context.Context, param *notifier.ParamExec) error {
	cfg := s.client.Config
	parser := cfg.Parser
	template := cfg.Template
	var errMsgs []string

	result := parser.Parse(param.CombinedOutput)
	if result.HasParseError {
		template = cfg.ParseErrorTemplate
	} else {
		if result.Error != nil {
			return result.Error
		}
		if result.Result == "" {
			return result.Error
		}
	}

	// Generate AI summary if summarizer is provided
	var aiSummary string
	// Only generate AI summary for failed applies (exit code != 0)
	if param.AISummarizer != nil && param.ExitCode != 0 {
		logrus.Info("AI summarizer enabled for apply failure, generating summary...")

		operationType := "apply"
		isSuccess := !result.HasError && param.ExitCode == 0

		applyData := map[string]interface{}{
			"Result":                 result.Result,
			"CreatedResources":       result.CreatedResources,
			"UpdatedResources":       result.UpdatedResources,
			"DeletedResources":       result.DeletedResources,
			"ReplacedResources":      result.ReplacedResources,
			"MovedResources":         result.MovedResources,
			"ImportedResources":      result.ImportedResources,
			"HasDestroy":             result.HasDestroy,
			"HasError":               result.HasError,
			"Warning":                result.Warning,
			"ChangeOutsideTerraform": result.OutsideTerraform,
			"ErrorMessages":          errMsgs,
			"ExitCode":               param.ExitCode,
			"CombinedOutput":         param.CombinedOutput,
			"OperationType":          operationType,
			"IsSuccess":              isSuccess,
		}

		summary, err := param.AISummarizer.GenerateSummary(ctx, applyData)
		if err != nil {
			logrus.WithError(err).Warn("failed to generate AI summary for apply")
		} else {
			logrus.WithField("length", len(summary)).Info("AI summary generated successfully for apply")
			aiSummary = summary
		}
	}

	template.SetValue(terraform.CommonTemplate{
		Result:                 result.Result,
		ChangedResult:          result.ChangedResult,
		ChangeOutsideTerraform: result.OutsideTerraform,
		Warning:                result.Warning,
		HasDestroy:             result.HasDestroy,
		HasError:               result.HasError,
		Link:                   cfg.CI.Link,
		UseRawOutput:           cfg.UseRawOutput,
		Vars:                   cfg.Vars,
		Templates:              cfg.Templates,
		Stdout:                 param.Stdout,
		Stderr:                 param.Stderr,
		CombinedOutput:         param.CombinedOutput,
		ExitCode:               param.ExitCode,
		ErrorMessages:          errMsgs,
		CreatedResources:       result.CreatedResources,
		UpdatedResources:       result.UpdatedResources,
		DeletedResources:       result.DeletedResources,
		ReplacedResources:      result.ReplacedResources,
		AISummary:              aiSummary,
		SummaryEnabled:         param.AISummarizer != nil,
	})

	body, err := template.Execute()
	if err != nil {
		return err
	}

	// Mask sensitive information
	body = mask.Mask(body, []*config.Mask{}) // TODO: add masks support if needed

	result.Result = mask.Mask(result.Result, []*config.Mask{}) // TODO: add masks support if needed

	// Prepare the message with title if configured
	var fullMessage string
	title := cfg.ApplyTitle
	if title == "" {
		title = cfg.Title // Fallback to default title
	}
	message := cfg.ApplyMessage
	if message == "" {
		message = cfg.Message // Fallback to default message
	}

	// Build the full message
	if title != "" {
		fullMessage = fmt.Sprintf("*%s*\n\n", title)
	}
	if message != "" {
		fullMessage += fmt.Sprintf("%s\n\n", message)
	}
	fullMessage += body

	// Only notify on apply failures if configured
	if !cfg.NotifyOnApplyError || param.ExitCode == 0 {
		logrus.WithFields(logrus.Fields{
			"exit_code":             param.ExitCode,
			"notify_on_apply_error": cfg.NotifyOnApplyError,
		}).Debug("Skipping Slack notification (apply succeeded or notification disabled)")
		return nil
	}

	logrus.Info("Apply failed, posting to Slack channel")

	// If using threads, send parent message then error details in thread
	if cfg.UseThreads {
		// Send parent summary message
		parentMessage := ""
		if title != "" {
			parentMessage = fmt.Sprintf("*%s*\n\n", title)
		}
		if message != "" {
			parentMessage += fmt.Sprintf("%s\n\n", message)
		}
		parentMessage += "❌ Terraform apply failed. See thread for details."

		logrus.Info("Sending parent message to Slack")
		timestamp, err := s.postMessageAndGetTimestamp(ctx, parentMessage, nil)
		if err != nil {
			return err
		}

		threadMessage := fmt.Sprintf("```\n%s\n```", result.Result)
		logrus.WithField("parent_ts", timestamp).Info("Sending error details in thread")
		return s.postMessage(ctx, threadMessage, &timestamp)
	}

	// Send single message with full content
	_, err = s.postMessageAndGetTimestamp(ctx, fullMessage, nil)
	return err
}

// postMessage posts a message to the configured Slack channel
func (s *NotifyService) postMessage(ctx context.Context, text string, parentTS *string) error {
	_, err := s.postMessageAndGetTimestamp(ctx, text, parentTS)
	return err
}

// postMessageAndGetTimestamp posts a message and returns the timestamp for threading
func (s *NotifyService) postMessageAndGetTimestamp(ctx context.Context, text string, parentTS *string) (string, error) {
	cfg := s.client.Config

	// Build message options
	options := []slackgo.MsgOption{
		slackgo.MsgOptionText(text, false),
	}

	// Set bot name if configured
	if cfg.BotName != "" {
		options = append(options, slackgo.MsgOptionUsername(cfg.BotName))
	}

	// If parentTS is provided, post as a threaded reply
	if parentTS != nil && *parentTS != "" {
		options = append(options, slackgo.MsgOptionTS(*parentTS))
	}

	logE := logrus.WithFields(logrus.Fields{
		"channel_id": cfg.ChannelID,
		"bot_name":   cfg.BotName,
		"threaded":   parentTS != nil && *parentTS != "",
	})

	logE.Debug("posting message to Slack")

	_, timestamp, err := s.client.PostMessageContext(
		ctx,
		cfg.ChannelID,
		options...,
	)
	if err != nil {
		return "", fmt.Errorf("failed to post Slack message: %w", err)
	}

	logE.WithField("timestamp", timestamp).Info("successfully posted message to Slack")
	return timestamp, nil
}

// PostThreadMessage posts a message as a threaded reply
func (s *NotifyService) PostThreadMessage(ctx context.Context, text string, parentTS string) error {
	return s.postMessage(ctx, text, &parentTS)
}
