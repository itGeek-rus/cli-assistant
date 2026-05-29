package cli

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/usecase"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

type App struct {
	cfg        config.Config
	log        *slog.Logger
	deployment *usecase.Deployment
	observe    *usecase.Observability
}

func NewApp(
	cfg config.Config,
	observe *usecase.Observability,
	deployment *usecase.Deployment,
	log *slog.Logger,
) *App {
	return &App{cfg: cfg, observe: observe, deployment: deployment, log: log}
}

func (a *App) Run() int {
	root := &cobra.Command{
		Use:           "assistant",
		Short:         "Personal DevOps CLI (GitOps + Observability)",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			if a.cfg.LogLevel == "debug" {
				a.log.Debug("running command", "command", cmd.CommandPath())
			}
		},
	}

	root.AddCommand(
		a.versionCmd(),
		newDeployCmd(a),
		newObserveCmd(a),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func (a *App) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "assistant dev (%s profile)\n", a.cfg.Profile)
		},
	}
}
