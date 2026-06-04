package cli

import (
	"fmt"

	"cli-assistant/internal/domain/deploy"
	"cli-assistant/pkg/output"

	"github.com/spf13/cobra"
)

func newDeployCmd(a *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "GitOps deployment operations",
	}
	cmd.AddCommand(newDeployListCmd(a), newDeployStatusCmd(a), newDeploySyncCmd(a), newDeployDiffCmd(a))
	return cmd
}

func newDeployListCmd(a *App) *cobra.Command {
	var (
		namespace  string
		namePrefix string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List GitOps applications",
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := a.deployment.List(cmd.Context(), deploy.ListFilter{
				Namespace:  namespace,
				NamePrefix: namePrefix,
			})
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintApplications(res.Applications)
		},
	}
	cmd.Flags().StringVar(&namespace, "namespace", "", "Filter by namespace")
	cmd.Flags().StringVar(&namePrefix, "name-prefix", "", "Filter by name prefix")
	return cmd
}

func newDeployStatusCmd(a *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [name]",
		Short: "Show GitOps application status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				res, err := a.deployment.List(cmd.Context(), deploy.ListFilter{})
				if err != nil {
					return err
				}
				printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
				return printer.PrintApplications(res.Applications)
			}
			res, err := a.deployment.Status(cmd.Context(), name)
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintApplication(res.Application)
		},
	}
	return cmd
}

func newDeploySyncCmd(a *App) *cobra.Command {
	var (
		dryRun bool
		prune  bool
		force  bool
		yes    bool
	)
	cmd := &cobra.Command{
		Use:   "sync [name]",
		Short: "Sync GitOps application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			confirmed := yes || dryRun
			if !dryRun && !yes {
				fmt.Fprint(cmd.ErrOrStderr(), "Proceed with sync? [y/N]: ")
				var answer string
				if _, err := fmt.Fscanln(cmd.InOrStdin(), &answer); err != nil {
					return fmt.Errorf("sync cancelled")
				}
				if answer != "y" && answer != "Y" {
					return fmt.Errorf("sync cancelled")
				}
				confirmed = true
			}
			res, err := a.deployment.Sync(cmd.Context(), name, deploy.SyncOptions{
				DryRun: dryRun,
				Prune:  prune,
				Force:  force,
			}, confirmed)
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintSync(res)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without applying")
	cmd.Flags().BoolVar(&prune, "prune", false, "Prune resources during sync")
	cmd.Flags().BoolVar(&force, "force", false, "Force sync")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func newDeployDiffCmd(a *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [name]",
		Short: "Show diff for GitOps application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			diff, err := a.deployment.Diff(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintDiff(diff)
		},
	}
	return cmd
}
