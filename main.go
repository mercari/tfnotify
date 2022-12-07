package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mercari/tfnotify/config"
	"github.com/mercari/tfnotify/notifier"
	"github.com/mercari/tfnotify/notifier/github"
	"github.com/mercari/tfnotify/notifier/gitlab"
	"github.com/mercari/tfnotify/notifier/slack"
	"github.com/mercari/tfnotify/notifier/typetalk"
	"github.com/mercari/tfnotify/terraform"

	"github.com/urfave/cli"
)

const (
	name        = "tfnotify"
	description = "Notify the execution result of terraform command"

	version = "unset"
)

type tfnotify struct {
	config                 config.Config
	context                *cli.Context
	parser                 terraform.Parser
	template               terraform.Template
	destroyWarningTemplate terraform.Template
	warnDestroy            bool
}

// Run sends the notification with notifier
func (t *tfnotify) Run() error {
	ciname := t.config.CI
	if t.context.GlobalString("ci") != "" {
		ciname = t.context.GlobalString("ci")
	}
	ciname = strings.ToLower(ciname)
	var ci CI
	var err error
	switch ciname {
	case "circleci", "circle-ci":
		ci, err = circleci()
		if err != nil {
			return err
		}
	case "travis", "travisci", "travis-ci":
		ci, err = travisci()
		if err != nil {
			return err
		}
	case "codebuild":
		ci, err = codebuild()
		if err != nil {
			return err
		}
	case "teamcity":
		ci, err = teamcity()
		if err != nil {
			return err
		}
	case "drone":
		ci, err = drone()
		if err != nil {
			return err
		}
	case "jenkins":
		ci, err = jenkins()
		if err != nil {
			return err
		}
	case "gitlabci", "gitlab-ci":
		ci, err = gitlabci()
		if err != nil {
			return err
		}
	case "github-actions":
		ci, err = githubActions()
		if err != nil {
			return err
		}
	case "cloud-build", "cloudbuild":
		ci, err = cloudbuild()
		if err != nil {
			return err
		}
	case "":
		return fmt.Errorf("CI service: required (e.g. circleci)")
	default:
		return fmt.Errorf("CI service %v: not supported yet", ci)
	}

	selectedNotifier := t.config.GetNotifierType()
	if t.context.GlobalString("notifier") != "" {
		selectedNotifier = t.context.GlobalString("notifier")
	}

	var notifier notifier.Notifier
	switch selectedNotifier {
	case "github":
		client, err := github.NewClient(github.Config{
			Token:   t.config.Notifier.Github.Token,
			BaseURL: t.config.Notifier.Github.BaseURL,
			Owner:   t.config.Notifier.Github.Repository.Owner,
			Repo:    t.config.Notifier.Github.Repository.Name,
			PR: github.PullRequest{
				Revision:              ci.PR.Revision,
				Number:                ci.PR.Number,
				Title:                 t.context.String("title"),
				Message:               t.context.String("message"),
				DestroyWarningTitle:   t.context.String("destroy-warning-title"),
				DestroyWarningMessage: t.context.String("destroy-warning-message"),
			},
			CI:                     ci.URL,
			Parser:                 t.parser,
			UseRawOutput:           t.config.Terraform.UseRawOutput,
			Template:               t.template,
			DestroyWarningTemplate: t.destroyWarningTemplate,
			WarnDestroy:            t.warnDestroy,
			ResultLabels: github.ResultLabels{
				AddOrUpdateLabel: t.config.Terraform.Plan.WhenAddOrUpdateOnly.Label,
				DestroyLabel:     t.config.Terraform.Plan.WhenDestroy.Label,
				NoChangesLabel:   t.config.Terraform.Plan.WhenNoChanges.Label,
				PlanErrorLabel:   t.config.Terraform.Plan.WhenPlanError.Label,
			},
		})
		if err != nil {
			return err
		}
		notifier = client.Notify
	case "gitlab":
		client, err := gitlab.NewClient(gitlab.Config{
			Token:     t.config.Notifier.Gitlab.Token,
			BaseURL:   t.config.Notifier.Gitlab.BaseURL,
			NameSpace: t.config.Notifier.Gitlab.Repository.Owner,
			Project:   t.config.Notifier.Gitlab.Repository.Name,
			MR: gitlab.MergeRequest{
				Revision: ci.PR.Revision,
				Number:   ci.PR.Number,
				Title:    t.context.String("title"),
				Message:  t.context.String("message"),
			},
			CI:       ci.URL,
			Parser:   t.parser,
			Template: t.template,
		})
		if err != nil {
			return err
		}
		notifier = client.Notify
	case "slack":
		client, err := slack.NewClient(slack.Config{
			Token:    t.config.Notifier.Slack.Token,
			Channel:  t.config.Notifier.Slack.Channel,
			Botname:  t.config.Notifier.Slack.Bot,
			Title:    t.context.String("title"),
			Message:  t.context.String("message"),
			CI:       ci.URL,
			Parser:   t.parser,
			Template: t.template,
		})
		if err != nil {
			return err
		}
		notifier = client.Notify
	case "typetalk":
		client, err := typetalk.NewClient(typetalk.Config{
			Token:    t.config.Notifier.Typetalk.Token,
			TopicID:  t.config.Notifier.Typetalk.TopicID,
			Title:    t.context.String("title"),
			Message:  t.context.String("message"),
			CI:       ci.URL,
			Parser:   t.parser,
			Template: t.template,
		})
		if err != nil {
			return err
		}
		notifier = client.Notify
	case "":
		return fmt.Errorf("notifier is missing")
	default:
		return fmt.Errorf("%s: not supported notifier yet", t.context.GlobalString("notifier"))
	}

	if notifier == nil {
		return fmt.Errorf("no notifier specified at all")
	}

	return NewExitError(notifier.Notify(tee(os.Stdin, os.Stdout)))
}

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Usage = description
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "ci", Usage: "name of CI to run tfnotify"},
		cli.StringFlag{Name: "config", Usage: "config path"},
		cli.StringFlag{Name: "notifier", Usage: "notification destination"},
	}
	app.Commands = []cli.Command{
		{
			Name:   "fmt",
			Usage:  "Parse stdin as a fmt result",
			Action: cmdFmt,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "title, t",
					Usage: "Specify the title to use for notification",
				},
				cli.StringFlag{
					Name:  "message, m",
					Usage: "Specify the message to use for notification",
				},
			},
		},
		{
			Name:   "drift",
			Usage:  "Parse stdin as a drift result",
			Action: cmdDrift,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "title, t",
					Usage: "Specify the title to use for notification",
				},
				cli.StringFlag{
					Name:  "message, m",
					Usage: "Specify the message to use for notification",
				},
			},
		},
		{
			Name:   "plan",
			Usage:  "Parse stdin as a plan result",
			Action: cmdPlan,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "title, t",
					Usage: "Specify the title to use for notification",
				},
				cli.StringFlag{
					Name:  "message, m",
					Usage: "Specify the message to use for notification",
				},
				cli.StringFlag{
					Name:  "destroy-warning-title",
					Usage: "Specify the title to use for destroy warning notification",
				},
				cli.StringFlag{
					Name:  "destroy-warning-message",
					Usage: "Specify the message to use for destroy warning notification",
				},
			},
		},
		{
			Name:   "apply",
			Usage:  "Parse stdin as a apply result",
			Action: cmdApply,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "title, t",
					Usage: "Specify the title to use for notification",
				},
				cli.StringFlag{
					Name:  "message, m",
					Usage: "Specify the message to use for notification",
				},
			},
		},
	}

	err := app.Run(os.Args)
	os.Exit(HandleExit(err))
}

