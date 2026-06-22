package cli

import "github.com/spf13/cobra"

func newServerCmd(a *App) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start REST API server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.httpServer.ListenAndServe(cmd.Context())
		},
	}
}
