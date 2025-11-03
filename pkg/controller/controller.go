package controller

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/mercari/tfnotify/v1/pkg/notifier"
	"github.com/mercari/tfnotify/v1/pkg/notifier/github"
	"github.com/mercari/tfnotify/v1/pkg/notifier/localfile"
	tmpl "github.com/mercari/tfnotify/v1/pkg/template"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
)

type Controller struct {
	Config             config.Config
	Parser             terraform.Parser
	Template           *terraform.Template
	ParseErrorTemplate *terraform.Template
	AISummarizer       AISummarizer
}

type AISummarizer interface {
	GenerateSummary(ctx context.Context, data interface{}) (string, error)
}

type Command struct {
	Cmd  string
	Args []string
}

func (c *Controller) renderTemplate(tpl string) (string, error) {
	tmpl, err := template.New("_").Funcs(tmpl.TxtFuncMap()).Parse(tpl)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, map[string]any{
		"Vars": c.Config.Vars,
	}); err != nil {
		return "", fmt.Errorf("render a label template: %w", err)
	}
	return buf.String(), nil
}

func (c *Controller) renderGitHubLabels() (github.ResultLabels, error) { //nolint:cyclop
	labels := github.ResultLabels{
		AddOrUpdateLabelColor: c.Config.Terraform.Plan.WhenAddOrUpdateOnly.Color,
		DestroyLabelColor:     c.Config.Terraform.Plan.WhenDestroy.Color,
		NoChangesLabelColor:   c.Config.Terraform.Plan.WhenNoChanges.Color,
		PlanErrorLabelColor:   c.Config.Terraform.Plan.WhenPlanError.Color,
	}

	target, ok := c.Config.Vars["target"]
	if !ok {
		target = ""
	}

	if labels.AddOrUpdateLabelColor == "" {
		labels.AddOrUpdateLabelColor = "1d76db" // blue
	}
	if labels.DestroyLabelColor == "" {
		labels.DestroyLabelColor = "d93f0b" // red
	}
	if labels.NoChangesLabelColor == "" {
		labels.NoChangesLabelColor = "0e8a16" // green
	}

	if !c.Config.Terraform.Plan.WhenAddOrUpdateOnly.DisableLabel {
		if c.Config.Terraform.Plan.WhenAddOrUpdateOnly.Label == "" {
			if target == "" {
				labels.AddOrUpdateLabel = "add-or-update"
			} else {
				labels.AddOrUpdateLabel = target + "/add-or-update"
			}
		} else {
			addOrUpdateLabel, err := c.renderTemplate(c.Config.Terraform.Plan.WhenAddOrUpdateOnly.Label)
			if err != nil {
				return labels, err
			}
			labels.AddOrUpdateLabel = addOrUpdateLabel
		}
	}

	if !c.Config.Terraform.Plan.WhenDestroy.DisableLabel {
		if c.Config.Terraform.Plan.WhenDestroy.Label == "" {
			if target == "" {
				labels.DestroyLabel = "destroy"
			} else {
				labels.DestroyLabel = target + "/destroy"
			}
		} else {
			destroyLabel, err := c.renderTemplate(c.Config.Terraform.Plan.WhenDestroy.Label)
			if err != nil {
				return labels, err
			}
			labels.DestroyLabel = destroyLabel
		}
	}

	if !c.Config.Terraform.Plan.WhenNoChanges.DisableLabel {
		if c.Config.Terraform.Plan.WhenNoChanges.Label == "" {
			if target == "" {
				labels.NoChangesLabel = "no-changes"
			} else {
				labels.NoChangesLabel = target + "/no-changes"
			}
		} else {
			nochangesLabel, err := c.renderTemplate(c.Config.Terraform.Plan.WhenNoChanges.Label)
			if err != nil {
				return labels, err
			}
			labels.NoChangesLabel = nochangesLabel
		}
	}

	if !c.Config.Terraform.Plan.WhenPlanError.DisableLabel {
		planErrorLabel, err := c.renderTemplate(c.Config.Terraform.Plan.WhenPlanError.Label)
		if err != nil {
			return labels, err
		}
		labels.PlanErrorLabel = planErrorLabel
	}

	return labels, nil
}

