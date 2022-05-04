package cmd

import "github.com/spf13/cobra"

var (
	deleteCmd = &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	deleteCmd.AddCommand(deleteCacheCmd)
}
