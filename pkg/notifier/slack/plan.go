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

// Plan posts Slack message for terraform plan results
func (s *NotifyService) Plan(ctx context.Context, param *notifier.ParamExec) error {
	cfg := s.client.Config
	parser := cfg.Parser
	template := cfg.Template

	result := parser.Parse(param.CombinedOutput)
	if result.HasParseError {
		template = cfg.ParseErrorTemplate
	} else {
		if result.Error != nil {
			return result.Error
		}
		if result.Result == "" {
			return nil
		}
	}

	// Generate AI summary if summarizer is provided
	var aiSummary string
	if param.AISummarizer != nil {
		logrus.Info("AI summarizer enabled for plan, generating summary...")

		operationType := "plan"
		isSuccess := !result.HasError && param.ExitCode == 0

		planData := map[string]interface{}{
			"Result":                 result.Result,
			"CreatedResources":       result.CreatedResources,
			"UpdatedResources":       result.UpdatedResources,
			"DeletedResources":       result.DeletedResources,
			"ReplacedResources":      result.ReplacedResources,
			"MovedResources":         result.MovedResources,
			"ImportedResources":      result.ImportedResources,
			"HasDestroy":             result.HasDestroy,
			"HasNoChanges":           result.HasNoChanges,
			"HasError":               result.HasError,
			"Warning":                result.Warning,
			"ChangeOutsideTerraform": result.OutsideTerraform,
			"ExitCode":               param.ExitCode,
			"CombinedOutput":         param.CombinedOutput,
			"OperationType":          operationType,
			"IsSuccess":              isSuccess,
		}

		summary, err := param.AISummarizer.GenerateSummary(ctx, planData)
		if err != nil {
			logrus.WithError(err).Warn("failed to generate AI summary for plan")
		} else {
			logrus.WithField("length", len(summary)).Info("AI summary generated successfully for plan")
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
		ErrorMessages:          []string{},
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
	title := cfg.PlanTitle
	if title == "" {
		title = cfg.Title // Fallback to default title
	}
	message := cfg.PlanMessage
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

	// Only notify on failures if configured
	shouldNotify := false
	if cfg.NotifyOnPlanError && param.ExitCode != 0 {
		shouldNotify = true
		logrus.WithField("exit_code", param.ExitCode).Info("Plan failed, notifying Slack (notify_on_plan_error enabled)")
	}

	if !shouldNotify {
		logrus.WithFields(logrus.Fields{
			"exit_code":            param.ExitCode,
			"has_destroy":          result.HasDestroy,
			"notify_on_plan_error": cfg.NotifyOnPlanError,
		}).Debug("Skipping Slack notification (no critical issues or notification disabled)")
		return nil
	}

	// If using threads, send parent message then error details in thread
	if cfg.UseThreads && param.ExitCode != 0 {
		// Send parent summary message
		parentMessage := fullMessage
		if title != "" {
			parentMessage = fmt.Sprintf("*%s*\n\n", title)
		}
		if message != "" {
			parentMessage += fmt.Sprintf("%s\n\n", message)
		}
		parentMessage += "❌ Terraform plan failed. See thread for details."

		logrus.Info("Sending parent message to Slack")
		timestamp, err := s.postMessageAndGetTimestamp(ctx, parentMessage, nil)
		if err != nil {
			return err
		}

		threadMessage := fmt.Sprintf("```\n%s\n```", result.Result)
		logrus.WithField("parent_ts", timestamp).Info("Sending error details in thread")
		return s.postMessage(ctx, threadMessage, &timestamp)
	}
	return err
}

// postMessageWithBlocks posts a message with Slack blocks for rich formatting
func (s *NotifyService) postMessageWithBlocks(ctx context.Context, blocks []slackgo.Block, parentTS *string) error {
	cfg := s.client.Config

	options := []slackgo.MsgOption{
		slackgo.MsgOptionBlocks(blocks...),
	}

	if cfg.BotName != "" {
		options = append(options, slackgo.MsgOptionUsername(cfg.BotName))
	}

	if parentTS != nil && *parentTS != "" {
		options = append(options, slackgo.MsgOptionTS(*parentTS))
	}

	logE := logrus.WithFields(logrus.Fields{
		"channel_id": cfg.ChannelID,
		"bot_name":   cfg.BotName,
		"threaded":   parentTS != nil && *parentTS != "",
		"blocks":     len(blocks),
	})

	logE.Debug("posting message with blocks to Slack")

	_, timestamp, err := s.client.PostMessageContext(
		ctx,
		cfg.ChannelID,
		options...,
	)
	if err != nil {
		return fmt.Errorf("failed to post Slack message with blocks: %w", err)
	}

	logE.WithField("timestamp", timestamp).Info("successfully posted message with blocks to Slack")
	return nil
}