func (c *Controller) getPlanNotifier(ctx context.Context) (notifier.Notifier, error) {
	labels := github.ResultLabels{}
	if !c.Config.Terraform.Plan.DisableLabel {
		a, err := c.renderGitHubLabels()
		if err != nil {
			return nil, err
		}
		labels = a
	}
	var gh *github.NotifyService
	if !c.Config.Terraform.Plan.DisableLabel || c.Config.Output == "" {
		client, err := github.NewClient(ctx, &github.Config{
			BaseURL:         c.Config.GHEBaseURL,
			GraphQLEndpoint: c.Config.GHEGraphQLEndpoint,
			Owner:           c.Config.CI.Owner,
			Repo:            c.Config.CI.Repo,
			PR: github.PullRequest{
				Revision: c.Config.CI.SHA,
				Number:   c.Config.CI.PRNumber,
			},
			CI:                 c.Config.CI.Link,
			Parser:             c.Parser,
			UseRawOutput:       c.Config.Terraform.UseRawOutput,
			Template:           c.Template,
			ParseErrorTemplate: c.ParseErrorTemplate,
			ResultLabels:       labels,
			Vars:               c.Config.Vars,
			EmbeddedVarNames:   c.Config.EmbeddedVarNames,
			Templates:          c.Config.Templates,
			Patch:              c.Config.PlanPatch,
			SkipNoChanges:      c.Config.Terraform.Plan.WhenNoChanges.DisableComment,
			IgnoreWarning:      c.Config.Terraform.Plan.IgnoreWarning,
			Masks:              c.Config.Masks,
		})
		if err != nil {
			return nil, err
		}
		gh = client.Notify
	}
	if c.Config.Output == "" {
		return gh, nil
	}
	// Write output to file instead of github comment
	client, err := localfile.NewClient(&localfile.Config{
		OutputFile:         c.Config.Output,
		Parser:             c.Parser,
		UseRawOutput:       c.Config.Terraform.UseRawOutput,
		CI:                 c.Config.CI.Link,
		Template:           c.Template,
		ParseErrorTemplate: c.ParseErrorTemplate,
		Vars:               c.Config.Vars,
		EmbeddedVarNames:   c.Config.EmbeddedVarNames,
		Templates:          c.Config.Templates,
		Masks:              c.Config.Masks,
		DisableLabel:       c.Config.Terraform.Plan.DisableLabel,
	}, gh)
	if err != nil {
		return nil, err
	}
	return client.Notify, nil
}

func (c *Controller) getApplyNotifier(ctx context.Context) (notifier.Notifier, error) {
	if c.Config.Output != "" {
		// Write output to file instead of github comment
		client, err := localfile.NewClient(&localfile.Config{
			OutputFile:         c.Config.Output,
			Parser:             c.Parser,
			UseRawOutput:       c.Config.Terraform.UseRawOutput,
			CI:                 c.Config.CI.Link,
			Template:           c.Template,
			ParseErrorTemplate: c.ParseErrorTemplate,
			Vars:               c.Config.Vars,
			EmbeddedVarNames:   c.Config.EmbeddedVarNames,
			Templates:          c.Config.Templates,
			Masks:              c.Config.Masks,
			DisableLabel:       c.Config.Terraform.Plan.DisableLabel,
		}, nil)
		if err != nil {
			return nil, err
		}
		return client.Notify, nil
	}
	client, err := github.NewClient(ctx, &github.Config{
		BaseURL:         c.Config.GHEBaseURL,
		GraphQLEndpoint: c.Config.GHEGraphQLEndpoint,
		Owner:           c.Config.CI.Owner,
		Repo:            c.Config.CI.Repo,
		PR: github.PullRequest{
			Revision: c.Config.CI.SHA,
			Number:   c.Config.CI.PRNumber,
		},
		CI:                 c.Config.CI.Link,
		Parser:             c.Parser,
		UseRawOutput:       c.Config.Terraform.UseRawOutput,
		Template:           c.Template,
		ParseErrorTemplate: c.ParseErrorTemplate,
		Vars:               c.Config.Vars,
		EmbeddedVarNames:   c.Config.EmbeddedVarNames,
		Templates:          c.Config.Templates,
		Patch:              c.Config.PlanPatch,
		SkipNoChanges:      c.Config.Terraform.Plan.WhenNoChanges.DisableComment,
		IgnoreWarning:      c.Config.Terraform.Plan.IgnoreWarning,
		Masks:              c.Config.Masks,
	})
	if err != nil {
		return nil, err
	}
	return client.Notify, nil
}
