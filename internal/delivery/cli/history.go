package cli

import (
	"cli-assistant/internal/domain/history"
	"context"

	"github.com/spf13/cobra"
)

func (a *App) wrapRunE(fn func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return a.recordRun(cmd.Context(), cmd.CommandPath(), func() error {
			return fn(cmd, args)
		})
	}
}

func (a *App) recordRun(ctx context.Context, command string, runFn func() error) error {
	err := runFn()
	status := "ok"
	errText := ""
	if err != nil {
		status = "error"
		errText = err.Error()
	}
	if a.hist != nil {
		_ = a.hist.SaveCommandRun(ctx, history.CommandRun{
			Command:   command,
			Profile:   a.cfg.Profile,
			Status:    status,
			ErrorText: errText,
		})
	}
	return err
}
