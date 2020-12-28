package github

import (
	"context"
	"net/http"

	"github.com/mercari/tfnotify/terraform"
)

// NotifyService handles communication with the notification related
// methods of GitHub API
type NotifyService service

// Notify posts comment optimized for notifications
func (g *NotifyService) Notify(body string) (exit int, err error) {
	cfg := g.client.Config
	parser := g.client.Config.Parser
	template := g.client.Config.Template

	result := parser.Parse(body)
	if result.Error != nil {
		return result.ExitCode, result.Error
	}
	if result.Result == "" {
		return result.ExitCode, result.Error
	}

	_, isPlan := parser.(*terraform.PlanParser)
	if isPlan {
		if result.HasDestroy && cfg.WarnDestroy {
			// Notify destroy warning as a new comment before normal plan result
			if err = g.notifyDestoryWarning(body, result); err != nil {
				return result.ExitCode, err
			}
		}
		if cfg.PR.IsNumber() && cfg.ResultLabels.HasAnyLabelDefined() {
			err = g.removeResultLabels()
			if err != nil {
				return result.ExitCode, err
			}
			var labelToAdd string

			if result.HasAddOrUpdateOnly {
				labelToAdd = cfg.ResultLabels.AddOrUpdateLabel
			} else if result.HasDestroy {
				labelToAdd = cfg.ResultLabels.DestroyLabel
			} else if result.HasNoChanges {
				labelToAdd = cfg.ResultLabels.NoChangesLabel
			} else if result.HasPlanError {
				labelToAdd = cfg.ResultLabels.PlanErrorLabel
			}

			if labelToAdd != "" {
				_, _, err = g.client.API.IssuesAddLabels(
					context.Background(),
					cfg.PR.Number,
					[]string{labelToAdd},
				)
				if err != nil {
					return result.ExitCode, err
				}
			}
		}
	}

	template.SetValue(terraform.CommonTemplate{
		Title:        cfg.PR.Title,
		Message:      cfg.PR.Message,
		Result:       result.Result,
		Body:         body,
		Link:         cfg.CI,
		UseRawOutput: cfg.UseRawOutput,
	})
	body, err = template.Execute()
	if err != nil {
		return result.ExitCode, err
	}

	value := template.GetValue()

	if cfg.PR.IsNumber() {
		g.client.Comment.DeleteDuplicates(value.Title)
	}

	_, isApply := parser.(*terraform.ApplyParser)
	if isApply {
		prNumber, err := g.client.Commits.MergedPRNumber(cfg.PR.Revision)
		if err == nil {
			cfg.PR.Number = prNumber
		} else if !cfg.PR.IsNumber() {
			commits, err := g.client.Commits.List(cfg.PR.Revision)
			if err != nil {
				return result.ExitCode, err
			}
			lastRevision, _ := g.client.Commits.lastOne(commits, cfg.PR.Revision)
			cfg.PR.Revision = lastRevision
		}
	}

	return result.ExitCode, g.client.Comment.Post(body, PostOptions{
		Number:   cfg.PR.Number,
		Revision: cfg.PR.Revision,
	})
}

func (g *NotifyService) notifyDestoryWarning(body string, result terraform.ParseResult) error {
	cfg := g.client.Config
	destroyWarningTemplate := g.client.Config.DestroyWarningTemplate
	destroyWarningTemplate.SetValue(terraform.CommonTemplate{
		Title:        cfg.PR.DestroyWarningTitle,
		Message:      cfg.PR.DestroyWarningMessage,
		Result:       result.Result,
		Body:         body,
		Link:         cfg.CI,
		UseRawOutput: cfg.UseRawOutput,
	})
	body, err := destroyWarningTemplate.Execute()
	if err != nil {
		return err
	}

	return g.client.Comment.Post(body, PostOptions{
		Number:   cfg.PR.Number,
		Revision: cfg.PR.Revision,
	})
}

func (g *NotifyService) removeResultLabels() error {
	cfg := g.client.Config
	labels, _, err := g.client.API.IssuesListLabels(context.Background(), cfg.PR.Number, nil)
	if err != nil {
		return err
	}

	for _, l := range labels {
		labelText := l.GetName()
		if cfg.ResultLabels.IsResultLabel(labelText) {
			resp, err := g.client.API.IssuesRemoveLabel(context.Background(), cfg.PR.Number, labelText)
			// Ignore 404 errors, which are from the PR not having the label
			if err != nil && resp.StatusCode != http.StatusNotFound {
				return err
			}
		}
	}

	return nil
}
