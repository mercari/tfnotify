package github

import (
	"context"
	"fmt"

	"github.com/mercari/tfnotify/v1/pkg/mask"
	"github.com/mercari/tfnotify/v1/pkg/notifier"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/sirupsen/logrus"
)

// Plan posts comment optimized for notifications
func (g *NotifyService) Plan(ctx context.Context, param *notifier.ParamExec) error { //nolint:cyclop
	cfg := g.client.Config
	parser := g.client.Config.Parser
	template := g.client.Config.Template
	var errMsgs []string

	if cfg.PR.Number == 0 && cfg.PR.Revision != "" {
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

	if cfg.PR.IsNumber() && cfg.ResultLabels.HasAnyLabelDefined() {
		errMsgs = append(errMsgs, g.UpdateLabels(ctx, result)...)
	}

	if cfg.IgnoreWarning {
		result.Warning = ""
	}

	// Generate AI summary if summarizer is provided
	var aiSummary string
	if param.AISummarizer != nil {
		logrus.Info("AI summarizer enabled, generating summary...")

		// Determine operation type and success status for template selection
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
		}).Debug("plan data for AI summary")

		summary, err := param.AISummarizer.GenerateSummary(ctx, planData)
		if err != nil {
			logrus.WithError(err).Warn("failed to generate AI summary")
		} else {
			logrus.WithField("length", len(summary)).Info("AI summary generated successfully")
			aiSummary = summary
		}
	} else {
		logrus.Debug("AI summarizer not configured, skipping AI summary generation")
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
		MovedResources:         result.MovedResources,
		ImportedResources:      result.ImportedResources,
		AISummary:              aiSummary,
		SummaryEnabled:         param.AISummarizer != nil,
	})
	body, err := template.Execute()
	if err != nil {
		return err
	}

	logE := logrus.WithFields(logrus.Fields{
		"program": "tfnotify",
	})

	embeddedComment, err := getEmbeddedComment(cfg, param.CIName, true)
	if err != nil {
		return err
	}
	logE.WithFields(logrus.Fields{
		"comment": embeddedComment,
	}).Debug("embedded HTML comment")
	// embed HTML tag to hide old comments
	body += embeddedComment

	body = mask.Mask(body, g.client.Config.Masks)

	if cfg.Patch && cfg.PR.Number != 0 {
		logE.Debug("try patching")
		comments, err := g.client.Comment.List(ctx, cfg.Owner, cfg.Repo, cfg.PR.Number)
		if err != nil {
			logE.WithError(err).Debug("list comments")
			// Post a new comment instead of patching an existing comment
			if err := g.client.Comment.Post(ctx, body, &PostOptions{
				Number:   cfg.PR.Number,
				Revision: cfg.PR.Revision,
			}); err != nil {
				return fmt.Errorf("post a comment: %w", err)
			}
			return nil
		}
		logE.WithField("size", len(comments)).Debug("list comments")
		comment := g.getPatchedComment(logE, comments, cfg.Vars["target"])
		if comment != nil {
			if comment.Body == body {
				logE.Debug("comment isn't changed")
				return nil
			}
			logE.WithField("comment_id", comment.DatabaseID).Debug("patch a comment")
			if err := g.client.Comment.Patch(ctx, body, int64(comment.DatabaseID)); err != nil {
				return fmt.Errorf("patch a comment: %w", err)
			}
			return nil
		}
	}

	if result.HasNoChanges && result.Warning == "" && len(errMsgs) == 0 && cfg.SkipNoChanges {
		logE.Debug("skip posting a comment because there is no change")
		return nil
	}

	logE.Debug("create a comment")
	if err := g.client.Comment.Post(ctx, body, &PostOptions{
		Number:   cfg.PR.Number,
		Revision: cfg.PR.Revision,
	}); err != nil {
		return fmt.Errorf("post a comment: %w", err)
	}
	return nil
}
