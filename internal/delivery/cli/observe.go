package cli

import "github.com/spf13/cobra"

func newObserveCmd(a *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "observe",
		Aliases: []string{"obs"},
		Short:   "Observability operations",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "health",
		Short: "Check observability stack health",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.observe.Health(cmd.Context())
		},
	})
	return cmd
}
