package github

import (
	"context"
	"fmt"

	"github.com/mercari/tfnotify/v1/pkg/mask"
	"github.com/mercari/tfnotify/v1/pkg/notifier"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/sirupsen/logrus"
)

// Apply posts comment optimized for notifications
func (g *NotifyService) Apply(ctx context.Context, param *notifier.ParamExec) error {
	cfg := g.client.Config
	parser := g.client.Config.Parser
	template := g.client.Config.Template
	var errMsgs []string

	if cfg.PR.Number == 0 {
		if prNumber, err := g.client.Commits.PRNumber(ctx, cfg.PR.Revision); err == nil {
			cfg.PR.Number = prNumber
		}
	}

	result := parser.Parse(param.CombinedOutput)
	if result.HasParseError {
		template = g.client.Config.ParseErrorTemplate
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

		// Determine operation type and success status for template selection
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
			"PRNumber":               cfg.PR.Number,
		}
		logrus.WithFields(logrus.Fields{
			"created":        len(result.CreatedResources),
			"updated":        len(result.UpdatedResources),
			"deleted":        len(result.DeletedResources),
			"replaced":       len(result.ReplacedResources),
			"has_error":      result.HasError,
			"error_count":    len(errMsgs),
			"exit_code":      param.ExitCode,
			"operation_type": operationType,
			"is_success":     isSuccess,
		}).Debug("apply data for AI summary")

		summary, err := param.AISummarizer.GenerateSummary(ctx, applyData)
		if err != nil {
			logrus.WithError(err).Warn("failed to generate AI summary for apply")
		} else {
			logrus.WithField("length", len(summary)).Info("AI summary generated successfully for apply")
			aiSummary = summary
		}
	} else {
		logrus.Debug("AI summarizer not configured, skipping AI summary generation for apply")
	}

	template.SetValue(terraform.CommonTemplate{
		Result:                 result.Result,
		ChangedResult:          result.ChangedResult,
		ChangeOutsideTerraform: result.OutsideTerraform,
		Warning:                result.Warning,
		HasDestroy:             result.HasDestroy,
		HasError:               result.HasError,
		Link:                   cfg.CI,
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
	})
	body, err := template.Execute()
	if err != nil {
		return err
	}

	logE := logrus.WithFields(logrus.Fields{
		"program": "tfnotify",
	})

	embeddedComment, err := getEmbeddedComment(cfg, param.CIName, false)
	if err != nil {
		return err
	}
	logE.WithFields(logrus.Fields{
		"comment": embeddedComment,
	}).Debug("embedded HTML comment")
	// embed HTML tag to hide old comments
	body += embeddedComment

	body = mask.Mask(body, g.client.Config.Masks)

	logE.Debug("create a comment")
	if err := g.client.Comment.Post(ctx, body, &PostOptions{
		Number:   cfg.PR.Number,
		Revision: cfg.PR.Revision,
	}); err != nil {
		return fmt.Errorf("post a comment: %w", err)
	}
	return nil
}
