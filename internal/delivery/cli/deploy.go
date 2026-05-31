package cli

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"cli-assistant/pkg/output"
	"errors"

	"github.com/spf13/cobra"
)

func newDeployCmd(a *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "GitOps deployment operations",
	}
	cmd.AddCommand(newDeployListCmd(a), newDeployStatusCmd(a))
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
	cmd.Flags().StringVarP(&namespace, "namespace", "", "", "Filter by namespace")
	cmd.Flags().StringVarP(&namePrefix, "name-prefix", "", "", "Filter by name prefix")
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
				if errors.Is(err, domain.ErrNotFound) {
					return err
				}
				return err
			}
			printer := output.NewPrinter(output.ParseFormat(a.cfg.Output), cmd.OutOrStdout())
			return printer.PrintApplication(res.Application)
		},
	}
	return cmd
}
