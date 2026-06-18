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
	inspector  *usecase.Inspector
	version    string
	commit     string
	buildDate  string
}

func NewApp(
	cfg config.Config,
	observe *usecase.Observability,
	deployment *usecase.Deployment,
	inspector *usecase.Inspector,
	log *slog.Logger,
	version, commit, buildDate string,
) *App {
	return &App{
		cfg: cfg, observe: observe, deployment: deployment,
		inspector: inspector, log: log,
		version: version, commit: commit, buildDate: buildDate,
	}
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
			fmt.Fprintf(cmd.OutOrStdout(),
				"assistant %s\nprofile: %s\ncommit: %s\nbuilt: %s\n",
				a.version, a.cfg.Profile, a.commit, a.buildDate,
			)
		},
	}
}