func newConfig(ctx *cli.Context) (cfg config.Config, err error) {
	confPath, err := cfg.Find(ctx.GlobalString("config"))
	if err != nil {
		return cfg, err
	}
	if err := cfg.LoadFile(confPath); err != nil {
		return cfg, err
	}
	if err := cfg.Validation(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func cmdFmt(ctx *cli.Context) error {
	cfg, err := newConfig(ctx)
	if err != nil {
		return err
	}
	t := &tfnotify{
		config:   cfg,
		context:  ctx,
		parser:   terraform.NewFmtParser(),
		template: terraform.NewFmtTemplate(cfg.Terraform.Fmt.Template),
	}
	return t.Run()
}

func cmdDrift(ctx *cli.Context) error {
	cfg, err := newConfig(ctx)
	if err != nil {
		return err
	}
	t := &tfnotify{
		config:   cfg,
		context:  ctx,
		parser:   terraform.NewDriftParser(),
		template: terraform.NewDriftTemplate(cfg.Terraform.Drift.Template),
	}
	return t.Run()
}

func cmdPlan(ctx *cli.Context) error {
	cfg, err := newConfig(ctx)
	if err != nil {
		return err
	}

	// If when_destroy is not defined in configuration, tfnotify should not notify it
	warnDestroy := cfg.Terraform.Plan.WhenDestroy.Template != ""

	t := &tfnotify{
		config:                 cfg,
		context:                ctx,
		parser:                 terraform.NewPlanParser(),
		template:               terraform.NewPlanTemplate(cfg.Terraform.Plan.Template),
		destroyWarningTemplate: terraform.NewDestroyWarningTemplate(cfg.Terraform.Plan.WhenDestroy.Template),
		warnDestroy:            warnDestroy,
	}
	return t.Run()
}

func cmdApply(ctx *cli.Context) error {
	cfg, err := newConfig(ctx)
	if err != nil {
		return err
	}
	t := &tfnotify{
		config:   cfg,
		context:  ctx,
		parser:   terraform.NewApplyParser(),
		template: terraform.NewApplyTemplate(cfg.Terraform.Apply.Template),
	}
	return t.Run()
}
