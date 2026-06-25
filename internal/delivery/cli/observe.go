package cli

import (
	"cli-assistant/internal/domain/observe"
	"cli-assistant/pkg/output"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newObserveCmd(a *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "observe",
		Aliases: []string{"obs"},
		Short:   "Observability operations",
	}
	cmd.AddCommand(
		newObserveHealthCmd(a),
		newObserveQueryCmd(a),
		newObserveAlertsCmd(a),
		newObserveLogsCmd(a),
	)
	return cmd
}

func newObserveHealthCmd(a *App) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check observability stack health",
		RunE: a.wrapRunE(func(cmd *cobra.Command, _ []string) error {
			res, err := a.observe.Health(cmd.Context())
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintStackHealth(res.Health)
		}),
	}
}

func newObserveQueryCmd(a *App) *cobra.Command {
	return &cobra.Command{
		Use:   "query [expr]",
		Short: "Run instant PromQL query",
		Args:  cobra.ExactArgs(1),
		RunE: a.wrapRunE(func(cmd *cobra.Command, args []string) error {
			res, err := a.observe.Query(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintQueryResult(res.Result)
		}),
	}
}

func newObserveAlertsCmd(a *App) *cobra.Command {
	var state string
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "List alerts",
		RunE: a.wrapRunE(func(cmd *cobra.Command, _ []string) error {
			filter := observe.AlertFilter{}
			if state != "" {
				filter.State = observe.AlertState(state)
			}
			res, err := a.observe.Alerts(cmd.Context(), filter)
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintAlerts(res.Alerts)
		}),
	}
	cmd.Flags().StringVar(&state, "state", "", "Filter by state: Firing, Pending, Inactive")
	return cmd
}

func newObserveLogsCmd(a *App) *cobra.Command {
	var since string
	var limit int
	cmd := &cobra.Command{
		Use:   "logs [query]",
		Short: "Query logs (LogQL)",
		Args:  cobra.ExactArgs(1),
		RunE: a.wrapRunE(func(cmd *cobra.Command, args []string) error {
			d, err := time.ParseDuration(since)
			if err != nil {
				return fmt.Errorf("invalid --since: %w", err)
			}
			res, err := a.observe.Logs(cmd.Context(), args[0], d, limit)
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintLogs(res)
		}),
	}
	cmd.Flags().StringVar(&since, "since", "1h", "Lookback window")
	cmd.Flags().IntVar(&limit, "limit", 100, "Max log lines")
	return cmd
}
