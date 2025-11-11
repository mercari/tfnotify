package controller

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/mattn/go-colorable"
	"github.com/mercari/tfnotify/v1/pkg/apperr"
	"github.com/mercari/tfnotify/v1/pkg/mask"
	"github.com/mercari/tfnotify/v1/pkg/notifier"
	"github.com/mercari/tfnotify/v1/pkg/platform"
)

// Apply sends the notification with notifier
func (c *Controller) Apply(ctx context.Context, command Command) error {
	if command.Cmd == "" {
		return errors.New("no command specified")
	}
	if err := platform.Complement(&c.Config); err != nil {
		return err
	}

	if err := c.Config.Validate(); err != nil {
		return err
	}

	ntf, err := c.getApplyNotifier(ctx)
	if err != nil {
		return err
	}

	if ntf == nil {
		return errors.New("no notifier specified at all")
	}

	cmd := exec.CommandContext(ctx, command.Cmd, command.Args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	combinedOutput := &bytes.Buffer{}
	uncolorizedStdout := colorable.NewNonColorable(stdout)
	uncolorizedStderr := colorable.NewNonColorable(stderr)
	uncolorizedCombinedOutput := colorable.NewNonColorable(combinedOutput)
	cmd.Stdout = io.MultiWriter(mask.NewWriter(os.Stdout, c.Config.Masks), uncolorizedStdout, uncolorizedCombinedOutput)
	cmd.Stderr = io.MultiWriter(mask.NewWriter(os.Stderr, c.Config.Masks), uncolorizedStderr, uncolorizedCombinedOutput)
	setCancel(cmd)
	_ = cmd.Run()

	return apperr.NewExitError(cmd.ProcessState.ExitCode(), ntf.Apply(ctx, &notifier.ParamExec{
		Stdout:         stdout.String(),
		Stderr:         stderr.String(),
		CombinedOutput: combinedOutput.String(),
		CIName:         c.Config.CI.Name,
		ExitCode:       cmd.ProcessState.ExitCode(),
		AISummarizer:   c.AISummarizer,
	}))
}
