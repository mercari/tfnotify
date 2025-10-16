package cli

import (
	"context"
	"os"

	"github.com/mercari/tfnotify/v1/pkg/controller"
	"github.com/mercari/tfnotify/v1/pkg/terraform"
	"github.com/urfave/cli/v3"
)

func cmdApply(ctx context.Context, cmd *cli.Command) error {
	logLevel := cmd.String("log-level")
	setLogLevel(logLevel)

	cfg, err := newConfig(cmd)
	if err != nil {
		return err
	}

	if logLevel == "" {
		logLevel = cfg.Log.Level
		setLogLevel(logLevel)
	}

	if err := parseOpts(cmd, &cfg, os.Environ()); err != nil {
		return err
	}

	t := &controller.Controller{
		Config:             cfg,
		Parser:             terraform.NewApplyParser(),
		Template:           terraform.NewApplyTemplate(cfg.Terraform.Apply.Template),
		ParseErrorTemplate: terraform.NewApplyParseErrorTemplate(cfg.Terraform.Apply.WhenParseError.Template),
	}

	args := cmd.Args()

	return t.Apply(ctx, controller.Command{
		Cmd:  args.First(),
		Args: args.Tail(),
	})
}
