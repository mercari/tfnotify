package cli

import (
	"context"

	"github.com/suzuki-shunsuke/urfave-cli-v3-util/helpall"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/vcmd"
	"github.com/urfave/cli/v3"
)

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func New(flags *LDFlags) *cli.Command {
	return helpall.With(vcmd.With(&cli.Command{
		Name:           "tfnotify",
		Usage:          "Notify the execution result of terraform command",
		Version:        flags.Version,
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "owner",
				Usage:   "GitHub Repository owner name",
				Sources: cli.EnvVars("TFCMT_REPO_OWNER"),
			},
			&cli.StringFlag{
				Name:    "repo",
				Usage:   "GitHub Repository name",
				Sources: cli.EnvVars("TFCMT_REPO_NAME"),
			},
			&cli.StringFlag{
				Name:    "sha",
				Usage:   "commit SHA (revision)",
				Sources: cli.EnvVars("TFCMT_SHA"),
			},
			&cli.StringFlag{
				Name:  "build-url",
				Usage: "build url",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "log level",
			},
			&cli.IntFlag{
				Name:    "pr",
				Usage:   "pull request number",
				Sources: cli.EnvVars("TFCMT_PR_NUMBER"),
			},
			&cli.StringFlag{
				Name:    "config",
				Usage:   "config path",
				Sources: cli.EnvVars("TFCMT_CONFIG"),
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "template variables. The format of value is '<name>:<value>'. You can refer to the variable in the comment and label template using {{.Vars.<variable name>}}.",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "specify file to output result instead of posting a comment",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "plan",
				ArgsUsage: " <command> <args>...",
				Usage:     "Run terraform plan and post a comment to GitHub commit, pull request, or issue",
				Description: `Run terraform plan and post a comment to GitHub commit, pull request, or issue.

$ tfnotify [<global options>] plan [-patch] [-skip-no-changes] -- terraform plan [<terraform plan options>]`,
				Action: cmdPlan,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "patch",
						Usage:   "update an existing comment instead of creating a new comment. If there is no existing comment, a new comment is created.",
						Sources: cli.EnvVars("TFCMT_PLAN_PATCH"),
					},
					&cli.BoolFlag{
						Name:    "skip-no-changes",
						Usage:   "If there is no change tfnotify updates a label but doesn't post a comment",
						Sources: cli.EnvVars("TFCMT_SKIP_NO_CHANGES"),
					},
					&cli.BoolFlag{
						Name:    "ignore-warning",
						Usage:   "If skip-no-changes is enabled, comment is posted even if there is a warning. If skip-no-changes is disabled, warning is removed from the comment.",
						Sources: cli.EnvVars("TFCMT_IGNORE_WARNING"),
					},
					&cli.BoolFlag{
						Name:    "disable-label",
						Usage:   "Disable to add or update a label",
						Sources: cli.EnvVars("TFCMT_DISABLE_LABEL"),
					},
				},
			},
			{
				Name:      "apply",
				ArgsUsage: " <command> <args>...",
				Usage:     "Run terraform apply and post a comment to GitHub commit, pull request, or issue",
				Description: `Run terraform apply and post a comment to GitHub commit, pull request, or issue.

$ tfnotify [<global options>] apply -- terraform apply [<terraform apply options>]`,
				Action: cmdApply,
			},
			vcmd.New(&vcmd.Command{
				Name:    "tfnotify",
				Version: flags.Version,
				SHA:     flags.Commit,
			}),
		},
	}, flags.Commit), nil)
}
