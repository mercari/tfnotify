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
