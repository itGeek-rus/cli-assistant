package cli

import "github.com/spf13/cobra"

func newDeployCmd(a *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "GitOps deployment operations",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show deployment status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.deployment.Status(cmd.Context())
		},
	})
	return cmd
}
