package localfile

import (
	"context"
	"fmt"

	"github.com/mercari/tfnotify/v1/pkg/mask"
	"github.com/mercari/tfnotify/v1/pkg/notifier"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/sirupsen/logrus"
)

// Plan posts comment optimized for notifications
func (g *NotifyService) Plan(ctx context.Context, param *notifier.ParamExec) error {
	cfg := g.client.Config
	parser := g.client.Config.Parser
	template := g.client.Config.Template
	var errMsgs []string

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

	logE := logrus.WithFields(logrus.Fields{
		"program": "tfnotify",
	})
	if !cfg.DisableLabel {
		logE.Debugf("updating labels")
		errMsgs = append(errMsgs, g.client.labeler.UpdateLabels(ctx, result)...)
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

	body = mask.Mask(body, g.client.Config.Masks)

	logE.Debug("write a plan output to a file")
	if err := g.client.Output.WriteToFile(body, cfg.OutputFile); err != nil {
		return fmt.Errorf("write a plan output to a file: %w", err)
	}

	return nil
}
